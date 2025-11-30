package tenant

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) OnboardTenant(ctx context.Context, req OnboardingRequest) (*OnboardingResponse, error) {
	tenantCode := generateTenantCode(req.CompanyName)
	externalID := generateExternalID()
	
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create tenant with 14-day trial
	var tenantID int
	trialEnds := time.Now().AddDate(0, 0, 14)
	err = tx.QueryRowContext(ctx, `
		INSERT INTO yt_tenants (tenant_code, company_name, subscription_tier, trial_ends_at)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		tenantCode, req.CompanyName, TierProfessional, trialEnds,
	).Scan(&tenantID)
	if err != nil {
		return nil, err
	}

	// Create AWS accounts
	accounts := make([]AWSAccount, 0, len(req.AWSAccounts))
	for _, acc := range req.AWSAccounts {
		roleARN := fmt.Sprintf("arn:aws:iam::%s:role/YuktiReadOnlyRole", acc.AccountID)
		
		var accID int
		err = tx.QueryRowContext(ctx, `
			INSERT INTO yt_aws_accounts (tenant_id, account_id, account_name, role_arn, external_id)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			tenantID, acc.AccountID, acc.AccountName, roleARN, externalID,
		).Scan(&accID)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, AWSAccount{
			ID:         accID,
			TenantID:   tenantID,
			AccountID:  acc.AccountID,
			AccountName: acc.AccountName,
			RoleARN:    roleARN,
			ExternalID: externalID,
			Status:     "pending",
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &OnboardingResponse{
		TenantCode:       tenantCode,
		RoleARN:          accounts[0].RoleARN,
		ExternalID:       externalID,
		SetupScript:      generateIAMSetupScript(accounts[0].AccountID, externalID),
		Accounts:         accounts,
		SubscriptionTier: TierProfessional,
	}, nil
}

func (s *Service) GetTenant(ctx context.Context, tenantCode string) (*Tenant, error) {
	var t Tenant
	err := s.db.QueryRowContext(ctx, `
		SELECT id, tenant_code, company_name, subscription_tier, status, created_at, trial_ends_at
		FROM yt_tenants WHERE tenant_code = $1`,
		tenantCode,
	).Scan(&t.ID, &t.TenantCode, &t.CompanyName, &t.SubscriptionTier, &t.Status, &t.CreatedAt, &t.TrialEndsAt)
	return &t, err
}

func (s *Service) ListAWSAccounts(ctx context.Context, tenantID int) ([]AWSAccount, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tenant_id, account_id, account_name, role_arn, external_id, status, 
		       COALESCE(last_sync, NOW()), COALESCE(error_message, '')
		FROM yt_aws_accounts WHERE tenant_id = $1`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []AWSAccount{}
	for rows.Next() {
		var acc AWSAccount
		if err := rows.Scan(&acc.ID, &acc.TenantID, &acc.AccountID, &acc.AccountName, 
			&acc.RoleARN, &acc.ExternalID, &acc.Status, &acc.LastSync, &acc.ErrorMessage); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func generateTenantCode(companyName string) string {
	clean := strings.ToLower(strings.ReplaceAll(companyName, " ", ""))
	if len(clean) > 8 {
		clean = clean[:8]
	}
	suffix := make([]byte, 4)
	rand.Read(suffix)
	return fmt.Sprintf("%s-%s", clean, hex.EncodeToString(suffix))
}

func generateExternalID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateIAMSetupScript(accountID, externalID string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"AWS": "arn:aws:iam::YUKTI_ACCOUNT:root"},
    "Action": "sts:AssumeRole",
    "Condition": {"StringEquals": {"sts:ExternalId": "%s"}}
  }]
}

# AWS CLI command to create role:
aws iam create-role --role-name YuktiReadOnlyRole \
  --assume-role-policy-document file://trust-policy.json

aws iam attach-role-policy --role-name YuktiReadOnlyRole \
  --policy-arn arn:aws:iam::aws:policy/ReadOnlyAccess`, externalID)
}
