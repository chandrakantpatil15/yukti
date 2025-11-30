package services

import (
	"context"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
	"github.com/shopspring/decimal"
	"yukti/internal/models"
)

type BatchProcessor struct {
	db          *gorm.DB
	workerCount int
	batchSize   int
}

type ProcessingResult struct {
	TotalRecords    int64           `json:"total_records"`
	ProcessedCount  int64           `json:"processed_count"`
	TotalCost       decimal.Decimal `json:"total_cost"`
	TotalSavings    decimal.Decimal `json:"total_savings"`
	ProcessingTime  time.Duration   `json:"processing_time"`
	RecordsPerSec   float64         `json:"records_per_sec"`
}

func NewBatchProcessor(db *gorm.DB) *BatchProcessor {
	return &BatchProcessor{
		db:          db,
		workerCount: 10, // Concurrent workers
		batchSize:   1000, // Records per batch
	}
}

// Process massive cost data in parallel batches
func (bp *BatchProcessor) ProcessCostData(ctx context.Context, days int) (*ProcessingResult, error) {
	startTime := time.Now()
	
	// Get total record count
	var totalRecords int64
	startDate := time.Now().AddDate(0, 0, -days)
	
	err := bp.db.Model(&models.ResourceCost{}).
		Where("date >= ?", startDate).
		Count(&totalRecords).Error
	if err != nil {
		return nil, err
	}

	log.Printf("Processing %d cost records with %d workers", totalRecords, bp.workerCount)

	// Create worker pool
	jobs := make(chan int64, bp.workerCount)
	results := make(chan ProcessingResult, bp.workerCount)
	
	var wg sync.WaitGroup
	
	// Start workers
	for i := 0; i < bp.workerCount; i++ {
		wg.Add(1)
		go bp.costWorker(ctx, jobs, results, &wg, startDate)
	}

	// Send batch jobs
	go func() {
		defer close(jobs)
		for offset := int64(0); offset < totalRecords; offset += int64(bp.batchSize) {
			select {
			case jobs <- offset:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Aggregate results
	var finalResult ProcessingResult
	finalResult.TotalRecords = totalRecords
	
	for result := range results {
		finalResult.ProcessedCount += result.ProcessedCount
		finalResult.TotalCost = finalResult.TotalCost.Add(result.TotalCost)
		finalResult.TotalSavings = finalResult.TotalSavings.Add(result.TotalSavings)
	}

	finalResult.ProcessingTime = time.Since(startTime)
	if finalResult.ProcessingTime.Seconds() > 0 {
		finalResult.RecordsPerSec = float64(finalResult.ProcessedCount) / finalResult.ProcessingTime.Seconds()
	}

	log.Printf("Processed %d records in %v (%.0f records/sec)", 
		finalResult.ProcessedCount, finalResult.ProcessingTime, finalResult.RecordsPerSec)

	return &finalResult, nil
}

func (bp *BatchProcessor) costWorker(ctx context.Context, jobs <-chan int64, results chan<- ProcessingResult, wg *sync.WaitGroup, startDate time.Time) {
	defer wg.Done()
	
	var workerResult ProcessingResult
	
	for offset := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			// Process batch
			var costs []models.ResourceCost
			err := bp.db.Where("date >= ?", startDate).
				Offset(int(offset)).
				Limit(bp.batchSize).
				Find(&costs).Error
			
			if err != nil {
				log.Printf("Worker error at offset %d: %v", offset, err)
				continue
			}

			// Process each cost record
			for _, cost := range costs {
				workerResult.ProcessedCount++
				workerResult.TotalCost = workerResult.TotalCost.Add(cost.CostUSD)
				
				// Calculate potential savings (30% average)
				savings := cost.CostUSD.Mul(decimal.NewFromFloat(0.3))
				workerResult.TotalSavings = workerResult.TotalSavings.Add(savings)
			}
		}
	}
	
	results <- workerResult
}

// Process metrics data with streaming
func (bp *BatchProcessor) ProcessMetricsData(ctx context.Context, hours int) (*ProcessingResult, error) {
	startTime := time.Now()
	
	var totalRecords int64
	startTimestamp := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	err := bp.db.Model(&models.ResourceMetrics{}).
		Where("timestamp >= ?", startTimestamp).
		Count(&totalRecords).Error
	if err != nil {
		return nil, err
	}

	log.Printf("Processing %d metrics records", totalRecords)

	// Stream processing with cursor
	var processedCount int64
	var totalCPU, totalMemory decimal.Decimal
	
	// Process in chunks to avoid memory issues
	for offset := int64(0); offset < totalRecords; offset += int64(bp.batchSize) {
		var metrics []models.ResourceMetrics
		err := bp.db.Where("timestamp >= ?", startTimestamp).
			Offset(int(offset)).
			Limit(bp.batchSize).
			Find(&metrics).Error
		
		if err != nil {
			return nil, err
		}

		for _, metric := range metrics {
			processedCount++
			totalCPU = totalCPU.Add(decimal.NewFromFloat(metric.CPUUtilization))
			totalMemory = totalMemory.Add(decimal.NewFromFloat(metric.MemoryUtilization))
		}

		// Progress logging
		if offset%10000 == 0 {
			log.Printf("Processed %d/%d metrics records", processedCount, totalRecords)
		}
	}

	result := &ProcessingResult{
		TotalRecords:   totalRecords,
		ProcessedCount: processedCount,
		TotalCost:      totalCPU,  // Using CPU as example metric
		TotalSavings:   totalMemory, // Using Memory as example metric
		ProcessingTime: time.Since(startTime),
	}

	if result.ProcessingTime.Seconds() > 0 {
		result.RecordsPerSec = float64(result.ProcessedCount) / result.ProcessingTime.Seconds()
	}

	return result, nil
}

// Generate optimization recommendations at scale
func (bp *BatchProcessor) GenerateRecommendations(ctx context.Context) (*ProcessingResult, error) {
	startTime := time.Now()
	
	// Get all resources for analysis
	var resources []models.Resource
	err := bp.db.Find(&resources).Error
	if err != nil {
		return nil, err
	}

	log.Printf("Generating recommendations for %d resources", len(resources))

	jobs := make(chan models.Resource, bp.workerCount)
	results := make(chan int, bp.workerCount)
	
	var wg sync.WaitGroup
	
	// Start recommendation workers
	for i := 0; i < bp.workerCount; i++ {
		wg.Add(1)
		go bp.recommendationWorker(ctx, jobs, results, &wg)
	}

	// Send resources to workers
	go func() {
		defer close(jobs)
		for _, resource := range resources {
			select {
			case jobs <- resource:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	var totalRecommendations int
	for count := range results {
		totalRecommendations += count
	}

	result := &ProcessingResult{
		TotalRecords:   int64(len(resources)),
		ProcessedCount: int64(totalRecommendations),
		ProcessingTime: time.Since(startTime),
	}

	if result.ProcessingTime.Seconds() > 0 {
		result.RecordsPerSec = float64(result.ProcessedCount) / result.ProcessingTime.Seconds()
	}

	log.Printf("Generated %d recommendations in %v", totalRecommendations, result.ProcessingTime)

	return result, nil
}

func (bp *BatchProcessor) recommendationWorker(ctx context.Context, jobs <-chan models.Resource, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	
	recommendationCount := 0
	
	for resource := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			// Generate recommendation based on resource characteristics
			if bp.shouldRecommend(resource) {
				recommendation := bp.createRecommendation(resource)
				
				err := bp.db.Create(&recommendation).Error
				if err == nil {
					recommendationCount++
				}
			}
		}
	}
	
	results <- recommendationCount
}

func (bp *BatchProcessor) shouldRecommend(resource models.Resource) bool {
	// Logic to determine if resource needs recommendation
	if resource.Environment == "prod" && time.Since(resource.LaunchTime).Hours() > 24*90 {
		return true // RI recommendation for long-running prod resources
	}
	if resource.Environment != "prod" {
		return true // Spot recommendation for non-prod
	}
	if resource.Status == "stopped" {
		return true // Termination recommendation
	}
	return false
}

func (bp *BatchProcessor) createRecommendation(resource models.Resource) models.OptimizationRecommendation {
	var recType string
	var currentCost, optimizedCost, savings decimal.Decimal
	var confidence float64

	if resource.Environment == "prod" {
		recType = "reserved_instance"
		currentCost = decimal.NewFromFloat(100.0)
		optimizedCost = decimal.NewFromFloat(70.0)
		savings = decimal.NewFromFloat(30.0)
		confidence = 0.85
	} else if resource.Status == "stopped" {
		recType = "termination"
		currentCost = decimal.NewFromFloat(50.0)
		optimizedCost = decimal.NewFromFloat(0.0)
		savings = decimal.NewFromFloat(50.0)
		confidence = 0.95
	} else {
		recType = "spot_instance"
		currentCost = decimal.NewFromFloat(50.0)
		optimizedCost = decimal.NewFromFloat(15.0)
		savings = decimal.NewFromFloat(35.0)
		confidence = 0.75
	}

	return models.OptimizationRecommendation{
		ResourceID:         resource.ID,
		RecommendationType: recType,
		CurrentCost:        currentCost,
		OptimizedCost:      optimizedCost,
		PotentialSavings:   savings,
		Confidence:         confidence,
		Status:            "active",
		CreatedAt:         time.Now(),
	}
}