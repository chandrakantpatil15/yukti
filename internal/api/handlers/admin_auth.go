package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"yukti/internal/security"

	"golang.org/x/crypto/bcrypt"
)

type AdminAuthHandler struct {
	db         *sql.DB
	jwtService *security.JWTService
}

func NewAdminAuthHandler(db *sql.DB, jwtSecret string) *AdminAuthHandler {
	return &AdminAuthHandler{
		db:         db,
		jwtService: security.NewJWTService(jwtSecret),
	}
}

type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AdminLoginResponse struct {
	Success bool        `json:"success"`
	Token   string      `json:"token,omitempty"`
	Admin   interface{} `json:"admin,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// AdminLogin authenticates platform admin
func (h *AdminAuthHandler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AdminLoginResponse{Success: false, Error: "Invalid request"})
		return
	}

	// Get admin user
	var adminID, passwordHash, role string
	var isActive bool
	err := h.db.QueryRow(`
		SELECT id, password_hash, role, is_active
		FROM yt_admin_users
		WHERE email = $1
	`, req.Email).Scan(&adminID, &passwordHash, &role, &isActive)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AdminLoginResponse{Success: false, Error: "Invalid credentials"})
		return
	}

	if err != nil {
		log.Printf("[ERROR] Admin login query failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AdminLoginResponse{Success: false, Error: "Login failed"})
		return
	}

	if !isActive {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(AdminLoginResponse{Success: false, Error: "Account is inactive"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AdminLoginResponse{Success: false, Error: "Invalid credentials"})
		return
	}

	// Generate JWT token (24 hours for admin)
	token, err := h.jwtService.GenerateToken(adminID, 0, "admin", req.Email, role, []string{"admin"}, 24*time.Hour)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AdminLoginResponse{Success: false, Error: "Failed to generate token"})
		return
	}

	// Update last login
	h.db.Exec(`
		UPDATE yt_admin_users 
		SET last_login_at = NOW(), last_login_ip = $1
		WHERE id = $2
	`, r.RemoteAddr, adminID)

	log.Printf("[ADMIN] Login successful: %s (role: %s)", req.Email, role)
	json.NewEncoder(w).Encode(AdminLoginResponse{
		Success: true,
		Token:   token,
		Admin: map[string]interface{}{
			"id":    adminID,
			"email": req.Email,
			"role":  role,
		},
	})
}
