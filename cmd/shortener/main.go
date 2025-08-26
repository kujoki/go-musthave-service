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
	"github.com/kujoki/go-musthave-service/internal/db"
	"github.com/kujoki/go-musthave-service/internal/repository"
	"github.com/kujoki/go-musthave-service/internal/handler"
	l "github.com/kujoki/go-musthave-service/internal/logger"
	"github.com/kujoki/go-musthave-service/internal/model"
	"github.com/kujoki/go-musthave-service/internal/service"
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

	data, err := cache.Load(cfg.FileStoragePath)
	if err != nil {
		sugar.Infow(err.Error(), "event", "read URL map")
		data = []model.Data{}
	}
	sugar.Infow("Read storage", "filename", cfg.FileStoragePath)

	dbConnection, _ := repository.CreatePool(cfg.DatabaseDSN)

	s := service.NewService(data)

	r := chi.NewRouter()

	
	r.Use(handler.GzipMiddlewareRequest)
	r.Use(handler.GzipMiddlewareResponse)

	postHandler := handler.WrapperPostSlash(cfg.BaseURL, s)
	postAPIShortHandler := handler.WrapperPostAPIShort(cfg.BaseURL, s)
	getHandler := handler.WrapperGetSlashURL(s)
	getPingHandler := handler.WrapperPingAPI(dbConnection)

	r.Post("/",  l.WithLogging(sugar, postHandler))
	r.Post("/api/shorten", l.WithLogging(sugar, postAPIShortHandler))
	r.Get("/{ID}", l.WithLogging(sugar, getHandler))
	r.Get("/ping", getPingHandler)

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
		sugar.Infow("Running server", "address", cfg.RunAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalw(err.Error(), "event", "start server")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan 
	cancel()
	<-ctx.Done()

	sugar.Infow("Save data before shutting down")

	data = s.AllData()
	if err := cache.Save(cfg.FileStoragePath, data); err != nil {
		sugar.Errorw(err.Error(), "event", "save URL map")
	}
	if dbConnection != nil {
		sugar.Infow("Close pool connection before shutting down")
		defer dbConnection.Close()
	}

    sugar.Infow("Shutting down server gracefully")
	return srv.Shutdown(context.Background())
}