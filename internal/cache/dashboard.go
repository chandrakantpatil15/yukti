package cache

import (
	"fmt"
	"time"
)

type DashboardCache struct {
	redis *RedisCache
}

func NewDashboardCache(redis *RedisCache) *DashboardCache {
	return &DashboardCache{redis: redis}
}

// SetDashboard caches dashboard data
func (d *DashboardCache) SetDashboard(tenantID int, data interface{}) error {
	key := fmt.Sprintf("dashboard:tenant:%d", tenantID)
	return d.redis.Set(key, data, 5*time.Minute)
}

// GetDashboard retrieves cached dashboard
func (d *DashboardCache) GetDashboard(tenantID int) (map[string]interface{}, error) {
	key := fmt.Sprintf("dashboard:tenant:%d", tenantID)
	var data map[string]interface{}
	err := d.redis.Get(key, &data)
	return data, err
}

// InvalidateDashboard clears dashboard cache
func (d *DashboardCache) InvalidateDashboard(tenantID int) error {
	key := fmt.Sprintf("dashboard:tenant:%d", tenantID)
	return d.redis.client.Del(d.redis.ctx, key).Err()
}

// SetResources caches resources list
func (d *DashboardCache) SetResources(tenantID int, data interface{}) error {
	key := fmt.Sprintf("resources:tenant:%d", tenantID)
	return d.redis.Set(key, data, 10*time.Minute)
}

// GetResources retrieves cached resources
func (d *DashboardCache) GetResources(tenantID int) ([]map[string]interface{}, error) {
	key := fmt.Sprintf("resources:tenant:%d", tenantID)
	var data []map[string]interface{}
	err := d.redis.Get(key, &data)
	return data, err
}

// SetFindings caches findings list
func (d *DashboardCache) SetFindings(tenantID int, data interface{}) error {
	key := fmt.Sprintf("findings:tenant:%d", tenantID)
	return d.redis.Set(key, data, 5*time.Minute)
}

// GetFindings retrieves cached findings
func (d *DashboardCache) GetFindings(tenantID int) ([]map[string]interface{}, error) {
	key := fmt.Sprintf("findings:tenant:%d", tenantID)
	var data []map[string]interface{}
	err := d.redis.Get(key, &data)
	return data, err
}

// InvalidateAll clears all cache for tenant
func (d *DashboardCache) InvalidateAll(tenantID int) error {
	keys := []string{
		fmt.Sprintf("dashboard:tenant:%d", tenantID),
		fmt.Sprintf("resources:tenant:%d", tenantID),
		fmt.Sprintf("findings:tenant:%d", tenantID),
	}
	return d.redis.client.Del(d.redis.ctx, keys...).Err()
}
