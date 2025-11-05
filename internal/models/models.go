package models

import (
	"time"
	"github.com/shopspring/decimal"
)

type Resource struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ResourceID   string    `json:"resource_id" gorm:"uniqueIndex;type:varchar(255)"`
	ResourceType string    `json:"resource_type"`
	InstanceType string    `json:"instance_type"`
	Region       string    `json:"region"`
	Status       string    `json:"status"`
	ProjectID    *uint     `json:"project_id"`
	Environment  string    `json:"environment"`
	LaunchTime   time.Time `json:"launch_time"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ResourceCost struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	ResourceID uint            `json:"resource_id"`
	Date       time.Time       `json:"date"`
	CostUSD    decimal.Decimal `json:"cost_usd" gorm:"type:decimal(10,4)"`
	UsageHours int             `json:"usage_hours"`
	DataSource string          `json:"data_source"`
	CreatedAt  time.Time       `json:"created_at"`
}

type OptimizationRecommendation struct {
	ID                uint            `json:"id" gorm:"primaryKey"`
	ResourceID        uint            `json:"resource_id"`
	RecommendationType string         `json:"recommendation_type"`
	CurrentCost       decimal.Decimal `json:"current_cost" gorm:"type:decimal(10,4)"`
	OptimizedCost     decimal.Decimal `json:"optimized_cost" gorm:"type:decimal(10,4)"`
	PotentialSavings  decimal.Decimal `json:"potential_savings" gorm:"type:decimal(10,4)"`
	Confidence        float64         `json:"confidence"`
	Status            string          `json:"status"`
	CreatedAt         time.Time       `json:"created_at"`
}

type AWSPricing struct {
	ID                  uint            `json:"id" gorm:"primaryKey"`
	InstanceType        string          `json:"instance_type" gorm:"index"`
	Region              string          `json:"region"`
	OS                  string          `json:"os"`
	PricePerHour        decimal.Decimal `json:"price_per_hour" gorm:"type:decimal(10,4)"`
	RI1YrNoUpfront      decimal.Decimal `json:"ri_1yr_no_upfront" gorm:"type:decimal(10,4)"`
	RI1YrPartialUpfront decimal.Decimal `json:"ri_1yr_partial_upfront" gorm:"type:decimal(10,4)"`
	SpotPriceAvg        decimal.Decimal `json:"spot_price_avg" gorm:"type:decimal(10,4)"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type ResourceMetrics struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ResourceID       uint      `json:"resource_id"`
	Timestamp        time.Time `json:"timestamp"`
	CPUUtilization   float64   `json:"cpu_utilization"`
	MemoryUtilization float64  `json:"memory_utilization"`
	CreatedAt        time.Time `json:"created_at"`
}