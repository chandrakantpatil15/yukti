package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"yukti/internal/api/routes"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("=== Week 7: API Gateway & REST Endpoints ===\n")

	db := connectDB()
	defer db.Close()

	// Setup API routes
	handler := routes.SetupRoutes(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 API Gateway starting on port %s\n", port)
	fmt.Println("\n📋 Available Endpoints:")
	fmt.Println("─────────────────────────────────────────")
	fmt.Printf("  GET  /health                      - Health check\n")
	fmt.Printf("  GET  /api/v1/resources            - List resources (paginated)\n")
	fmt.Printf("  GET  /api/v1/resources/stats      - Resource statistics\n")
	fmt.Printf("  GET  /api/v1/recommendations      - List recommendations\n")
	fmt.Println("\n🔐 Authentication:")
	fmt.Println("  Header: X-API-Key: <tenant-code>_<api-key>")
	fmt.Println("\n⚡ Rate Limiting:")
	fmt.Println("  100 requests per minute per API key")
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Printf("\n✅ Server ready at http://localhost:%s\n\n", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func connectDB() *sql.DB {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://chandrakantpatil@localhost:5432/yukti?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}
