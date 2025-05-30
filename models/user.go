package models
import (
	"github.com/google/uuid"
	"time"
)


type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token string `json:"token"`
		Refresh_token string `json:"refresh_token"`
		ChirpyRed bool `json:"is_chirpy_red"`
}