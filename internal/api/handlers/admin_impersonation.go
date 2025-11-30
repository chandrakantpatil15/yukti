package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"yukti/internal/api/middleware"
	"yukti/internal/config"
	"yukti/internal/services"
)

type AdminImpersonationHandler struct {
	db                   *sql.DB
	impersonationService *services.ImpersonationService
}

func NewAdminImpersonationHandler(db *sql.DB) *AdminImpersonationHandler {
	secrets := config.GetSecrets()
	return &AdminImpersonationHandler{
		db:                   db,
		impersonationService: services.NewImpersonationService(db, secrets.JWTSecret),
	}
}

type ImpersonateRequest struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Reason   string `json:"reason"`
}

type ImpersonateResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ImpersonateUser starts impersonation session
func (h *AdminImpersonationHandler) ImpersonateUser(w http.ResponseWriter, r *http.Request) {
	adminID, _ := middleware.GetAdminUserID(r.Context())

	var req ImpersonateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ImpersonateResponse{Success: false, Error: "Invalid request"})
		return
	}

	if req.Reason == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ImpersonateResponse{Success: false, Error: "Reason is required"})
		return
	}

	// Start impersonation
	token, err := h.impersonationService.StartImpersonation(adminID, req.UserID, req.TenantID, req.Reason)
	if err != nil {
		log.Printf("[ERROR] Impersonation failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ImpersonateResponse{Success: false, Error: err.Error()})
		return
	}

	// Log to audit trail
	h.db.Exec(`
		INSERT INTO yt_admin_audit_logs (admin_user_id, action, resource_type, resource_id, tenant_id, target_user_id, details)
		VALUES ($1, 'impersonate_user', 'user', $2, $3, $2, $4)
	`, adminID, req.UserID, req.TenantID, `{"reason":"`+req.Reason+`"}`)

	log.Printf("[ADMIN] Impersonation started: admin=%s, user=%s, tenant=%s, reason=%s", adminID, req.UserID, req.TenantID, req.Reason)
	json.NewEncoder(w).Encode(ImpersonateResponse{
		Success: true,
		Token:   token,
		Message: "Impersonation started",
	})
}

type EndImpersonationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// EndImpersonation ends active impersonation session
func (h *AdminImpersonationHandler) EndImpersonation(w http.ResponseWriter, r *http.Request) {
	adminID, _ := middleware.GetAdminUserID(r.Context())

	err := h.impersonationService.EndImpersonation(adminID)
	if err != nil {
		log.Printf("[ERROR] Failed to end impersonation: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(EndImpersonationResponse{Success: false, Error: "Failed to end impersonation"})
		return
	}

	// Log to audit trail
	h.db.Exec(`
		INSERT INTO yt_admin_audit_logs (admin_user_id, action, resource_type, details)
		VALUES ($1, 'end_impersonation', 'session', '{}')
	`, adminID)

	log.Printf("[ADMIN] Impersonation ended: admin=%s", adminID)
	json.NewEncoder(w).Encode(EndImpersonationResponse{Success: true, Message: "Impersonation ended"})
}

type AdminUsersHandler struct {
	db *sql.DB
}

func NewAdminUsersHandler(db *sql.DB) *AdminUsersHandler {
	return &AdminUsersHandler{db: db}
}

type UserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	IsActive  bool   `json:"is_active"`
	Tenants   int    `json:"tenants"`
	CreatedAt string `json:"created_at"`
}

type ListUsersResponse struct {
	Success bool       `json:"success"`
	Users   []UserInfo `json:"users"`
	Total   int        `json:"total"`
}

// ListUsers returns all users across all tenants
func (h *AdminUsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT 
			u.id,
			u.email,
			COALESCE(u.first_name, ''),
			COALESCE(u.last_name, ''),
			u.is_active,
			u.created_at,
			COALESCE((SELECT COUNT(*) FROM yt_tenant_users tu WHERE tu.user_id = u.id), 0)
		FROM yt_users u
		ORDER BY u.created_at DESC
	`)

	if err != nil {
		log.Printf("[ERROR] Failed to list users: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ListUsersResponse{Success: false})
		return
	}
	defer rows.Close()

	users := []UserInfo{}
	for rows.Next() {
		var u UserInfo
		rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.IsActive, &u.CreatedAt, &u.Tenants)
		users = append(users, u)
	}

	json.NewEncoder(w).Encode(ListUsersResponse{
		Success: true,
		Users:   users,
		Total:   len(users),
	})
}

type UserActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// SuspendUser suspends a user account
func (h *AdminUsersHandler) SuspendUser(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	userID = strings.TrimSuffix(userID, "/suspend")

	adminID, _ := middleware.GetAdminUserID(r.Context())

	_, err := h.db.Exec(`UPDATE yt_users SET is_active = false WHERE id = $1`, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to suspend user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(UserActionResponse{Success: false, Error: "Failed to suspend user"})
		return
	}

	// Log admin action
	h.db.Exec(`
		INSERT INTO yt_admin_audit_logs (admin_user_id, action, resource_type, resource_id, target_user_id)
		VALUES ($1, 'suspend_user', 'user', $2, $2)
	`, adminID, userID)

	log.Printf("[ADMIN] User suspended: %s by admin %s", userID, adminID)
	json.NewEncoder(w).Encode(UserActionResponse{Success: true, Message: "User suspended"})
}

// ActivateUser activates a suspended user
func (h *AdminUsersHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	userID = strings.TrimSuffix(userID, "/activate")

	adminID, _ := middleware.GetAdminUserID(r.Context())

	_, err := h.db.Exec(`UPDATE yt_users SET is_active = true WHERE id = $1`, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to activate user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(UserActionResponse{Success: false, Error: "Failed to activate user"})
		return
	}

	// Log admin action
	h.db.Exec(`
		INSERT INTO yt_admin_audit_logs (admin_user_id, action, resource_type, resource_id, target_user_id)
		VALUES ($1, 'activate_user', 'user', $2, $2)
	`, adminID, userID)

	log.Printf("[ADMIN] User activated: %s by admin %s", userID, adminID)
	json.NewEncoder(w).Encode(UserActionResponse{Success: true, Message: "User activated"})
}
