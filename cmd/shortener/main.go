package main

import (
	"net/http"
	"log"
	"math/rand"
	"time"
	"io"
	"strings"
)

var originShortUrlMap = make(map[string][]byte)

func validateHeaders(r *http.Request) bool {
    contentType := r.Header.Get("Content-Type")
    return contentType == "text/plain"
}

func createShoterUrl (originUrl string) []byte {
	const sumbols = "zxcvbnmasdfghjklqwertyuiopZXCVBNMASDFGHJKLQWERTYUIOP1234567890"
	const lenUrl = 5

	log.Println("check if url was generated")
	shortUrl, ok := originShortUrlMap[originUrl]
	log.Printf("the result of check was received: %t \n", ok)
	if !ok {
		shortUrl := make([]byte, lenUrl)
		for i := 0; i < lenUrl; i++ {
			shortUrl[i] = sumbols[rand.Intn(len(sumbols))]
		}
		originShortUrlMap[originUrl] = shortUrl
		log.Printf("url %s has been saved to map \n", shortUrl)
	} else {
		log.Println("url for this value was already generated")
	}
	return shortUrl
}

func reverseMap (searchValue string) (originUrl string, ok bool) {
	for k, v := range originShortUrlMap {
		if string(v) == searchValue {
			return k, true
		}
	}
	return "", false
}


func slashUrl(w http.ResponseWriter, req *http.Request) () {
	var validHeaders bool = validateHeaders(req)
	if !validHeaders {
		log.Println("uncorrect headers")
		http.Error(w, "Uncorrect headers were suggested!", http.StatusBadRequest)
		return 
	}

	reqData, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("failed to extract body contents")
		http.Error(w, "failed to extract body contents", http.StatusBadRequest)
		return
	}
	
	originUrl := strings.Split(string(reqData), "=")[1]
	log.Printf("origin URL is extracted: %s \n", originUrl)

	shortUrl := createShoterUrl(originUrl)
	log.Printf("a request was received for URL %s: %s \n", originUrl, shortUrl)
	fullUrl := "http://localhost:8080/" + string(shortUrl)

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullUrl))
	log.Println("processing POST request was completed")
}

func getslashUrl (w http.ResponseWriter, req *http.Request) () {
	var validHeaders bool = validateHeaders(req)
	if !validHeaders {
		log.Println("uncorrect headers")
		http.Error(w, "Uncorrect headers were suggested!", http.StatusBadRequest)
		return 
	}
	shortUrl := req.URL.Path
	if len(shortUrl) > 0 {
		shortUrl = shortUrl[1:]
	} else {
		log.Println("inappropriate user behavior - there is no url")
		http.Error(w, "there is no url", http.StatusBadRequest)
		return
	}
	log.Printf("the short url is %s \n", shortUrl)

	originUrl, ok := reverseMap(shortUrl)
	log.Printf("the result of check was received: %t \n", ok)

	if !ok {
		log.Printf("value for this url doesn't exist in map")
		http.Error(w, "value for this url doesn't exist in map!", http.StatusBadRequest)
		return 
	}

	w.Header().Set("Location", originUrl)
	w.WriteHeader(http.StatusTemporaryRedirect)
	log.Println("processing GET request was completed")
}

func choiceHand(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" &&  req.Method == http.MethodPost{
		log.Println("post api request was recieved")
		slashUrl(w, req)
		return
	} else if req.Method == http.MethodGet {
		log.Println("get api request was recieved")
		getslashUrl(w, req)
		return
	}
	http.Error(w, "this request are not allowed!", http.StatusBadRequest)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	mux := http.NewServeMux()
	
	mux.HandleFunc(`/`, choiceHand)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}