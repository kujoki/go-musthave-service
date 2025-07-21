package main

import (
	"net/http"
	"log"
	"github.com/go-chi/chi/v5"
	"github.com/kujoki/go-musthave-service/internal/handler"
	"github.com/kujoki/go-musthave-service/internal/config"
)

func main() {
	config.ParseFlags()

	if err := run(); err != nil {
        panic(err)
    }
}

func run() error {
	r := chi.NewRouter()

	log.Println("Running server on", config.RunAddr)

	r.Route("/", func(r chi.Router) {
		r.Post("/", handler.WrapperPostSlash(config.BaseURL)) 
		r.Get("/{id}", handler.GetSlashURL)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`this request are not allowed!`))
	})

	return http.ListenAndServe(config.RunAddr, r)
}