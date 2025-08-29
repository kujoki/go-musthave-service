package service

import (
	"log"
	"math/rand"
	"sync"
	"github.com/kujoki/go-musthave-service/internal/model"
	"github.com/kujoki/go-musthave-service/internal/storage"
)

const symbols = "zxcvbnmasdfghjklqwertyuiopZXCVBNMASDFGHJKLQWERTYUIOP1234567890"

type Service struct {
    repo   storage.URLRepository
    lenURL int
    mu     sync.Mutex
}

func NewService(repo storage.URLRepository, lenURL int) *Service {
    return &Service{
        repo:   repo,
        lenURL: lenURL,
    }
}

func CreateURLMap(data []model.Data) (map[string]string, int) {
	URLMap := make(map[string]string)
	lenURL := 5
	for _, item := range data {
        URLMap[item.ShortURL] = item.OriginalURL
		lenURL = len(item.ShortURL)
    }
    return URLMap, lenURL
}


func generateRandomString(length int) string {
    b := make([]byte, length)
    for i := range b {
        b[i] = symbols[rand.Intn(len(symbols))]
    }
    return string(b)
}


func (s *Service) ReverseMap(shortURL string) (string, bool, error) {
    longURL, ok, err := s.repo.GetLongURL(shortURL)
    return longURL, ok, err
}


func (s *Service) CreateShortURL(originURL string) (string, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    if short, found, _ := s.repo.GetShortURL(originURL); found {
        return short, nil
    }

    var shortURL string
    for {
        shortURL = generateRandomString(s.lenURL)

        if _, found, _ := s.repo.GetLongURL(shortURL); !found {
            if err := s.repo.SaveURL(shortURL, originURL); err != nil {
                return "", err
            }
            log.Printf("url %s has been saved -> %s\n", shortURL, originURL)
            return shortURL, nil
        }
    }
}


func (s *Service) AllData() map[string]string {
    return s.repo.GetAll()
}
