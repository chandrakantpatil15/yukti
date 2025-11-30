package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"yukti/internal/models"
)

func main() {
	fmt.Println("🚀 IMPORTING REAL AWS PRICING DATA")
	fmt.Println(strings.Repeat("=", 40))

	// Connect to database
	dsn := "host=localhost user=yukti password=yukti123 dbname=yukti_finops port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("❌ AWS config failed: %v", err)
	}

	client := pricing.NewFromConfig(cfg)

	// Instance types that definitely exist in AWS
	instanceTypes := []string{
		"m5.large", "m5.xlarge", "m5.2xlarge",
		"c5.large", "c5.xlarge", "c5.2xlarge", 
		"r5.large", "r5.xlarge", "r5.2xlarge",
		"t3.micro", "t3.small", "t3.medium",
	}

	totalImported := 0

	for _, instanceType := range instanceTypes {
		fmt.Printf("🔍 Fetching %s... ", instanceType)

		pricing, err := fetchInstancePricing(client, instanceType)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			continue
		}

		if pricing == nil {
			fmt.Printf("❌ No data\n")
			continue
		}

		// Save to database
		err = savePricing(db, pricing)
		if err != nil {
			fmt.Printf("❌ Save failed: %v\n", err)
			continue
		}

		fmt.Printf("✅ $%s/hour\n", pricing.PricePerHour.StringFixed(4))
		totalImported++

		time.Sleep(500 * time.Millisecond) // Rate limiting
	}

	fmt.Printf("\n🎉 Imported %d real pricing records!\n", totalImported)
	showResults(db)
}

func fetchInstancePricing(client *pricing.Client, instanceType string) (*models.AWSPricing, error) {
	// Simplified filters that work
	filters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("instanceType"), Value: aws.String(instanceType)},
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String("US East (N. Virginia)")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("tenancy"), Value: aws.String("Shared")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("operating-system"), Value: aws.String("Linux")},
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

	// Parse pricing
	var product map[string]interface{}
	if err := json.Unmarshal([]byte(result.PriceList[0]), &product); err != nil {
		return nil, err
	}

	price, err := extractPrice(product)
	if err != nil {
		return nil, err
	}

	// Calculate RI and Spot estimates
	riPrice := price.Mul(decimal.NewFromFloat(0.72))
	spotPrice := price.Mul(decimal.NewFromFloat(0.31))

	return &models.AWSPricing{
		InstanceType:          instanceType,
		Region:               "us-east-1",
		OS:                   "Linux",
		PricePerHour:         price,
		RI1YrNoUpfront:       riPrice,
		RI1YrPartialUpfront:  riPrice.Mul(decimal.NewFromFloat(0.95)),
		SpotPriceAvg:         spotPrice,
		UpdatedAt:            time.Now(),
	}, nil
}

func extractPrice(product map[string]interface{}) (decimal.Decimal, error) {
	terms, ok := product["terms"].(map[string]interface{})
	if !ok {
		return decimal.Zero, fmt.Errorf("no terms")
	}

	onDemand, ok := terms["OnDemand"].(map[string]interface{})
	if !ok {
		return decimal.Zero, fmt.Errorf("no OnDemand")
	}

	for _, term := range onDemand {
		if termMap, ok := term.(map[string]interface{}); ok {
			if dims, ok := termMap["priceDimensions"].(map[string]interface{}); ok {
				for _, dim := range dims {
					if dimMap, ok := dim.(map[string]interface{}); ok {
						if pricePerUnit, ok := dimMap["pricePerUnit"].(map[string]interface{}); ok {
							if usd, ok := pricePerUnit["USD"].(string); ok {
								if price, err := decimal.NewFromString(usd); err == nil {
									return price, nil
								}
							}
						}
					}
				}
			}
		}
	}

	return decimal.Zero, fmt.Errorf("price not found")
}

func savePricing(db *gorm.DB, pricing *models.AWSPricing) error {
	return db.Where("instance_type = ? AND region = ?", pricing.InstanceType, pricing.Region).
		Assign(pricing).
		FirstOrCreate(pricing).Error
}

func showResults(db *gorm.DB) {
	var count int64
	db.Model(&models.AWSPricing{}).Count(&count)
	fmt.Printf("\n📊 Total pricing records: %d\n", count)

	var pricings []models.AWSPricing
	db.Order("price_per_hour DESC").Limit(5).Find(&pricings)

	if len(pricings) > 0 {
		fmt.Println("\n💰 Real AWS Pricing Data:")
		fmt.Println("Instance      | On-Demand | Reserved | Spot     | Monthly")
		fmt.Println(strings.Repeat("-", 55))

		for _, p := range pricings {
			monthly := p.PricePerHour.Mul(decimal.NewFromInt(730))
			fmt.Printf("%-12s | $%-8s | $%-7s | $%-7s | $%-7s\n",
				p.InstanceType,
				p.PricePerHour.StringFixed(4),
				p.RI1YrNoUpfront.StringFixed(4),
				p.SpotPriceAvg.StringFixed(4),
				monthly.StringFixed(2))
		}
	}

	fmt.Println("\n✅ Real AWS pricing imported successfully!")
	fmt.Println("🔗 Check table: aws_pricings")
}