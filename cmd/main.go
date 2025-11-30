package main

import (
	"log"
	"os"

	"yukti/internal/api"
	"yukti/internal/config"
	"yukti/internal/database"
)

func main() {
	log.Printf("[INFO] ========================================")
	log.Printf("[INFO] Yukti FinOps Platform Starting...")
	log.Printf("[INFO] ========================================")
	
	log.Printf("[INFO] Loading configuration...")
	cfg := config.Load()
	log.Printf("[INFO] Configuration loaded successfully")
	
	log.Printf("[INFO] Loading secrets...")
	config.LoadSecrets()
	log.Printf("[INFO] Secrets loaded successfully")
	
	log.Printf("[INFO] Connecting to database...")
	_, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[FATAL] Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Printf("[INFO] Database connection established")

	log.Printf("[INFO] Initializing API server...")
	server := api.NewServer(database.DB)
	
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("[FATAL] PORT environment variable not set. Check .env.ports file.")
	}
	
	log.Printf("[INFO] ========================================")
	log.Printf("[INFO] Server starting on port %s", port)
	log.Printf("[INFO] Health check: http://localhost:%s/health", port)
	log.Printf("[INFO] Admin API: http://localhost:%s/api/admin/*", port)
	log.Printf("[INFO] Customer API: http://localhost:%s/api/customers/*", port)
	log.Printf("[INFO] ========================================")
	
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}
}