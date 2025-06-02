package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/joho/godotenv"
	"github.com/leonardoklaser/Chirpy/handlers"
	"github.com/leonardoklaser/Chirpy/internal/config"
	_ "github.com/lib/pq"
)

func configureRoutes(router *http.ServeMux, cfg *config.ApiConfig ){

	router.HandleFunc("POST /admin/reset", cfg.HandleReset())

	router.HandleFunc("POST /api/validate_chirp", handlers.HandlerValidateChirp)

	router.HandleFunc("POST /api/users", handlers.PostUser)

	router.HandleFunc("POST /api/chirps", cfg.MiddlewareAuth(handlers.PostChirps))

	router.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.MiddlewareAuth(handlers.DeleteChirpById))

	router.HandleFunc("GET /api/chirps", handlers.ListChirps)

	router.HandleFunc("GET /api/chirps/{id}", handlers.GetChirp)

	router.HandleFunc("POST /api/login", handlers.LoginUser)

	router.HandleFunc("POST /api/revoke", handlers.RevokeRefreshToken)

	router.HandleFunc("POST /api/refresh", handlers.RefreshToken)

	router.HandleFunc("PUT /api/users", cfg.MiddlewareAuth(handlers.UpdateUser))

	router.HandleFunc("POST /api/polka/webhooks", cfg.MiddlewarePolka(handlers.PolkaWebhook))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading env: %s", err.Error())
		return
	}

	cfg, err := config.New()
	if err != nil {
		fmt.Printf("Error setting configuration: %s", err.Error())
		return
	}
	router := http.NewServeMux()

	configureRoutes(router, cfg)

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
