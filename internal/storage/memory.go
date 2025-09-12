package storage

import (
    "log"
    "sync"
	"context"
    "github.com/kujoki/go-musthave-service/internal/model"
)

type MemoryRepository struct {
    Data map[string]string
    URLUserMap map[string][]string
    mu   sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
    return &MemoryRepository{
        Data: make(map[string]string),
        URLUserMap: make(map[string][]string),
    }
}

func (r *MemoryRepository) GetLongURL(shortURL string) (string, bool, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    longURL, ok := r.Data[shortURL]
    return longURL, ok, nil
}

func (r *MemoryRepository) GetShortURL(longURL string) (string, bool, error) {
	r.mu.RLock()
    defer r.mu.RUnlock()
    for short, long := range r.Data {
        if long == longURL {
            return short, true, nil
        }
    }
    return "", false, nil
}

func (r *MemoryRepository) SaveURL(shortURL string, longURL string, userUUID string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.Data[shortURL] = longURL
    r.URLUserMap[userUUID] = append(r.URLUserMap[userUUID], shortURL)
    return nil
}

func (r *MemoryRepository) GetAll() map[string]string {
    r.mu.RLock()
    defer r.mu.RUnlock()
    copyData := make(map[string]string, len(r.Data))
    for k, v := range r.Data {
        copyData[k] = v
    }
    return copyData
}

func (r *MemoryRepository) Close() {

}

func (r *MemoryRepository) Ping(ctx context.Context) bool {
	return false
}

func (r *MemoryRepository) GetUserURLs(userUUID string) ([]model.UserURL, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    shortURLs, ok := r.URLUserMap[userUUID]
    log.Println("shorlURLs", shortURLs)
    if !ok {
        return []model.UserURL{}, nil
    }
    userURLs := []model.UserURL{}
    for _, shortURL := range shortURLs {
        longURL := r.Data[shortURL]
        userURLs = append(
            userURLs, model.UserURL {
                ShortURL: shortURL,
                OriginalURL: longURL,
            },
        )
    }
    return userURLs, nil
}