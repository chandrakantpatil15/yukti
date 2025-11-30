package security

import (
	"database/sql"
	"fmt"
)

type SecretsManager struct {
	db         *sql.DB
	encryption *EncryptionService
}

func NewSecretsManager(db *sql.DB, encryptionKey string) *SecretsManager {
	return &SecretsManager{
		db:         db,
		encryption: NewEncryptionService(encryptionKey),
	}
}

func (s *SecretsManager) StoreSecret(tenantID int, key, value, secretType string) error {
	encrypted, err := s.encryption.Encrypt(value)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO yt_secrets (tenant_id, secret_key, secret_value_encrypted, secret_type)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, secret_key) 
		DO UPDATE SET secret_value_encrypted = $3, updated_at = NOW()`,
		tenantID, key, encrypted, secretType,
	)

	return err
}

func (s *SecretsManager) GetSecret(tenantID int, key string) (string, error) {
	var encrypted string
	err := s.db.QueryRow(`
		SELECT secret_value_encrypted FROM yt_secrets
		WHERE tenant_id = $1 AND secret_key = $2`,
		tenantID, key,
	).Scan(&encrypted)

	if err != nil {
		return "", fmt.Errorf("secret not found")
	}

	return s.encryption.Decrypt(encrypted)
}

func (s *SecretsManager) DeleteSecret(tenantID int, key string) error {
	_, err := s.db.Exec(`
		DELETE FROM yt_secrets WHERE tenant_id = $1 AND secret_key = $2`,
		tenantID, key,
	)
	return err
}
