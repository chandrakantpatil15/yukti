cd /Users/chandrakantpatil/workspace/yukti

# 1. Update main.go to use BACKEND_PORT
cat > main_fix.go << 'EOF'
package main

import (
    "fmt"
    "os"
    "io/ioutil"
)

func main() {
    content, err := ioutil.ReadFile("main.go")
    if err != nil {
        panic(err)
    }
    
    newContent := string(content)
    // Replace PORT check with BACKEND_PORT
    newContent = `package main

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
    
    port := os.Getenv("BACKEND_PORT")  // CHANGED HERE
    if port == "" {
        log.Fatal("[FATAL] BACKEND_PORT environment variable not set. Check .env.ports file.")  // CHANGED HERE
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
}`
    
    err = ioutil.WriteFile("main.go", []byte(newContent), 0644)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("✅ main.go updated to use BACKEND_PORT")
}
EOF

go run main_fix.go
rm main_fix.go

# 2. Add CORS to .env.ports
echo "CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173" >> .env.ports

# 3. Fix frontend .env.local
cat > frontend/.env.local << 'EOF'
REACT_APP_API_URL=http://localhost:8081
REACT_APP_API_BASE=http://localhost:8081
EOF

# 4. Fix hardcoded 8080 URLs
find frontend/src -type f \( -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.test.tsx" \) \
  -exec sed -i '' 's|http://localhost:8080|http://localhost:8081|g' {} \;

# 5. Load environment and start backend
echo "=== LOADING ENVIRONMENT ==="
export $(grep -v '^#' .env.ports | xargs)
echo "BACKEND_PORT: $BACKEND_PORT"
echo "FRONTEND_PORT: $FRONTEND_PORT"

# 6. Start backend
pkill -f "main.go"
go run main.go &

echo ""
echo "✅ Fixed! Using BACKEND_PORT=8081"
echo "✅ Backend: http://localhost:8081"
echo "✅ Frontend: http://localhost:3000"
