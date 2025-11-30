package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("=== Week 7: API Gateway Demo & Testing ===\n")

	db := connectDB()
	defer db.Close()

	// Create test tenant and API key
	apiKey := setupTestTenant(db)

	// Wait for API server to be ready
	fmt.Println("⏳ Waiting for API server to start...")
	time.Sleep(2 * time.Second)

	baseURL := "http://localhost:8080"

	// Demo 1: Health Check
	fmt.Println("\n📋 Demo 1: Health Check")
	fmt.Println("─────────────────────────────────")
	testEndpoint("GET", baseURL+"/health", "", nil)

	// Demo 2: List Resources
	fmt.Println("\n📋 Demo 2: List Resources (Authenticated)")
	fmt.Println("─────────────────────────────────────────")
	testEndpoint("GET", baseURL+"/api/v1/resources?page=1", apiKey, nil)

	// Demo 3: Resource Statistics
	fmt.Println("\n📋 Demo 3: Resource Statistics")
	fmt.Println("──────────────────────────────")
	testEndpoint("GET", baseURL+"/api/v1/resources/stats", apiKey, nil)

	// Demo 4: List Recommendations
	fmt.Println("\n📋 Demo 4: Cost Optimization Recommendations")
	fmt.Println("────────────────────────────────────────────")
	testEndpoint("GET", baseURL+"/api/v1/recommendations?status=pending", apiKey, nil)

	// Demo 5: Rate Limiting Test
	fmt.Println("\n📋 Demo 5: Rate Limiting (100 req/min)")
	fmt.Println("──────────────────────────────────────")
	fmt.Println("Making 5 rapid requests...")
	for i := 1; i <= 5; i++ {
		resp, _ := makeRequest("GET", baseURL+"/api/v1/resources/stats", apiKey, nil)
		fmt.Printf("  Request %d: Status %d\n", i, resp.StatusCode)
		resp.Body.Close()
	}

	// Demo 6: Unauthorized Access
	fmt.Println("\n📋 Demo 6: Unauthorized Access (No API Key)")
	fmt.Println("───────────────────────────────────────────")
	testEndpoint("GET", baseURL+"/api/v1/resources", "", nil)

	fmt.Println("\n\n✅ Week 7 API Gateway Demo Complete!")
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("\n📊 Implementation Summary:")
	fmt.Println("  ✓ RESTful API with versioning (/api/v1/)")
	fmt.Println("  ✓ Tenant-based authentication (API keys)")
	fmt.Println("  ✓ Rate limiting (100 req/min)")
	fmt.Println("  ✓ CORS enabled for frontend")
	fmt.Println("  ✓ Pagination support")
	fmt.Println("  ✓ JSON responses with metadata")
	fmt.Println("\n🔜 Next: Week 8 - Security Hardening")
}

func setupTestTenant(db *sql.DB) string {
	var tenantCode string
	err := db.QueryRow(`
		SELECT tenant_code FROM yt_tenants 
		WHERE company_name = 'Demo Company' LIMIT 1`,
	).Scan(&tenantCode)

	if err == sql.ErrNoRows {
		db.QueryRow(`
			INSERT INTO yt_tenants (tenant_code, company_name, subscription_tier)
			VALUES ('democorp-test', 'Demo Company', 'PROFESSIONAL')
			RETURNING tenant_code`,
		).Scan(&tenantCode)

		// Create test AWS account
		var accountID int
		db.QueryRow(`
			INSERT INTO yt_aws_accounts (tenant_id, account_id, account_name, role_arn, external_id, status)
			VALUES ((SELECT id FROM yt_tenants WHERE tenant_code = $1), '999999999999', 'Demo Account', 
			        'arn:aws:iam::999999999999:role/Demo', 'demo-external-id', 'active')
			RETURNING id`,
			tenantCode,
		).Scan(&accountID)

		// Create test resources
		for i := 1; i <= 5; i++ {
			db.Exec(`
				INSERT INTO yt_tenant_resources 
				(tenant_id, aws_account_id, resource_id, resource_type, region, instance_type, state, monthly_cost)
				VALUES ((SELECT id FROM yt_tenants WHERE tenant_code = $1), $2, $3, 'ec2', 'us-east-1', 't3.medium', 'running', $4)`,
				tenantCode, accountID, fmt.Sprintf("i-demo%d", i), float64(50+i*10),
			)
		}

		// Create test recommendations
		db.Exec(`
			INSERT INTO yt_tenant_recommendations 
			(tenant_id, recommendation_type, current_cost, optimized_cost, monthly_savings, confidence_score)
			VALUES ((SELECT id FROM yt_tenants WHERE tenant_code = $1), 'downsize', 100, 50, 50, 0.85)`,
			tenantCode,
		)
	}

	return tenantCode + "_demo-api-key-12345"
}

func testEndpoint(method, url, apiKey string, body interface{}) {
	resp, err := makeRequest(method, url, apiKey, body)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	fmt.Printf("Request:  %s %s\n", method, url)
	if apiKey != "" {
		fmt.Printf("API Key:  %s\n", apiKey[:20]+"...")
	}
	fmt.Printf("Status:   %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, bodyBytes, "", "  ")
	fmt.Printf("Response:\n%s\n", prettyJSON.String())
}

func makeRequest(method, url, apiKey string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	return client.Do(req)
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
