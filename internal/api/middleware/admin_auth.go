package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"yukti/internal/config"
	"yukti/internal/models"
	"yukti/internal/security"
)

const (
	AdminUserIDKey contextKey = "admin_user_id"
	AdminRoleKey   contextKey = "admin_role"
	AdminEmailKey  contextKey = "admin_email"
)

type AdminAuthMiddleware struct {
	db         *sql.DB
	jwtService *security.JWTService
}

func NewAdminAuthMiddleware(db *sql.DB) *AdminAuthMiddleware {
	secrets := config.GetSecrets()
	return &AdminAuthMiddleware{
		db:         db,
		jwtService: security.NewJWTService(secrets.JWTSecret),
	}
}

// RequireAdmin validates admin JWT and checks permissions
func (m *AdminAuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Missing Authorization header",
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid Authorization header format",
			})
			return
		}

		token := parts[1]
		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid or expired token",
			})
			return
		}

		// Verify admin user exists and is active
		var isActive bool
		var role string
		err = m.db.QueryRow(`
			SELECT is_active, role FROM yt_admin_users WHERE id::text = $1
		`, claims.UserID).Scan(&isActive, &role)

		if err != nil {
			log.Printf("[ERROR] Admin user %s not found: %v", claims.UserID, err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid admin token",
			})
			return
		}

		if !isActive {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Admin account is inactive",
			})
			return
		}

		// Set admin context
		ctx := context.WithValue(r.Context(), AdminUserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, AdminRoleKey, role)
		ctx = context.WithValue(ctx, AdminEmailKey, claims.Email)

		log.Printf("[ADMIN] Request: user=%s, role=%s, path=%s", claims.UserID, role, r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdminPermission checks specific admin permission
func (m *AdminAuthMiddleware) RequireAdminPermission(permission models.AdminPermission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(AdminRoleKey)
			if role == nil {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Unauthorized",
				})
				return
			}

			adminRole := models.AdminRole(role.(string))
			if !models.HasAdminPermission(adminRole, permission) {
				log.Printf("[WARN] Admin lacks permission: role=%s, permission=%s", role, permission)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Insufficient permissions",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetAdminUserID extracts admin user ID from context
func GetAdminUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(AdminUserIDKey).(string)
	return userID, ok
}

// GetAdminRole extracts admin role from context
func GetAdminRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(AdminRoleKey).(string)
	return role, ok
}
