package storage

import (
	"context"
	"github.com/kujoki/go-musthave-service/internal/model")

type URLRepository interface {
	GetShortURL(longURL string) (model.ShortURLResult, bool, error)
    GetLongURL(shortURL string) (model.LongURLResult, bool, error)
    SaveURL(shortURL, longURL string, userUUID string) error
    GetAll() map[string]model.URLRecord
	Ping(ctx context.Context) bool
	GetUserURLs(userUUID string) ([]model.UserURL, error)
	DeleteURL(context context.Context, tasks []model.Task) error
	Close()
}