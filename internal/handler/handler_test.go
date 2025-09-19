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
	"github.com/kujoki/go-musthave-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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

	cfg := config.Config{
        BaseURL:         "http://localhost:8080",
    }

	testSugar := zap.NewExample().Sugar()
	repo := storage.NewMemoryRepository()
	s := service.NewService(repo, 5, *testSugar)

	builder := handler.NewJWTBuild("secret", "must-test-service", "auth_user", *testSugar)
		repo.Data["OfsO5"] = model.URLRecord{
		LongURL: "https://practicum.yandex.ru/",
		IsDeleted: false,
		UserUUID: "Kate",
	}

	r := chi.NewRouter()
    r.Use(handler.AuthMiddleware(builder))
	r.Post("/", handler.WrapperPostSlash(cfg.BaseURL, s))

    ts := httptest.NewServer(r)
    defer ts.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := strings.NewReader(test.url)
        	request, err := http.NewRequest(http.MethodPost, ts.URL+"/", body)
			require.NoError(t, err)

			if test.userCookie.isSet {
				setCookieTest(test.userCookie.hasUserUUID, builder, request)
			}

			request.Header.Set("Content-Type", "text/plain")

			resp, err := ts.Client().Do(request)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)
			assert.True(
				t,
				strings.HasPrefix(resp.Header.Get("Content-Type"), test.want.contentType),
				"expected Content-Type to start with %s, got %s",
				test.want.contentType,
				resp.Header.Get("Content-Type"),
			)

			resBody, err := io.ReadAll(resp.Body)
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
	cfg := config.Config{
        SecretToken:     "test-secret",
        ApplicationName: "test-app",
        UserCookieName:  "auth",
    }

	testSugar := zap.NewExample().Sugar()
	repo := storage.NewMemoryRepository()
	s := service.NewService(repo, 5, *testSugar)

	builder := handler.NewJWTBuild(cfg.SecretToken, cfg.ApplicationName, cfg.UserCookieName, *testSugar)
	repo.Data["OfsO5"] = model.URLRecord{
		LongURL: "https://practicum.yandex.ru/",
		IsDeleted: false,
		UserUUID: "Kate",
	}

	r := chi.NewRouter()
	r.Use(handler.AuthMiddleware(builder))
	r.Get("/{ID}", handler.WrapperGetSlashURL(s))

	ts := httptest.NewServer(r)
    defer ts.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
        	request, err := http.NewRequest(http.MethodGet, ts.URL+"/"+test.url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "text/plain")
			client := ts.Client() // auto redirect 
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}

			resp, err := client.Do(request)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)
			assert.True(
				t,
				strings.HasPrefix(resp.Header.Get("Content-Type"), test.want.contentType),
				"expected Content-Type to start with %s, got %s",
				test.want.contentType,
				resp.Header.Get("Content-Type"),
			)
			assert.Equal(t, test.want.location, resp.Header.Get("Location"))
		})
	}
}

func TestPostAPIShortHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "POST; status code 201",
			url:  "https://github.com/golang-standards/project-layout/blob/master/README_ru.md",
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name: "POST; status code 409",
			url:  "https://practicum.yandex.ru/", // already used
			want: want{
				code:        409,
				contentType: "application/json",
			},
		},
		{
			name: "POST; status code 400", // Bad Request
			url:  "",
			want: want{
				code:        400,
				contentType: "application/json",
			},
		},
	}

	cfg := config.Config{
		SecretToken:     "test-secret",
		ApplicationName: "test-app",
		UserCookieName:  "auth",
		BaseURL:         "http://localhost:8080",
	}

	testSugar := zap.NewExample().Sugar()
	repo := storage.NewMemoryRepository()
	s := service.NewService(repo, 5, *testSugar)

	builder := handler.NewJWTBuild(cfg.SecretToken, cfg.ApplicationName, cfg.UserCookieName, *testSugar)

	repo.Data["OfsO5"] = model.URLRecord{
		LongURL:   "https://practicum.yandex.ru/",
		IsDeleted: false,
		UserUUID:  "Kate",
	}

	r := chi.NewRouter()
	r.Use(handler.AuthMiddleware(builder))
	r.Post("/api/shorten", handler.WrapperPostAPIShort(cfg.BaseURL, s))

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(map[string]string{
				"url": test.url,
			})
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten", bytes.NewReader(jsonBody))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)
			assert.True(
				t,
				strings.HasPrefix(resp.Header.Get("Content-Type"), test.want.contentType),
				"expected Content-Type to start with %s, got %s",
				test.want.contentType,
				resp.Header.Get("Content-Type"),
			)

			resBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			log.Println("Response body:", string(resBody))

			if test.want.code == http.StatusCreated {
				assert.NotEmpty(t, strings.TrimSpace(string(resBody)), "Expected not empty for 201")
			}
		})
	}
}


func TestPostBatchAPI(t *testing.T) {
	type want struct {
		code        int
		contentType string
		respData    []model.BatchResponse
	}

	tests := []struct {
		name    string
		reqData []model.BatchRequest
		want    want
	}{
		{
			name: "POST; status code 201",
			reqData: []model.BatchRequest{
				{
					CorrelationID: "uuid",
					OriginalURL:   "https://practicum.yandex.ru/",
				},
			},
			want: want{
				code:        201,
				contentType: "application/json",
				respData: []model.BatchResponse{
					{
						CorrelationID: "uuid",
						ShortURL:      "http://localhost:8080/OfsO5",
					},
				},
			},
		},
	}

	cfg := config.Config{
		SecretToken:     "test-secret",
		ApplicationName: "test-app",
		UserCookieName:  "auth",
		BaseURL:         "http://localhost:8080",
	}

	testSugar := zap.NewExample().Sugar()
	repo := storage.NewMemoryRepository()
	s := service.NewService(repo, 5, *testSugar)

	builder := handler.NewJWTBuild(cfg.SecretToken, cfg.ApplicationName, cfg.UserCookieName, *testSugar)

	repo.Data["OfsO5"] = model.URLRecord{
		LongURL:   "https://practicum.yandex.ru/",
		IsDeleted: false,
		UserUUID:  "Kate",
	}

	r := chi.NewRouter()
	r.Use(handler.AuthMiddleware(builder))
	r.Post("/api/shorten/batch", handler.WrapperPostBatchAPI(cfg.BaseURL, s))

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(test.reqData)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten/batch", bytes.NewReader(jsonBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)

			assert.True(
				t,
				strings.HasPrefix(resp.Header.Get("Content-Type"), test.want.contentType),
				"expected Content-Type to start with %s, got %s",
				test.want.contentType,
				resp.Header.Get("Content-Type"),
			)

			resBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			log.Println("Response body:", string(resBody))

			if test.want.code == http.StatusCreated {
				assert.NotEmpty(t, strings.TrimSpace(string(resBody)), "Expected not empty for 201")

				var gotResp []model.BatchResponse
				err := json.Unmarshal(resBody, &gotResp)
				require.NoError(t, err)
				assert.Equal(t, test.want.respData, gotResp)
			}
		})
	}
}


func TestDeleteHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	type cookie struct {
		isSet       bool
		hasUserUUID bool
	}

	tests := []struct {
		name       string
		shortURLs  []string
		userCookie cookie
		want       want
	}{
		{
			name:      "DELETE; status code 202",
			shortURLs: []string{"6qxTVvsy", "RTfd56hn", "Jlfd67ds"},
			userCookie: cookie{
				isSet:       true,
				hasUserUUID: true,
			},
			want: want{
				code:        202,
				contentType: "application/json",
			},
		},
	}

	cfg := config.Config{
		SecretToken:     "test-secret",
		ApplicationName: "test-app",
		UserCookieName:  "auth",
	}

	testSugar := zap.NewExample().Sugar()
	repo := storage.NewMemoryRepository()
	s := service.NewService(repo, 5, *testSugar)

	builder := handler.NewJWTBuild(cfg.SecretToken, cfg.ApplicationName, cfg.UserCookieName, *testSugar)

	repo.Data["6qxTVvsy"] = model.URLRecord{
		LongURL:   "https://example1.com",
		IsDeleted: false,
		UserUUID:  "Kate",
	}
	repo.Data["RTfd56hn"] = model.URLRecord{
		LongURL:   "https://example2.com",
		IsDeleted: false,
		UserUUID:  "Kate",
	}
	repo.Data["Jlfd67ds"] = model.URLRecord{
		LongURL:   "https://example3.com",
		IsDeleted: false,
		UserUUID:  "Kate",
	}

	r := chi.NewRouter()
	r.Use(handler.AuthMiddleware(builder))
	r.Delete("/api/user/urls", handler.WrapperDeleteURLs(s))

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.shortURLs)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodDelete, ts.URL+"/api/user/urls", bytes.NewReader(body))
			require.NoError(t, err)

			if test.userCookie.isSet {
				setCookieTest(test.userCookie.hasUserUUID, builder, request)
			}

			request.Header.Set("Content-Type", "application/json")

			resp, err := ts.Client().Do(request)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)

			assert.True(
				t,
				strings.HasPrefix(resp.Header.Get("Content-Type"), test.want.contentType),
				"expected Content-Type to start with %s, got %s",
				test.want.contentType,
				resp.Header.Get("Content-Type"),
			)

			resBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			log.Println("Response body:", string(resBody))
		})
	}
}
