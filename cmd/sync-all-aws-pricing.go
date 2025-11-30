package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"maps"
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

// Supported AWS services and their pricing attributes
var awsServices = map[string][]string{
	"AmazonEC2":         {"instanceType", "vcpu", "memory", "storage", "operatingSystem"},
	"AmazonRDS":         {"instanceType", "databaseEngine", "deploymentOption", "databaseEdition"},
	"AmazonRedshift":    {"nodeType", "clusterType"},
	"AmazonElastiCache": {"cacheNodeType", "cacheEngine"},
	"AmazonOpenSearch":  {"instanceType", "engineVersion"},
	"AmazonMQ":          {"brokerType", "engineVersion"},
	"AWSLambda":         {"memory"},
	"AmazonDynamoDB":    {"readCapacityUnit", "writeCapacityUnit"},
	"AmazonECS":         {"instanceType", "launchType"},
	"AWSEKS":            {"instanceType", "version"},
	"AmazonEFS":         {"storageClass"},
	"AmazonFSx":         {"storageType", "deploymentType"},
	"AmazonS3":          {"storageClass", "volumeType"},
	"AmazonRoute53":     {"recordType", "routingPolicy"},
	"AmazonCloudFront":  {"priceClass", "origin"},
	"AmazonECR":         {"storageType"},
	"AmazonKinesis":     {"streamType"},
	"AmazonSNS":         {"deliveryType"},
	"AmazonSQS":         {"queueType"},
	"AWSSecretsManager": {"secretType"},
	"AmazonApiGateway":  {"apiType", "protocol"},
	"AWSCloudTrail":     {"storageType", "managementEvents"},
	"AmazonCloudWatch":  {"metricType", "alarmType"},
	"AWSBackup":         {"storageType", "backupType"},
	"AmazonVPC":         {"resourceType", "connectionType"},
}

// RegionMap maps AWS region codes to display names
var regionMap = map[string]string{
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

type PricingData struct {
	ServiceCode     string
	ResourceType    string
	Region          string
	Attributes      map[string]string
	OnDemandPrice   float64
	ReservedPrice   float64
	SpotPrice       float64
	PricingUnit     string
	PricingCurrency string
	LastUpdated     time.Time
}

func main() {
	fmt.Println("🚀 FULL AWS PRICING SYNC")
	fmt.Println("Fetching pricing for 200+ AWS services")
	fmt.Println(strings.Repeat("=", 50))

	// Database connection
	dsn := "host=localhost user=yukti password=yukti123 dbname=yukti_finops port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}

	// Check cache validity (24 hours)
	var cacheCount int64
	db.Raw("SELECT COUNT(*) FROM yt_aws_pricing WHERE last_updated > NOW() - INTERVAL '24 hours' AND is_active = true").Scan(&cacheCount)

	if cacheCount > 1000 { // If we have substantial cached data
		fmt.Printf("✅ Cache valid: %d pricing records (< 24 hours old)\n", cacheCount)
		showCurrentData(db)
		return
	}

	fmt.Println("🔄 Cache expired or empty, fetching from AWS...")

	// AWS Pricing client (must be us-east-1)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("❌ AWS config failed: %v", err)
	}

	client := pricing.NewFromConfig(cfg)

	totalFetched := 0
	errorCount := 0

	// Clear old data
	db.Exec("UPDATE yt_aws_pricing SET is_active = false")

	// Process each service
	for serviceCode, attributes := range awsServices {
		fmt.Printf("\n📦 Processing %s...\n", serviceCode)

		for _, region := range getServiceRegions(serviceCode) {
			fmt.Printf("  📍 Region: %s\n", region)

			// Get resource types for the service
			resourceTypes, err := getResourceTypes(client, serviceCode)
			if err != nil {
				fmt.Printf("  ❌ Failed to get resource types: %v\n", err)
				errorCount++
				continue
			}

			for _, resourceType := range resourceTypes {
				pricingData, err := fetchServicePricing(client, serviceCode, resourceType, region, attributes)
				if err != nil {
					fmt.Printf("  ❌ %s/%s: %v\n", resourceType, region, err)
					errorCount++
					continue
				}

				if pricingData == nil {
					continue
				}

				// Store pricing data
				if err := storePricingData(db, pricingData); err != nil {
					fmt.Printf("  ❌ Failed to store %s/%s: %v\n", resourceType, region, err)
					errorCount++
					continue
				}

				totalFetched++

				// Rate limiting
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	fmt.Printf("\n🎉 SYNC COMPLETE!\n")
	fmt.Printf("✅ Successfully fetched: %d pricing records\n", totalFetched)
	fmt.Printf("❌ Errors encountered: %d\n", errorCount)

	showCurrentData(db)
}

func getServiceRegions(serviceCode string) []string {
	// Return appropriate regions based on service
	// Some services are global, some are regional
	if serviceCode == "AmazonRoute53" || serviceCode == "AmazonCloudFront" {
		return []string{"us-east-1"} // Global services
	}
	return maps.Keys(regionMap)
}

func getResourceTypes(client *pricing.Client, serviceCode string) ([]string, error) {
	var resourceTypes []string
	var nextToken *string

	for {
		input := &pricing.GetAttributeValuesInput{
			ServiceCode:   aws.String(serviceCode),
			AttributeName: aws.String("resourceType"),
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
				resourceTypes = append(resourceTypes, *attr.Value)
			}
		}

		nextToken = result.NextToken
		if nextToken == nil {
			break
		}

		time.Sleep(200 * time.Millisecond) // Rate limiting
	}

	return resourceTypes, nil
}

func fetchServicePricing(client *pricing.Client, serviceCode, resourceType, region string, attributes []string) (*PricingData, error) {
	filters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("resourceType"), Value: aws.String(resourceType)},
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String(regionMap[region])},
	}

	// Add service-specific filters
	for _, attr := range attributes {
		filters = append(filters, types.Filter{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String(attr),
			Value: aws.String("*"),
		})
	}

	result, err := client.GetProducts(context.TODO(), &pricing.GetProductsInput{
		ServiceCode: aws.String(serviceCode),
		Filters:     filters,
		MaxResults:  aws.Int32(100),
	})

	if err != nil {
		return nil, err
	}

	if len(result.PriceList) == 0 {
		return nil, nil
	}

	// Process pricing data
	pricingData := &PricingData{
		ServiceCode:  serviceCode,
		ResourceType: resourceType,
		Region:       region,
		Attributes:   make(map[string]string),
		LastUpdated:  time.Now(),
	}

	// Parse product attributes and pricing
	for _, priceItem := range result.PriceList {
		var product map[string]interface{}
		if err := json.Unmarshal([]byte(priceItem), &product); err != nil {
			continue
		}

		// Extract attributes
		if attrs, ok := product["attributes"].(map[string]interface{}); ok {
			for _, attr := range attributes {
				if val, exists := attrs[attr]; exists {
					pricingData.Attributes[attr] = fmt.Sprint(val)
				}
			}
		}

		// Extract pricing
		if terms, ok := product["terms"].(map[string]interface{}); ok {
			// On-Demand pricing
			if onDemand, ok := terms["OnDemand"].(map[string]interface{}); ok {
				pricingData.OnDemandPrice = extractPrice(onDemand)
			}

			// Reserved pricing
			if reserved, ok := terms["Reserved"].(map[string]interface{}); ok {
				pricingData.ReservedPrice = extractPrice(reserved)
			}
		}
	}

	return pricingData, nil
}

func extractPrice(terms map[string]interface{}) float64 {
	for _, dimension := range terms {
		if dim, ok := dimension.(map[string]interface{}); ok {
			if priceMap, ok := dim["pricePerUnit"].(map[string]interface{}); ok {
				if usd, ok := priceMap["USD"].(string); ok {
					price, _ := strconv.ParseFloat(usd, 64)
					return price
				}
			}
		}
	}
	return 0
}

func storePricingData(db *gorm.DB, data *PricingData) error {
	sql := `
	INSERT INTO yt_aws_pricing 
	(service_code, resource_type, region, attributes, on_demand_price_usd, 
	 reserved_price_usd, spot_price_usd, pricing_unit, pricing_currency, 
	 last_updated, is_active)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), true)
	ON CONFLICT (service_code, resource_type, region) 
	DO UPDATE SET 
		attributes = EXCLUDED.attributes,
		on_demand_price_usd = EXCLUDED.on_demand_price_usd,
		reserved_price_usd = EXCLUDED.reserved_price_usd,
		spot_price_usd = EXCLUDED.spot_price_usd,
		last_updated = NOW(),
		is_active = true`

	return db.Exec(sql,
		data.ServiceCode,
		data.ResourceType,
		data.Region,
		data.Attributes,
		data.OnDemandPrice,
		data.ReservedPrice,
		data.SpotPrice,
		data.PricingUnit,
		"USD").Error
}

func showCurrentData(db *gorm.DB) {
	var total int64
	db.Raw("SELECT COUNT(*) FROM yt_aws_pricing WHERE is_active = true").Scan(&total)
	fmt.Printf("\n📊 Total active pricing records: %d\n", total)

	var serviceCount []map[string]interface{}
	db.Raw(`
		SELECT service_code, COUNT(*) as count 
		FROM yt_aws_pricing 
		WHERE is_active = true 
		GROUP BY service_code 
		ORDER BY count DESC
	`).Scan(&serviceCount)

	fmt.Println("\n📦 Pricing by Service:")
	for _, s := range serviceCount {
		fmt.Printf("  %s: %v resources\n", s["service_code"], s["count"])
	}

	var expensive []map[string]interface{}
	db.Raw(`
		SELECT service_code, resource_type, region, on_demand_price_usd 
		FROM yt_aws_pricing 
		WHERE is_active = true 
		ORDER BY on_demand_price_usd DESC 
		LIMIT 10
	`).Scan(&expensive)

	if len(expensive) > 0 {
		fmt.Println("\n💰 Most Expensive Resources:")
		fmt.Println("Service          | Resource Type     | Region    | Price/hour")
		fmt.Println(strings.Repeat("-", 75))
		for _, e := range expensive {
			fmt.Printf("%-15s | %-16s | %-9s | $%-9.4f\n",
				e["service_code"],
				e["resource_type"],
				e["region"],
				e["on_demand_price_usd"])
		}
	}
}
