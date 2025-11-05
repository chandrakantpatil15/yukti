package main

import (
	"log"
	"os"

	"github.com/cloudcostoptimizer/yukti/internal/api"
	"github.com/cloudcostoptimizer/yukti/internal/config"
	"github.com/cloudcostoptimizer/yukti/internal/database"
)

func main() {
	cfg := config.Load()
	
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	server := api.NewServer(db)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting Yukti FinOps server on port %s", port)
	if err := server.Run(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}