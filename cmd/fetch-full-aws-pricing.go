package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AWSPricing struct {
	ID                        uint            `gorm:"primaryKey"`
	InstanceType              string          `gorm:"not null"`
	Region                    string          `gorm:"not null"`
	OS                        string          `gorm:"default:Linux"`
	VCPU                      *int
	MemoryGB                  *decimal.Decimal
	Storage                   *string
	NetworkPerformance        *string
	OnDemandPriceUSD          decimal.Decimal `gorm:"not null"`
	Reserved1YrNoUpfront      *decimal.Decimal
	Reserved1YrPartialUpfront *decimal.Decimal
	Reserved3YrNoUpfront      *decimal.Decimal
	SpotPriceAvg              *decimal.Decimal
	LastUpdated               time.Time `gorm:"default:now()"`
	IsActive                  bool      `gorm:"default:true"`
}

func (AWSPricing) TableName() string {
	return "yt_aws_pricing"
}

func main() {
	fmt.Println("🚀 FETCHING FULL AWS EC2 PRICING DUMP")
	fmt.Println(strings.Repeat("=", 50))

	// Connect to database
	dsn := "host=localhost user=yukti password=yukti123 dbname=yukti_finops port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}

	// Check cache
	var count int64
	db.Model(&AWSPricing{}).Where("last_updated > ?", time.Now().Add(-24*time.Hour)).Count(&count)
	if count > 0 {
		fmt.Printf("✅ Cache valid: %d pricing records (< 24 hours old)\n", count)
		fmt.Println("⏭️  Skipping API call to save resources")
		return
	}

	fmt.Println("🔄 Cache expired or empty, fetching from AWS Pricing API...")

	// AWS Pricing client (must be us-east-1)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("❌ AWS config failed: %v", err)
	}

	client := pricing.NewFromConfig(cfg)

	// Get all EC2 instance types
	instanceTypes, err := getAllInstanceTypes(client)
	if err != nil {
		log.Fatalf("❌ Failed to get instance types: %v", err)
	}

	fmt.Printf("📋 Found %d instance types to fetch pricing for\n", len(instanceTypes))

	// Regions to fetch pricing for
	regions := []string{
		"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1",
	}

	totalFetched := 0
	
	// Clear old data
	db.Exec("UPDATE yt_aws_pricing SET is_active = false")

	for _, region := range regions {
		fmt.Printf("\n📍 Fetching pricing for region: %s\n", region)
		
		for i, instanceType := range instanceTypes {
			if i > 0 && i%10 == 0 {
				fmt.Printf("  📊 Progress: %d/%d instance types\n", i, len(instanceTypes))
			}

			pricing, err := fetchInstancePricing(client, instanceType, region)
			if err != nil {
				log.Printf("❌ %s in %s: %v", instanceType, region, err)
				continue
			}

			if pricing == nil {
				continue
			}

			// Save to database
			err = savePricing(db, pricing)
			if err != nil {
				log.Printf("❌ Save failed for %s: %v", instanceType, err)
				continue
			}

			totalFetched++
			
			// Rate limiting (AWS API limits)
			time.Sleep(200 * time.Millisecond)
		}
	}

	fmt.Printf("\n🎉 SUCCESS: Fetched %d pricing records from AWS!\n", totalFetched)
	showSummary(db)
}

func getAllInstanceTypes(client *pricing.Client) ([]string, error) {
	// Get all available instance types from AWS
	result, err := client.GetAttributeValues(context.TODO(), &pricing.GetAttributeValuesInput{
		ServiceCode:   aws.String("AmazonEC2"),
		AttributeName: aws.String("instanceType"),
		MaxResults:    aws.Int32(100),
	})
	if err != nil {
		return nil, err
	}

	var instanceTypes []string
	for _, attr := range result.AttributeValues {
		if attr.Value != nil {
			instanceTypes = append(instanceTypes, *attr.Value)
		}
	}

	return instanceTypes, nil
}

func fetchInstancePricing(client *pricing.Client, instanceType, region string) (*AWSPricing, error) {
	locationMap := map[string]string{
		"us-east-1":      "US East (N. Virginia)",
		"us-west-2":      "US West (Oregon)",
		"eu-west-1":      "Europe (Ireland)",
		"ap-southeast-1": "Asia Pacific (Singapore)",
	}

	location := locationMap[region]
	if location == "" {
		return nil, fmt.Errorf("unsupported region: %s", region)
	}

	filters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("instanceType"), Value: aws.String(instanceType)},
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String(location)},
		{Type: types.FilterTypeTermMatch, Field: aws.String("tenancy"), Value: aws.String("Shared")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("operating-system"), Value: aws.String("Linux")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("preInstalledSw"), Value: aws.String("NA")},
	}

	result, err := client.GetProducts(context.TODO(), &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters:     filters,
		MaxResults:  aws.Int32(1),
	})

	if err != nil {
		return nil, err
	}

	if len(result.PriceList) == 0 {
		return nil, nil
	}

	// Parse AWS response
	var product map[string]interface{}
	if err := json.Unmarshal([]byte(result.PriceList[0]), &product); err != nil {
		return nil, err
	}

	// Extract pricing and specs
	onDemandPrice, err := extractOnDemandPrice(product)
	if err != nil {
		return nil, err
	}

	specs := extractInstanceSpecs(product)

	// Calculate Reserved and Spot estimates
	reserved1yr := onDemandPrice.Mul(decimal.NewFromFloat(0.72))
	spotPrice := onDemandPrice.Mul(decimal.NewFromFloat(0.31))

	pricing := &AWSPricing{
		InstanceType:              instanceType,
		Region:                   region,
		OS:                       "Linux",
		OnDemandPriceUSD:         onDemandPrice,
		Reserved1YrNoUpfront:     &reserved1yr,
		Reserved1YrPartialUpfront: &reserved1yr,
		SpotPriceAvg:             &spotPrice,
		LastUpdated:              time.Now(),
		IsActive:                 true,
	}

	// Add specs if available
	if specs.VCPU != nil {
		pricing.VCPU = specs.VCPU
	}
	if specs.Memory != nil {
		pricing.MemoryGB = specs.Memory
	}
	if specs.Storage != nil {
		pricing.Storage = specs.Storage
	}
	if specs.Network != nil {
		pricing.NetworkPerformance = specs.Network
	}

	return pricing, nil
}

type InstanceSpecs struct {
	VCPU    *int
	Memory  *decimal.Decimal
	Storage *string
	Network *string
}

func extractInstanceSpecs(product map[string]interface{}) InstanceSpecs {
	specs := InstanceSpecs{}

	if productAttrs, ok := product["product"].(map[string]interface{}); ok {
		if attributes, ok := productAttrs["attributes"].(map[string]interface{}); ok {
			// Extract vCPU
			if vcpuStr, ok := attributes["vcpu"].(string); ok {
				if vcpu, err := strconv.Atoi(vcpuStr); err == nil {
					specs.VCPU = &vcpu
				}
			}

			// Extract memory
			if memoryStr, ok := attributes["memory"].(string); ok {
				// Parse "8 GiB" format
				memoryStr = strings.ReplaceAll(memoryStr, " GiB", "")
				memoryStr = strings.ReplaceAll(memoryStr, ",", "")
				if memory, err := decimal.NewFromString(memoryStr); err == nil {
					specs.Memory = &memory
				}
			}

			// Extract storage
			if storageStr, ok := attributes["storage"].(string); ok {
				specs.Storage = &storageStr
			}

			// Extract network performance
			if networkStr, ok := attributes["networkPerformance"].(string); ok {
				specs.Network = &networkStr
			}
		}
	}

	return specs
}

func extractOnDemandPrice(product map[string]interface{}) (decimal.Decimal, error) {
	terms, ok := product["terms"].(map[string]interface{})
	if !ok {
		return decimal.Zero, fmt.Errorf("no terms found")
	}

	onDemand, ok := terms["OnDemand"].(map[string]interface{})
	if !ok {
		return decimal.Zero, fmt.Errorf("no OnDemand pricing found")
	}

	for _, term := range onDemand {
		if termMap, ok := term.(map[string]interface{}); ok {
			if priceDimensions, exists := termMap["priceDimensions"].(map[string]interface{}); exists {
				for _, dimension := range priceDimensions {
					if dimMap, ok := dimension.(map[string]interface{}); ok {
						if pricePerUnit, exists := dimMap["pricePerUnit"].(map[string]interface{}); exists {
							if usdPrice, exists := pricePerUnit["USD"].(string); exists {
								price, err := decimal.NewFromString(usdPrice)
								if err == nil && !price.IsZero() {
									return price, nil
								}
							}
						}
					}
				}
			}
		}
	}

	return decimal.Zero, fmt.Errorf("USD price not found")
}

func savePricing(db *gorm.DB, pricing *AWSPricing) error {
	return db.Where("instance_type = ? AND region = ? AND os = ?",
		pricing.InstanceType, pricing.Region, pricing.OS).
		Assign(pricing).
		FirstOrCreate(pricing).Error
}

func showSummary(db *gorm.DB) {
	var total int64
	db.Model(&AWSPricing{}).Where("is_active = true").Count(&total)
	fmt.Printf("\n📊 Total active pricing records: %d\n", total)

	var pricings []AWSPricing
	db.Where("is_active = true").Order("on_demand_price_usd DESC").Limit(5).Find(&pricings)

	if len(pricings) > 0 {
		fmt.Println("\n💰 Most expensive instances:")
		fmt.Println("Instance Type    | Region    | On-Demand | Reserved | Spot")
		fmt.Println(strings.Repeat("-", 60))

		for _, p := range pricings {
			reserved := "N/A"
			spot := "N/A"
			if p.Reserved1YrNoUpfront != nil {
				reserved = "$" + p.Reserved1YrNoUpfront.StringFixed(4)
			}
			if p.SpotPriceAvg != nil {
				spot = "$" + p.SpotPriceAvg.StringFixed(4)
			}

			fmt.Printf("%-15s | %-9s | $%-8s | %-8s | %s\n",
				p.InstanceType,
				p.Region,
				p.OnDemandPriceUSD.StringFixed(4),
				reserved,
				spot)
		}
	}

	fmt.Println("\n✅ AWS pricing data cached for 24 hours!")
	fmt.Println("🔗 Table: yt_aws_pricing")
}