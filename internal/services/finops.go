package services

import (
	"time"
	"gorm.io/gorm"
	"github.com/shopspring/decimal"
	"github.com/cloudcostoptimizer/yukti/internal/models"
)

type FinOpsService struct {
	db *gorm.DB
}

func NewFinOpsService(db *gorm.DB) *FinOpsService {
	return &FinOpsService{db: db}
}

type CostSummary struct {
	TotalCost     decimal.Decimal `json:"total_cost"`
	ResourceCount int64           `json:"resource_count"`
	Period        string          `json:"period"`
}

func (s *FinOpsService) GetCostSummary(days int) (*CostSummary, error) {
	var result struct {
		TotalCost     decimal.Decimal
		ResourceCount int64
	}

	startDate := time.Now().AddDate(0, 0, -days)
	
	err := s.db.Model(&models.ResourceCost{}).
		Select("COALESCE(SUM(cost_usd), 0) as total_cost, COUNT(DISTINCT resource_id) as resource_count").
		Where("date >= ?", startDate).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return &CostSummary{
		TotalCost:     result.TotalCost,
		ResourceCount: result.ResourceCount,
		Period:        "last_30_days",
	}, nil
}

func (s *FinOpsService) GetOptimizationRecommendations() ([]models.OptimizationRecommendation, error) {
	var recommendations []models.OptimizationRecommendation
	
	err := s.db.Preload("Resource").
		Where("status = ?", "active").
		Order("potential_savings DESC").
		Limit(10).
		Find(&recommendations).Error

	return recommendations, err
}