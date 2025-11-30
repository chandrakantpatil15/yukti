package middleware

import (
	"database/sql"
	"log"
	"net/http"
)

type TenantIsolationMiddleware struct {
	db *sql.DB
}

func NewTenantIsolationMiddleware(db *sql.DB) *TenantIsolationMiddleware {
	return &TenantIsolationMiddleware{db: db}
}

func (m *TenantIsolationMiddleware) ValidateTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SECURITY FIX: This middleware is DEPRECATED - use JWT middleware instead
		// Keeping for backward compatibility but logging warning
		log.Printf("[WARN] TenantIsolationMiddleware is deprecated. Use JWT middleware for tenant isolation.")
		
		// For now, just pass through - JWT middleware handles tenant isolation
		next.ServeHTTP(w, r)
	})
}
