package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"yukti/internal/models"
)

type RoleAuthMiddleware struct {
	db *sql.DB
}

func NewRoleAuthMiddleware(db *sql.DB) *RoleAuthMiddleware {
	return &RoleAuthMiddleware{db: db}
}

// RequireRole checks if user has required role for tenant
func (m *RoleAuthMiddleware) RequireRole(allowedRoles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserID(r.Context())
			if !ok {
				log.Printf("[ERROR] User ID not found in context")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Unauthorized",
				})
				return
			}

			tenantIDVal, ok := GetTenantID(r.Context())
			if !ok {
				log.Printf("[ERROR] Tenant ID not found in context")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Unauthorized",
				})
				return
			}
			tenantID := strconv.Itoa(tenantIDVal)

			// Get user's role for this tenant
			var role string
			var isActive bool
			err := m.db.QueryRow(`
				SELECT role, is_active 
				FROM yt_tenant_users 
				WHERE user_id = $1 AND tenant_id = $2
			`, userID, tenantID).Scan(&role, &isActive)

			if err != nil {
				log.Printf("[ERROR] User %s not found in tenant %s: %v", userID, tenantID, err)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Access denied",
				})
				return
			}

			if !isActive {
				log.Printf("[WARN] User %s is inactive in tenant %s", userID, tenantID)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "User is inactive",
				})
				return
			}

			// Check if user has required role
			userRole := models.Role(role)
			hasRole := false
			for _, allowedRole := range allowedRoles {
				if userRole == allowedRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				log.Printf("[WARN] User %s with role %s attempted to access resource requiring %v", userID, role, allowedRoles)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Insufficient permissions",
				})
				return
			}

			// Add role to context
			ctx := context.WithValue(r.Context(), RoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission checks if user has specific permission
func (m *RoleAuthMiddleware) RequirePermission(permission models.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserID(r.Context())
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Unauthorized",
				})
				return
			}

			tenantIDVal, ok := GetTenantID(r.Context())
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Unauthorized",
				})
				return
			}
			tenantID := strconv.Itoa(tenantIDVal)

			// Get user's role
			var role string
			err := m.db.QueryRow(`
				SELECT role FROM yt_tenant_users 
				WHERE user_id = $1 AND tenant_id = $2 AND is_active = true
			`, userID, tenantID).Scan(&role)

			if err != nil {
				log.Printf("[ERROR] User %s not found in tenant %s: %v", userID, tenantID, err)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Access denied",
				})
				return
			}

			// Check permission
			if !models.HasPermission(models.Role(role), permission) {
				log.Printf("[WARN] User %s with role %s lacks permission %s", userID, role, permission)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Insufficient permissions",
				})
				return
			}

			ctx := context.WithValue(r.Context(), RoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
