package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/leonardoklaser/Chirpy/internal/auth"
	"github.com/leonardoklaser/Chirpy/internal/database"
	"github.com/leonardoklaser/Chirpy/utils"
)

type contextKey string

const UserIDKey contextKey = "userID"
const TokenKey contextKey = "Token"

type ApiConfig struct {
	Environment       string
	FileServerHits *atomic.Int32
	DB             *database.Queries
	SecretKey      string
	PolkaKey       string
}

var instance *ApiConfig

func createDatabaseInstance() (*database.Queries, error) {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return &database.Queries{}, fmt.Errorf("error opening database connection: %s", err.Error())
	}

	return database.New(db), nil
}

func New() (*ApiConfig, error) {
	if instance == nil {
		db, err := createDatabaseInstance()
		if err != nil {
			return &ApiConfig{}, nil
		}

		instance = &ApiConfig{
			Environment:       os.Getenv("PLATFORM"),
			FileServerHits: &atomic.Int32{},
			DB:             db,
			SecretKey:      os.Getenv("APP_SECRET"),
			PolkaKey:       os.Getenv("POLKA_KEY"),
		}
		instance.FileServerHits.Store(0)
	}

	return instance, nil
}

func (cfg *ApiConfig) MiddlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(&req.Header)
		if err != nil {
			utils.RespondWithError(resp, http.StatusUnauthorized, err.Error())
			return
		}

		userId, err := auth.ValidateJWT(token, cfg.SecretKey)
		if err != nil {
			utils.RespondWithError(resp, http.StatusUnauthorized, "Unauthorized")
			return
		}
		ctx := context.WithValue(req.Context(), UserIDKey, userId)
		ctx = context.WithValue(ctx, TokenKey, token)

		next.ServeHTTP(resp, req.WithContext(ctx))
	})
}

func (cfg *ApiConfig) MiddlewarePolka(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		APIKey, err := auth.GetAPIKey(&req.Header)
		if err != nil {
			utils.RespondWithError(resp, http.StatusUnauthorized, err.Error())
			return
		}

		if cfg.PolkaKey != APIKey {
			utils.RespondWithError(resp, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next.ServeHTTP(resp, req)
	})
}

func (cfg *ApiConfig) HandleReset() http.HandlerFunc {
	return http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		if cfg.Environment != "dev" {
			utils.RespondWithError(resp, http.StatusForbidden, "Forbidden")
			return
		}
		_, err := cfg.DB.DeleteUsers(r.Context())
		if err != nil {
			utils.RespondWithError(resp, http.StatusInternalServerError ,err.Error())
			return
		}
		cfg.FileServerHits.Store(0)
		log.Println("Reset endpoint called. Everything was wiped")
		resp.WriteHeader(http.StatusOK)
	})
}