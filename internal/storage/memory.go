package storage

import (
    "sync"
	"context"
)

type MemoryRepository struct {
    Data map[string]string
    mu   sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
    return &MemoryRepository{
        Data: make(map[string]string),
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

func (r *MemoryRepository) SaveURL(shortURL, longURL string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.Data[shortURL] = longURL
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
