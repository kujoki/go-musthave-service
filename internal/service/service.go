package service

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"
	"github.com/kujoki/go-musthave-service/internal/model"
	"github.com/kujoki/go-musthave-service/internal/storage"
)

const symbols = "zxcvbnmasdfghjklqwertyuiopZXCVBNMASDFGHJKLQWERTYUIOP1234567890"

type Service struct {
	Repo   storage.URLRepository
	lenURL int
    ChTask chan model.Task
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewService(repo storage.URLRepository, lenURL int) *Service {
	ctx, cancel := context.WithCancel(context.Background())
    s := &Service{
        Repo:   repo,
        lenURL: lenURL,
        ChTask: make(chan model.Task, 1024),
        ctx:    ctx,
        cancel: cancel,
    }

    go s.DeleteURLs(ctx)
    return s
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = symbols[rand.Intn(len(symbols))]
	}
	return string(b)
}

func (s *Service) CheckOriginURLValue(shortURL string) (model.LongURLResult, bool, error) {
	longURLRes, ok, err := s.Repo.GetLongURL(shortURL)
	return longURLRes, ok, err
}

func (s *Service) CheckShortURLValue(longURL string) (model.ShortURLResult, bool, error) {
	ShortURLResult, ok, err := s.Repo.GetShortURL(longURL)
	return ShortURLResult, ok, err
}

func (s *Service) CreateShortURL(originURL string, userUUID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if shortRes, found, _ := s.Repo.GetShortURL(originURL); found && !shortRes.IsDeleted  {
		return shortRes.ShortURL, nil
	}

	var shortURL string
	for {
		shortURL = generateRandomString(s.lenURL)

		if _, found, _ := s.Repo.GetLongURL(shortURL); !found {
			if err := s.Repo.SaveURL(shortURL, originURL, userUUID); err != nil {
				return "", err
			}
			log.Printf("url %s has been saved -> %s\n", shortURL, originURL)
			return shortURL, nil
		}
	}
}

func (s *Service) AllData() map[string]model.URLRecord {
	return s.Repo.GetAll()
}

func (s *Service) GetURLByUser(userUUID string) ([]model.UserURL, error) {
    userURLs, err := s.Repo.GetUserURLs(userUUID)
	return userURLs, err
}

func (s *Service) Stop() {
	log.Println("stop service")
    s.cancel()
}

func (s *Service) DeleteURLs(ctx context.Context) { 

    ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

    var messages []model.Task

	flush := func() {
        if len(messages) == 0 {
            return
        }
        if err := s.Repo.DeleteURL(ctx, messages); err != nil {
            log.Println("cannot delete messages:", err)
            return
        }
        messages = messages[:0] 
    }

    for {
        select {
        case <-ctx.Done():
			log.Println("context is done, so flush")
            flush()
            return

        case msg := <-s.ChTask:
            messages = append(messages, msg)

            if len(messages) >= 50 {
				log.Println("too many messages, so flush")
                flush()
            }

        case <-ticker.C:
            flush()
        }
    }
}