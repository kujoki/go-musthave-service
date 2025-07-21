package main

import (
	"net/http"
	"log"
	"github.com/go-chi/chi/v5"
	"github.com/kujoki/go-musthave-service/internal/handler"
)

func main() {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Post("/", handler.SlashURL) 
		r.Get("/{id}", handler.GetSlashURL)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`this request are not allowed!`))
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}