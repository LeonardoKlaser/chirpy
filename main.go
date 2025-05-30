package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/leonardoklaser/Chirpy/handlers"
	"github.com/leonardoklaser/Chirpy/internal/config"
	_ "github.com/lib/pq"
)



func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Printf("Erro ao inserir corpo da requisicao")
	}

}

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

func configureRoutes(router *http.ServeMux, cfg *config.ApiConfig ){
	// fileServer := http.FileServer(http.Dir("."))

	// strippedFileServerHandler := http.StripPrefix("/app/", fileServer)
	// metricsWrappedFileServerHandler := cfg.MiddlewareMetricsInc(strippedFileServerHandler)

	//router.HandleFunc("/app/", metricsWrappedFileServerHandler.ServeHTTP)

	router.HandleFunc("GET /api/healthz", handleHealthz)

	router.HandleFunc("GET /admin/metrics", cfg.HandleMetrics)

	router.HandleFunc("POST /admin/reset", handlers.DeleteUsers)

	router.HandleFunc("POST /api/validate_chirp", handlers.HandlerValidateChirp)

	router.HandleFunc("POST /api/users", handlers.PostUser)

	router.HandleFunc("POST /api/chirps", handlers.PostChirps)

	router.HandleFunc("DELETE /api/chirps/{chirpID}", handlers.DeleteChirpById)

	router.HandleFunc("GET /api/chirps", handlers.ListChirps)

	router.HandleFunc("GET /api/chirps/{id}", handlers.GetChirp)

	router.HandleFunc("POST /api/login", handlers.LoginUser)

	router.HandleFunc("POST /api/revoke", handlers.RevokeRefreshToken)

	router.HandleFunc("POST /api/refresh", handlers.RefreshToken)

	router.HandleFunc("PUT /api/users", handlers.UpdateUser)

	router.HandleFunc("POST /api/polka/webhooks", handlers.PolkaWebhook)
}

func main() {

	cfg, err := config.New()
	if err != nil {
		fmt.Printf("Error setting configuration: %s", err.Error())
		return
	}
	router := http.NewServeMux()

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
