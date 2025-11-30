package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

// RefreshToken generates new access token using refresh token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request format",
		})
		return
	}

	// Validate refresh token
	userID, tenantID, err := h.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get user details
	var email, role string
	err = h.db.QueryRow(`
		SELECT email, role FROM yt_users WHERE id = $1
	`, userID).Scan(&email, &role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to fetch user details",
		})
		return
	}

	// Generate new access token (15 minutes)
	token, err := h.jwtService.GenerateToken(userID, tenantID, "tenant-"+strconv.Itoa(tenantID), email, role, []string{"read", "write"}, 15*time.Minute)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to generate token",
		})
		return
	}

	// Log activity
	h.tokenService.LogSessionActivity(userID, tenantID, "refresh", r.RemoteAddr, r.UserAgent())

	log.Printf("Token refreshed for user: %s", email)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   token,
	})
}

// Logout revokes refresh token and blacklists access token
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	json.NewDecoder(r.Body).Decode(&req)

	// Get JWT from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && len(authHeader) > 7 {
		tokenString := authHeader[7:] // Remove "Bearer "
		
		// Validate and extract claims
		claims, err := h.jwtService.ValidateToken(tokenString)
		if err == nil {
			// Blacklist the access token
			expiresAt := time.Unix(claims.ExpiresAt, 0)
			h.tokenService.BlacklistToken(claims.JTI, claims.UserID, claims.TenantID, expiresAt, "user_logout")
			
			// Log activity
			h.tokenService.LogSessionActivity(claims.UserID, claims.TenantID, "logout", r.RemoteAddr, r.UserAgent())
		}
	}

	// Revoke refresh token if provided
	if req.RefreshToken != "" {
		h.tokenService.RevokeRefreshToken(req.RefreshToken)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}

// LogoutAll revokes all user sessions
func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	// Get JWT from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) <= 7 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No authorization token provided",
		})
		return
	}

	tokenString := authHeader[7:]
	claims, err := h.jwtService.ValidateToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid token",
		})
		return
	}

	// Revoke all refresh tokens for user
	if err := h.tokenService.RevokeAllUserTokens(claims.UserID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to revoke sessions",
		})
		return
	}

	// Log activity
	h.tokenService.LogSessionActivity(claims.UserID, claims.TenantID, "logout_all", r.RemoteAddr, r.UserAgent())

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "All sessions logged out successfully",
	})
}
