package storage

import (
    "log"
    "sync"
	"context"
    "github.com/kujoki/go-musthave-service/internal/model"
)

type MemoryRepository struct {
    Data map[string]model.URLRecord
    mu   sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
    return &MemoryRepository{
        Data: make(map[string]model.URLRecord),
    }
}

func (r *MemoryRepository) GetLongURL(shortURL string) (model.LongURLResult, bool, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var longURLRes model.LongURLResult

    rec, ok := r.Data[shortURL]
    if !ok {
        return longURLRes, false, nil
    }
    longURLRes.OriginURL = rec.LongURL
    longURLRes.IsDeleted = rec.IsDeleted
    return longURLRes, true, nil
}

func (r *MemoryRepository) GetShortURL(longURL string) (model.ShortURLResult, bool, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var shortURLRes model.ShortURLResult

    for short, rec := range r.Data {
        if rec.LongURL == longURL {
            shortURLRes.ShortURL = short
            shortURLRes.IsDeleted = rec.IsDeleted
            return shortURLRes, true, nil
        }
    }
    return shortURLRes, false, nil
}

func (r *MemoryRepository) SaveURL(shortURL, longURL, userUUID string) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.Data[shortURL] = model.URLRecord{
        LongURL:   longURL,
        UserUUID: userUUID,
        IsDeleted: false,
    }
    return nil
}

func (r *MemoryRepository) GetAll() map[string]model.URLRecord {
    r.mu.RLock()
    defer r.mu.RUnlock()

    copyData := make(map[string]model.URLRecord, len(r.Data))
    for k, v := range r.Data {
        if !v.IsDeleted {
            copyData[k] = v
        }
    }
    return copyData
}

func (r *MemoryRepository) DeleteURL(ctx context.Context, tasks []model.Task) error {
    if len(tasks) == 0 {
        return nil
    }
    log.Println("start deleting process")
    r.mu.Lock()
    defer r.mu.Unlock()

    for _, task := range tasks {
        rec, ok := r.Data[task.Item]
        if !ok {
            log.Println("there is no such item ", task.Item)
            continue
        }

        rec.IsDeleted = true
        r.Data[task.Item] = rec
        log.Println("the rec was deleted ", task.Item)
    }
    log.Println("process of deleting was finished")
    return nil
}

func (r *MemoryRepository) GetUserURLs(userUUID string) ([]model.UserURL, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    userURLs := []model.UserURL{}
    for short, rec := range r.Data {
        if rec.UserUUID == userUUID && !rec.IsDeleted {
            userURLs = append(userURLs, model.UserURL{
                ShortURL:    short,
                OriginalURL: rec.LongURL,
            })
        }
    }
    return userURLs, nil
}

func (r *MemoryRepository) Close() {}
func (r *MemoryRepository) Ping(ctx context.Context) bool { return false }
