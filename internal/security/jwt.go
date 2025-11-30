package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type JWTClaims struct {
	UserID     string   `json:"user_id"`     // UUID string
	TenantID   int      `json:"tenant_id"`
	TenantCode string   `json:"tenant_code"`
	Email      string   `json:"email"`
	Role       string   `json:"role"`        // admin, editor, viewer
	Scopes     []string `json:"scopes"`
	IssuedAt   int64    `json:"iat"`
	ExpiresAt  int64    `json:"exp"`
}

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

// GenerateToken generates a JWT token with user and tenant information
func (j *JWTService) GenerateToken(userID string, tenantID int, tenantCode, email, role string, scopes []string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:     userID,
		TenantID:   tenantID,
		TenantCode: tenantCode,
		Email:      email,
		Role:       role,
		Scopes:     scopes,
		IssuedAt:   now.Unix(),
		ExpiresAt:  now.Add(expiresIn).Unix(),
	}

	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	message := headerB64 + "." + claimsB64
	signature := j.sign(message)

	return message + "." + signature, nil
}

func (j *JWTService) ValidateToken(token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	message := parts[0] + "." + parts[1]
	signature := parts[2]

	if j.sign(message) != signature {
		return nil, fmt.Errorf("invalid signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid claims encoding")
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("invalid claims format")
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

func (j *JWTService) sign(message string) string {
	h := hmac.New(sha256.New, j.secret)
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
