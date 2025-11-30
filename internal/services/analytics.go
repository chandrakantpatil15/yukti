package services

import (
	"time"
	"gorm.io/gorm"
	"github.com/shopspring/decimal"
	"yukti/internal/models"
)

type AnalyticsService struct {
	db *gorm.DB
}

type CostByResourceType struct {
	ResourceType string          `json:"resource_type"`
	TotalCost    decimal.Decimal `json:"total_cost"`
	Count        int64           `json:"count"`
}

type CostTrend struct {
	Date      time.Time       `json:"date"`
	TotalCost decimal.Decimal `json:"total_cost"`
}

type UtilizationMetrics struct {
	ResourceID      string  `json:"resource_id"`
	AvgCPU          float64 `json:"avg_cpu"`
	AvgMemory       float64 `json:"avg_memory"`
	ResourceType    string  `json:"resource_type"`
	Environment     string  `json:"environment"`
}

func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

func (s *AnalyticsService) GetCostByResourceType(days int) ([]CostByResourceType, error) {
	var results []CostByResourceType
	
	startDate := time.Now().AddDate(0, 0, -days)
	
	err := s.db.Model(&models.ResourceCost{}).
		Select("r.resource_type, COALESCE(SUM(rc.cost_usd), 0) as total_cost, COUNT(DISTINCT rc.resource_id) as count").
		Joins("JOIN resources r ON rc.resource_id = r.id").
		Where("rc.date >= ?", startDate).
		Group("r.resource_type").
		Scan(&results).Error

	return results, err
}

func (s *AnalyticsService) GetCostTrend(days int) ([]CostTrend, error) {
	var results []CostTrend
	
	startDate := time.Now().AddDate(0, 0, -days)
	
	err := s.db.Model(&models.ResourceCost{}).
		Select("date, COALESCE(SUM(cost_usd), 0) as total_cost").
		Where("date >= ?", startDate).
		Group("date").
		Order("date ASC").
		Scan(&results).Error

	return results, err
}

func (s *AnalyticsService) GetUtilizationMetrics(days int) ([]UtilizationMetrics, error) {
	var results []UtilizationMetrics
	
	startDate := time.Now().AddDate(0, 0, -days)
	
	err := s.db.Model(&models.ResourceMetrics{}).
		Select(`
			r.resource_id,
			AVG(rm.cpu_utilization) as avg_cpu,
			AVG(rm.memory_utilization) as avg_memory,
			r.resource_type,
			r.environment
		`).
		Joins("JOIN resources r ON rm.resource_id = r.id").
		Where("rm.timestamp >= ?", startDate).
		Group("r.id, r.resource_id, r.resource_type, r.environment").
		Scan(&results).Error

	return results, err
}

func (s *AnalyticsService) GetTopCostResources(days int, limit int) ([]models.Resource, error) {
	var resources []models.Resource
	
	startDate := time.Now().AddDate(0, 0, -days)
	
	err := s.db.Model(&models.Resource{}).
		Select("resources.*, COALESCE(SUM(rc.cost_usd), 0) as total_cost").
		Joins("LEFT JOIN resource_costs rc ON resources.id = rc.resource_id AND rc.date >= ?", startDate).
		Group("resources.id").
		Order("total_cost DESC").
		Limit(limit).
		Find(&resources).Error

	return resources, err
}