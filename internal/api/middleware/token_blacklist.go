package middleware

import (
	"database/sql"
	"net/http"
	"strings"

	"yukti/internal/security"
)

type TokenBlacklistMiddleware struct {
	tokenService *security.TokenService
	jwtService   *security.JWTService
}

func NewTokenBlacklistMiddleware(db *sql.DB, jwtSecret string) *TokenBlacklistMiddleware {
	return &TokenBlacklistMiddleware{
		tokenService: security.NewTokenService(db),
		jwtService:   security.NewJWTService(jwtSecret),
	}
}

func (m *TokenBlacklistMiddleware) CheckBlacklist(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			next.ServeHTTP(w, r)
			return
		}

		tokenString := authHeader[7:]
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Check if token is blacklisted
		if m.tokenService.IsTokenBlacklisted(claims.JTI) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"Token has been revoked"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
