package services

import (
	"gorm.io/gorm"
	"github.com/shopspring/decimal"
)

type SimpleAWSService struct {
	db *gorm.DB
}

func NewSimpleAWSService(db *gorm.DB) *SimpleAWSService {
	return &SimpleAWSService{db: db}
}

// Mock AWS resource sync for demo
func (s *SimpleAWSService) SyncEC2Resources() error {
	// This would normally call AWS APIs
	// For now, return success (data comes from SQL scripts)
	return nil
}

// Mock cost data sync for demo
func (s *SimpleAWSService) SyncCostData(days int) error {
	// This would normally call AWS Cost Explorer
	// For now, return success (data comes from SQL scripts)
	return nil
}

// Mock spot pricing for demo
func (s *SimpleAWSService) GetSpotPricing(instanceTypes []string) (map[string]decimal.Decimal, error) {
	pricing := make(map[string]decimal.Decimal)
	
	// Mock spot prices (30% of on-demand)
	mockPrices := map[string]string{
		"t3.micro":  "0.0031",
		"t3.small":  "0.0062",
		"m5.large":  "0.0288",
		"c5.xlarge": "0.051",
	}
	
	for _, instanceType := range instanceTypes {
		if price, exists := mockPrices[instanceType]; exists {
			spotPrice, _ := decimal.NewFromString(price)
			pricing[instanceType] = spotPrice
		}
	}
	
	return pricing, nil
}