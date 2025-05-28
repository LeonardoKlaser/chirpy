package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"text/template"
	"time"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/leonardoklaser/Chirpy/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct{
	fileserverHits atomic.Int32
	DB *database.Queries
	Environment string
}

type errorResponse struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, statusCode int, msg string){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := errorResponse{Error: msg}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}

func respondWithJson(w http.ResponseWriter, statusCode int, data interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if data != nil {
			if err := json.NewEncoder(w).Encode(data); err != nil{
				log.Printf("Error encoding json response: %v", err)
			}
		}
}

func formatProfane(originalText string, sanitizedText string) (string, error){
	words := strings.Split(sanitizedText, " ")
	wordsToReturn := strings.Split(originalText, " ")
	if len(words) != len(wordsToReturn){
		return "", errors.New("words whit diferent sizes, cant format it")
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

//func (cfg *apiConfig) handleReset(w http.ResponseWriter, _ *http.Request){
//	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
//	w.WriteHeader(http.StatusOK)
//	cfg.fileserverHits.Store(0)
//}

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
	type ResponseType struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := ReturnType{}
	err := decoder.Decode(&params)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(params.Body) > 140{
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	replacer := strings.NewReplacer("kerfuffle", "****", "sharbert", "****", "fornax", "****")
	cleanedBody := replacer.Replace(strings.ToLower(params.Body))
	responseCleanedBody, err := formatProfane(params.Body, cleanedBody)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Error formatting profane words")
		return
	}

	response := ResponseType{CleanedBody: responseCleanedBody}
	respondWithJson(w, http.StatusOK, response)


}

func (cfg *apiConfig) PostUser(w http.ResponseWriter, r *http.Request ){
	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	type requestBody struct{
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	log.Printf(params.Email)
	user, err := cfg.DB.CreateUser(r.Context(), params.Email)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Error to insert new User in Database")
		return
	}

	userToReturn := User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email}
	respondWithJson(w, http.StatusCreated, userToReturn)


}

func (cfg *apiConfig) DeleteUsers(w http.ResponseWriter, r *http.Request){
	if cfg.Environment != "dev" {
		respondWithError(w, http.StatusForbidden, "This action is available only on development environment")
		return
	}
	_, err := cfg.DB.DeleteUsers(r.Context())
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Error to delete user")
		return
	}
	
	var nullInterface interface{}
	respondWithJson(w, http.StatusOK, nullInterface)
}

func main() {

	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil{
		log.Printf("cant connect to the database")
	}
	dbQueries := database.New(db)
	environment := os.Getenv("PLATFORM")
	apiCfg := apiConfig{
		DB : dbQueries,
		Environment: environment,
	}
	router := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))

	strippedFileServerHandler := http.StripPrefix("/app/", fileServer)
	metricsWrappedFileServerHandler := apiCfg.MiddlewareMetricsInc(strippedFileServerHandler)

	router.HandleFunc("/app/", metricsWrappedFileServerHandler.ServeHTTP)

	router.HandleFunc("GET /api/healthz", handleHealthz) 

	router.HandleFunc("GET /admin/metrics", apiCfg.HandleMetrics)

	router.HandleFunc("POST /admin/reset", apiCfg.DeleteUsers)

	router.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	router.HandleFunc("POST /api/users", apiCfg.PostUser)

	server := &http.Server{
		Addr:   ":8080",
		Handler: router,
	}

	log.Println("Starting server on :8080")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
