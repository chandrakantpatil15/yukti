package cache

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

type OTPCache struct {
	redis *RedisCache
}

func NewOTPCache(redis *RedisCache) *OTPCache {
	return &OTPCache{redis: redis}
}

// GenerateOTP creates and stores a 6-digit OTP
func (o *OTPCache) GenerateOTP(email string) (string, error) {
	// Generate 6-digit OTP
	max := big.NewInt(999999)
	min := big.NewInt(100000)
	n, err := rand.Int(rand.Reader, max.Sub(max, min))
	if err != nil {
		return "", err
	}
	code := fmt.Sprintf("%06d", n.Add(n, min).Int64())

	// Store in Redis with 10-minute expiry
	key := fmt.Sprintf("otp:%s", email)
	err = o.redis.Set(key, code, 10*time.Minute)
	if err != nil {
		return "", err
	}

	return code, nil
}

// ValidateOTP checks if OTP is valid
func (o *OTPCache) ValidateOTP(email, code string) (bool, error) {
	key := fmt.Sprintf("otp:%s", email)

	var storedCode string
	err := o.redis.Get(key, &storedCode)
	if err != nil {
		return false, fmt.Errorf("OTP expired or not found")
	}

	if storedCode != code {
		return false, fmt.Errorf("invalid OTP")
	}

	// Delete OTP after successful validation
	o.redis.client.Del(o.redis.ctx, key)

	return true, nil
}

// CanResendOTP checks if resend is allowed (rate limiting)
func (o *OTPCache) CanResendOTP(email string) (bool, error) {
	key := fmt.Sprintf("otp:resend:%s", email)

	// Check if resend cooldown exists
	val, err := o.redis.client.Get(o.redis.ctx, key).Result()
	if err == nil && val != "" {
		return false, fmt.Errorf("please wait 60 seconds before resending")
	}

	// Set 60-second cooldown
	o.redis.client.Set(o.redis.ctx, key, "1", 60*time.Second)

	return true, nil
}
