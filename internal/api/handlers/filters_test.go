package handlers

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupFilterTestDB(t *testing.T) (*sql.DB, int) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/yukti_test?sslmode=disable"
	}

	gormDB, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return nil, 0
	}

	db := gormDB.DB()

	// Create test tenant
	var tenantID int
	err = db.QueryRow(`
		INSERT INTO yt_tenants (tenant_code, company_name, subscription_tier, status)
		VALUES ('test-tenant', 'Test Company', 'PROFESSIONAL', 'active')
		RETURNING id
	`).Scan(&tenantID)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	// Insert test resources
	_, err = db.Exec(`
		INSERT INTO yt_tenant_resources (tenant_id, aws_account_id, resource_id, resource_type, region, tags)
		VALUES 
			($1, 1, 'i-123', 'ec2', 'us-east-1', '{"Environment": "prod", "Team": "backend"}'::jsonb),
			($1, 1, 'i-456', 'ec2', 'us-west-2', '{"Environment": "dev", "Team": "frontend"}'::jsonb),
			($1, 1, 'db-prod', 'rds', 'us-east-1', '{"Environment": "prod"}'::jsonb)
	`, tenantID)
	if err != nil {
		t.Logf("Note: Could not insert test resources (table may not exist): %v", err)
	}

	// Insert test cost data
	_, err = db.Exec(`
		INSERT INTO yt_cost_data (id, tenant_id, date, service, cost)
		VALUES 
			('cost-1', $1, CURRENT_DATE, 'EC2', 1000.00),
			('cost-2', $1, CURRENT_DATE, 'RDS', 500.00),
			('cost-3', $1, CURRENT_DATE, 'S3', 200.00)
	`, tenantID)
	if err != nil {
		t.Logf("Note: Could not insert test cost data (table may not exist): %v", err)
	}

	return db, tenantID
}

func TestFilterHandler_GetResourceTypes(t *testing.T) {
	db, tenantID := setupFilterTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	handler := NewFilterHandler(db)

	t.Run("returns distinct resource types", func(t *testing.T) {
		// Stub: Test GetResourceTypes endpoint
		// 1. Create httptest request with tenant_id in context
		// 2. Call handler.GetResourceTypes
		// 3. Verify response contains resource types
		// 4. Verify counts are correct
		t.Skip("TODO: Implement with httptest and context")
	})
}

func TestFilterHandler_GetTags(t *testing.T) {
	db, tenantID := setupFilterTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	handler := NewFilterHandler(db)

	t.Run("returns tag keys and values", func(t *testing.T) {
		// Stub: Test GetTags endpoint
		// 1. Create request with tenant_id
		// 2. Call handler.GetTags
		// 3. Verify tag_keys array
		// 4. Verify tag_values object
		t.Skip("TODO: Implement with httptest")
	})
}

func TestFilterHandler_GetServices(t *testing.T) {
	db, tenantID := setupFilterTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	handler := NewFilterHandler(db)

	t.Run("returns distinct services with costs", func(t *testing.T) {
		// Stub: Test GetServices endpoint
		t.Skip("TODO: Implement with httptest")
	})
}

