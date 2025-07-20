package service

import (
	"log"
	"math/rand"
	"time"
)

var randGen = rand.New(rand.NewSource(time.Now().UnixNano()))

var OriginShortURLMap = make(map[string]string)

func ReverseMap(shortURL string) (originURL string, ok bool) {
	originURL, ok = OriginShortURLMap[shortURL]
	return originURL, ok
}


func CreateShortURL(originURL string) string {
	const symbols = "zxcvbnmasdfghjklqwertyuiopZXCVBNMASDFGHJKLQWERTYUIOP1234567890"
	const lenURL = 5

	log.Println("check if url was generated")

	for short, origin := range OriginShortURLMap {
		if origin == originURL {
			log.Println("url for this value was already generated")
			return short
		}
	}

	b := make([]byte, lenURL)
	for i := 0; i < lenURL; i++ {
		b[i] = symbols[randGen.Intn(len(symbols))]
	}
	shortURL := string(b)

	OriginShortURLMap[shortURL] = originURL
	log.Printf("url %s has been saved to map \n", shortURL)

	return shortURL
}