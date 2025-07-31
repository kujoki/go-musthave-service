package handler

import (
	"io"
	"log"
	"net/http"
	"strings"
	"github.com/go-chi/chi/v5"
	"github.com/kujoki/go-musthave-service/internal/service"
)

func validateHeaders(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	return strings.HasPrefix(contentType, "text/plain")
}

func WrapperPostSlash(baseURL string, s *service.Service) http.HandlerFunc {
	log.Printf("base url is %s \n", baseURL)
	return func(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	validHeaders := validateHeaders(req)

	if !validHeaders {
		log.Println("incorrect headers")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`incorrect headers were suggested!`))
		return
	}

	reqData, err := io.ReadAll(req.Body)
	if err != nil || len(reqData) == 0 {
		log.Println("failed to extract body contents")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`failed to extract body contents`))
		return
	}

	originURL := strings.TrimSpace(string(reqData))
	if originURL == "" {
		log.Println("empty URL in body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("origin URL is extracted: %s \n", originURL)

	shortURL := s.CreateShortURL(originURL)
	log.Printf("a request was received for URL %s: %s \n", originURL, shortURL)
	fullURL := baseURL + "/" + shortURL

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullURL))
	log.Println("processing POST request was completed")
	}
}

func WrapperGetSlashURL(s *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		
		shortURL := chi.URLParam(req, "ID")
		if len(shortURL) == 0 {
			log.Println("inappropriate user behavior - there is no url")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`there is no url`))
			return
		}

		log.Printf("the short url is %s \n", shortURL)

		originURL, ok := s.ReverseMap(shortURL)
		log.Printf("the result of check was received: %t \n", ok)

		if !ok || originURL == "" {
			log.Printf("value for this url doesn't exist in map")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`value for this url doesn't exist in map!`))
			return
		}

		w.Header().Set("Location", originURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		log.Println("processing GET request was completed")
		}
	}
