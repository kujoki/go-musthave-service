package main

import (
	"net/http"
	"log"
	"math/rand"
	"time"
	"io"
	"strings"
)

var originshortURLMap = make(map[string][]byte)

var randGen = rand.New(rand.NewSource(time.Now().UnixNano()))

func validateHeaders(r *http.Request) bool {
    contentType := r.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "text/plain")
}

func createShortURL (originURL string) []byte {
	const symbols = "zxcvbnmasdfghjklqwertyuiopZXCVBNMASDFGHJKLQWERTYUIOP1234567890"
	const lenURL = 5

	log.Println("check if url was generated")
	shortURL, ok := originshortURLMap[originURL]
	log.Printf("the result of check was received: %t \n", ok)
	if !ok {
		shortURL := make([]byte, lenURL)
		for i := 0; i < lenURL; i++ {
			shortURL[i] = symbols[randGen.Intn(len(symbols))]
		}
		originshortURLMap[originURL] = shortURL
		log.Printf("url %s has been saved to map \n", shortURL)
	} else {
		log.Println("url for this value was already generated")
	}
	return shortURL
}

func reverseMap (searchValue string) (originURL string, ok bool) {
	for k, v := range originshortURLMap {
		if string(v) == searchValue {
			return k, true
		}
	}
	return "", false
}


func slashURL(w http.ResponseWriter, req *http.Request) () {
	validHeaders := validateHeaders(req)
	w.Header().Set("Content-Type", "text/plain")

	if !validHeaders {
		log.Println("incorrect headers")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`incorrect headers were suggested!`))
		return 
	}

	reqData, err := io.ReadAll(req.Body)
	if err != nil || len(reqData) == 0 {
		log.Println("failed to extract body contents")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`failed to extract body contents`))
		return
	}
	parts := strings.SplitN(string(reqData), "=", 2)

	if len(parts) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid url format"))
		return
	}
	originURL := parts[1]

	log.Printf("origin URL is extracted: %s \n", originURL)

	shortURL := createShortURL(originURL)
	log.Printf("a request was received for URL %s: %s \n", originURL, shortURL)
	fullURL := "http://localhost:8080/" + string(shortURL)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullURL))
	log.Println("processing POST request was completed")
}

func getSlashURL(w http.ResponseWriter, req *http.Request) () {
	validHeaders := validateHeaders(req)

	w.Header().Set("Content-Type", "text/plain")

	if !validHeaders {
		log.Println("incorrect headers")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`incorrect headers were suggested!`))
		return 
	}
	shortURL := req.URL.Path
	if len(shortURL) > 0 {
		shortURL = shortURL[1:]
	} else {
		log.Println("inappropriate user behavior - there is no url")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`there is no url`))
		return
	}
	log.Printf("the short url is %s \n", shortURL)

	originURL, ok := reverseMap(shortURL)
	log.Printf("the result of check was received: %t \n", ok)

	if !ok || originURL == "" {
		log.Printf("value for this url doesn't exist in map")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`value for this url doesn't exist in map!`))
		return 
	}

	w.Header().Set("Location", originURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	log.Println("processing GET request was completed")
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" &&  req.Method == http.MethodPost{
		log.Println("post api request was recieved")
		slashURL(w, req)
		return
	} else if req.Method == http.MethodGet {
		log.Println("get api request was recieved")
		getSlashURL(w, req)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`this request are not allowed!`))
}

func main() {
	mux := http.NewServeMux()
	
	mux.HandleFunc(`/`, handleRequest)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}