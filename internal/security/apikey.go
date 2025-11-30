package security

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type APIKeyService struct {
	db *sql.DB
}

func NewAPIKeyService(db *sql.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

func (s *APIKeyService) GenerateAPIKey(tenantID int, keyName string, scopes []string, expiresIn time.Duration) (string, error) {
	keyBytes := make([]byte, 32)
	rand.Read(keyBytes)
	apiKey := hex.EncodeToString(keyBytes)

	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])
	keyPrefix := apiKey[:8]

	var expiresAt *time.Time
	if expiresIn > 0 {
		exp := time.Now().Add(expiresIn)
		expiresAt = &exp
	}

	_, err := s.db.Exec(`
		INSERT INTO yt_api_keys (tenant_id, key_name, key_hash, key_prefix, scopes, expires_at)
		VALUES ($1, $2, $3, $4, $5::text[], $6)`,
		tenantID, keyName, keyHash, keyPrefix, fmt.Sprintf("{%s}", strings.Join(scopes, ",")), expiresAt,
	)

	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func (s *APIKeyService) ValidateAPIKey(apiKey string) (int, []string, error) {
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])

	var tenantID int
	var scopes []string
	var expiresAt sql.NullTime
	var revoked bool

	err := s.db.QueryRow(`
		SELECT tenant_id, scopes, expires_at, revoked
		FROM yt_api_keys
		WHERE key_hash = $1`,
		keyHash,
	).Scan(&tenantID, &scopes, &expiresAt, &revoked)

	if err != nil {
		return 0, nil, fmt.Errorf("invalid API key")
	}

	if revoked {
		return 0, nil, fmt.Errorf("API key revoked")
	}

	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return 0, nil, fmt.Errorf("API key expired")
	}

	s.db.Exec(`UPDATE yt_api_keys SET last_used = NOW() WHERE key_hash = $1`, keyHash)

	return tenantID, scopes, nil
}

func (s *APIKeyService) RevokeAPIKey(apiKey string) error {
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])

	_, err := s.db.Exec(`UPDATE yt_api_keys SET revoked = TRUE WHERE key_hash = $1`, keyHash)
	return err
}
