package middleware

import (
	"context"
	"database/sql"
	"net/http"
)

type contextKey string

const TenantIDKey contextKey = "tenant_id"

type AuthMiddleware struct {
	db *sql.DB
}

func NewAuthMiddleware(db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

// TenantAuth is DEPRECATED - use JWT middleware instead
func (m *AuthMiddleware) TenantAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func GetTenantID(ctx context.Context) (int, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(int)
	return tenantID, ok
}
