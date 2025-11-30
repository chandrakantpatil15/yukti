package main

import (
	"log"
	"net/http"
	"yukti/internal/api/handlers"
	"yukti/internal/database"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Connect to database
	db, err := database.Connect("postgresql://chandrakantpatil@localhost:5432/yukti_finops?sslmode=disable")
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Create auth handler
	authHandler := handlers.NewAuthHandler(db, "yukti-secret-key-change-in-production")

	// Create simple router
	router := mux.NewRouter()
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// Add CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	log.Printf("Test server starting on :8082")
	log.Fatal(http.ListenAndServe(":8082", c.Handler(router)))
}