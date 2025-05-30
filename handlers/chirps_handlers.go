package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sort"
	"github.com/google/uuid"
	"github.com/leonardoklaser/Chirpy/internal/auth"
	"github.com/leonardoklaser/Chirpy/internal/config"
	"github.com/leonardoklaser/Chirpy/internal/database"
	"github.com/leonardoklaser/Chirpy/models"
	"github.com/leonardoklaser/Chirpy/utils"
)


func HandlerValidateChirp(w http.ResponseWriter, r *http.Request) {

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
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(params.Body) > 140 {
		utils.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	replacer := strings.NewReplacer("kerfuffle", "****", "sharbert", "****", "fornax", "****")
	cleanedBody := replacer.Replace(strings.ToLower(params.Body))
	responseCleanedBody, err := utils.FormatProfane(params.Body, cleanedBody)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error formatting profane words")
		return
	}

	response := ResponseType{CleanedBody: responseCleanedBody}
	utils.RespondWithJson(w, http.StatusOK, response)

}


func ListChirps(w http.ResponseWriter, r *http.Request) {
	
	var err error

	cfg, err := config.New()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest ,"Error to retrieve server configurations")
	}

	author := r.URL.Query().Get("author_id")
	var chirps []database.Chirp
	
	if author == ""{
		chirps, err = cfg.DB.GetAllChirps(r.Context())
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error to list all chirps: %v", err))
			return
		}
	}else{
		authoruuid, err := uuid.Parse(author)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error to retrieve Author Id : %v ", err))
			return
		}	
		chirps, err = cfg.DB.GetChirpsByUserId(r.Context(), authoruuid)
		if err != nil {
                        utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error to list all chirps: %v", err))
                        return
                }
	}
	var returnChirps []models.Chirp

	for _, val := range chirps {
		newChirp := models.Chirp{
			ID:        val.ID,
			CreatedAt: val.CreatedAt,
			UpdatedAt: val.UpdatedAt,
			Body:      val.Body,
			UserId:    val.UserID,
		}
		returnChirps = append(returnChirps, newChirp)
	}
	

	sortSlice := r.URL.Query().Get("sort")
	if sortSlice == "desc"{
		sort.Slice(returnChirps, func(i, j int) bool { return returnChirps[i].CreatedAt.After(returnChirps[j].CreatedAt)})
	}else {
		sort.Slice(returnChirps, func(i, j int) bool { return returnChirps[i].CreatedAt.Before(returnChirps[j].CreatedAt)})
	}
	utils.RespondWithListJson(w, http.StatusOK, returnChirps)
}

func GetChirp(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.New()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest ,"Error to retrieve server configurations")
	}

	id := r.PathValue("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error to retrieve chirp Id : %v ", err))
		return
	}

	chirp, err := cfg.DB.GetChirpById(r.Context(), uid)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Chirp with ID %s not found", id))
		return
	}

	chirpToReturn := models.Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
	utils.RespondWithJson(w, http.StatusOK, chirpToReturn)
}

func PostChirps(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.New()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest ,"Error to retrieve server configurations")
	}

	type requestBody struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or missing Bearer token")
		return
	}
	if token == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "Bearer token is empty")
		return
	}
	
	uuidUser, err := auth.ValidateJWT(token, cfg.SecretKey)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid Bearer token: %v", err))
			return
	}
	
	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err = decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Input")
		return
	}

	if len(params.Body) > 140 {
		utils.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	replacer := strings.NewReplacer("kerfuffle", "****", "sharbert", "****", "fornax", "****")
	cleanedBody := replacer.Replace(strings.ToLower(params.Body))
	responseCleanedBody, err := utils.FormatProfane(params.Body, cleanedBody)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error to format profane words")
		return
	}

	createChirpParam := database.CreateChirpParams{Body: responseCleanedBody, UserID: uuidUser}

	chirp, err := cfg.DB.CreateChirp(r.Context(), createChirpParam)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error to Create new Chirp: %v, json: %v", err, createChirpParam) )
		return
	}

	response := models.Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserId: chirp.UserID}
	utils.RespondWithJson(w, http.StatusCreated, response)

}


func DeleteChirpById(w http.ResponseWriter, r *http.Request){
	
	cfg, err := config.New()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest ,"Error to retrieve server configurations")
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or missing Bearer token")
		return
	}
	if token == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "Bearer token is empty")
		return
	}

	uuidUser, err := auth.ValidateJWT(token, cfg.SecretKey)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid Bearer token: %v", err))
			return
	}


	id := r.PathValue("chirpID")
	uuid, err := uuid.Parse(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error to retrieve chirp Id : %v ", err))
		return
	}

	chirp, err := cfg.DB.GetChirpById(r.Context(), uuid)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Chirp with ID %s not found", id))
		return
	}

	if uuidUser != chirp.UserID {
		utils.RespondWithError(w, http.StatusForbidden, fmt.Sprintf("You can just delete chirps posteds by your userID: %v", uuid))
                return
	}

 	_, err = cfg.DB.DeleteChirpById(r.Context(), uuid)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error to delete chirp : %v ", err))
                return
	}
	
	var nullInterface interface{}

	utils.RespondWithJson(w, http.StatusNoContent, nullInterface)
}

