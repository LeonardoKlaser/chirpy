package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct{
	fileserverHits atomic.Int32
}

func handleHealthz (w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_,err := w.Write([]byte("OK"))
	if err != nil{
		log.Printf("Erro ao inserir corpo da requisicao")
	}

}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	
	
}

func main() {

	router := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))

	strippedFileServerHandler := http.StripPrefix("/app/", fileServer)

	router.HandleFunc("/app/", func(w http.ResponseWriter, r *http.Request) {
		strippedFileServerHandler.ServeHTTP(w, r)
	})

	router.HandleFunc("/healthz", handleHealthz) 

	server := &http.Server{
		Addr:   ":8080",
		Handler: router,
	}

	log.Println("Starting server on :8080")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
