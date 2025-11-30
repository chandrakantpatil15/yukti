package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"yukti/internal/api/middleware"
	"yukti/internal/models"
)

type TeamHandler struct {
	db *sql.DB
}

func NewTeamHandler(db *sql.DB) *TeamHandler {
	return &TeamHandler{db: db}
}

type InviteUserRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type InviteUserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	InviteID string `json:"invite_id,omitempty"`
}

// InviteUser creates a new user invitation
func (h *TeamHandler) InviteUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	tenantID, _ := middleware.GetTenantID(r.Context())

	var req InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(InviteUserResponse{Success: false, Message: "Invalid request"})
		return
	}

	// Validate role
	if req.Role != "admin" && req.Role != "editor" && req.Role != "viewer" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(InviteUserResponse{Success: false, Message: "Invalid role"})
		return
	}

	// Check if user already exists in tenant
	var existingUser string
	err := h.db.QueryRow(`
		SELECT u.id FROM yt_users u
		JOIN yt_tenant_users tu ON u.id = tu.user_id
		WHERE u.email = $1 AND tu.tenant_id = $2
	`, req.Email, strconv.Itoa(tenantID)).Scan(&existingUser)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(InviteUserResponse{Success: false, Message: "User already in tenant"})
		return
	}

	// Generate invitation token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	// Create invitation
	var inviteID string
	err = h.db.QueryRow(`
		INSERT INTO yt_user_invitations (tenant_id, email, role, invitation_token, invited_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, strconv.Itoa(tenantID), req.Email, req.Role, token, userID, time.Now().Add(7*24*time.Hour)).Scan(&inviteID)

	if err != nil {
		log.Printf("[ERROR] Failed to create invitation: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(InviteUserResponse{Success: false, Message: "Failed to create invitation"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(InviteUserResponse{
		Success: true,
		Message: "Invitation sent",
		InviteID: inviteID,
	})
}

type TeamMember struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	JoinedAt  time.Time `json:"joined_at"`
}

type ListMembersResponse struct {
	Success bool         `json:"success"`
	Members []TeamMember `json:"members"`
}

// ListMembers returns all team members
func (h *TeamHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	rows, err := h.db.Query(`
		SELECT u.id, u.email, u.first_name, u.last_name, tu.role, tu.is_active, tu.joined_at
		FROM yt_tenant_users tu
		JOIN yt_users u ON tu.user_id = u.id
		WHERE tu.tenant_id = $1
		ORDER BY tu.role, u.email
	`, strconv.Itoa(tenantID))

	if err != nil {
		log.Printf("[ERROR] Failed to list members: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ListMembersResponse{Success: false})
		return
	}
	defer rows.Close()

	members := []TeamMember{}
	for rows.Next() {
		var m TeamMember
		var firstName, lastName sql.NullString
		err := rows.Scan(&m.UserID, &m.Email, &firstName, &lastName, &m.Role, &m.IsActive, &m.JoinedAt)
		if err != nil {
			continue
		}
		if firstName.Valid {
			m.FirstName = firstName.String
		}
		if lastName.Valid {
			m.LastName = lastName.String
		}
		members = append(members, m)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListMembersResponse{Success: true, Members: members})
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}

type UpdateRoleResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UpdateRole changes a user's role
func (h *TeamHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	tenantID, _ := middleware.GetTenantID(r.Context())
	targetUserID := strings.TrimPrefix(r.URL.Path, "/api/v1/team/members/")
	targetUserID = strings.TrimSuffix(targetUserID, "/role")

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UpdateRoleResponse{Success: false, Message: "Invalid request"})
		return
	}

	// Get current user's role
	var currentRole string
	h.db.QueryRow(`SELECT role FROM yt_tenant_users WHERE user_id = $1 AND tenant_id = $2`, 
		userID, strconv.Itoa(tenantID)).Scan(&currentRole)

	// Get target user's role
	var targetRole string
	h.db.QueryRow(`SELECT role FROM yt_tenant_users WHERE user_id = $1 AND tenant_id = $2`, 
		targetUserID, strconv.Itoa(tenantID)).Scan(&targetRole)

	// Check permissions
	if !models.CanManageUser(models.Role(currentRole), models.Role(targetRole)) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(UpdateRoleResponse{Success: false, Message: "Insufficient permissions"})
		return
	}

	// Update role
	_, err := h.db.Exec(`
		UPDATE yt_tenant_users SET role = $1, updated_at = NOW()
		WHERE user_id = $2 AND tenant_id = $3
	`, req.Role, targetUserID, strconv.Itoa(tenantID))

	if err != nil {
		log.Printf("[ERROR] Failed to update role: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(UpdateRoleResponse{Success: false, Message: "Failed to update role"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UpdateRoleResponse{Success: true, Message: "Role updated"})
}

type RemoveUserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RemoveUser removes a user from the tenant
func (h *TeamHandler) RemoveUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	tenantID, _ := middleware.GetTenantID(r.Context())
	targetUserID := strings.TrimPrefix(r.URL.Path, "/api/v1/team/members/")

	// Get current user's role
	var currentRole string
	h.db.QueryRow(`SELECT role FROM yt_tenant_users WHERE user_id = $1 AND tenant_id = $2`, 
		userID, strconv.Itoa(tenantID)).Scan(&currentRole)

	// Get target user's role
	var targetRole string
	err := h.db.QueryRow(`SELECT role FROM yt_tenant_users WHERE user_id = $1 AND tenant_id = $2`, 
		targetUserID, strconv.Itoa(tenantID)).Scan(&targetRole)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(RemoveUserResponse{Success: false, Message: "User not found"})
		return
	}

	// Cannot remove owner
	if targetRole == "owner" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(RemoveUserResponse{Success: false, Message: "Cannot remove owner"})
		return
	}

	// Check permissions
	if !models.CanManageUser(models.Role(currentRole), models.Role(targetRole)) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(RemoveUserResponse{Success: false, Message: "Insufficient permissions"})
		return
	}

	// Remove user
	_, err = h.db.Exec(`DELETE FROM yt_tenant_users WHERE user_id = $1 AND tenant_id = $2`, 
		targetUserID, strconv.Itoa(tenantID))

	if err != nil {
		log.Printf("[ERROR] Failed to remove user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RemoveUserResponse{Success: false, Message: "Failed to remove user"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RemoveUserResponse{Success: true, Message: "User removed"})
}

type Invitation struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type ListInvitationsResponse struct {
	Success     bool         `json:"success"`
	Invitations []Invitation `json:"invitations"`
}

// ListInvitations returns pending invitations
func (h *TeamHandler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())

	rows, err := h.db.Query(`
		SELECT id, email, role, status, expires_at, created_at
		FROM yt_user_invitations
		WHERE tenant_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`, strconv.Itoa(tenantID))

	if err != nil {
		log.Printf("[ERROR] Failed to list invitations: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ListInvitationsResponse{Success: false})
		return
	}
	defer rows.Close()

	invitations := []Invitation{}
	for rows.Next() {
		var inv Invitation
		rows.Scan(&inv.ID, &inv.Email, &inv.Role, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt)
		invitations = append(invitations, inv)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListInvitationsResponse{Success: true, Invitations: invitations})
}

type AcceptInviteRequest struct {
	Token string `json:"token"`
}

type AcceptInviteResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	TenantID  string `json:"tenant_id,omitempty"`
	TenantName string `json:"tenant_name,omitempty"`
}

// AcceptInvite validates token and adds user to tenant
func (h *TeamHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var req AcceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AcceptInviteResponse{Success: false, Message: "Invalid request"})
		return
	}

	var inviteID, tenantID, role string
	var expiresAt time.Time
	var status string

	err := h.db.QueryRow(`
		SELECT id, tenant_id, role, expires_at, status
		FROM yt_user_invitations
		WHERE invitation_token = $1
	`, req.Token).Scan(&inviteID, &tenantID, &role, &expiresAt, &status)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(AcceptInviteResponse{Success: false, Message: "Invalid invitation"})
		return
	}

	if status != "pending" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AcceptInviteResponse{Success: false, Message: fmt.Sprintf("Invitation already %s", status)})
		return
	}

	if time.Now().After(expiresAt) {
		h.db.Exec(`UPDATE yt_user_invitations SET status = 'expired' WHERE id = $1`, inviteID)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AcceptInviteResponse{Success: false, Message: "Invitation expired"})
		return
	}

	// Add user to tenant
	_, err = h.db.Exec(`
		INSERT INTO yt_tenant_users (user_id, tenant_id, role, is_active, joined_at)
		VALUES ($1, $2, $3, true, NOW())
		ON CONFLICT (user_id, tenant_id) DO NOTHING
	`, userID, tenantID, role)

	if err != nil {
		log.Printf("[ERROR] Failed to add user to tenant: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AcceptInviteResponse{Success: false, Message: "Failed to accept invitation"})
		return
	}

	// Mark invitation as accepted
	_, err = h.db.Exec(`
		UPDATE yt_user_invitations 
		SET status = 'accepted', accepted_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, inviteID)

	if err != nil {
		log.Printf("[ERROR] Failed to update invitation: %v", err)
	}

	// Get tenant name
	var tenantName string
	h.db.QueryRow(`SELECT company_name FROM yt_customers WHERE id = $1`, tenantID).Scan(&tenantName)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AcceptInviteResponse{
		Success: true,
		Message: "Invitation accepted",
		TenantID: tenantID,
		TenantName: tenantName,
	})
}

type GetInviteDetailsResponse struct {
	Success    bool   `json:"success"`
	Email      string `json:"email,omitempty"`
	Role       string `json:"role,omitempty"`
	TenantName string `json:"tenant_name,omitempty"`
	ExpiresAt  string `json:"expires_at,omitempty"`
	Message    string `json:"message,omitempty"`
}

// GetInviteDetails retrieves invitation info by token (public endpoint)
func (h *TeamHandler) GetInviteDetails(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GetInviteDetailsResponse{Success: false, Message: "Missing token"})
		return
	}

	var email, role, tenantID string
	var expiresAt time.Time
	var status string

	err := h.db.QueryRow(`
		SELECT email, role, tenant_id, expires_at, status
		FROM yt_user_invitations
		WHERE invitation_token = $1
	`, token).Scan(&email, &role, &tenantID, &expiresAt, &status)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(GetInviteDetailsResponse{Success: false, Message: "Invalid invitation"})
		return
	}

	if status != "pending" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GetInviteDetailsResponse{Success: false, Message: fmt.Sprintf("Invitation %s", status)})
		return
	}

	if time.Now().After(expiresAt) {
		h.db.Exec(`UPDATE yt_user_invitations SET status = 'expired' WHERE invitation_token = $1`, token)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GetInviteDetailsResponse{Success: false, Message: "Invitation expired"})
		return
	}

	var tenantName string
	h.db.QueryRow(`SELECT company_name FROM yt_customers WHERE id = $1`, tenantID).Scan(&tenantName)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetInviteDetailsResponse{
		Success: true,
		Email: email,
		Role: role,
		TenantName: tenantName,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

type ResendInviteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ResendInvite sends invitation email again
func (h *TeamHandler) ResendInvite(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	inviteID := strings.TrimPrefix(r.URL.Path, "/api/v1/team/invitations/")
	inviteID = strings.TrimSuffix(inviteID, "/resend")

	var email, token string
	var expiresAt time.Time
	var status string

	err := h.db.QueryRow(`
		SELECT email, invitation_token, expires_at, status
		FROM yt_user_invitations
		WHERE id = $1 AND tenant_id = $2
	`, inviteID, strconv.Itoa(tenantID)).Scan(&email, &token, &expiresAt, &status)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ResendInviteResponse{Success: false, Message: "Invitation not found"})
		return
	}

	if status != "pending" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResendInviteResponse{Success: false, Message: fmt.Sprintf("Invitation %s", status)})
		return
	}

	if time.Now().After(expiresAt) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResendInviteResponse{Success: false, Message: "Invitation expired"})
		return
	}

	// TODO: Resend email via EmailService
	log.Printf("[INFO] Resending invitation to %s", email)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResendInviteResponse{Success: true, Message: "Invitation resent"})
}

type RevokeInvitationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RevokeInvitation cancels a pending invitation
func (h *TeamHandler) RevokeInvitation(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := middleware.GetTenantID(r.Context())
	inviteID := strings.TrimPrefix(r.URL.Path, "/api/v1/team/invitations/")

	_, err := h.db.Exec(`
		UPDATE yt_user_invitations 
		SET status = 'revoked', updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2 AND status = 'pending'
	`, inviteID, strconv.Itoa(tenantID))

	if err != nil {
		log.Printf("[ERROR] Failed to revoke invitation: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RevokeInvitationResponse{Success: false, Message: "Failed to revoke"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RevokeInvitationResponse{Success: true, Message: "Invitation revoked"})
}
