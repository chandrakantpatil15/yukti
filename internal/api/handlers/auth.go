package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"yukti/internal/models"
	"yukti/internal/security"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db        *sql.DB
	gormDB    *gorm.DB
	jwtService *security.JWTService
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	// Initialize GORM for user operations
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		log.Printf("[ERROR] Failed to initialize GORM: %v", err)
		// Continue with sql.DB only
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "yukti-secret-key-change-in-production" // Default for development
		log.Printf("[WARN] JWT_SECRET not set, using default (change in production!)")
	}

	return &AuthHandler{
		db:         db,
		gormDB:     gormDB,
		jwtService: security.NewJWTService(jwtSecret),
	}
}

// SignupRequest represents the signup request payload
type SignupRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	CompanyName string `json:"company_name,omitempty"`
}

// SignupResponse represents the signup response
type SignupResponse struct {
	Success bool   `json:"success"`
	UserID  string `json:"user_id,omitempty"`
	TenantID int   `json:"tenant_id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Signup handles user registration
// POST /api/v1/auth/signup
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Signup request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to decode signup request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SignupResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SignupResponse{
			Success: false,
			Error:   "Email and password are required",
		})
		return
	}

	if len(req.Password) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SignupResponse{
			Success: false,
			Error:   "Password must be at least 8 characters",
		})
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user already exists (across all tenants for now, or we can scope by tenant later)
	var existingUserID string
	err := h.db.QueryRow("SELECT id::text FROM yt_users WHERE email = $1 AND is_active = true", req.Email).Scan(&existingUserID)
	if err == nil {
		log.Printf("[WARN] User with email %s already exists", req.Email)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(SignupResponse{
			Success: false,
			Error:   "User with this email already exists",
		})
		return
	}

	// Determine tenant: create new or use existing
	// For now, we'll create a new tenant for each signup (can be enhanced later)
	var tenantID int
	var tenantCode string

	if req.CompanyName == "" {
		req.CompanyName = "Company " + req.Email
	}

	// Generate tenant code from company name
	tenantCode = generateTenantCode(req.CompanyName)

	// Check if tenant with this code exists
	var existingTenantID int
	err = h.db.QueryRow("SELECT id FROM yt_tenants WHERE tenant_code = $1", tenantCode).Scan(&existingTenantID)
	if err == nil {
		tenantID = existingTenantID
		log.Printf("[INFO] Using existing tenant %d for signup", tenantID)
	} else {
		// Create new tenant
		err = h.db.QueryRow(`
			INSERT INTO yt_tenants (tenant_code, company_name, subscription_tier, status, trial_ends_at)
			VALUES ($1, $2, 'FREE', 'active', NOW() + INTERVAL '30 days')
			RETURNING id
		`, tenantCode, req.CompanyName).Scan(&tenantID)
		if err != nil {
			log.Printf("[ERROR] Failed to create tenant: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(SignupResponse{
				Success: false,
				Error:   "Failed to create tenant",
			})
			return
		}
		log.Printf("[INFO] Created new tenant %d for signup", tenantID)
	}

	// Check if this is the first user for the tenant
	isFirstUser, err := models.GetFirstUserForTenant(h.gormDB, tenantID)
	if err != nil {
		log.Printf("[ERROR] Failed to check first user: %v", err)
		isFirstUser = false
	}

	// Determine role: admin for first user, otherwise default to viewer (can be changed by admin)
	role := "viewer"
	if isFirstUser {
		role = "admin"
		log.Printf("[INFO] First user for tenant %d, assigning admin role", tenantID)
	}

	// Create user using GORM
	user, err := models.CreateUser(h.gormDB, tenantID, req.Email, req.Password, role)
	if err != nil {
		log.Printf("[ERROR] Failed to create user: %v", err)
		// Check if it's a unique constraint violation
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(SignupResponse{
				Success: false,
				Error:   "User with this email already exists in this tenant",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SignupResponse{
			Success: false,
			Error:   "Failed to create user",
		})
		return
	}

	log.Printf("[INFO] Successfully created user %s for tenant %d", user.ID, tenantID)
	json.NewEncoder(w).Encode(SignupResponse{
		Success:  true,
		UserID:   user.ID.String(),
		TenantID: tenantID,
		Message:  "User created successfully",
	})
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	TenantCode string `json:"tenant_code" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Success   bool      `json:"success"`
	Token     string    `json:"token,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	User      *UserInfo `json:"user,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// UserInfo represents user information in login response
type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	TenantID int    `json:"tenant_id"`
}

// Login handles user authentication
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Login request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to decode login request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Validate input
	if req.TenantCode == "" || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Tenant code, email, and password are required",
		})
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Get tenant ID from tenant code
	var tenantID int
	var tenantStatus string
	err := h.db.QueryRow(`
		SELECT id, status FROM yt_tenants WHERE tenant_code = $1
	`, req.TenantCode).Scan(&tenantID, &tenantStatus)
	if err != nil {
		log.Printf("[WARN] Invalid tenant code: %s", req.TenantCode)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Invalid tenant code or email",
		})
		return
	}

	if tenantStatus != "active" {
		log.Printf("[WARN] Tenant %d is not active (status: %s)", tenantID, tenantStatus)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Tenant account is suspended",
		})
		return
	}

	// Get user by email and tenant
	userRepo := models.NewUserRepository(h.db)
	user, err := userRepo.GetUserByEmailTenantSQL(tenantID, req.Email)
	if err != nil {
		log.Printf("[WARN] User not found or inactive: %s in tenant %d", req.Email, tenantID)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Invalid tenant code or email",
		})
		return
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		log.Printf("[WARN] Invalid password for user %s", req.Email)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Invalid tenant code or email",
		})
		return
	}

	// Generate JWT token
	expiresIn := 24 * time.Hour // 24 hours
	token, err := h.jwtService.GenerateToken(
		user.ID.String(),
		user.TenantID,
		req.TenantCode,
		user.Email,
		user.Role,
		[]string{"read", "write"}, // Default scopes
		expiresIn,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to generate token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	log.Printf("[INFO] User %s logged in successfully for tenant %d", req.Email, tenantID)
	json.NewEncoder(w).Encode(LoginResponse{
		Success:   true,
		Token:     token,
		ExpiresAt: time.Now().Add(expiresIn),
		User: &UserInfo{
			ID:       user.ID.String(),
			Email:    user.Email,
			Role:     user.Role,
			TenantID: user.TenantID,
		},
	})
}

// Logout handles user logout (stateless - client deletes token)
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Logout request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// Stateless logout - just return success
	// Client should delete token from storage
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}

// CreateAPIKeyRequest represents API key creation request
type CreateAPIKeyRequest struct {
	KeyName  string   `json:"key_name" validate:"required"`
	Scopes   []string `json:"scopes,omitempty"`
	ExpiresInDays int  `json:"expires_in_days,omitempty"` // Optional expiration
}

// CreateAPIKeyResponse represents API key creation response
type CreateAPIKeyResponse struct {
	Success bool   `json:"success"`
	APIKey  string `json:"api_key,omitempty"`
	KeyID   string `json:"key_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CreateAPIKey handles API key creation (admin only)
// POST /api/v1/auth/api-keys
func (h *AuthHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] CreateAPIKey request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// Get tenant_id and role from context (set by JWT middleware)
	tenantID, ok := r.Context().Value("tenant_id").(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(CreateAPIKeyResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	role, ok := r.Context().Value("role").(string)
	if !ok || role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(CreateAPIKeyResponse{
			Success: false,
			Error:   "Admin access required",
		})
		return
	}

	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateAPIKeyResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	if req.KeyName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateAPIKeyResponse{
			Success: false,
			Error:   "Key name is required",
		})
		return
	}

	// Use existing API key service
	apiKeyService := security.NewAPIKeyService(h.db)
	
	var expiresIn time.Duration
	if req.ExpiresInDays > 0 {
		expiresIn = time.Duration(req.ExpiresInDays) * 24 * time.Hour
	}

	if req.Scopes == nil {
		req.Scopes = []string{"read", "write"}
	}

	apiKey, err := apiKeyService.GenerateAPIKey(tenantID, req.KeyName, req.Scopes, expiresIn)
	if err != nil {
		log.Printf("[ERROR] Failed to generate API key: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CreateAPIKeyResponse{
			Success: false,
			Error:   "Failed to create API key",
		})
		return
	}

	// Format: tenant-code_api-key
	var tenantCode string
	h.db.QueryRow("SELECT tenant_code FROM yt_tenants WHERE id = $1", tenantID).Scan(&tenantCode)
	formattedKey := tenantCode + "_" + apiKey

	log.Printf("[INFO] API key created for tenant %d", tenantID)
	json.NewEncoder(w).Encode(CreateAPIKeyResponse{
		Success: true,
		APIKey:  formattedKey,
		KeyID:   req.KeyName,
	})
}

// Helper function to generate tenant code from company name
func generateTenantCode(companyName string) string {
	// Simple slug generation
	code := strings.ToLower(companyName)
	code = strings.ReplaceAll(code, " ", "-")
	code = strings.ReplaceAll(code, "_", "-")
	// Remove special characters
	var result strings.Builder
	for _, r := range code {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	code = result.String()
	
	// Add random suffix to ensure uniqueness
	randomSuffix := uuid.New().String()[:8]
	return code + "-" + randomSuffix
}

