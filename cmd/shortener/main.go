package main

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"github.com/kujoki/go-musthave-service/internal/handler"
	"github.com/kujoki/go-musthave-service/internal/config"
	"github.com/kujoki/go-musthave-service/internal/service"
	l "github.com/kujoki/go-musthave-service/internal/logger"
)

var sugar zap.SugaredLogger

func main() {
	cfg := config.ParseFlags()
    
    if err := run(cfg); err != nil {
		sugar.Fatalw(err.Error(), "event", "start server")
    }
}


func run(cfg *config.Config) error {
	logger, err := zap.NewDevelopment()
	if err != nil {
        panic(err)
    }
    defer logger.Sync()
	sugar = *logger.Sugar()

	s := service.NewService()

	r := chi.NewRouter()

	
	r.Use(handler.GzipMiddleware)
	r.Use(middleware.Compress(5, "application/json", "text/html"))

	postHandler := handler.WrapperPostSlash(cfg.BaseURL, s)
	postAPIShortHandler := handler.WrapperPostAPIShort(cfg.BaseURL, s)
	getHandler := handler.WrapperGetSlashURL(s)

	r.Post("/",  l.WithLogging(sugar, postHandler))
	r.Post("/api/shorten", l.WithLogging(sugar, postAPIShortHandler))
	r.Get("/{ID}", l.WithLogging(sugar, getHandler))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`this request are not allowed!`))
	})

	sugar.Infow("Running server", "address", cfg.RunAddr)

	return http.ListenAndServe(cfg.RunAddr, r)
}