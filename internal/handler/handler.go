package handler

import (
	"io"
	"log"
	"net/http"
	"strings"
	"github.com/kujoki/go-musthave-service/internal/service"
)

func validateHeaders(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "text/plain")
}

func SlashURL(w http.ResponseWriter, req *http.Request) {
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

	shortURL := service.CreateShortURL(originURL)
	log.Printf("a request was received for URL %s: %s \n", originURL, shortURL)
	fullURL := "http://localhost:8080/" + shortURL

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullURL))
	log.Println("processing POST request was completed")
}

func GetSlashURL(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	
	shortURL := req.URL.Path
	if len(shortURL) > 0 {
		shortURL = shortURL[1:]
	} else {
		log.Println("inappropriate user behavior - there is no url")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`there is no url`))
		return
	}
	log.Printf("the short url is %s \n", shortURL)

	originURL, ok := service.ReverseMap(shortURL)
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
