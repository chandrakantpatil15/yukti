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
	"yukti/internal/services"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db         *sql.DB
	gormDB     *gorm.DB
	jwtService *security.JWTService
	otpService *services.OTPService
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
		otpService: services.NewOTPService(db),
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
	OTPCode string `json:"otp_code,omitempty"` // Only in dev mode
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

	// Send OTP for email verification
	otpCode, err := h.otpService.SendOTPAndGetCode(req.Email)
	if err != nil {
		log.Printf("[ERROR] Failed to send OTP: %v", err)
		// Don't fail signup, just log the error
	}

	log.Printf("[INFO] Successfully created user %s for tenant %d", user.ID, tenantID)
	
	// Return OTP code in dev mode only
	response := SignupResponse{
		Success:  true,
		UserID:   user.ID.String(),
		TenantID: tenantID,
		Message:  "User created successfully. Please verify your email.",
	}
	
	// Show OTP in dev mode (when JWT_SECRET is default)
	if os.Getenv("JWT_SECRET") == "" {
		response.OTPCode = otpCode
		log.Printf("[DEV] OTP code for %s: %s", req.Email, otpCode)
	}
	
	json.NewEncoder(w).Encode(response)
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Success   bool      `json:"success"`
	Token     string    `json:"token,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	User      *UserInfoAuth `json:"user,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// UserInfoAuth represents user information in login response
type UserInfoAuth struct {
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
	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Email and password are required",
		})
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Get user by email (email is unique across all tenants)
	userRepo := models.NewUserRepository(h.db)
	user, err := userRepo.GetUserByEmailSQL(req.Email)
	if err != nil {
		log.Printf("[WARN] User not found or inactive: %s", req.Email)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Invalid email or password",
		})
		return
	}

	// Check tenant status
	var tenantStatus string
	err = h.db.QueryRow(`
		SELECT status FROM yt_tenants WHERE id = $1
	`, user.TenantID).Scan(&tenantStatus)
	if err != nil || tenantStatus != "active" {
		log.Printf("[WARN] Tenant %d is not active (status: %s)", user.TenantID, tenantStatus)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Account is suspended",
		})
		return
	}

	// Check if email is verified
	if !user.EmailVerified {
		log.Printf("[WARN] Email not verified for user %s", req.Email)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Please verify your email before logging in",
		})
		return
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		log.Printf("[WARN] Invalid password for user %s", req.Email)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Success: false,
			Error:   "Invalid email or password",
		})
		return
	}

	// Get tenant code for JWT
	var tenantCode string
	h.db.QueryRow("SELECT tenant_code FROM yt_tenants WHERE id = $1", user.TenantID).Scan(&tenantCode)

	// Generate JWT token
	expiresIn := 24 * time.Hour // 24 hours
	token, err := h.jwtService.GenerateToken(
		user.ID.String(),
		user.TenantID,
		tenantCode,
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

	log.Printf("[INFO] User %s logged in successfully for tenant %d", req.Email, user.TenantID)
	json.NewEncoder(w).Encode(LoginResponse{
		Success:   true,
		Token:     token,
		ExpiresAt: time.Now().Add(expiresIn),
		User: &UserInfoAuth{
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


// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required"`
}

// VerifyEmailResponse represents email verification response
type VerifyEmailResponse struct {
	Success   bool          `json:"success"`
	Token     string        `json:"token,omitempty"`
	ExpiresAt time.Time     `json:"expires_at,omitempty"`
	User      *UserInfoAuth `json:"user,omitempty"`
	Error     string        `json:"error,omitempty"`
}

// VerifyEmail handles email verification with OTP
// POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] VerifyEmail request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to decode verify email request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(VerifyEmailResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Verify OTP
	err := h.otpService.VerifyOTP(req.Email, req.Code)
	if err != nil {
		log.Printf("[WARN] OTP verification failed for %s: %v", req.Email, err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(VerifyEmailResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Mark email as verified
	_, err = h.db.Exec(`
		UPDATE yt_users 
		SET email_verified = true, updated_at = NOW()
		WHERE email = $1
	`, req.Email)
	if err != nil {
		log.Printf("[ERROR] Failed to update email_verified: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(VerifyEmailResponse{
			Success: false,
			Error:   "Failed to verify email",
		})
		return
	}

	// Get user details
	userRepo := models.NewUserRepository(h.db)
	user, err := userRepo.GetUserByEmailSQL(req.Email)
	if err != nil {
		log.Printf("[ERROR] Failed to get user after verification: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(VerifyEmailResponse{
			Success: false,
			Error:   "Email verified but failed to generate token",
		})
		return
	}

	// Get tenant code for JWT
	var tenantCode string
	h.db.QueryRow("SELECT tenant_code FROM yt_tenants WHERE id = $1", user.TenantID).Scan(&tenantCode)

	// Generate JWT token
	expiresIn := 24 * time.Hour
	token, err := h.jwtService.GenerateToken(
		user.ID.String(),
		user.TenantID,
		tenantCode,
		user.Email,
		user.Role,
		[]string{"read", "write"},
		expiresIn,
	)
	if err != nil {
		log.Printf("[ERROR] Failed to generate token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(VerifyEmailResponse{
			Success: false,
			Error:   "Email verified but failed to generate token",
		})
		return
	}

	log.Printf("[INFO] Email verified successfully for %s", req.Email)
	json.NewEncoder(w).Encode(VerifyEmailResponse{
		Success:   true,
		Token:     token,
		ExpiresAt: time.Now().Add(expiresIn),
		User: &UserInfoAuth{
			ID:       user.ID.String(),
			Email:    user.Email,
			Role:     user.Role,
			TenantID: user.TenantID,
		},
	})
}

// ResendOTPRequest represents resend OTP request
type ResendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResendOTPResponse represents resend OTP response
type ResendOTPResponse struct {
	Success bool   `json:"success"`
	OTPCode string `json:"otp_code,omitempty"` // Only in dev mode
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ResendOTP handles OTP resend
// POST /api/v1/auth/resend-code
func (h *AuthHandler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] ResendOTP request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	var req ResendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to decode resend OTP request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResendOTPResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM yt_users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil || !exists {
		log.Printf("[WARN] User not found for resend OTP: %s", req.Email)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ResendOTPResponse{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	// Send new OTP
	otpCode, err := h.otpService.SendOTPAndGetCode(req.Email)
	if err != nil {
		log.Printf("[ERROR] Failed to resend OTP: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ResendOTPResponse{
			Success: false,
			Error:   "Failed to send OTP",
		})
		return
	}

	response := ResendOTPResponse{
		Success: true,
		Message: "OTP sent successfully",
	}

	// Show OTP in dev mode
	if os.Getenv("JWT_SECRET") == "" {
		response.OTPCode = otpCode
		log.Printf("[DEV] Resent OTP code for %s: %s", req.Email, otpCode)
	}

	json.NewEncoder(w).Encode(response)
}
