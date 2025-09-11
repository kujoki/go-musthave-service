package service

import (
	"github.com/kujoki/go-musthave-service/internal/model"
	"github.com/kujoki/go-musthave-service/internal/storage"
	"log"
	"math/rand"
	"sync"
)

const symbols = "zxcvbnmasdfghjklqwertyuiopZXCVBNMASDFGHJKLQWERTYUIOP1234567890"

type Service struct {
	Repo   storage.URLRepository
	lenURL int
	mu     sync.Mutex
}

func NewService(repo storage.URLRepository, lenURL int) *Service {
	return &Service{
		Repo:   repo,
		lenURL: lenURL,
	}
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = symbols[rand.Intn(len(symbols))]
	}
	return string(b)
}

func (s *Service) CheckOriginURLValue(shortURL string) (string, bool, error) {
	longURL, ok, err := s.Repo.GetLongURL(shortURL)
	return longURL, ok, err
}

func (s *Service) CheckShortURLValue(longURL string) (string, bool, error) {
	shortURL, ok, err := s.Repo.GetShortURL(longURL)
	return shortURL, ok, err
}

func (s *Service) CreateShortURL(originURL string, userUUID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if short, found, _ := s.Repo.GetShortURL(originURL); found {
		return short, nil
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

func (s *Service) AllData() map[string]string {
	return s.Repo.GetAll()
}

func (s *Service) GetURLByUser(userUUID string) ([]model.UserURL, error) {
    userURLs, err := s.Repo.GetUserURLs(userUUID)
	return userURLs, err
}
