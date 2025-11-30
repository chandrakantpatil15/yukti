package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(addr, password string, db int) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     20,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	return &RedisCache{
		client: rdb,
		ctx:    context.Background(),
	}
}

func (r *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(r.ctx, key, data, expiration).Err()
}

func (r *RedisCache) Get(key string, dest interface{}) error {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	return json.Unmarshal([]byte(val), dest)
}

func (r *RedisCache) SetPricingData(region, instanceType string, pricing interface{}) error {
	key := fmt.Sprintf("pricing:%s:%s", region, instanceType)
	return r.Set(key, pricing, 24*time.Hour)
}

func (r *RedisCache) GetPricingData(region, instanceType string, dest interface{}) error {
	key := fmt.Sprintf("pricing:%s:%s", region, instanceType)
	return r.Get(key, dest)
}

func (r *RedisCache) SetResourceData(customerID, resourceID string, resource interface{}) error {
	key := fmt.Sprintf("resource:%s:%s", customerID, resourceID)
	return r.Set(key, resource, 1*time.Hour)
}

func (r *RedisCache) GetResourceData(customerID, resourceID string, dest interface{}) error {
	key := fmt.Sprintf("resource:%s:%s", customerID, resourceID)
	return r.Get(key, dest)
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}