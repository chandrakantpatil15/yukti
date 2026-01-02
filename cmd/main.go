package main

import (
	"fmt"
	"log"
	"os"

	"yukti/internal/api"
	"yukti/internal/cache"
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
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[FATAL] Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Printf("[INFO] Database connection established")

	log.Printf("[INFO] Connecting to Redis...")
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		redisPassword = "yukti123"
	}
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	_ = cache.NewRedisCache(redisAddr, redisPassword, 0)
	log.Printf("[INFO] Redis connection established at %s", redisAddr)

	log.Printf("[INFO] Initializing cache services...")
	_ = cache.NewSessionCache(nil)
	_ = cache.NewOTPCache(nil)
	_ = cache.NewDashboardCache(nil)
	_ = cache.NewRateLimiter(nil)
	log.Printf("[INFO] Cache services initialized")

	log.Printf("[INFO] Initializing API server...")
	server := api.NewServer(db)

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		log.Fatal("[FATAL] BACKEND_PORT environment variable not set. Check .env.ports file.")
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
