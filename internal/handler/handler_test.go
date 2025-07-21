package handler_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"github.com/kujoki/go-musthave-service/internal/handler"
	"github.com/kujoki/go-musthave-service/internal/service"
	"io"
	"log"
	"strings"
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := strings.NewReader(test.url)

			request := httptest.NewRequest(http.MethodPost, "/", body)
			request.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
            handler.SlashURL(w, request)

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
	// service.OriginShortURLMap["https://practicum.yandex.ru/"] = []byte("OfsO5")

	service.OriginShortURLMap["OfsO5"] = "https://practicum.yandex.ru/"

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
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/" + test.url, nil)
			request.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
            handler.GetSlashURL(w, request)

            res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
		})
	}
}