package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"text/template"
	"strings"
	"errors"
)

type apiConfig struct{
	fileserverHits atomic.Int32
}

func formatProfane(originalText string, sanitizedText string) (string, error){
	words := strings.Split(sanitizedText, " ")
	wordsToReturn := strings.Split(originalText, " ")
	if len(words) != len(wordsToReturn){
		return "", errors.New("Words whit diferent sizes, cant format it")
	}
	for i:=0; i < len(words); i++{
		if(strings.Contains(words[i], "****")){
			wordsToReturn[i] = words[i]
		}
	}
	return strings.Join(wordsToReturn, " "), nil
}


func handleHealthz (w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_,err := w.Write([]byte("OK"))
	if err != nil{
		log.Printf("Erro ao inserir corpo da requisicao")
	}

}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, _ *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

// func (cfg *apiConfig) HandleMetrics(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
// 	w.WriteHeader(http.StatusOK)
// 	metricsBody := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())
// 	_, err := w.Write([]byte(metricsBody))
// 	if err != nil {
// 		log.Printf("Error writing response body: %v", err)
// 		return
// 	}
// }

func (cfg *apiConfig) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	metricsBody := cfg.fileserverHits.Load()

	tmpl, err := template.ParseFiles("countHits.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, metricsBody)
}

func (cfg *apiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type ReturnType struct{
		Body string `json:"body"`
	}

	type ErrorType struct {
		Error string `json:"error"`
	}
	type ResponseType struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := ReturnType{}
	err := decoder.Decode(&params)
	if err != nil{
		log.Printf("Error decoding request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := ErrorType{Error: "Something went wrong"}
		jsonResponse, _ := json.Marshal(errorResponse)
		_, err := w.Write(jsonResponse)
		if err != nil{
			log.Printf("Error writing response body: %v", err)
		}
		return
	}

	if len(params.Body) > 140{
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := ErrorType{Error: "Chirp is too long"}
		jsonResponse, _ := json.Marshal(errorResponse)
		_, err := w.Write(jsonResponse)
		if err != nil{
			log.Printf("Error writing response body: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	replacer := strings.NewReplacer("kerfuffle", "****", "sharbert", "****", "fornax", "****")
	cleanedBody := replacer.Replace(strings.ToLower(params.Body))
	responseCleanedBody, err := formatProfane(params.Body, cleanedBody)
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		errorFormatStruct := ErrorType{Error: err.Error()}
		errorReturn, _ := json.Marshal(errorFormatStruct)
		_, err := w.Write(errorReturn)
		if err != nil {
			log.Printf("Error to format cleaned body")
		}
	}
	response := ResponseType{CleanedBody: responseCleanedBody}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Printf("Error writing response body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}


}

func main() {

	var apiCfg apiConfig
	router := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))

	strippedFileServerHandler := http.StripPrefix("/app/", fileServer)
	metricsWrappedFileServerHandler := apiCfg.MiddlewareMetricsInc(strippedFileServerHandler)

	router.HandleFunc("/app/", metricsWrappedFileServerHandler.ServeHTTP)

	router.HandleFunc("GET /api/healthz", handleHealthz) 

	router.HandleFunc("GET /admin/metrics", apiCfg.HandleMetrics)

	router.HandleFunc("POST /admin/reset", apiCfg.handleReset)

	router.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

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
