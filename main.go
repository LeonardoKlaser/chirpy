package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/leonardoklaser/Chirpy/internal/auth"
	"github.com/leonardoklaser/Chirpy/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	Environment    string
	SecretKey      string
}

type errorResponse struct {
	Error string `json:"error"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func respondWithError(w http.ResponseWriter, statusCode int, msg string) {
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
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding json response: %v", err)
		}
	}
}

func respondWithListJson(w http.ResponseWriter, statusCode int, data []Chirp) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding json response: %v", err)
		}
	}
}

func formatProfane(originalText string, sanitizedText string) (string, error) {
	words := strings.Split(sanitizedText, " ")
	wordsToReturn := strings.Split(originalText, " ")
	if len(words) != len(wordsToReturn) {
		return "", errors.New("words whit diferent sizes, cant format it")
	}
	for i := 0; i < len(words); i++ {
		if strings.Contains(words[i], "****") {
			wordsToReturn[i] = words[i]
		}
	}
	return strings.Join(wordsToReturn, " "), nil
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
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
	type ReturnType struct {
		Body string `json:"body"`
	}
	type ResponseType struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := ReturnType{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	replacer := strings.NewReplacer("kerfuffle", "****", "sharbert", "****", "fornax", "****")
	cleanedBody := replacer.Replace(strings.ToLower(params.Body))
	responseCleanedBody, err := formatProfane(params.Body, cleanedBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error formatting profane words")
		return
	}

	response := ResponseType{CleanedBody: responseCleanedBody}
	respondWithJson(w, http.StatusOK, response)

}

func (cfg *apiConfig) PostUser(w http.ResponseWriter, r *http.Request) {
	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	passwordHashed, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error to has password")
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, Password: passwordHashed})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error to insert new User in Database")
		return
	}

	userToReturn := User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email}
	respondWithJson(w, http.StatusCreated, userToReturn)

}

func (cfg *apiConfig) DeleteUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.Environment != "dev" {
		respondWithError(w, http.StatusForbidden, "This action is available only on development environment")
		return
	}
	_, err := cfg.DB.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error to delete user: %v", err))
		return
	}

	var nullInterface interface{}
	respondWithJson(w, http.StatusOK, nullInterface)
}

func (cfg *apiConfig) ListChirps(w http.ResponseWriter, r *http.Request) {

	var chirps []database.Chirp
	chirps, err := cfg.DB.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error to list all chirps: %v", err))
		return
	}

	var returnChirps []Chirp

	for _, val := range chirps {
		newChirp := Chirp{
			ID:        val.ID,
			CreatedAt: val.CreatedAt,
			UpdatedAt: val.UpdatedAt,
			Body:      val.Body,
			UserId:    val.UserID,
		}
		returnChirps = append(returnChirps, newChirp)
	}

	respondWithListJson(w, http.StatusOK, returnChirps)
}

func (cfg *apiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error to retrieve chirp Id : %v ", err))
		return
	}

	chirp, err := cfg.DB.GetChirpById(r.Context(), uid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Chirp with ID %s not found", id))
		return
	}

	chirpToReturn := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
	respondWithJson(w, http.StatusOK, chirpToReturn)
}

func (cfg *apiConfig) PostChirps(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	type responseBody struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing Bearer token")
		return
	}
	if token == "" {
		respondWithError(w, http.StatusUnauthorized, "Bearer token is empty")
		return
	}
	
	uuidUser, err := auth.ValidateJWT(token, cfg.SecretKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid Bearer token: %v", err))
			return
	}
	
	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Input")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	replacer := strings.NewReplacer("kerfuffle", "****", "sharbert", "****", "fornax", "****")
	cleanedBody := replacer.Replace(strings.ToLower(params.Body))
	responseCleanedBody, err := formatProfane(params.Body, cleanedBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error to format profane words")
		return
	}

	createChirpParam := database.CreateChirpParams{Body: responseCleanedBody, UserID: uuidUser}

	chirp, err := cfg.DB.CreateChirp(r.Context(), createChirpParam)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error to Create new Chirp: %v, json: %v", err, createChirpParam) )
		return
	}

	response := responseBody{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserId: chirp.UserID}
	respondWithJson(w, http.StatusCreated, response)

}

func (cfg *apiConfig) LoginUser(w http.ResponseWriter, r *http.Request) {
	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token string `json:"token"`
		Refresh_token string `json:"refresh_token"`
	}

	type requestBody struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	user, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Email nao cadastrado")
		return
	}

	err = auth.CheckPasswordHash(user.Password, params.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect password")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.SecretKey, time.Duration(3600)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating JWT token")
		return
	}

	refresh_token, _ := auth.MakeRefreshToken()
	args := database.CreateRefreshTokenParams{
		Token: refresh_token,
		UserID: user.ID,
		ExpiresAt: sql.NullTime{
			Time: time.Now().Add((24 * time.Hour) * 60 ),
			Valid: true,
		},
	}
	_, err = cfg.DB.CreateRefreshToken(r.Context(), args)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating refresh Token")
		return
	}

	respondWithJson(w, http.StatusOK, User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email, Token: token, Refresh_token: refresh_token})

}

func (cfg *apiConfig) RefreshToken(w http.ResponseWriter, r *http.Request){

	type responseBody struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing Bearer token")
		return
	}
	if token == "" {
		respondWithError(w, http.StatusUnauthorized, "Bearer token is empty")
		return
	}

	exist, err := cfg.DB.GetValidRefreshToken(r.Context(),token)
	if err != nil && !exist{
		respondWithError(w, http.StatusUnauthorized, "Token invalid")
		return
	}

	user, err := cfg.DB.GetUserForValidRefreshToken(r.Context(),token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid for this User")
		return
	}

	tokenAcces, err := auth.MakeJWT(user.ID, cfg.SecretKey, time.Duration(3600)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating JWT token")
		return
	}
	
	returnToken := responseBody{Token: tokenAcces} 

	respondWithJson(w, http.StatusOK, returnToken )
}

func (cfg *apiConfig) RevokeRefreshToken(w http.ResponseWriter, r *http.Request){
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing Bearer token")
		return
	}
	if token == "" {
		respondWithError(w, http.StatusUnauthorized, "Bearer token is empty")
		return
	}

	exist, err := cfg.DB.GetValidRefreshToken(r.Context(),token)
	if err != nil && !exist {
		respondWithError(w, http.StatusUnauthorized, "Token invalid")
		return
	}

	_, err = cfg.DB.RevokeRefreshToken(r.Context(),token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error to revoke token")
		return
	}
	var nullInterface interface{}
	respondWithJson(w, http.StatusNoContent, nullInterface)
}

func main() {

	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Printf("cant connect to the database")
	}
	dbQueries := database.New(db)
	environment := os.Getenv("PLATFORM")
	secretKey := os.Getenv("SECRET_KEY")
	apiCfg := apiConfig{
		DB:          dbQueries,
		Environment: environment,
		SecretKey:   secretKey,
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

	router.HandleFunc("POST /api/chirps", apiCfg.PostChirps)

	router.HandleFunc("GET /api/chirps", apiCfg.ListChirps)

	router.HandleFunc("GET /api/chirps/{id}", apiCfg.GetChirp)

	router.HandleFunc("POST /api/login", apiCfg.LoginUser)

	router.HandleFunc("POST /api/revoke", apiCfg.RevokeRefreshToken)

	router.HandleFunc("POST /api/refresh", apiCfg.RefreshToken)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Println("Starting server on :8080")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
