package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"github.com/golang-jwt/jwt/v4"
	"github.com/go-chi/chi/v5"
	"github.com/kujoki/go-musthave-service/internal/handler"
	"github.com/kujoki/go-musthave-service/internal/model"
	"github.com/kujoki/go-musthave-service/internal/service"
	"github.com/kujoki/go-musthave-service/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setCookieTest(hasUserUUID bool, builder *handler.JWTBuilder, request *http.Request) () {
	var tokenString string
	type Claims struct {
    	jwt.RegisteredClaims
	}
	if hasUserUUID {
		userUUID := builder.GenerateUserUUID()
		tokenString, _ = builder.SignUserUUID(userUUID)
		request.AddCookie(&http.Cookie{
		Name:  builder.UserCookieName,
		Value: tokenString,
	})
	} else {
		token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims {
			RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 3)),
			},
		},
		)
		tokenString, _ = builder.Sign(token)
		request.AddCookie(&http.Cookie{
		Name:  builder.UserCookieName,
		Value: tokenString,
		})
	}
}


func TestPostHandler(t *testing.T) {
	type want struct {
		code int
		contentType string
	}
	type cookie struct {
		isSet bool
		hasUserUUID bool
	}
	tests := []struct {
		name string
		url string
		userCookie cookie
		want want
	}{
		{
			name: "POST; status code 201",
			url: "https://github.com/golang-standards/project-layout/blob/master/README_ru.md",
			userCookie: cookie{
				isSet: false,
				hasUserUUID: false,
			},
			want: want{
				code: 201,
				contentType: "text/plain",
			},
		}, 
		{
			name: "POST; status code 400", // Bad Request
			url: "",
			userCookie: cookie{
				isSet: false,
				hasUserUUID: false,
			},
			want: want{
				code: 400,
				contentType: "text/plain",
			},
		},
		{
			name: "POST; status code 409", // was used
			url: "https://practicum.yandex.ru/",
			userCookie: cookie{
				isSet: false,
				hasUserUUID: false,
			},			
			want: want{
				code: 409,
				contentType: "text/plain",
			},
		},
		{
			name: "POST; has cookie without userUUID", // has no auth
			url: "https://without.auth.yandex.ru/",
			userCookie: cookie{
				isSet: true,
				hasUserUUID: false,
			},			
			want: want{
				code: 401,
				contentType: "text/plain",
			},
		},
		{
			name: "POST; has cookie with userUUID", // has auth
			url: "https://auth.yandex.ru/",
			userCookie: cookie{
				isSet: true,
				hasUserUUID: true,
			},			
			want: want{
				code: 201,
				contentType: "text/plain",
			},
		},
	}
	repo := storage.NewMemoryRepository()
	s := service.NewService(repo, 5)
	builder := handler.NewJWTBuild("secret", "must-test-service", "auth_user")
		repo.Data["OfsO5"] = model.URLRecord{
		LongURL: "https://practicum.yandex.ru/",
		IsDeleted: false,
		UserUUID: "Kate",
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := strings.NewReader(test.url)
			request := httptest.NewRequest(http.MethodPost, "/", body)

			if test.userCookie.isSet {
				setCookieTest(test.userCookie.hasUserUUID, builder, request)
			}

			request.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
			postSlashHandler := handler.WrapperPostSlash("http://localhost:8080", s, builder)
            postSlashHandler(w, request)

            res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			resBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			log.Println("Response body:", string(resBody))

			if test.want.code == http.StatusCreated {
				assert.NotEmpty(t, strings.TrimSpace(string(resBody)), "Expected not empty for 201")
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	type want struct {
		code int
		contentType string
		location string
	}
	tests := []struct {
		name string
		url string
		want want
	}{
		{
			name: "GET; status code 307",
			url: "OfsO5", 
			want: want{
				code: 307,
				contentType: "text/plain",
				location: "https://practicum.yandex.ru/",
			},
		},
		{
			name: "GET; status code 400",
			url: "11111",
			want: want{
				code: 400,
				contentType: "text/plain",
				location: "",
			},
		},
	}
	repo := storage.NewMemoryRepository()
	s := service.NewService(repo, 5)
	builder := handler.NewJWTBuild("secret", "must-test-service", "auth_user")
	repo.Data["OfsO5"] = model.URLRecord{
		LongURL: "https://practicum.yandex.ru/",
		IsDeleted: false,
		UserUUID: "Kate",
	}

	r := chi.NewRouter()
	r.Get("/{ID}", handler.WrapperGetSlashURL(s, builder))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/" + test.url, nil)
			request.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
            getSlashHandler := handler.WrapperGetSlashURL(s, builder)
			getSlashHandler(w, request)

            res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
		})
	}
}

func TestPostAPIShortHandler(t *testing.T) {
	type want struct {
		code int
		contentType string
	}
	tests := []struct {
		name string
		url string
		want want
	}{
		{
			name: "POST; status code 201",
			url: "https://github.com/golang-standards/project-layout/blob/master/README_ru.md",
			want: want{
				code: 201,
				contentType: "application/json",
			},
		},
		{
			name: "POST; status code 409",
			url: "https://practicum.yandex.ru/", // was used
			want: want{
				code: 409,
				contentType: "application/json",
			},
		}, 
		{
			name: "POST; status code 400", // Bad Request
			url: "",
			want: want{
				code: 400,
				contentType: "application/json",
			},
		},
	}
	repo := storage.NewMemoryRepository()
	s := service.NewService(repo, 5)
	builder := handler.NewJWTBuild("secret", "must-test-service", "auth_user")
	repo.Data["OfsO5"] = model.URLRecord{
		LongURL: "https://practicum.yandex.ru/",
		IsDeleted: false,
		UserUUID: "Kate",
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(map[string]string{
				"url": test.url,
			})
			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(jsonBody))
			request.Header.Set("Content-Type", "application/jsonn")

			w := httptest.NewRecorder()
			postSlashHandler := handler.WrapperPostAPIShort("http://localhost:8080", s, builder)
            postSlashHandler(w, request)

            res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			resBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			log.Println("Response body:", string(resBody))

			if test.want.code == http.StatusCreated {
				assert.NotEmpty(t, strings.TrimSpace(string(resBody)), "Expected not empty for 201")
			}
		},
		)
	}
}

func TesPostBatchAPI(t *testing.T) {
	type want struct {
		code int
		contentType string
		respData []model.BatchResponse
	}
	tests := []struct {
		name string
		reqData []model.BatchRequest
		want want
	}{
		{
			name: "POST; status code 201",
			reqData: []model.BatchRequest{
				{
					CorrelationID: "uuid",
					OriginalURL: "https://practicum.yandex.ru/",
			},
			},
			want: want{
				code: 201,
				contentType: "application/json",
				respData: []model.BatchResponse{
					{
						CorrelationID: "uuid",
						ShortURL: "http://localhost:8080/OfsO5",
				},
			},
		},
		},
	}
	repo := storage.NewMemoryRepository()
	s := service.NewService(repo, 5)
	builder := handler.NewJWTBuild("secret", "must-test-service", "auth_user")
	repo.Data["OfsO5"] = model.URLRecord{
		LongURL: "https://practicum.yandex.ru/",
		IsDeleted: false,
		UserUUID: "Kate",
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(test.reqData)
			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(jsonBody))
			request.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			postBatchHandler := handler.WrapperPostBatchAPI("http://localhost:8080", s, builder)
            postBatchHandler(w, request)

            res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			resBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			log.Println("Response body:", string(resBody))

			if test.want.code == http.StatusCreated {
				assert.NotEmpty(t, strings.TrimSpace(string(resBody)), "Expected not empty for 201")
			}
		},
		)
	}
}

func TestDeleteHandler(t *testing.T) {
	type want struct {
		code int
		contentType string
	}
	type cookie struct {
		isSet bool
		hasUserUUID bool
	}
	tests := []struct {
		name string
		shortURLs []string
		userCookie cookie
		want want
	}{
		{
			name: "DELETE; status code 202",
			shortURLs: []string{"6qxTVvsy", "RTfd56hn", "Jlfd67ds"},
			userCookie: cookie{
				isSet: true,
				hasUserUUID: true,
			},
			want: want{
				code: 202,
				contentType: "application/json",
			},
		}, 
	}
	repo := storage.NewMemoryRepository()
	s := service.NewService(repo, 5)
	builder := handler.NewJWTBuild("secret", "must-test-service", "auth_user")
		repo.Data["OfsO5"] = model.URLRecord{
		LongURL: "https://practicum.yandex.ru/",
		IsDeleted: false,
		UserUUID: "Kate",
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.shortURLs)
			assert.NoError(t, err)

			request := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
			
			if test.userCookie.isSet {
				setCookieTest(test.userCookie.hasUserUUID, builder, request)
			}

			request.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
			postSlashHandler := handler.WrapperDeleteURLs(s, builder)
            postSlashHandler(w, request)

            res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			resBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			log.Println("Response body:", string(resBody))
		})
	}
}
