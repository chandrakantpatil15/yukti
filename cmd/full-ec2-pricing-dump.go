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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🚀 FULL EC2 PRICING DUMP FROM AWS")
	fmt.Println("This will fetch ALL EC2 instance types pricing")
	fmt.Println(strings.Repeat("=", 50))

	// Database connection
	dsn := "host=localhost user=yukti password=yukti123 dbname=yukti_finops port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Database failed: %v", err)
	}

	// Check 24-hour cache
	var cacheCount int64
	db.Raw("SELECT COUNT(*) FROM yt_aws_pricing WHERE last_updated > NOW() - INTERVAL '24 hours' AND is_active = true").Scan(&cacheCount)
	
	if cacheCount > 100 { // If we have substantial cached data
		fmt.Printf("✅ Cache valid: %d pricing records (< 24 hours old)\n", cacheCount)
		fmt.Println("⏭️  Skipping AWS API calls to save resources")
		showCurrentData(db)
		return
	}

	fmt.Println("🔄 Cache expired or empty, fetching FULL DUMP from AWS...")

	// AWS Pricing client (must be us-east-1)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("❌ AWS config failed: %v", err)
	}

	client := pricing.NewFromConfig(cfg)

	// Step 1: Get ALL EC2 instance types from AWS
	fmt.Println("📋 Step 1: Getting ALL EC2 instance types from AWS...")
	allInstanceTypes, err := getAllEC2InstanceTypes(client)
	if err != nil {
		log.Fatalf("❌ Failed to get instance types: %v", err)
	}

	fmt.Printf("✅ Found %d total EC2 instance types\n", len(allInstanceTypes))

	// Step 2: Fetch pricing for all regions
	regions := []string{
		"us-east-1", "us-west-2", "us-west-1", "us-east-2",
		"eu-west-1", "eu-central-1", "eu-west-2", "eu-west-3",
		"ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ap-northeast-2",
		"ap-south-1", "ca-central-1", "sa-east-1",
	}

	totalFetched := 0
	totalAttempts := 0

	// Clear old data
	fmt.Println("🗑️  Clearing old pricing data...")
	db.Exec("UPDATE yt_aws_pricing SET is_active = false WHERE last_updated < NOW() - INTERVAL '24 hours'")

	for _, region := range regions {
		fmt.Printf("\n📍 Region: %s\n", region)
		
		for i, instanceType := range allInstanceTypes {
			totalAttempts++
			
			if i > 0 && i%50 == 0 {
				fmt.Printf("  📊 Progress: %d/%d instance types (%d successful)\n", i, len(allInstanceTypes), totalFetched)
			}

			pricing, err := fetchInstancePricing(client, instanceType, region)
			if err != nil {
				if i < 10 { // Only show first few errors to avoid spam
					fmt.Printf("  ❌ %s: %v\n", instanceType, err)
				}
				continue
			}

			if pricing == nil {
				continue
			}

			// Insert into database
			err = insertPricing(db, pricing)
			if err != nil {
				if i < 10 {
					fmt.Printf("  ❌ Save %s: %v\n", instanceType, err)
				}
				continue
			}

			totalFetched++
			
			// Rate limiting - AWS has strict limits
			time.Sleep(100 * time.Millisecond)
		}
		
		fmt.Printf("  ✅ %s completed\n", region)
	}

	fmt.Printf("\n🎉 FULL DUMP COMPLETE!\n")
	fmt.Printf("📊 Total attempts: %d\n", totalAttempts)
	fmt.Printf("✅ Successfully fetched: %d pricing records\n", totalFetched)
	fmt.Printf("📈 Success rate: %.1f%%\n", float64(totalFetched)/float64(totalAttempts)*100)

	showCurrentData(db)
}

func getAllEC2InstanceTypes(client *pricing.Client) ([]string, error) {
	var allTypes []string
	var nextToken *string

	for {
		input := &pricing.GetAttributeValuesInput{
			ServiceCode:   aws.String("AmazonEC2"),
			AttributeName: aws.String("instanceType"),
			MaxResults:    aws.Int32(100),
		}
		
		if nextToken != nil {
			input.NextToken = nextToken
		}

		result, err := client.GetAttributeValues(context.TODO(), input)
		if err != nil {
			return nil, err
		}

		for _, attr := range result.AttributeValues {
			if attr.Value != nil {
				allTypes = append(allTypes, *attr.Value)
			}
		}

		nextToken = result.NextToken
		if nextToken == nil {
			break
		}

		time.Sleep(200 * time.Millisecond) // Rate limiting
	}

	return allTypes, nil
}

func fetchInstancePricing(client *pricing.Client, instanceType, region string) (*PricingData, error) {
	// Map region codes to AWS location names
	locationMap := map[string]string{
		"us-east-1":      "US East (N. Virginia)",
		"us-west-2":      "US West (Oregon)",
		"us-west-1":      "US West (N. California)",
		"us-east-2":      "US East (Ohio)",
		"eu-west-1":      "Europe (Ireland)",
		"eu-central-1":   "Europe (Frankfurt)",
		"eu-west-2":      "Europe (London)",
		"eu-west-3":      "Europe (Paris)",
		"ap-southeast-1": "Asia Pacific (Singapore)",
		"ap-southeast-2": "Asia Pacific (Sydney)",
		"ap-northeast-1": "Asia Pacific (Tokyo)",
		"ap-northeast-2": "Asia Pacific (Seoul)",
		"ap-south-1":     "Asia Pacific (Mumbai)",
		"ca-central-1":   "Canada (Central)",
		"sa-east-1":      "South America (Sao Paulo)",
	}

	location := locationMap[region]
	if location == "" {
		return nil, fmt.Errorf("unsupported region: %s", region)
	}

	filters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("instanceType"), Value: aws.String(instanceType)},
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String(location)},
		{Type: types.FilterTypeTermMatch, Field: aws.String("tenancy"), Value: aws.String("Shared")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("operatingSystem"), Value: aws.String("Linux")},
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

	// Extract pricing
	price, err := extractOnDemandPrice(product)
	if err != nil {
		return nil, err
	}

	// Extract specs
	vcpu, memory, storage, network := extractFullSpecs(product)

	return &PricingData{
		InstanceType:    instanceType,
		Region:         region,
		OS:             "Linux",
		VCPU:           vcpu,
		MemoryGB:       memory,
		Storage:        storage,
		Network:        network,
		OnDemandPrice:  price,
		ReservedPrice:  price * 0.72, // Typical 28% RI discount
		SpotPrice:      price * 0.31, // Typical 69% spot discount
	}, nil
}

type PricingData struct {
	InstanceType   string
	Region         string
	OS             string
	VCPU           *int
	MemoryGB       *float64
	Storage        *string
	Network        *string
	OnDemandPrice  float64
	ReservedPrice  float64
	SpotPrice      float64
}

func extractOnDemandPrice(product map[string]interface{}) (float64, error) {
	terms, ok := product["terms"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("no terms")
	}

	onDemand, ok := terms["OnDemand"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("no OnDemand")
	}

	for _, term := range onDemand {
		if termMap, ok := term.(map[string]interface{}); ok {
			if dims, ok := termMap["priceDimensions"].(map[string]interface{}); ok {
				for _, dim := range dims {
					if dimMap, ok := dim.(map[string]interface{}); ok {
						if pricePerUnit, ok := dimMap["pricePerUnit"].(map[string]interface{}); ok {
							if usd, ok := pricePerUnit["USD"].(string); ok {
								return strconv.ParseFloat(usd, 64)
							}
						}
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("price not found")
}

func extractFullSpecs(product map[string]interface{}) (*int, *float64, *string, *string) {
	if productAttrs, ok := product["product"].(map[string]interface{}); ok {
		if attributes, ok := productAttrs["attributes"].(map[string]interface{}); ok {
			var vcpu *int
			var memory *float64
			var storage *string
			var network *string

			// vCPU
			if vcpuStr, ok := attributes["vcpu"].(string); ok {
				if v, err := strconv.Atoi(vcpuStr); err == nil {
					vcpu = &v
				}
			}

			// Memory
			if memoryStr, ok := attributes["memory"].(string); ok {
				memoryStr = strings.ReplaceAll(memoryStr, " GiB", "")
				memoryStr = strings.ReplaceAll(memoryStr, ",", "")
				if m, err := strconv.ParseFloat(memoryStr, 64); err == nil {
					memory = &m
				}
			}

			// Storage
			if storageStr, ok := attributes["storage"].(string); ok {
				storage = &storageStr
			}

			// Network
			if networkStr, ok := attributes["networkPerformance"].(string); ok {
				network = &networkStr
			}

			return vcpu, memory, storage, network
		}
	}
	return nil, nil, nil, nil
}

func insertPricing(db *gorm.DB, pricing *PricingData) error {
	sql := `
	INSERT INTO yt_aws_pricing 
	(instance_type, region, os, vcpu, memory_gb, storage, network_performance, 
	 on_demand_price_usd, reserved_1yr_no_upfront, spot_price_avg, last_updated, is_active)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), true)
	ON CONFLICT (instance_type, region, os) 
	DO UPDATE SET 
		vcpu = EXCLUDED.vcpu,
		memory_gb = EXCLUDED.memory_gb,
		storage = EXCLUDED.storage,
		network_performance = EXCLUDED.network_performance,
		on_demand_price_usd = EXCLUDED.on_demand_price_usd,
		reserved_1yr_no_upfront = EXCLUDED.reserved_1yr_no_upfront,
		spot_price_avg = EXCLUDED.spot_price_avg,
		last_updated = NOW(),
		is_active = true
	`

	return db.Exec(sql,
		pricing.InstanceType,
		pricing.Region,
		pricing.OS,
		pricing.VCPU,
		pricing.MemoryGB,
		pricing.Storage,
		pricing.Network,
		pricing.OnDemandPrice,
		pricing.ReservedPrice,
		pricing.SpotPrice,
	).Error
}

func showCurrentData(db *gorm.DB) {
	var total int64
	db.Raw("SELECT COUNT(*) FROM yt_aws_pricing WHERE is_active = true").Scan(&total)
	fmt.Printf("\n📊 Total active pricing records: %d\n", total)

	var regionCount []map[string]interface{}
	db.Raw("SELECT region, COUNT(*) as count FROM yt_aws_pricing WHERE is_active = true GROUP BY region ORDER BY count DESC").Scan(&regionCount)

	fmt.Println("\n🌍 Pricing by Region:")
	for _, r := range regionCount {
		fmt.Printf("  %s: %v instances\n", r["region"], r["count"])
	}

	var expensive []map[string]interface{}
	db.Raw("SELECT instance_type, region, on_demand_price_usd, vcpu, memory_gb FROM yt_aws_pricing WHERE is_active = true ORDER BY on_demand_price_usd DESC LIMIT 10").Scan(&expensive)

	if len(expensive) > 0 {
		fmt.Println("\n💰 Most Expensive Instances:")
		fmt.Println("Instance Type    | Region    | Price/hour | vCPU | Memory")
		fmt.Println(strings.Repeat("-", 60))
		for _, e := range expensive {
			fmt.Printf("%-15s | %-9s | $%-9.4f | %-4v | %vGB\n",
				e["instance_type"],
				e["region"],
				e["on_demand_price_usd"],
				e["vcpu"],
				e["memory_gb"])
		}
	}

	fmt.Println("\n✅ Full EC2 pricing dump cached for 24 hours!")
	fmt.Println("🔗 Table: yt_aws_pricing")
	fmt.Println("⏰ Next refresh: 24 hours from now")
}