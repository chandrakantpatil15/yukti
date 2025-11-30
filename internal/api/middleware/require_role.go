package middleware

import (
	"encoding/json"
	"log"
	"net/http"
)

// RequireRole creates middleware that checks if user has one of the allowed roles
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	// Build map for O(1) lookup
	allowedMap := make(map[string]bool)
	for _, role := range allowedRoles {
		allowedMap[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get role from context (set by JWT middleware)
			role, ok := GetRole(r.Context())
			if !ok {
				log.Printf("[WARN] Role not found in context for path: %s", r.URL.Path)
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Unauthorized - role not found",
				})
				return
			}

			// Check if role is allowed
			if !allowedMap[role] {
				log.Printf("[WARN] Insufficient permissions: role=%s, required=%v, path=%s", role, allowedRoles, r.URL.Path)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Forbidden - insufficient permissions",
					"required_role": allowedRoles,
					"user_role": role,
				})
				return
			}

			log.Printf("[DEBUG] Role check passed: role=%s, path=%s", role, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

