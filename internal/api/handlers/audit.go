package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type AuditHandler struct {
	db *sql.DB
}

func NewAuditHandler(db *sql.DB) *AuditHandler {
	return &AuditHandler{db: db}
}

func (h *AuditHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] GetAuditLogs called from IP: %s", r.RemoteAddr)
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "100"
	}

	log.Printf("[DEBUG] Fetching audit logs with limit: %s", limit)
	rows, err := h.db.Query(`
		SELECT id, user_id, action, resource_type, resource_id, tenant_id, 
		       ip_address, user_agent, metadata, created_at
		FROM yt_audit_logs
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		log.Printf("[ERROR] Failed to query audit logs: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	defer rows.Close()

	logs := []map[string]interface{}{}
	for rows.Next() {
		var id, action, resourceType string
		var adminUser, resourceID, ipAddress, userAgent sql.NullString
		var tenantID sql.NullInt64
		var details sql.NullString
		var createdAt sql.NullTime

		if err := rows.Scan(&id, &adminUser, &action, &resourceType, &resourceID, &tenantID, 
			&ipAddress, &userAgent, &details, &createdAt); err != nil {
			log.Printf("[WARN] Failed to scan audit log row: %v", err)
			continue
		}

		tenantIDStr := ""
		if tenantID.Valid {
			tenantIDStr = fmt.Sprintf("%d", tenantID.Int64)
		}

		logs = append(logs, map[string]interface{}{
			"id":            id,
			"admin_user":    adminUser.String,
			"action":        action,
			"resource_type": resourceType,
			"resource_id":   resourceID.String,
			"tenant_id":     tenantIDStr,
			"ip_address":    ipAddress.String,
			"user_agent":    userAgent.String,
			"details":       details.String,
			"created_at":    createdAt.Time.Format("2006-01-02 15:04:05"),
		})
	}

	log.Printf("[INFO] Successfully fetched %d audit logs", len(logs))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"logs":    logs,
	})
}
