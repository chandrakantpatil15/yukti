package security

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"
)

type TokenService struct {
	db *sql.DB
}

func NewTokenService(db *sql.DB) *TokenService {
	return &TokenService{db: db}
}

// GenerateRefreshToken creates a new refresh token
func (s *TokenService) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// StoreRefreshToken saves refresh token to database
func (s *TokenService) StoreRefreshToken(userID string, tenantID int, token, deviceInfo, ipAddress string, expiresAt time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO yt_refresh_tokens (user_id, tenant_id, token, expires_at, device_info, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, tenantID, token, expiresAt, deviceInfo, ipAddress)
	return err
}

// ValidateRefreshToken checks if refresh token is valid
func (s *TokenService) ValidateRefreshToken(token string) (userID string, tenantID int, err error) {
	var expiresAt time.Time
	var isRevoked bool

	err = s.db.QueryRow(`
		SELECT user_id, tenant_id, expires_at, is_revoked
		FROM yt_refresh_tokens
		WHERE token = $1
	`, token).Scan(&userID, &tenantID, &expiresAt, &isRevoked)

	if err == sql.ErrNoRows {
		return "", 0, errors.New("invalid refresh token")
	}
	if err != nil {
		return "", 0, err
	}
	if isRevoked {
		return "", 0, errors.New("refresh token has been revoked")
	}
	if time.Now().After(expiresAt) {
		return "", 0, errors.New("refresh token has expired")
	}

	// Update last used timestamp
	s.db.Exec(`UPDATE yt_refresh_tokens SET last_used_at = NOW() WHERE token = $1`, token)

	return userID, tenantID, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (s *TokenService) RevokeRefreshToken(token string) error {
	_, err := s.db.Exec(`UPDATE yt_refresh_tokens SET is_revoked = true WHERE token = $1`, token)
	return err
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (s *TokenService) RevokeAllUserTokens(userID string) error {
	_, err := s.db.Exec(`UPDATE yt_refresh_tokens SET is_revoked = true WHERE user_id = $1`, userID)
	return err
}

// BlacklistToken adds JWT token to blacklist
func (s *TokenService) BlacklistToken(jti, userID string, tenantID int, expiresAt time.Time, reason string) error {
	_, err := s.db.Exec(`
		INSERT INTO yt_token_blacklist (token_jti, user_id, tenant_id, expires_at, reason)
		VALUES ($1, $2, $3, $4, $5)
	`, jti, userID, tenantID, expiresAt, reason)
	return err
}

// IsTokenBlacklisted checks if JWT token is blacklisted
func (s *TokenService) IsTokenBlacklisted(jti string) bool {
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM yt_token_blacklist WHERE token_jti = $1)`, jti).Scan(&exists)
	return err == nil && exists
}

// LogSessionActivity logs user session activity
func (s *TokenService) LogSessionActivity(userID string, tenantID int, action, ipAddress, userAgent string) error {
	_, err := s.db.Exec(`
		INSERT INTO yt_session_activity (user_id, tenant_id, action, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, tenantID, action, ipAddress, userAgent)
	return err
}

// CleanupExpiredTokens removes expired tokens from database
func (s *TokenService) CleanupExpiredTokens() error {
	_, err := s.db.Exec(`DELETE FROM yt_refresh_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`DELETE FROM yt_token_blacklist WHERE expires_at < NOW()`)
	return err
}
