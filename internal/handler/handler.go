package handler

import (
	"io"
	"log"
	"net/http"
	"encoding/json"
	"strings"
	"github.com/go-chi/chi/v5"
	"github.com/kujoki/go-musthave-service/internal/storage"
	"github.com/kujoki/go-musthave-service/internal/service"
	"github.com/kujoki/go-musthave-service/internal/model"
)

func WrapperPostSlash(baseURL string, s *service.Service) http.HandlerFunc {
	log.Printf("base url is %s \n", baseURL)
	return func(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

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

	shortURL, _ := s.CreateShortURL(originURL)
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

		originURL, ok, _ := s.ReverseMap(shortURL)
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

func WrapperPostAPIShort(baseURL string, s *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		prefix := "application/json"
		w.Header().Set("Content-Type", prefix)

		log.Println("decoding request")

		var jsonReq model.Request
		dec := json.NewDecoder(req.Body)
		
		if err := dec.Decode(&jsonReq); err != nil {
			log.Println("error during encoding")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if jsonReq.URL == "" {
			log.Println("empty URL in body")
			w.WriteHeader(http.StatusBadRequest)
		return
		}
		
		shortURL, _ := s.CreateShortURL(jsonReq.URL)
		log.Printf("a request was received for URL %s: %s \n", jsonReq.URL, shortURL)
		fullURL := baseURL + "/" + shortURL
    
		resp := model.Response{
			Result: fullURL,
		}
        
		w.WriteHeader(http.StatusCreated)
		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			log.Println("error encoding response")
			return
		}
		log.Println("sending HTTP 201 response")
	}
}

func WrapperPingAPI(repo storage.URLRepository) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        ctx := req.Context()
        res := repo.Ping(ctx)
		if !res {
            w.WriteHeader(http.StatusInternalServerError)
        } else {
            w.WriteHeader(http.StatusOK)
        }
    }
}

func WrapperPostBatchAPI(baseURL string, s *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		prefix := "application/json"
		w.Header().Set("Content-Type", prefix)

		log.Println("decoding request")

		var jsonReq []model.BatchRequest
		json.NewDecoder(req.Body).Decode(&jsonReq)
        if len(jsonReq) == 0 {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

		var jsonResp []model.BatchResponse
		for _, item := range jsonReq {
            shortURL, _ := s.CreateShortURL(item.OriginalURL)
			fullURL := baseURL + "/" + shortURL
            jsonResp = append(jsonResp, model.BatchResponse{
                CorrelationID: item.CorrelationID,
                ShortURL: fullURL,
            })
        }

        w.WriteHeader(http.StatusCreated)
		enc := json.NewEncoder(w)
		if err := enc.Encode(jsonResp); err != nil {
			log.Println("error encoding response")
			return
		}
		log.Println("sending HTTP 201 response")
	}
}