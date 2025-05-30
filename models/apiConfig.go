package models

import (
	"sync/atomic"

	"github.com/leonardoklaser/Chirpy/internal/database"
)

type ApiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	Environment    string
	SecretKey      string
	PolkaKey       string
}