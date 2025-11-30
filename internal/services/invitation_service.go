package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

type InvitationService struct {
	db           *sql.DB
	emailService *EmailService
}

func NewInvitationService(db *sql.DB, emailService *EmailService) *InvitationService {
	return &InvitationService{
		db:           db,
		emailService: emailService,
	}
}

type Invitation struct {
	ID        string
	TenantID  string
	Email     string
	Role      string
	Token     string
	InvitedBy string
	ExpiresAt time.Time
}

// CreateInvitation generates token and sends email
func (s *InvitationService) CreateInvitation(tenantID, email, role, invitedByUserID string) (*Invitation, error) {
	// Generate token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Create invitation
	var inviteID string
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	err := s.db.QueryRow(`
		INSERT INTO yt_user_invitations (tenant_id, email, role, invitation_token, invited_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, tenantID, email, role, token, invitedByUserID, expiresAt).Scan(&inviteID)

	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Get tenant name
	var tenantName string
	s.db.QueryRow(`SELECT company_name FROM yt_customers WHERE id = $1`, tenantID).Scan(&tenantName)

	// Send invitation email
	inviteURL := fmt.Sprintf("http://localhost:3000/accept-invite?token=%s", token)
	err = s.emailService.SendInvitationEmail(email, tenantName, inviteURL)
	if err != nil {
		log.Printf("[WARN] Failed to send invitation email: %v", err)
	}

	return &Invitation{
		ID:        inviteID,
		TenantID:  tenantID,
		Email:     email,
		Role:      role,
		Token:     token,
		InvitedBy: invitedByUserID,
		ExpiresAt: expiresAt,
	}, nil
}

// AcceptInvitation validates token and adds user to tenant
func (s *InvitationService) AcceptInvitation(token, userID string) error {
	var inviteID, tenantID, role string
	var expiresAt time.Time
	var status string

	err := s.db.QueryRow(`
		SELECT id, tenant_id, role, expires_at, status
		FROM yt_user_invitations
		WHERE invitation_token = $1
	`, token).Scan(&inviteID, &tenantID, &role, &expiresAt, &status)

	if err != nil {
		return fmt.Errorf("invalid invitation token")
	}

	if status != "pending" {
		return fmt.Errorf("invitation already %s", status)
	}

	if time.Now().After(expiresAt) {
		s.db.Exec(`UPDATE yt_user_invitations SET status = 'expired' WHERE id = $1`, inviteID)
		return fmt.Errorf("invitation expired")
	}

	// Add user to tenant
	_, err = s.db.Exec(`
		INSERT INTO yt_tenant_users (user_id, tenant_id, role, is_active, joined_at)
		VALUES ($1, $2, $3, true, NOW())
		ON CONFLICT (user_id, tenant_id) DO NOTHING
	`, userID, tenantID, role)

	if err != nil {
		return fmt.Errorf("failed to add user to tenant: %w", err)
	}

	// Mark invitation as accepted
	_, err = s.db.Exec(`
		UPDATE yt_user_invitations 
		SET status = 'accepted', accepted_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, inviteID)

	return err
}

// GetInvitationByToken retrieves invitation details
func (s *InvitationService) GetInvitationByToken(token string) (*Invitation, error) {
	var inv Invitation
	var expiresAt time.Time

	err := s.db.QueryRow(`
		SELECT id, tenant_id, email, role, invitation_token, invited_by, expires_at
		FROM yt_user_invitations
		WHERE invitation_token = $1 AND status = 'pending'
	`, token).Scan(&inv.ID, &inv.TenantID, &inv.Email, &inv.Role, &inv.Token, &inv.InvitedBy, &expiresAt)

	if err != nil {
		return nil, fmt.Errorf("invitation not found")
	}

	inv.ExpiresAt = expiresAt

	if time.Now().After(expiresAt) {
		s.db.Exec(`UPDATE yt_user_invitations SET status = 'expired' WHERE id = $1`, inv.ID)
		return nil, fmt.Errorf("invitation expired")
	}

	return &inv, nil
}

// ResendInvitation sends invitation email again
func (s *InvitationService) ResendInvitation(inviteID, tenantID string) error {
	var email, token string
	var expiresAt time.Time
	var status string

	err := s.db.QueryRow(`
		SELECT email, invitation_token, expires_at, status
		FROM yt_user_invitations
		WHERE id = $1 AND tenant_id = $2
	`, inviteID, tenantID).Scan(&email, &token, &expiresAt, &status)

	if err != nil {
		return fmt.Errorf("invitation not found")
	}

	if status != "pending" {
		return fmt.Errorf("invitation already %s", status)
	}

	if time.Now().After(expiresAt) {
		return fmt.Errorf("invitation expired")
	}

	// Get tenant name
	var tenantName string
	s.db.QueryRow(`SELECT company_name FROM yt_customers WHERE id = $1`, tenantID).Scan(&tenantName)

	// Resend email
	inviteURL := fmt.Sprintf("http://localhost:3000/accept-invite?token=%s", token)
	return s.emailService.SendInvitationEmail(email, tenantName, inviteURL)
}

// ExpireOldInvitations marks expired invitations
func (s *InvitationService) ExpireOldInvitations() error {
	_, err := s.db.Exec(`
		UPDATE yt_user_invitations 
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'pending' AND expires_at < NOW()
	`)
	return err
}
