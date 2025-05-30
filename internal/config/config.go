package config

import (
	"database/sql"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/leonardoklaser/Chirpy/internal/database"
)

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