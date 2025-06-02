package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/leonardoklaser/Chirpy/internal/auth"
	"github.com/leonardoklaser/Chirpy/internal/config"
	"github.com/leonardoklaser/Chirpy/utils"
)

func RefreshToken(w http.ResponseWriter, r *http.Request){

	cfg, err := config.New()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest ,"Error to retrieve server configurations")
	}

	type responseBody struct {
		Token string `json:"token"`
	}

	token, ok := r.Context().Value(config.TokenKey).(string)
	if !ok {
		utils.RespondWithError(w, 500, fmt.Sprintf("omg you're so bad at this"))
	}

	exist, err := cfg.DB.GetValidRefreshToken(r.Context(),token)
	if err != nil && !exist{
		utils.RespondWithError(w, http.StatusUnauthorized, "Token invalid")
		return
	}

	user, err := cfg.DB.GetUserForValidRefreshToken(r.Context(),token)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Token invalid for this User")
		return
	}

	tokenAcces, err := auth.MakeJWT(user.ID, cfg.SecretKey, time.Duration(3600)*time.Second)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error generating JWT token")
		return
	}
	
	returnToken := responseBody{Token: tokenAcces} 

	utils.RespondWithJson(w, http.StatusOK, returnToken )
}

func RevokeRefreshToken(w http.ResponseWriter, r *http.Request){

	cfg, err := config.New()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest ,"Error to retrieve server configurations")
	}
	
	token, ok := r.Context().Value(config.TokenKey).(string)
	if !ok {
		utils.RespondWithError(w, 500, fmt.Sprintf("omg you're so bad at this"))
	}

	exist, err := cfg.DB.GetValidRefreshToken(r.Context(),token)
	if err != nil && !exist {
		utils.RespondWithError(w, http.StatusUnauthorized, "Token invalid")
		return
	}

	_, err = cfg.DB.RevokeRefreshToken(r.Context(),token)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Error to revoke token")
		return
	}
	var nullInterface interface{}
	utils.RespondWithJson(w, http.StatusNoContent, nullInterface)
}