package main

import (
	"net/http"
	"log"
	"github.com/kujoki/go-musthave-service/internal/handler"
)

func handleRequest(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" &&  req.Method == http.MethodPost{
		log.Println("post api request was recieved")
		handler.SlashURL(w, req)
		return
	} else if req.Method == http.MethodGet {
		log.Println("get api request was recieved")
		handler.GetSlashURL(w, req)
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