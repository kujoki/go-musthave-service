package storage

import "context"

type URLRepository interface {
	GetShortURL(longURL string) (string, bool, error)
    GetLongURL(shortURL string) (string, bool, error)
    SaveURL(shortURL, longURL string) error
    GetAll() map[string]string
	Ping(ctx context.Context) bool
	Close()
}