package handlers

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"yukti/internal/models"
)

// setupTestDB creates an in-memory test database connection
// In a real test, you'd use a test database or sqlmock
func setupTestDB(t *testing.T) (*sql.DB, *gorm.DB) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/yukti_test?sslmode=disable"
	}

	gormDB, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return nil, nil
	}

	// Auto-migrate test schema
	err = gormDB.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return gormDB.DB(), gormDB
}

func TestAuthHandler_Signup(t *testing.T) {
	db, gormDB := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	handler := NewAuthHandler(db)

	// Test cases would go here
	// Example structure:
	t.Run("valid signup creates user and tenant", func(t *testing.T) {
		// This is a stub - implement with httptest
		// 1. Create request with valid email/password
		// 2. Call handler.Signup
		// 3. Verify user created in database
		// 4. Verify tenant created
		// 5. Verify first user gets admin role
		t.Skip("TODO: Implement with httptest")
	})

	t.Run("duplicate email returns error", func(t *testing.T) {
		// Stub for duplicate email test
		t.Skip("TODO: Implement with httptest")
	})

	t.Run("invalid email format returns error", func(t *testing.T) {
		// Stub for validation test
		t.Skip("TODO: Implement with httptest")
	})

	t.Run("short password returns error", func(t *testing.T) {
		// Stub for password validation
		t.Skip("TODO: Implement with httptest")
	})
}

func TestAuthHandler_Login(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	handler := NewAuthHandler(db)

	t.Run("valid credentials returns JWT token", func(t *testing.T) {
		// Stub: Test login flow
		// 1. Create user via signup or direct DB insert
		// 2. Call handler.Login with correct credentials
		// 3. Verify JWT token returned
		// 4. Verify token contains correct claims
		t.Skip("TODO: Implement with httptest")
	})

	t.Run("invalid credentials returns error", func(t *testing.T) {
		// Stub: Test invalid login
		t.Skip("TODO: Implement with httptest")
	})

	t.Run("inactive user cannot login", func(t *testing.T) {
		// Stub: Test inactive user
		t.Skip("TODO: Implement with httptest")
	})
}

func TestUserModel(t *testing.T) {
	_, gormDB := setupTestDB(t)
	if gormDB == nil {
		return
	}

	t.Run("CreateUser hashes password", func(t *testing.T) {
		user, err := models.CreateUser(gormDB, 1, "test@example.com", "password123", "admin")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if user.PasswordHash == "password123" {
			t.Error("Password was not hashed")
		}

		if !user.CheckPassword("password123") {
			t.Error("Password check failed for correct password")
		}

		if user.CheckPassword("wrongpassword") {
			t.Error("Password check passed for wrong password")
		}
	})

	t.Run("GetUserByEmailTenant finds user", func(t *testing.T) {
		// Create user first
		_, err := models.CreateUser(gormDB, 1, "findme@example.com", "password123", "viewer")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		user, err := models.GetUserByEmailTenant(gormDB, 1, "findme@example.com")
		if err != nil {
			t.Fatalf("Failed to find user: %v", err)
		}

		if user.Email != "findme@example.com" {
			t.Errorf("Expected email findme@example.com, got %s", user.Email)
		}
	})
}

