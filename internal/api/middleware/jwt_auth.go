package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"yukti/internal/security"

	"os"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	TenantIDKey contextKey = "tenant_id"
	RoleKey     contextKey = "role"
	EmailKey    contextKey = "email"
)

type JWTAuthMiddleware struct {
	jwtService *security.JWTService
	db         *sql.DB
}

func NewJWTAuthMiddleware(db *sql.DB) *JWTAuthMiddleware {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "yukti-secret-key-change-in-production"
		log.Printf("[WARN] JWT_SECRET not set, using default (change in production!)")
	}

	return &JWTAuthMiddleware{
		jwtService: security.NewJWTService(jwtSecret),
		db:         db,
	}
}

// RequireAuth validates JWT token and sets user context
func (m *JWTAuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("[WARN] Missing Authorization header from IP: %s", r.RemoteAddr)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Missing Authorization header",
			})
			return
		}

		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("[WARN] Invalid Authorization header format from IP: %s", r.RemoteAddr)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid Authorization header format. Expected: Bearer <token>",
			})
			return
		}

		token := parts[1]

		// Validate token
		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			log.Printf("[WARN] Invalid token from IP: %s, error: %v", r.RemoteAddr, err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid or expired token",
			})
			return
		}

		// Verify user is still active
		var isActive bool
		err = m.db.QueryRow(`
			SELECT is_active FROM yt_users WHERE id::text = $1 AND tenant_id = $2
		`, claims.UserID, claims.TenantID).Scan(&isActive)
		if err != nil || !isActive {
			log.Printf("[WARN] User %s is inactive or not found", claims.UserID)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "User account is inactive",
			})
			return
		}

		// Verify tenant is still active
		var tenantStatus string
		err = m.db.QueryRow(`
			SELECT status FROM yt_tenants WHERE id = $1
		`, claims.TenantID).Scan(&tenantStatus)
		if err != nil || tenantStatus != "active" {
			log.Printf("[WARN] Tenant %d is inactive or not found", claims.TenantID)
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Tenant account is suspended",
			})
			return
		}

		// Set context values
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)

		log.Printf("[DEBUG] JWT authenticated: user=%s, tenant=%d, role=%s", claims.UserID, claims.TenantID, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetTenantID extracts tenant ID from context (compatible with existing code)
func GetTenantID(ctx context.Context) (int, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(int)
	return tenantID, ok
}

// GetRole extracts role from context
func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok
}

// GetEmail extracts email from context
func GetEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(EmailKey).(string)
	return email, ok
}

