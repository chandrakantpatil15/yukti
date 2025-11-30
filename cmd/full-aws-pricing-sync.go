package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SyncStatus represents the current state of pricing sync
type SyncStatus struct {
	ID               uint      `gorm:"primarykey"`
	SyncStartTime    time.Time `gorm:"not null"`
	SyncEndTime      *time.Time
	Status           string    `gorm:"not null"` // running, completed, failed
	TotalRecords     int       `gorm:"default:0"`
	ProcessedRecords int       `gorm:"default:0"`
	ErrorCount       int       `gorm:"default:0"`
	LastService      string    `gorm:"size:100"`
	LastRegion       string    `gorm:"size:50"`
	ErrorMessage     string    `gorm:"type:text"`
	CreatedAt        time.Time `gorm:"default:now()"`
}

func (SyncStatus) TableName() string {
	return "yt_aws_pricing_sync_status"
}

// FullAWSPricingData represents a pricing record in the database
type FullAWSPricingData struct {
	gorm.Model
	ServiceCode      string          `gorm:"not null"`
	ResourceType     string          `gorm:"not null"`
	Region           string          `gorm:"not null"`
	OS               string          `gorm:"default:Linux"`
	Attributes       json.RawMessage `gorm:"type:jsonb"`
	OnDemandPriceUSD float64         `gorm:"not null"`
	ReservedPriceUSD float64
	SpotPriceUSD     float64
	PricingUnit      string    `gorm:"default:Hrs"`
	PricingCurrency  string    `gorm:"default:USD"`
	LastUpdated      time.Time `gorm:"default:now()"`
	IsActive         bool      `gorm:"default:true"`
	BatchID          string    `gorm:"index"` // For tracking batch updates
}

func (FullAWSPricingData) TableName() string {
	return "yt_aws_pricing"
}

// ServiceConfig represents configuration for a single AWS service
type ServiceConfig struct {
	Attributes []string
	Regions    []string
}

// AWS service categories and their attributes
var fullSyncServiceMap = map[string]ServiceConfig{
	"AmazonEC2": {
		Attributes: []string{"instanceType", "vcpu", "memory", "storage", "operatingSystem"},
		Regions:    nil, // nil means all regions
	},
	"AmazonRDS": {
		Attributes: []string{"instanceType", "databaseEngine", "deploymentOption", "databaseEdition"},
		Regions:    nil,
	},
	"AmazonRedshift": {
		Attributes: []string{"nodeType", "clusterType"},
		Regions:    nil,
	},
	"AmazonElastiCache": {
		Attributes: []string{"cacheNodeType", "cacheEngine"},
		Regions:    nil,
	},
	"AmazonOpenSearch": {
		Attributes: []string{"instanceType", "engineVersion"},
		Regions:    nil,
	},
	"AmazonMQ": {
		Attributes: []string{"brokerType", "engineVersion"},
		Regions:    nil,
	},
	"AWSLambda": {
		Attributes: []string{"memory"},
		Regions:    nil,
	},
	"AmazonDynamoDB": {
		Attributes: []string{"readCapacityUnit", "writeCapacityUnit"},
		Regions:    nil,
	},
	"AmazonECS": {
		Attributes: []string{"launchType"},
		Regions:    nil,
	},
	"AWSEKS": {
		Attributes: []string{"version"},
		Regions:    nil,
	},
	// Global services
	"AmazonRoute53": {
		Attributes: []string{"recordType", "routingPolicy"},
		Regions:    []string{"us-east-1"},
	},
	"AmazonCloudFront": {
		Attributes: []string{"priceClass", "origin"},
		Regions:    []string{"us-east-1"},
	},
}

// RegionDisplayMap maps AWS region codes to display names
var fullSyncRegionMap = map[string]string{
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

// Configuration constants
const (
	batchSize         = 1000 // Number of records to process in a batch
	maxConcurrentJobs = 5    // Maximum number of concurrent service fetches
	rateLimitDelay    = 100 * time.Millisecond
	syncInterval      = 24 * time.Hour
	maxRetries        = 3
)

// RunFullPricingSync performs a comprehensive pricing sync for all AWS services
func RunFullPricingSync(ctx context.Context) error {
	for {
		if err := performSingleSync(ctx); err != nil {
			log.Printf("❌ Sync error: %v", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(syncInterval):
			continue
		}
	}
}

func performSingleSync(ctx context.Context) error {
	dsn := "host=localhost user=yukti password=yukti123 dbname=yukti_finops port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("database connection failed: %v", err)
	}

	// Initialize sync status
	syncStatus := &SyncStatus{
		SyncStartTime: time.Now(),
		Status:        "running",
	}
	if err := db.Create(syncStatus).Error; err != nil {
		return fmt.Errorf("failed to create sync status: %v", err)
	}

	// Check cache validity
	var cacheCount int64
	db.Model(&FullAWSPricingData{}).Where("last_updated > ?",
		time.Now().Add(-syncInterval)).Count(&cacheCount)

	if cacheCount > 1000 {
		log.Printf("✅ Cache valid: %d pricing records (< 24 hours old)", cacheCount)
		return nil
	}

	// AWS Pricing client setup
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		updateSyncStatus(db, syncStatus, "failed", fmt.Sprintf("AWS config failed: %v", err))
		return err
	}

	client := pricing.NewFromConfig(cfg)
	batchID := time.Now().Format("20060102150405")

	// Process services in parallel with limited concurrency
	semaphore := make(chan struct{}, maxConcurrentJobs)
	var wg sync.WaitGroup
	errChan := make(chan error, len(fullSyncServiceMap))

	for serviceCode, service := range fullSyncServiceMap {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(svc string, cfg ServiceConfig) {
			defer func() {
				<-semaphore // Release semaphore
				wg.Done()
			}()

			if err := processService(ctx, client, db, svc, cfg, batchID, syncStatus); err != nil {
				errChan <- fmt.Errorf("service %s failed: %v", svc, err)
			}
		}(serviceCode, service)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Collect any errors
	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		err := fmt.Errorf("sync completed with errors: %s", strings.Join(errors, "; "))
		updateSyncStatus(db, syncStatus, "failed", err.Error())
		return err
	}

	// Mark old records as inactive
	if err := db.Model(&FullAWSPricingData{}).
		Where("batch_id != ?", batchID).
		Update("is_active", false).Error; err != nil {
		log.Printf("⚠️ Failed to mark old records as inactive: %v", err)
	}

	// Update sync status
	now := time.Now()
	syncStatus.SyncEndTime = &now
	syncStatus.Status = "completed"
	db.Save(syncStatus)

	log.Printf("🎉 Sync completed successfully. BatchID: %s", batchID)
	return nil
}

func processService(ctx context.Context, client *pricing.Client, db *gorm.DB,
	serviceCode string, config ServiceConfig, batchID string, status *SyncStatus) error {

	regions := config.Regions
	if regions == nil {
		regions = getFullSyncRegionList()
	}

	for _, region := range regions {
		status.LastService = serviceCode
		status.LastRegion = region
		db.Save(status)

		resourceTypes, err := getFullSyncResourceTypes(ctx, client, serviceCode)
		if err != nil {
			status.ErrorCount++
			db.Save(status)
			return fmt.Errorf("failed to get resource types: %v", err)
		}

		var batch []*FullAWSPricingData
		for _, resourceType := range resourceTypes {
			pricing, err := fetchFullSyncServicePricing(ctx, client, serviceCode, resourceType, region, config.Attributes)
			if err != nil {
				status.ErrorCount++
				db.Save(status)
				continue
			}

			if pricing != nil {
				pricing.BatchID = batchID
				batch = append(batch, pricing)

				if len(batch) >= batchSize {
					if err := saveBatch(db, batch); err != nil {
						status.ErrorCount++
						db.Save(status)
						return err
					}
					status.ProcessedRecords += len(batch)
					db.Save(status)
					batch = nil
				}
			}

			time.Sleep(rateLimitDelay)
		}

		// Save any remaining records
		if len(batch) > 0 {
			if err := saveBatch(db, batch); err != nil {
				status.ErrorCount++
				db.Save(status)
				return err
			}
			status.ProcessedRecords += len(batch)
			db.Save(status)
		}
	}

	return nil
}

func saveBatch(db *gorm.DB, batch []*FullAWSPricingData) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, record := range batch {
			if err := tx.Create(record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func updateSyncStatus(db *gorm.DB, status *SyncStatus, newStatus string, errorMsg string) {
	if newStatus == "failed" || newStatus == "completed" {
		now := time.Now()
		status.SyncEndTime = &now
	}
	status.Status = newStatus
	status.ErrorMessage = errorMsg
	db.Save(status)
}

// GetLatestSyncStatus returns the status of the most recent sync operation
func GetLatestSyncStatus(db *gorm.DB) (*SyncStatus, error) {
	var status SyncStatus
	err := db.Order("sync_start_time DESC").First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Get total record counts by service
func GetSyncServiceCounts(db *gorm.DB) (map[string]int64, error) {
	type ServiceCount struct {
		ServiceCode string
		Count       int64
	}
	var counts []ServiceCount
	err := db.Model(&FullAWSPricingData{}).
		Where("is_active = true").
		Select("service_code, count(*) as count").
		Group("service_code").
		Find(&counts).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, count := range counts {
		result[count.ServiceCode] = count.Count
	}
	return result, nil
}

// getFullSyncRegionList returns the list of AWS regions
func getFullSyncRegionList() []string {
	regions := make([]string, 0, len(fullSyncRegionMap))
	for region := range fullSyncRegionMap {
		regions = append(regions, region)
	}
	return regions
}

// getFullSyncResourceTypes fetches resource types for a service with retries
func getFullSyncResourceTypes(ctx context.Context, client *pricing.Client, serviceCode string) ([]string, error) {
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

		result, err := client.GetAttributeValues(ctx, input)
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

		time.Sleep(rateLimitDelay)
	}

	return resourceTypes, nil
}

// fetchFullSyncServicePricing fetches pricing for a specific service/resource/region combination
func fetchFullSyncServicePricing(ctx context.Context, client *pricing.Client, serviceCode, resourceType, region string, attributes []string) (*FullAWSPricingData, error) {
	filters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("resourceType"), Value: aws.String(resourceType)},
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String(fullSyncRegionMap[region])},
	}

	for _, attr := range attributes {
		filters = append(filters, types.Filter{
			Type:  types.FilterTypeTermMatch,
			Field: aws.String(attr),
			Value: aws.String("*"),
		})
	}

	result, err := client.GetProducts(ctx, &pricing.GetProductsInput{
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

	pricingData := &FullAWSPricingData{
		ServiceCode:  serviceCode,
		ResourceType: resourceType,
		Region:       region,
		LastUpdated:  time.Now(),
		IsActive:     true,
	}

	var product map[string]interface{}
	if err := json.Unmarshal([]byte(result.PriceList[0]), &product); err != nil {
		return nil, err
	}

	attrs := make(map[string]interface{})
	if productAttrs, ok := product["attributes"].(map[string]interface{}); ok {
		for _, attr := range attributes {
			if val, exists := productAttrs[attr]; exists {
				attrs[attr] = val
			}
		}
	}
	attributesJSON, _ := json.Marshal(attrs)
	pricingData.Attributes = attributesJSON

	if terms, ok := product["terms"].(map[string]interface{}); ok {
		if onDemand, ok := terms["OnDemand"].(map[string]interface{}); ok {
			pricingData.OnDemandPriceUSD = extractFullSyncPrice(onDemand)
		}
		if reserved, ok := terms["Reserved"].(map[string]interface{}); ok {
			pricingData.ReservedPriceUSD = extractFullSyncPrice(reserved)
		}
		if spot, ok := terms["Spot"].(map[string]interface{}); ok {
			pricingData.SpotPriceUSD = extractFullSyncPrice(spot)
		}
	}

	return pricingData, nil
}

// extractFullSyncPrice extracts the price value from pricing terms
func extractFullSyncPrice(terms map[string]interface{}) float64 {
	for _, dimension := range terms {
		if dim, ok := dimension.(map[string]interface{}); ok {
			if priceDims, ok := dim["priceDimensions"].(map[string]interface{}); ok {
				for _, pd := range priceDims {
					if pdMap, ok := pd.(map[string]interface{}); ok {
						if pricePerUnit, ok := pdMap["pricePerUnit"].(map[string]interface{}); ok {
							if usd, ok := pricePerUnit["USD"].(string); ok {
								price, _ := strconv.ParseFloat(usd, 64)
								return price
							}
						}
					}
				}
			}
		}
	}
	return 0
}
