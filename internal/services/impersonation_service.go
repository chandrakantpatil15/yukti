package services

import (
	"database/sql"
	"fmt"
	"time"

	"yukti/internal/security"
)

type ImpersonationService struct {
	db         *sql.DB
	jwtService *security.JWTService
}

func NewImpersonationService(db *sql.DB, jwtSecret string) *ImpersonationService {
	return &ImpersonationService{
		db:         db,
		jwtService: security.NewJWTService(jwtSecret),
	}
}

// StartImpersonation creates impersonation session and generates JWT
func (s *ImpersonationService) StartImpersonation(adminUserID, targetUserID, tenantID, reason string) (string, error) {
	// Get target user info
	var email, role string
	err := s.db.QueryRow(`
		SELECT u.email, tu.role
		FROM yt_users u
		JOIN yt_tenant_users tu ON u.id = tu.user_id
		WHERE u.id = $1 AND tu.tenant_id = $2
	`, targetUserID, tenantID).Scan(&email, &role)

	if err != nil {
		return "", fmt.Errorf("target user not found in tenant")
	}

	// Create impersonation session
	var sessionID string
	err = s.db.QueryRow(`
		INSERT INTO yt_impersonation_sessions (admin_user_id, target_user_id, tenant_id, reason, is_active)
		VALUES ($1, $2, $3, $4, true)
		RETURNING id
	`, adminUserID, targetUserID, tenantID, reason).Scan(&sessionID)

	if err != nil {
		return "", fmt.Errorf("failed to create impersonation session: %w", err)
	}

	// Generate JWT for target user (1 hour)
	tenantIDInt := 0 // Will be set from tenant_id
	token, err := s.jwtService.GenerateToken(targetUserID, tenantIDInt, tenantID, email, role, []string{"impersonated"}, 1*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// EndImpersonation ends active impersonation session
func (s *ImpersonationService) EndImpersonation(adminUserID string) error {
	_, err := s.db.Exec(`
		UPDATE yt_impersonation_sessions
		SET is_active = false, ended_at = NOW()
		WHERE admin_user_id = $1 AND is_active = true
	`, adminUserID)

	return err
}

// GetActiveImpersonation returns active impersonation session
func (s *ImpersonationService) GetActiveImpersonation(adminUserID string) (map[string]interface{}, error) {
	var sessionID, targetUserID, tenantID, reason string
	var startedAt time.Time

	err := s.db.QueryRow(`
		SELECT id, target_user_id, tenant_id, reason, started_at
		FROM yt_impersonation_sessions
		WHERE admin_user_id = $1 AND is_active = true
		ORDER BY started_at DESC
		LIMIT 1
	`, adminUserID).Scan(&sessionID, &targetUserID, &tenantID, &reason, &startedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"session_id":     sessionID,
		"target_user_id": targetUserID,
		"tenant_id":      tenantID,
		"reason":         reason,
		"started_at":     startedAt,
	}, nil
}
