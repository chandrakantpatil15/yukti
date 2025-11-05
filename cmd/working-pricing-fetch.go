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
	OnDemandPriceUSD          decimal.Decimal `gorm:"not null"`
	Reserved1YrNoUpfront      *decimal.Decimal
	SpotPriceAvg              *decimal.Decimal
	LastUpdated               time.Time `gorm:"default:now()"`
	IsActive                  bool      `gorm:"default:true"`
}

func (AWSPricing) TableName() string {
	return "yt_aws_pricing"
}

func main() {
	fmt.Println("🚀 WORKING AWS PRICING FETCH")
	fmt.Println(strings.Repeat("=", 40))

	// Database connection
	dsn := "host=localhost user=yukti password=yukti123 dbname=yukti_finops port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Database failed: %v", err)
	}

	// AWS client
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("❌ AWS config failed: %v", err)
	}

	client := pricing.NewFromConfig(cfg)

	// Common instance types that definitely exist
	instanceTypes := []string{
		"t3.micro", "t3.small", "t3.medium", "t3.large",
		"m5.large", "m5.xlarge", "m5.2xlarge",
		"c5.large", "c5.xlarge", "c5.2xlarge",
		"r5.large", "r5.xlarge", "r5.2xlarge",
	}

	totalFetched := 0

	for _, instanceType := range instanceTypes {
		fmt.Printf("🔍 %s... ", instanceType)

		pricing, err := fetchPricing(client, instanceType)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			continue
		}

		if pricing == nil {
			fmt.Printf("❌ No data\n")
			continue
		}

		// Save to database
		err = db.Where("instance_type = ? AND region = ?", pricing.InstanceType, pricing.Region).
			Assign(pricing).
			FirstOrCreate(pricing).Error

		if err != nil {
			fmt.Printf("❌ Save failed: %v\n", err)
			continue
		}

		fmt.Printf("✅ $%s/hour\n", pricing.OnDemandPriceUSD.StringFixed(4))
		totalFetched++

		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("\n🎉 Fetched %d pricing records!\n", totalFetched)

	// Show results
	var pricings []AWSPricing
	db.Order("on_demand_price_usd DESC").Limit(5).Find(&pricings)

	if len(pricings) > 0 {
		fmt.Println("\n💰 Pricing Data:")
		for _, p := range pricings {
			fmt.Printf("%s: $%s/hour\n", p.InstanceType, p.OnDemandPriceUSD.StringFixed(4))
		}
	}
}

func fetchPricing(client *pricing.Client, instanceType string) (*AWSPricing, error) {
	// Simplified filters
	filters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("instanceType"), Value: aws.String(instanceType)},
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String("US East (N. Virginia)")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("tenancy"), Value: aws.String("Shared")},
	}

	result, err := client.GetProducts(context.TODO(), &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters:     filters,
		MaxResults:  aws.Int32(5), // Get more results to find Linux
	})

	if err != nil {
		return nil, err
	}

	if len(result.PriceList) == 0 {
		return nil, nil
	}

	// Find Linux pricing
	for _, productJSON := range result.PriceList {
		var product map[string]interface{}
		if err := json.Unmarshal([]byte(productJSON), &product); err != nil {
			continue
		}

		// Check if this is Linux
		if productAttrs, ok := product["product"].(map[string]interface{}); ok {
			if attributes, ok := productAttrs["attributes"].(map[string]interface{}); ok {
				if os, ok := attributes["operatingSystem"].(string); ok && os != "Linux" {
					continue // Skip non-Linux
				}
			}
		}

		// Extract price
		price, err := extractPrice(product)
		if err != nil {
			continue
		}

		// Extract specs
		vcpu, memory := extractSpecs(product)

		// Calculate estimates
		reserved := price.Mul(decimal.NewFromFloat(0.72))
		spot := price.Mul(decimal.NewFromFloat(0.31))

		return &AWSPricing{
			InstanceType:         instanceType,
			Region:              "us-east-1",
			OS:                  "Linux",
			VCPU:                vcpu,
			MemoryGB:            memory,
			OnDemandPriceUSD:    price,
			Reserved1YrNoUpfront: &reserved,
			SpotPriceAvg:        &spot,
			LastUpdated:         time.Now(),
			IsActive:            true,
		}, nil
	}

	return nil, fmt.Errorf("no Linux pricing found")
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
								return decimal.NewFromString(usd)
							}
						}
					}
				}
			}
		}
	}

	return decimal.Zero, fmt.Errorf("price not found")
}

func extractSpecs(product map[string]interface{}) (*int, *decimal.Decimal) {
	if productAttrs, ok := product["product"].(map[string]interface{}); ok {
		if attributes, ok := productAttrs["attributes"].(map[string]interface{}); ok {
			var vcpu *int
			var memory *decimal.Decimal

			if vcpuStr, ok := attributes["vcpu"].(string); ok {
				if v, err := strconv.Atoi(vcpuStr); err == nil {
					vcpu = &v
				}
			}

			if memoryStr, ok := attributes["memory"].(string); ok {
				memoryStr = strings.ReplaceAll(memoryStr, " GiB", "")
				memoryStr = strings.ReplaceAll(memoryStr, ",", "")
				if m, err := decimal.NewFromString(memoryStr); err == nil {
					memory = &m
				}
			}

			return vcpu, memory
		}
	}
	return nil, nil
}