package main

import (
	"net/http"
	"log"
	"github.com/go-chi/chi/v5"
	"github.com/kujoki/go-musthave-service/internal/handler"
	"github.com/kujoki/go-musthave-service/internal/config"
	"github.com/kujoki/go-musthave-service/internal/service"
)

func main() {
	cfg := config.ParseFlags()
    
    if err := run(cfg); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}


func run(cfg *config.Config) error {
	s := service.NewService()

	r := chi.NewRouter()

	log.Println("Running server on", cfg.RunAddr)

	r.Post("/", handler.WrapperPostSlash(cfg.BaseURL, s))
	r.Get("/{ID}", handler.WrapperGetSlashURL(s))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`this request are not allowed!`))
	})

	return http.ListenAndServe(cfg.RunAddr, r)
}