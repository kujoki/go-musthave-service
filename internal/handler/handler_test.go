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

	"github.com/go-chi/chi/v5"
	"github.com/kujoki/go-musthave-service/internal/handler"
	"github.com/kujoki/go-musthave-service/internal/model"
	"github.com/kujoki/go-musthave-service/internal/service"
	"github.com/kujoki/go-musthave-service/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostHandler(t *testing.T) {
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
				contentType: "text/plain",
			},
		}, 
		{
			name: "POST; status code 400", // Bad Request
			url: "",
			want: want{
				code: 400,
				contentType: "text/plain",
			},
		},
		{
			name: "POST; status code 409", // was used
			url: "https://practicum.yandex.ru/",
			want: want{
				code: 409,
				contentType: "text/plain",
			},
		},
	}
	repo := storage.NewMemoryRepository()
	s := service.NewService(repo, 5)
	repo.Data["OfsO5"] = "https://practicum.yandex.ru/"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := strings.NewReader(test.url)

			request := httptest.NewRequest(http.MethodPost, "/", body)
			request.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
			postSlashHandler := handler.WrapperPostSlash("http://localhost:8080", s)
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
	repo.Data["OfsO5"] = "https://practicum.yandex.ru/"

	r := chi.NewRouter()
	r.Get("/{ID}", handler.WrapperGetSlashURL(s))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/" + test.url, nil)
			request.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
            getSlashHandler := handler.WrapperGetSlashURL(s)
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
	repo.Data["OfsO5"] = "https://practicum.yandex.ru/"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(map[string]string{
				"url": test.url,
			})
			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(jsonBody))
			request.Header.Set("Content-Type", "application/jsonn")

			w := httptest.NewRecorder()
			postSlashHandler := handler.WrapperPostAPIShort("http://localhost:8080", s)
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
	repo.Data["OfsO5"] = "https://practicum.yandex.ru/"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(test.reqData)
			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(jsonBody))
			request.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			postBatchHandler := handler.WrapperPostBatchAPI("http://localhost:8080", s)
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