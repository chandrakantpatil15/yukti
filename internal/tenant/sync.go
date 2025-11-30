package tenant

import (
	"context"
	"database/sql"
)

type SyncService struct {
	db *sql.DB
}

func NewSyncService(db *sql.DB) *SyncService {
	return &SyncService{db: db}
}

func (s *SyncService) SyncAWSAccount(ctx context.Context, accountID int) error {
	var acc AWSAccount
	err := s.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, role_arn, external_id
		FROM yt_aws_accounts WHERE id = $1`,
		accountID,
	).Scan(&acc.ID, &acc.TenantID, &acc.AccountID, &acc.RoleARN, &acc.ExternalID)
	if err != nil {
		return err
	}

	// TODO: Implement AWS SDK AssumeRole integration
	// For demo, mark as active
	_, err = s.db.ExecContext(ctx, `
		UPDATE yt_aws_accounts 
		SET status = 'active', last_sync = NOW(), error_message = NULL
		WHERE id = $1`,
		accountID,
	)

	return err
}

func (s *SyncService) updateSyncError(ctx context.Context, accountID int, err error) error {
	_, execErr := s.db.ExecContext(ctx, `
		UPDATE yt_aws_accounts 
		SET status = 'error', error_message = $1, last_sync = NOW()
		WHERE id = $2`,
		err.Error(), accountID,
	)
	if execErr != nil {
		return execErr
	}
	return err
}

func (s *SyncService) SyncAllTenants(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id FROM yt_aws_accounts 
		WHERE status IN ('pending', 'active')
		AND (last_sync IS NULL OR last_sync < NOW() - INTERVAL '1 hour')`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var accountID int
		if err := rows.Scan(&accountID); err != nil {
			continue
		}
		go s.SyncAWSAccount(context.Background(), accountID)
	}

	return nil
}
