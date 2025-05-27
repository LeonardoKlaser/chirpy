package main

import (
	"log"
	"net/http"
)


func main() {

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:   ":8080",
		Handler: mux,
	}

	log.Println("Starting server on :8080")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}