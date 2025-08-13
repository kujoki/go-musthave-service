package service

import (
	"log"
	"math/rand"
	"sync"
	"strconv"
	"github.com/kujoki/go-musthave-service/internal/model"
)

const symbols = "zxcvbnmasdfghjklqwertyuiopZXCVBNMASDFGHJKLQWERTYUIOP1234567890"

type Service struct {
	URLMap map[string]string
	lenURL int
	mu sync.RWMutex
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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

func NewService(data []model.Data) *Service {
	URLMap, lenURL := CreateURLMap(data)
	return &Service{
		URLMap: URLMap,
		lenURL: lenURL,
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

	log.Println("check if url was generated")
	log.Println("URLMap", s.URLMap)
	for short, origin := range s.URLMap {
		if origin == originURL {
			log.Println("url for this value was already generated")
			return short
		}
	}

	var shortURL string

	for {
		shortURL = generateRandomString(s.lenURL)
		if _, exists := s.URLMap[shortURL]; !exists {
			s.URLMap[shortURL] = originURL
			log.Printf("url %s has been saved to map \n", shortURL)
			return shortURL
		}
	}
}

func (s *Service) AllData() []model.Data {
    s.mu.RLock()
    defer s.mu.RUnlock()

    data := make([]model.Data, 0, len(s.URLMap))
	i := 1 
    for short, orig := range s.URLMap {
        data = append(data, model.Data{
            UUID:        strconv.Itoa(i),
            ShortURL:    short,
            OriginalURL: orig,
        })
		i++
    }
    return data
}