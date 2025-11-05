package services

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/cloudcostoptimizer/yukti/internal/models"
)

type AWSPricingService struct {
	db            *gorm.DB
	pricingClient *pricing.Client
}

type PriceData struct {
	OnDemand string `json:"OnDemand"`
	Reserved string `json:"Reserved"`
}

func NewAWSPricingService(db *gorm.DB) (*AWSPricingService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return nil, err
	}

	return &AWSPricingService{
		db:            db,
		pricingClient: pricing.NewFromConfig(cfg),
	}, nil
}

// Import real AWS pricing data
func (s *AWSPricingService) ImportEC2Pricing(ctx context.Context, region string) error {
	log.Printf("Importing AWS EC2 pricing for region: %s", region)

	// Common EC2 instance types to fetch pricing for
	instanceTypes := []string{
		"m5.large", "m5.xlarge", "m5.2xlarge", "m5.4xlarge", "m5.8xlarge", "m5.12xlarge", "m5.16xlarge", "m5.24xlarge",
		"c5.large", "c5.xlarge", "c5.2xlarge", "c5.4xlarge", "c5.9xlarge", "c5.12xlarge", "c5.18xlarge", "c5.24xlarge",
		"r5.large", "r5.xlarge", "r5.2xlarge", "r5.4xlarge", "r5.8xlarge", "r5.12xlarge", "r5.16xlarge", "r5.24xlarge",
		"p3.2xlarge", "p3.8xlarge", "p3.16xlarge",
		"i3.large", "i3.xlarge", "i3.2xlarge", "i3.4xlarge", "i3.8xlarge", "i3.16xlarge",
	}

	for _, instanceType := range instanceTypes {
		err := s.fetchInstancePricing(ctx, instanceType, region)
		if err != nil {
			log.Printf("Error fetching pricing for %s: %v", instanceType, err)
			continue
		}
		time.Sleep(100 * time.Millisecond) // Rate limiting
	}

	log.Printf("Completed importing AWS pricing data")
	return nil
}

func (s *AWSPricingService) fetchInstancePricing(ctx context.Context, instanceType, region string) error {
	// Build filters for AWS Pricing API
	filters := []types.Filter{
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("ServiceCode"),
			Value: aws.String("AmazonEC2"),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("instanceType"),
			Value: aws.String(instanceType),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("location"),
			Value: aws.String(s.getLocationName(region)),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("tenancy"),
			Value: aws.String("Shared"),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("operating-system"),
			Value: aws.String("Linux"),
		},
		{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String("preInstalledSw"),
			Value: aws.String("NA"),
		},
	}

	input := &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters:     filters,
		MaxResults:  aws.Int32(10),
	}

	result, err := s.pricingClient.GetProducts(ctx, input)
	if err != nil {
		return err
	}

	for _, product := range result.PriceList {
		err := s.parseAndStorePricing(product, instanceType, region)
		if err != nil {
			log.Printf("Error parsing pricing for %s: %v", instanceType, err)
		}
	}

	return nil
}

func (s *AWSPricingService) parseAndStorePricing(productJSON string, instanceType, region string) error {
	var product map[string]interface{}
	if err := json.Unmarshal([]byte(productJSON), &product); err != nil {
		return err
	}

	terms, ok := product["terms"].(map[string]interface{})
	if !ok {
		return nil
	}

	var onDemandPrice, reservedPrice, spotPrice decimal.Decimal

	// Parse On-Demand pricing
	if onDemand, exists := terms["OnDemand"].(map[string]interface{}); exists {
		onDemandPrice = s.extractPrice(onDemand)
	}

	// Parse Reserved Instance pricing
	if reserved, exists := terms["Reserved"].(map[string]interface{}); exists {
		reservedPrice = s.extractPrice(reserved)
	}

	// Estimate spot price (typically 70% discount from on-demand)
	if !onDemandPrice.IsZero() {
		spotPrice = onDemandPrice.Mul(decimal.NewFromFloat(0.3))
	}

	// Store in database
	awsPricing := models.AWSPricing{
		InstanceType:          instanceType,
		Region:               region,
		OS:                   "Linux",
		PricePerHour:         onDemandPrice,
		RI1YrNoUpfront:       reservedPrice,
		RI1YrPartialUpfront:  reservedPrice.Mul(decimal.NewFromFloat(0.9)), // Estimate
		SpotPriceAvg:         spotPrice,
		UpdatedAt:            time.Now(),
	}

	// Upsert pricing data
	return s.db.Where("instance_type = ? AND region = ? AND os = ?", 
		instanceType, region, "Linux").
		Assign(&awsPricing).
		FirstOrCreate(&awsPricing).Error
}

func (s *AWSPricingService) extractPrice(terms map[string]interface{}) decimal.Decimal {
	for _, term := range terms {
		if termMap, ok := term.(map[string]interface{}); ok {
			if priceDimensions, exists := termMap["priceDimensions"].(map[string]interface{}); exists {
				for _, dimension := range priceDimensions {
					if dimMap, ok := dimension.(map[string]interface{}); ok {
						if pricePerUnit, exists := dimMap["pricePerUnit"].(map[string]interface{}); exists {
							if usdPrice, exists := pricePerUnit["USD"].(string); exists {
								if price, err := decimal.NewFromString(usdPrice); err == nil {
									return price
								}
							}
						}
					}
				}
			}
		}
	}
	return decimal.Zero
}

func (s *AWSPricingService) getLocationName(region string) string {
	locationMap := map[string]string{
		"us-east-1":      "US East (N. Virginia)",
		"us-east-2":      "US East (Ohio)",
		"us-west-1":      "US West (N. California)",
		"us-west-2":      "US West (Oregon)",
		"eu-west-1":      "Europe (Ireland)",
		"eu-central-1":   "Europe (Frankfurt)",
		"ap-southeast-1": "Asia Pacific (Singapore)",
		"ap-northeast-1": "Asia Pacific (Tokyo)",
	}

	if location, exists := locationMap[region]; exists {
		return location
	}
	return "US East (N. Virginia)" // Default
}

// Get current pricing for an instance type
func (s *AWSPricingService) GetInstancePricing(instanceType, region string) (*models.AWSPricing, error) {
	var pricing models.AWSPricing
	err := s.db.Where("instance_type = ? AND region = ?", instanceType, region).
		First(&pricing).Error
	
	if err == gorm.ErrRecordNotFound {
		// If not found, try to fetch from AWS
		ctx := context.Background()
		if fetchErr := s.fetchInstancePricing(ctx, instanceType, region); fetchErr == nil {
			// Try again after fetching
			err = s.db.Where("instance_type = ? AND region = ?", instanceType, region).
				First(&pricing).Error
		}
	}
	
	return &pricing, err
}

// Refresh all pricing data
func (s *AWSPricingService) RefreshAllPricing(ctx context.Context) error {
	regions := []string{"us-east-1", "us-west-2", "eu-west-1"}
	
	for _, region := range regions {
		log.Printf("Refreshing pricing for region: %s", region)
		if err := s.ImportEC2Pricing(ctx, region); err != nil {
			log.Printf("Error refreshing pricing for %s: %v", region, err)
		}
	}
	
	return nil
}