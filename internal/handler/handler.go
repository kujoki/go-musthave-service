package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kujoki/go-musthave-service/internal/model"
	"github.com/kujoki/go-musthave-service/internal/service"
	"github.com/kujoki/go-musthave-service/internal/storage"
)

func GetOrCreateUserUUID(b JWTBuilder, w http.ResponseWriter, req *http.Request) (string, error) {
	var userUUID string
	var err error
	authHeader := req.Header.Get("Authorization")

	if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			userUUID, _ = b.ParseUserUUID(tokenString)
		}
	if userUUID != "" {
		return userUUID, nil
	}

    cookie, err := req.Cookie(b.UserCookieName)
    if err != nil {
        log.Println("no cookie/ auth header, generating new one")
        userUUID := b.GenerateUserUUID()
        token, err := b.SignUserUUID(userUUID)
        if err != nil {
			log.Println("error during sign uuid")
            return "", err
        }
        http.SetCookie(w, &http. Cookie{
            Name:     b.UserCookieName,
            Value:    token,
            Path:     "/",
            HttpOnly: true,
            Secure:   true, 
        })
		log.Println("set cookie")
		w.Header().Set("Authorization", "Bearer "+token)
		log.Println("set authorization header")
        return userUUID, nil
    }
    userUUID, err = b.ParseUserUUID(cookie.Value)
    if err != nil {
        log.Println("cookie parsing failed, issuing new one")
        newUUID := b.GenerateUserUUID()
        token, err := b.SignUserUUID(newUUID)
        if err != nil {
			log.Println("error during sign uuid")
            return "", err
        }
        http.SetCookie(w, &http.Cookie{
            Name:     b.UserCookieName,
            Value:    token,
            Path:     "/",
            HttpOnly: true,
            Secure:   true,
        })
        return "", ErrNoUserUUID
    }

    log.Println("valid cookie with user UUID")
    return userUUID, nil
}

func CheckExistURL(s *service.Service, originURL string) (string, error) {
	shortURLRes, ok, err := s.CheckShortURLValue(originURL)
	if err != nil {
		log.Println("there is an error during checking URL existing")
		return "", err
	}
	if ok {
		if !shortURLRes.IsDeleted {
			return shortURLRes.ShortURL, model.ErrURLExists
		}
	}
	return "", nil
}

func WrapperPostSlash(baseURL string, s *service.Service, b *JWTBuilder) http.HandlerFunc {
	log.Printf("base url is %s \n", baseURL)
	return func(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	
	userUUID, err := GetOrCreateUserUUID(*b, w, req)
	log.Println("userUUID was got")
	if err != nil {
		log.Println("there is an authorization error")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("there is an authorization error"))
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

	shortURL, err := CheckExistURL(s, originURL)

	var fullURL string
	if errors.Is(err, model.ErrURLExists) {
		w.WriteHeader(http.StatusConflict)
		fullURL, _ = url.JoinPath(baseURL, shortURL)
		w.Write([]byte(fullURL))
		return
	}  else if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
	}

	shortURL, err = s.CreateShortURL(originURL, userUUID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	log.Printf("a request was received for URL %s: %s \n", originURL, shortURL)
	fullURL, _ = url.JoinPath(baseURL, shortURL)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullURL))
	log.Println("processing POST request was completed")
	}
}

func WrapperGetSlashURL(s *service.Service, b *JWTBuilder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		_, err := GetOrCreateUserUUID(*b, w, req)
		log.Println("userUUID was got")
		if err != nil {
			log.Println("there is an authorization error")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("there is an authorization error"))
			return
		}
		
		shortURL := chi.URLParam(req, "ID")
		if len(shortURL) == 0 {
			log.Println("inappropriate user behavior - there is no url")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`there is no url`))
			return
		}

		log.Printf("the short url is %s \n", shortURL)

		originURLRes, ok, _ := s.CheckOriginURLValue(shortURL)
		log.Printf("the result of check was received: %t \n", ok)

		if !ok || originURLRes.OriginURL == "" {
			log.Printf("value for this url doesn't exist in map")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`value for this url doesn't exist in map!`))
			return
		}
		if ok && originURLRes.IsDeleted {
			log.Println("the long URL was deleted")
			w.WriteHeader(http.StatusGone)
			return
		}

		w.Header().Set("Location", originURLRes.OriginURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
		log.Println("processing GET request was completed")
		}
	}

func WrapperPostAPIShort(baseURL string, s *service.Service, b *JWTBuilder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		prefix := "application/json"
		w.Header().Set("Content-Type", prefix)

		userUUID, err := GetOrCreateUserUUID(*b, w, req)
		log.Println("userUUID was got")
		if err != nil {
			log.Println("there is an authorization error")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("there is an authorization error"))
			return
		}

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

		var fullURL string
		var resp model.Response
		var enc *json.Encoder

		shortURL, err := CheckExistURL(s, jsonReq.URL)
		if errors.Is(err, model.ErrURLExists) {
			w.WriteHeader(http.StatusConflict)
			fullURL, _ = url.JoinPath(baseURL, shortURL)
			resp = model.Response{
				Result: fullURL,
			}
			enc = json.NewEncoder(w)
			_ = enc.Encode(resp)
			return
		} else if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		shortURL, err = s.CreateShortURL(jsonReq.URL, userUUID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("a request was received for URL %s: %s \n", jsonReq.URL, shortURL)
		fullURL, _ = url.JoinPath(baseURL, shortURL)
    
		resp = model.Response{
			Result: fullURL,
		}
        
		w.WriteHeader(http.StatusCreated)
		enc = json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			log.Println("error encoding response")
			return
		}
		log.Println("sending HTTP 201 response")
	}
}

func WrapperPingAPI(repo storage.URLRepository, b *JWTBuilder) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
		_, err := GetOrCreateUserUUID(*b, w, req)
		if err != nil {
			log.Println("there is an authorization error")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("there is an authorization error"))
			return
		}
        ctx := req.Context()
        res := repo.Ping(ctx)
		if !res {
            w.WriteHeader(http.StatusInternalServerError)
        } else {
            w.WriteHeader(http.StatusOK)
        }
    }
}

func WrapperPostBatchAPI(baseURL string, s *service.Service, b *JWTBuilder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		prefix := "application/json"
		w.Header().Set("Content-Type", prefix)
		userUUID, err := GetOrCreateUserUUID(*b, w, req)
		log.Println("userUUID was got")
		if err != nil {
			log.Println("there is an authorization error")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("there is an authorization error"))
			return
		}
		log.Println("decoding request")

		var jsonReq []model.BatchRequest
		json.NewDecoder(req.Body).Decode(&jsonReq)
        if len(jsonReq) == 0 {
            w.WriteHeader(http.StatusBadRequest) 
            return
        }

		var jsonResp []model.BatchResponse
		for _, item := range jsonReq {
            shortURL, _ := s.CreateShortURL(item.OriginalURL, userUUID)

			fullURL, _ := url.JoinPath(baseURL, shortURL)
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

func WrapperGetUsers(baseURL string, s *service.Service, b *JWTBuilder) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
		prefix := "application/json"
		w.Header().Set("Content-Type", prefix)

		userUUID, err := GetOrCreateUserUUID(*b, w, req)
		log.Println("got userUUID", userUUID)
		if err != nil {
			log.Println("there is an authorization error")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("there is an authorization error"))
			return
		}
		var data []model.UserURL
		data, err = s.GetURLByUser(userUUID)
		if err != nil {
			log.Println("can't get user's URLs during error ", err)
			w.WriteHeader(http.StatusBadRequest)
			return 
		}
        if len(data) == 0 {
            log.Println("no URLs for user")
            w.WriteHeader(http.StatusNoContent)
            return
        }
		for i := range data {
    		data[i].ShortURL, _ = url.JoinPath(baseURL, data[i].ShortURL)
		}

        w.WriteHeader(http.StatusOK)
        if err := json.NewEncoder(w).Encode(data); err != nil {
            log.Println("error encoding response:", err)
        }
	}
}

func WrapperDeleteURLs(s *service.Service, b *JWTBuilder) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		prefix := "application/json"
		w.Header().Set("Content-Type", prefix)

		userUUID, err := GetOrCreateUserUUID(*b, w, req)
		log.Println("got userUUID", userUUID)
		if err != nil {
			log.Println("there is an authorization error")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("there is an authorization error"))
			return
		}
		var URLs []string
		if err := json.NewDecoder(req.Body).Decode(&URLs); err != nil {
			log.Println("failed to parse JSON for delete:", err)
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		for _, URL := range URLs {
			log.Println("create task from URL ", URL)
			item := &model.Task{
				UserUUID: userUUID,
				Item: URL,
			}
    		s.ChTask <- *item
		}
		
		w.WriteHeader(http.StatusAccepted)
	}
}