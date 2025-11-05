package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"github.com/cloudcostoptimizer/yukti/internal/models"
)

type PricingService struct {
	db *gorm.DB
}

type AWSPricing struct {
	InstanceType string          `json:"instance_type"`
	Region       string          `json:"region"`
	OS           string          `json:"os"`
	OnDemand     decimal.Decimal `json:"on_demand"`
	Reserved1Yr  decimal.Decimal `json:"reserved_1yr"`
	SpotPrice    decimal.Decimal `json:"spot_price"`
}

func NewPricingService(db *gorm.DB) *PricingService {
	return &PricingService{db: db}
}

// Fetch latest AWS pricing from public API
func (s *PricingService) UpdatePricingData() error {
	// AWS Pricing API endpoint (simplified - real implementation uses AWS Price List API)
	url := "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/us-east-1/index.json"
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch pricing data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse AWS pricing JSON (simplified structure)
	var pricingData map[string]interface{}
	if err := json.Unmarshal(body, &pricingData); err != nil {
		return err
	}

	// Process and store pricing data
	return s.processPricingData(pricingData)
}

func (s *PricingService) processPricingData(data map[string]interface{}) error {
	// Simplified processing - real implementation parses complex AWS pricing JSON
	instanceTypes := []string{"t3.micro", "t3.small", "m5.large", "c5.xlarge"}
	
	for _, instanceType := range instanceTypes {
		pricing := models.AWSPricing{
			InstanceType: instanceType,
			Region:       "us-east-1",
			OS:           "Linux",
			// These would be parsed from actual AWS pricing data
			PricePerHour:        s.getOnDemandPrice(instanceType),
			RI1YrNoUpfront:      s.getReservedPrice(instanceType, 0.7),
			RI1YrPartialUpfront: s.getReservedPrice(instanceType, 0.6),
			SpotPriceAvg:        s.getSpotPrice(instanceType, 0.3),
			UpdatedAt:           time.Now(),
		}

		s.db.Where("instance_type = ? AND region = ? AND os = ?", 
			pricing.InstanceType, pricing.Region, pricing.OS).
			FirstOrCreate(&pricing)
	}

	return nil
}

// Calculate optimization recommendations
func (s *PricingService) GenerateOptimizationRecommendations() error {
	var resources []models.Resource
	s.db.Where("status = ?", "running").Find(&resources)

	for _, resource := range resources {
		// Get current pricing
		var pricing models.AWSPricing
		s.db.Where("instance_type = ? AND region = ?", 
			resource.InstanceType, resource.Region).First(&pricing)

		// Calculate potential savings
		currentCost := pricing.PricePerHour.Mul(decimal.NewFromInt(24 * 30)) // Monthly
		riCost := pricing.RI1YrNoUpfront.Mul(decimal.NewFromInt(24 * 30))
		spotCost := pricing.SpotPriceAvg.Mul(decimal.NewFromInt(24 * 30))

		// Reserved Instance recommendation
		if currentCost.GreaterThan(riCost) {
			recommendation := models.OptimizationRecommendation{
				ResourceID:         resource.ID,
				RecommendationType: "reserved_instance",
				CurrentCost:        currentCost,
				OptimizedCost:      riCost,
				PotentialSavings:   currentCost.Sub(riCost),
				Confidence:         0.85,
				Status:            "active",
			}
			s.db.Create(&recommendation)
		}

		// Spot instance recommendation for dev/test
		if resource.Environment != "prod" && currentCost.GreaterThan(spotCost) {
			recommendation := models.OptimizationRecommendation{
				ResourceID:         resource.ID,
				RecommendationType: "spot_instance",
				CurrentCost:        currentCost,
				OptimizedCost:      spotCost,
				PotentialSavings:   currentCost.Sub(spotCost),
				Confidence:         0.75,
				Status:            "active",
			}
			s.db.Create(&recommendation)
		}
	}

	return nil
}

func (s *PricingService) getOnDemandPrice(instanceType string) decimal.Decimal {
	// Simplified pricing - real implementation fetches from AWS API
	prices := map[string]string{
		"t3.micro":  "0.0104",
		"t3.small":  "0.0208",
		"m5.large":  "0.096",
		"c5.xlarge": "0.17",
	}
	
	if price, exists := prices[instanceType]; exists {
		p, _ := decimal.NewFromString(price)
		return p
	}
	return decimal.NewFromFloat(0.05) // Default
}

func (s *PricingService) getReservedPrice(instanceType string, discount float64) decimal.Decimal {
	onDemand := s.getOnDemandPrice(instanceType)
	return onDemand.Mul(decimal.NewFromFloat(discount))
}

func (s *PricingService) getSpotPrice(instanceType string, discount float64) decimal.Decimal {
	onDemand := s.getOnDemandPrice(instanceType)
	return onDemand.Mul(decimal.NewFromFloat(discount))
}