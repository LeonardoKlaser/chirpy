package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/leonardoklaser/Chirpy/internal/auth"
	"github.com/leonardoklaser/Chirpy/internal/config"
	"github.com/leonardoklaser/Chirpy/internal/database"
	"github.com/leonardoklaser/Chirpy/models"
	"github.com/leonardoklaser/Chirpy/utils"
)

func UpdateUser(w http.ResponseWriter, r *http.Request) {

	cfg, err := config.New()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest ,"Error to retrieve server configurations")
	}

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err = decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	uuidUser, ok := r.Context().Value(config.UserIDKey).(uuid.UUID)
	if !ok {
		utils.RespondWithError(w, 500, fmt.Sprintf("omg you're so bad at this"))
	}

	passwordHashed, err := auth.HashPassword(params.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Error to has password")
		return
	}

	args := database.UpdateUserByIdParams{
		Email:    params.Email,
		Password: passwordHashed,
		ID:       uuidUser,
	}

	user, err := cfg.DB.UpdateUserById(r.Context(), args)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error to update user: %v", err))
		return
	}

	utils.RespondWithJson(w, http.StatusOK, models.User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email, ChirpyRed: user.IsChirpyRed.Bool})

}

func PostUser(w http.ResponseWriter, r *http.Request) {

	cfg, err := config.New()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest ,"Error to retrieve server configurations")
	}

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err = decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	passwordHashed, err := auth.HashPassword(params.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Error to has password")
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, Password: passwordHashed})
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Error to insert new User in Database")
		return
	}

	userToReturn := models.User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email, ChirpyRed: user.IsChirpyRed.Bool}
	utils.RespondWithJson(w, http.StatusCreated, userToReturn)

}

func DeleteUsers(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.New()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest ,"Error to retrieve server configurations")
	}

	if cfg.Environment != "dev" {
		utils.RespondWithError(w, http.StatusForbidden, "This action is available only on development environment")
		return
	}
	_, err = cfg.DB.DeleteUsers(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error to delete user: %v", err))
		return
	}

	var nullInterface interface{}
	utils.RespondWithJson(w, http.StatusOK, nullInterface)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.New()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest ,"Error to retrieve server configurations")
	}

	type requestBody struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err = decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	user, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Email nao cadastrado")
		return
	}

	err = auth.CheckPasswordHash(user.Password, params.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Incorrect password")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.SecretKey, time.Duration(3600)*time.Second)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error generating JWT token")
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Error generating refresh Token")
		return
	}

	utils.RespondWithJson(w, http.StatusOK, models.User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email, ChirpyRed: user.IsChirpyRed.Bool ,Token: token, Refresh_token: refresh_token})

}