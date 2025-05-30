package handlers

import (
	"net/http"
	"github.com/leonardoklaser/Chirpy/internal/config"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/leonardoklaser/Chirpy/internal/auth"
	"github.com/leonardoklaser/Chirpy/utils"
)

func PolkaWebhook(w http.ResponseWriter, r *http.Request){

	cfg, err := config.New()
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest ,"Error to retrieve server configurations")
	}

	type DataUserId struct {
                UserId uuid.UUID `json:"user_id"`
        }

	type requestBody struct{
		Event string `json:"event"`
		Data DataUserId `json:"data"`
	}

	apiPolka, err := auth.GetAPIKey(r.Header)
	if err != nil{
		utils.RespondWithError(w, http.StatusUnauthorized, "Api token not found at header request")
                return
	}

	if cfg.PolkaKey != apiPolka{
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
                return
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err = decoder.Decode(&params)
	if err != nil{
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	var nullInterface interface{}
	if params.Event == "user.upgraded" {
		_,err := cfg.DB.UpgradeToRed(r.Context(), params.Data.UserId)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
                	return
		}
	}


	utils.RespondWithJson(w, http.StatusNoContent, nullInterface)

}
