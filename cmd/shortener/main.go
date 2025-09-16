package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/kujoki/go-musthave-service/internal/config"
	"github.com/kujoki/go-musthave-service/internal/handler"
	l "github.com/kujoki/go-musthave-service/internal/logger"
	"github.com/kujoki/go-musthave-service/internal/model"
	"github.com/kujoki/go-musthave-service/internal/service"
	"github.com/kujoki/go-musthave-service/internal/storage"
	"github.com/kujoki/go-musthave-service/migrations"
	"go.uber.org/zap"
)

func main() {
	cfg := config.ParseFlags()
    var sugar zap.SugaredLogger

    if err := run(cfg, sugar); err != nil {
		sugar.Fatalw(err.Error(), "event", "start server")
    }
}


func run(cfg *config.Config, sugar zap.SugaredLogger) error {
	logger, err := zap.NewDevelopment()
	if err != nil {
        return fmt.Errorf("failed to create logger: %w", err)
    }
    defer logger.Sync()
	sugar = *logger.Sugar()

	var repo storage.URLRepository
	var data []model.Data

	if cfg.DatabaseDSN != "" {
		migrations.RunMigrations(cfg.DatabaseDSN)
		postgresRepo, err := storage.NewPostgresRepository(cfg.DatabaseDSN)
		if err != nil {
			sugar.Fatalf("failed to connect to postgres: %v", err)
		}
		repo = postgresRepo
	} else {
		memoryRepo := storage.NewMemoryRepository()
		repo = memoryRepo
		if cfg.FileStoragePath != "" {
			loaded, err := storage.Load(cfg.FileStoragePath)
			if err != nil {
				sugar.Infow(err.Error(), "event", "read URL map")
				memoryRepo.Data = make(map[string]model.URLRecord)
			} else {
				memoryRepo.Data = model.DataSliceToMap(loaded)
			}
			sugar.Infow("read storage", "filename", cfg.FileStoragePath)
		} else {
			memoryRepo.Data = make(map[string]model.URLRecord)
		}
	}
	
	s := service.NewService(repo, cfg.ShortURLLen)

	r := chi.NewRouter()

	
	r.Use(handler.GzipMiddlewareRequest)
	r.Use(handler.GzipMiddlewareResponse)

	builder := handler.NewJWTBuild(cfg.SecretToken, cfg.ApplicationName, cfg.UserCookieName)

	postHandler := handler.WrapperPostSlash(cfg.BaseURL, s, builder)
	postAPIShortHandler := handler.WrapperPostAPIShort(cfg.BaseURL, s, builder)
	postBatchAPIHandler := handler.WrapperPostBatchAPI(cfg.BaseURL, s, builder)
	getHandler := handler.WrapperGetSlashURL(s, builder)
	getPingHandler := handler.WrapperPingAPI(repo, builder)
	getUsersURL := handler.WrapperGetUsers(cfg.BaseURL, s, builder)
	deleteUsersURL := handler.WrapperDeleteURLs(s, builder)

	r.Post("/",  l.WithLogging(sugar, postHandler))
	r.Post("/api/shorten", l.WithLogging(sugar, postAPIShortHandler))
	r.Post("/api/shorten/batch", l.WithLogging(sugar, postBatchAPIHandler))
	r.Get("/{ID}", l.WithLogging(sugar, getHandler))
	r.Get("/ping", getPingHandler)
	r.Get("/api/user/urls", getUsersURL)
	r.Delete("/api/user/urls", l.WithLogging(sugar, deleteUsersURL))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`this request are not allowed!`))
	})

	srv := &http.Server{
        Addr:    cfg.RunAddr,
        Handler: r,
    }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sugar.Infow("running server", "address", cfg.RunAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalw(err.Error(), "event", "start server")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	s.Stop()
	cancel()
	<-ctx.Done()

	sugar.Infow("save data before shutting down")
	
	if cfg.FileStoragePath != "" {
		data = model.MapToDataSlice(repo.GetAll())
		if err := storage.Save(cfg.FileStoragePath, data); err != nil {
			sugar.Errorw(err.Error(), "event", "save URL map")
		}
	}

	repo.Close()
	sugar.Infow("close repository")

    sugar.Infow("shutting down server gracefully")
	return srv.Shutdown(context.Background())
}