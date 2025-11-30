package whitelist

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateWhitelist(ctx context.Context, tenantID, userEmail string, req CreateWhitelistRequest) (*Whitelist, error) {
	costImpact, _ := s.estimateCostImpact(ctx, tenantID, req)
	
	status := StatusActive
	if costImpact > 1000.00 {
		status = StatusPendingApproval
	}

	var expiresAt *time.Time
	if req.ExpiresInDays != nil {
		expiry := time.Now().AddDate(0, 0, *req.ExpiresInDays)
		expiresAt = &expiry
	}

	whitelist := &Whitelist{
		ID:                    uuid.New().String(),
		TenantID:              tenantID,
		WhitelistType:         req.WhitelistType,
		ResourceARN:           req.ResourceARN,
		TagKey:                req.TagKey,
		TagValue:              req.TagValue,
		ServiceName:           req.ServiceName,
		RecommendationType:    req.RecommendationType,
		Reason:                req.Reason,
		BusinessJustification: req.BusinessJustification,
		CostImpactMonthly:     costImpact,
		CreatedBy:             userEmail,
		CreatedAt:             time.Now(),
		ExpiresAt:             expiresAt,
		Status:                status,
	}

	query := `
		INSERT INTO yt_whitelists 
		(id, tenant_id, whitelist_type, resource_arn, tag_key, tag_value, 
		 service_name, recommendation_type, reason, business_justification, 
		 cost_impact_monthly, created_by, created_at, expires_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := s.db.ExecContext(ctx, query,
		whitelist.ID, whitelist.TenantID, whitelist.WhitelistType,
		whitelist.ResourceARN, whitelist.TagKey, whitelist.TagValue,
		whitelist.ServiceName, whitelist.RecommendationType,
		whitelist.Reason, whitelist.BusinessJustification,
		whitelist.CostImpactMonthly, whitelist.CreatedBy,
		whitelist.CreatedAt, whitelist.ExpiresAt, whitelist.Status,
	)

	return whitelist, err
}

func (s *Service) IsResourceWhitelisted(ctx context.Context, tenantID, resourceARN, recommendationType string) (bool, error) {
	serviceName := extractServiceFromARN(resourceARN)

	query := `
		SELECT COUNT(*) FROM yt_whitelists
		WHERE tenant_id = $1
		AND status = 'active'
		AND (expires_at IS NULL OR expires_at > NOW())
		AND (
			(whitelist_type = 'resource' AND resource_arn = $2)
			OR (whitelist_type = 'recommendation_type' AND recommendation_type = $3)
			OR (whitelist_type = 'service' AND service_name = $4)
		)
	`

	var count int
	err := s.db.QueryRowContext(ctx, query, tenantID, resourceARN, recommendationType, serviceName).Scan(&count)
	return count > 0, err
}

func (s *Service) ListWhitelists(ctx context.Context, tenantID string) ([]Whitelist, error) {
	query := `
		SELECT id, tenant_id, whitelist_type, resource_arn, reason, 
		       cost_impact_monthly, created_by, created_at, expires_at, status
		FROM yt_whitelists
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var whitelists []Whitelist
	for rows.Next() {
		var w Whitelist
		err := rows.Scan(&w.ID, &w.TenantID, &w.WhitelistType, &w.ResourceARN,
			&w.Reason, &w.CostImpactMonthly, &w.CreatedBy, &w.CreatedAt,
			&w.ExpiresAt, &w.Status)
		if err != nil {
			continue
		}
		whitelists = append(whitelists, w)
	}

	return whitelists, nil
}

func (s *Service) RevokeWhitelist(ctx context.Context, whitelistID, userEmail, reason string) error {
	query := `
		UPDATE yt_whitelists
		SET status = 'revoked', revoked_by = $2, revoked_at = $3, revoked_reason = $4
		WHERE id = $1
	`
	_, err := s.db.ExecContext(ctx, query, whitelistID, userEmail, time.Now(), reason)
	return err
}

func (s *Service) estimateCostImpact(ctx context.Context, tenantID string, req CreateWhitelistRequest) (float64, error) {
	if req.WhitelistType == WhitelistTypeResource && req.ResourceARN != nil {
		query := `
			SELECT COALESCE(SUM(potential_savings_monthly), 0)
			FROM yt_tenant_recommendations
			WHERE tenant_id = $1 AND resource_arn = $2 AND status = 'open'
		`
		var total float64
		err := s.db.QueryRowContext(ctx, query, tenantID, *req.ResourceARN).Scan(&total)
		return total, err
	}
	return 0, nil
}

func extractServiceFromARN(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}
