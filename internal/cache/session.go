package cache

import (
	"fmt"
	"time"
)

type SessionCache struct {
	redis *RedisCache
}

func NewSessionCache(redis *RedisCache) *SessionCache {
	return &SessionCache{redis: redis}
}

type SessionData struct {
	UserID    string `json:"user_id"`
	TenantID  int    `json:"tenant_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expires_at"`
}

// SetSession stores JWT session data
func (s *SessionCache) SetSession(userID string, data SessionData) error {
	key := fmt.Sprintf("session:%s", userID)
	return s.redis.Set(key, data, 24*time.Hour)
}

// GetSession retrieves JWT session data
func (s *SessionCache) GetSession(userID string) (*SessionData, error) {
	key := fmt.Sprintf("session:%s", userID)
	var data SessionData
	err := s.redis.Get(key, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// DeleteSession invalidates session (logout)
func (s *SessionCache) DeleteSession(userID string) error {
	key := fmt.Sprintf("session:%s", userID)
	return s.redis.client.Del(s.redis.ctx, key).Err()
}

// RefreshSession extends session TTL
func (s *SessionCache) RefreshSession(userID string) error {
	key := fmt.Sprintf("session:%s", userID)
	return s.redis.client.Expire(s.redis.ctx, key, 24*time.Hour).Err()
}
