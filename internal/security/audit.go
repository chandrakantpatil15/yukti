package security

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type AuditService struct {
	db *sql.DB
}

type AuditLog struct {
	TenantID     *int
	UserID       string
	Action       string
	ResourceType string
	ResourceID   string
	IPAddress    string
	UserAgent    string
	Method       string
	Path         string
	StatusCode   int
	ErrorMessage string
	Metadata     map[string]interface{}
}

func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) Log(log AuditLog) error {
	metadata, _ := json.Marshal(log.Metadata)

	_, err := s.db.Exec(`
		INSERT INTO yt_audit_logs 
		(tenant_id, user_id, action, resource_type, resource_id, ip_address, 
		 user_agent, request_method, request_path, status_code, error_message, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		log.TenantID, log.UserID, log.Action, log.ResourceType, log.ResourceID,
		log.IPAddress, log.UserAgent, log.Method, log.Path, log.StatusCode,
		log.ErrorMessage, metadata,
	)

	return err
}

func (s *AuditService) LogRequest(r *http.Request, tenantID *int, statusCode int, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	s.Log(AuditLog{
		TenantID:   tenantID,
		Action:     "api_request",
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: statusCode,
		ErrorMessage: errMsg,
	})
}

func (s *AuditService) GetAuditLogs(tenantID int, limit int) ([]AuditLog, error) {
	rows, err := s.db.Query(`
		SELECT action, resource_type, resource_id, ip_address, request_method, 
		       request_path, status_code, created_at
		FROM yt_audit_logs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2`,
		tenantID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []AuditLog{}
	for rows.Next() {
		var log AuditLog
		var createdAt time.Time
		rows.Scan(&log.Action, &log.ResourceType, &log.ResourceID, &log.IPAddress,
			&log.Method, &log.Path, &log.StatusCode, &createdAt)
		logs = append(logs, log)
	}

	return logs, nil
}
