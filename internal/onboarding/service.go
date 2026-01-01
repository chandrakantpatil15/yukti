package onboarding

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateCustomer(ctx context.Context, companyName, email string) (*Customer, error) {
	customer := &Customer{
		ID:               uuid.New().String(),
		TenantID:         uuid.New().String(),
		CompanyName:      companyName,
		Email:            email,
		OnboardingStatus: StatusPending,
		OnboardingStep:   StepAWSConnection,
		CreatedAt:        time.Now(),
	}

	query := `
		INSERT INTO yt_customers (id, tenant_id, company_name, email, onboarding_status, onboarding_step, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.ExecContext(ctx, query,
		customer.ID, customer.TenantID, customer.CompanyName, customer.Email,
		customer.OnboardingStatus, customer.OnboardingStep, customer.CreatedAt,
	)

	return customer, err
}

func (s *Service) SaveAWSConnection(ctx context.Context, conn *AWSConnection) error {
	query := `
		INSERT INTO yt_aws_connections (tenant_id, account_id, role_arn, external_id, regions, verified, last_verified_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id) DO UPDATE SET
			account_id = $2, role_arn = $3, external_id = $4, regions = $5, verified = $6, last_verified_at = $7
	`
	_, err := s.db.ExecContext(ctx, query,
		conn.TenantID, conn.AccountID, conn.RoleARN, conn.ExternalID,
		pq.Array(conn.Regions), conn.Verified, conn.LastVerifiedAt,
	)
	return err
}

func (s *Service) GetAWSConnection(ctx context.Context, tenantID string) (*AWSConnection, error) {
	query := `
		SELECT tenant_id, account_id, role_arn, external_id, regions, verified, last_verified_at
		FROM yt_aws_connections WHERE tenant_id = $1
	`
	var conn AWSConnection
	var regions pq.StringArray
	err := s.db.QueryRowContext(ctx, query, tenantID).Scan(
		&conn.TenantID, &conn.AccountID, &conn.RoleARN, &conn.ExternalID,
		&regions, &conn.Verified, &conn.LastVerifiedAt,
	)
	if err != nil {
		return nil, err
	}
	conn.Regions = []string(regions)
	return &conn, nil
}

func (s *Service) SaveMetricsIntegration(ctx context.Context, integration *MetricsIntegration) error {
	query := `
		INSERT INTO yt_metrics_integrations (tenant_id, source, endpoint, verified, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.ExecContext(ctx, query,
		integration.TenantID, integration.Source, integration.Endpoint,
		integration.Verified, integration.CreatedAt,
	)
	return err
}

func (s *Service) UpdateOnboardingStep(ctx context.Context, tenantID string, step OnboardingStep) error {
	query := `UPDATE yt_customers SET onboarding_step = $1 WHERE tenant_id = $2`
	_, err := s.db.ExecContext(ctx, query, step, tenantID)
	return err
}

func (s *Service) CompleteOnboarding(ctx context.Context, tenantID string) error {
	now := time.Now()
	query := `UPDATE yt_customers SET onboarding_status = $1, completed_at = $2 WHERE tenant_id = $3`
	_, err := s.db.ExecContext(ctx, query, StatusCompleted, now, tenantID)
	return err
}

func (s *Service) GetCustomer(ctx context.Context, tenantID string) (*Customer, error) {
	query := `
		SELECT id, tenant_id, company_name, email, onboarding_status, onboarding_step, created_at, completed_at
		FROM yt_customers WHERE tenant_id = $1
	`
	var customer Customer
	err := s.db.QueryRowContext(ctx, query, tenantID).Scan(
		&customer.ID, &customer.TenantID, &customer.CompanyName, &customer.Email,
		&customer.OnboardingStatus, &customer.OnboardingStep, &customer.CreatedAt, &customer.CompletedAt,
	)
	return &customer, err
}

func (s *Service) GenerateExternalID(tenantID string) string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("yukti-%s-%s", tenantID, base64.URLEncoding.EncodeToString(b)[:12])
}

// GetDB returns the database connection for scanner access
func (s *Service) GetDB() *sql.DB {
	return s.db
}
