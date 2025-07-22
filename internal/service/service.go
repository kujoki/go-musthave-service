package service

import (
	"log"
	"math/rand"
	"sync"
)

const symbols = "zxcvbnmasdfghjklqwertyuiopZXCVBNMASDFGHJKLQWERTYUIOP1234567890"

type Service struct {
	URLMap map[string]string
	mu sync.RWMutex
}

func NewService() *Service {
	return &Service{
		URLMap: make(map[string]string),
	}
}

func (s *Service) ReverseMap(shortURL string) (originURL string, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	originURL, ok = s.URLMap[shortURL]
	return originURL, ok
}

func generateRandomString(length int) string {
	b := make([]byte, length)
    for i := range b {
        b[i] = symbols[rand.Intn(len(symbols))]
    }
    return string(b)
}


func (s *Service) CreateShortURL(originURL string) string {
	s.mu.Lock() 
	defer s.mu.Unlock()
	const lenURL = 5

	log.Println("check if url was generated")

	for short, origin := range s.URLMap {
		if origin == originURL {
			log.Println("url for this value was already generated")
			return short
		}
	}

	var shortURL string

	for {
		shortURL = generateRandomString(lenURL)
		if _, exists := s.URLMap[shortURL]; !exists {
			s.URLMap[shortURL] = originURL
			log.Printf("url %s has been saved to map \n", shortURL)
			return shortURL
		}
	}
}