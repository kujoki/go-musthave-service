package storage

import (
	"context"
	"github.com/kujoki/go-musthave-service/internal/model")

type URLRepository interface {
	GetShortURL(longURL string) (string, bool, error)
    GetLongURL(shortURL string) (string, bool, error)
    SaveURL(shortURL, longURL string, userUUID string) error
    GetAll() map[string]string
	Ping(ctx context.Context) bool
	GetUserURLs(userUUID string) ([]model.UserURL, error)
	Close()
}