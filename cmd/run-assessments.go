package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	_ "github.com/lib/pq"
)

type ResourceMetrics struct {
	ResourceARN    string
	InstanceID     string
	InstanceType   string
	Region         string
	AvgCPU         float64
	MaxCPU         float64
	AvgMemory      float64
	MaxMemory      float64
	CPUVariance    float64
	IdlePercentage float64
	DataPoints     int
}

type Assessment struct {
	ResourceARN              string
	Category                 string
	UsagePattern            string
	OptimizationScore       float64
	RecommendedAction       string
	RecommendedInstanceType string
	PotentialSavings        float64
	CurrentHourlyCost       float64
	ProjectedHourlyCost     float64
	ConfidenceScore         int
}

func main() {
	var windowHours = flag.Int("window", 24, "Assessment window in hours")
	var dryRun = flag.Bool("dry-run", false, "Print assessments without storing")
	flag.Parse()

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/yukti?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Failed to load AWS config:", err)
	}

	cwClient := cloudwatch.NewFromConfig(cfg)

	fmt.Printf("Running resource assessments (window: %d hours)\n", *windowHours)

	// Get all active resources
	resources, err := getActiveResources(db)
	if err != nil {
		log.Fatal("Failed to get resources:", err)
	}

	fmt.Printf("Found %d resources to assess\n", len(resources))

	// Process each resource
	for _, resource := range resources {
		fmt.Printf("Assessing %s (%s)...\n", resource.InstanceID, resource.InstanceType)

		// Collect metrics from CloudWatch
		metrics, err := collectMetrics(cwClient, resource, *windowHours)
		if err != nil {
			fmt.Printf("  Error collecting metrics: %v\n", err)
			continue
		}

		if metrics.DataPoints < 10 {
			fmt.Printf("  Insufficient data points (%d), skipping\n", metrics.DataPoints)
			continue
		}

		// Generate assessment
		assessment := generateAssessment(metrics)

		if *dryRun {
			printAssessment(assessment)
		} else {
			// Store assessment
			if err := storeAssessment(db, assessment, *windowHours); err != nil {
				fmt.Printf("  Error storing assessment: %v\n", err)
			} else {
				fmt.Printf("  ✓ Assessment completed: %s (Score: %.2f)\n", 
					assessment.Category, assessment.OptimizationScore)
			}
		}
	}

	fmt.Println("Assessment completed")
}

func getActiveResources(db *sql.DB) ([]ResourceMetrics, error) {
	query := `
		SELECT r.resource_arn, r.instance_id, r.instance_type, r.region,
		       COALESCE(p.on_demand_price_usd, 0) as hourly_cost
		FROM yt_aws_resources r
		LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
		WHERE r.sync_status = 'active' AND r.state = 'running'`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []ResourceMetrics
	for rows.Next() {
		var r ResourceMetrics
		var hourlyCost float64
		err := rows.Scan(&r.ResourceARN, &r.InstanceID, &r.InstanceType, &r.Region, &hourlyCost)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func collectMetrics(cwClient *cloudwatch.Client, resource ResourceMetrics, windowHours int) (*ResourceMetrics, error) {
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(windowHours) * time.Hour)

	// Collect CPU metrics
	cpuMetrics, err := getMetricStatistics(cwClient, "AWS/EC2", "CPUUtilization", 
		resource.InstanceID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU metrics: %w", err)
	}

	// Calculate CPU statistics
	if len(cpuMetrics) > 0 {
		resource.AvgCPU = calculateAverage(cpuMetrics)
		resource.MaxCPU = calculateMax(cpuMetrics)
		resource.CPUVariance = calculateVariance(cpuMetrics)
		resource.IdlePercentage = calculateIdlePercentage(cpuMetrics, 10.0) // < 10% CPU = idle
		resource.DataPoints = len(cpuMetrics)
	}

	// For now, estimate memory from CPU (CloudWatch doesn't provide memory by default)
	// In production, you'd use CloudWatch Agent or custom metrics
	resource.AvgMemory = resource.AvgCPU * 0.8  // Rough estimation
	resource.MaxMemory = resource.MaxCPU * 0.8

	return &resource, nil
}

func getMetricStatistics(cwClient *cloudwatch.Client, namespace, metricName, instanceID string, 
	startTime, endTime time.Time) ([]float64, error) {
	
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String(metricName),
		Dimensions: []types.Dimension{
			{
				Name:  aws.String("InstanceId"),
				Value: aws.String(instanceID),
			},
		},
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(endTime),
		Period:     aws.Int32(300), // 5-minute intervals
		Statistics: []types.Statistic{types.StatisticAverage},
	}

	result, err := cwClient.GetMetricStatistics(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var values []float64
	for _, datapoint := range result.Datapoints {
		if datapoint.Average != nil {
			values = append(values, *datapoint.Average)
		}
	}

	return values, nil
}

func generateAssessment(metrics *ResourceMetrics) Assessment {
	assessment := Assessment{
		ResourceARN:   metrics.ResourceARN,
		ConfidenceScore: 85,
	}

	// Classify utilization category
	assessment.Category = classifyUtilization(metrics.AvgCPU, metrics.MaxCPU, 
		metrics.AvgMemory, metrics.CPUVariance)

	// Determine usage pattern
	assessment.UsagePattern = determineUsagePattern(metrics.AvgCPU, metrics.MaxCPU, 
		metrics.CPUVariance, metrics.IdlePercentage)

	// Calculate optimization score
	assessment.OptimizationScore = calculateOptimizationScore(assessment.Category, 
		metrics.AvgCPU, metrics.AvgMemory)

	// Generate recommendations
	assessment.RecommendedAction, assessment.RecommendedInstanceType = 
		generateRecommendations(assessment.Category, metrics.InstanceType, 
			metrics.AvgCPU, metrics.AvgMemory)

	return assessment
}

func classifyUtilization(avgCPU, maxCPU, avgMemory, cpuVariance float64) string {
	// Batch workload pattern
	if maxCPU > 80 && avgCPU < 30 && cpuVariance > 25 {
		return "batch"
	}
	
	// Overutilized
	if avgCPU > 80 || avgMemory > 80 {
		return "overutilized"
	}
	
	// Underutilized
	if avgCPU < 20 && avgMemory < 25 {
		return "underutilized"
	}
	
	// Intermittent/Bursty
	if cpuVariance > 20 {
		return "intermittent"
	}
	
	// Idle
	if avgCPU < 5 && avgMemory < 10 {
		return "idle"
	}
	
	return "normal"
}

func determineUsagePattern(avgCPU, maxCPU, cpuVariance, idlePercentage float64) string {
	if idlePercentage > 70 {
		return "scheduled"
	}
	if cpuVariance > 25 {
		return "bursty"
	}
	if maxCPU > 80 && avgCPU < 30 {
		return "batch"
	}
	if cpuVariance < 15 {
		return "steady"
	}
	return "variable"
}

func calculateOptimizationScore(category string, avgCPU, avgMemory float64) float64 {
	switch category {
	case "idle":
		return 0.95
	case "underutilized":
		return 0.85
	case "batch":
		return 0.75
	case "intermittent":
		return 0.65
	case "overutilized":
		return 0.30
	default:
		return 0.50
	}
}

func generateRecommendations(category, currentType string, avgCPU, avgMemory float64) (string, string) {
	switch category {
	case "idle":
		return "terminate", ""
	case "underutilized":
		return "downsize", suggestSmallerInstance(currentType)
	case "overutilized":
		return "upsize", suggestLargerInstance(currentType)
	case "batch":
		return "spot", currentType
	case "intermittent":
		return "burstable", suggestBurstableInstance(currentType)
	default:
		return "monitor", currentType
	}
}

func suggestSmallerInstance(currentType string) string {
	// Simple instance type downsizing logic
	downsizeMap := map[string]string{
		"m5.large":   "m5.medium",
		"m5.xlarge":  "m5.large",
		"m5.2xlarge": "m5.xlarge",
		"t3.medium":  "t3.small",
		"t3.large":   "t3.medium",
	}
	if smaller, exists := downsizeMap[currentType]; exists {
		return smaller
	}
	return currentType
}

func suggestLargerInstance(currentType string) string {
	// Simple instance type upsizing logic
	upsizeMap := map[string]string{
		"m5.medium": "m5.large",
		"m5.large":  "m5.xlarge",
		"m5.xlarge": "m5.2xlarge",
		"t3.small":  "t3.medium",
		"t3.medium": "t3.large",
	}
	if larger, exists := upsizeMap[currentType]; exists {
		return larger
	}
	return currentType
}

func suggestBurstableInstance(currentType string) string {
	// Convert to burstable instance types
	burstableMap := map[string]string{
		"m5.medium": "t3.medium",
		"m5.large":  "t3.large",
		"m5.xlarge": "t3.xlarge",
	}
	if burstable, exists := burstableMap[currentType]; exists {
		return burstable
	}
	return currentType
}

func storeAssessment(db *sql.DB, assessment Assessment, windowHours int) error {
	query := `
		INSERT INTO yt_resource_assessments 
		(resource_arn, assessment_timestamp, assessment_window_hours, 
		 utilization_category, usage_pattern, optimization_score, 
		 recommended_action, recommended_instance_type, confidence_score,
		 assessment_window_start, assessment_window_end)
		VALUES ($1, NOW(), $2, $3, $4, $5, $6, $7, $8, 
		        NOW() - INTERVAL '%d hours', NOW())
		ON CONFLICT (resource_arn) DO UPDATE SET
		assessment_timestamp = NOW(),
		utilization_category = EXCLUDED.utilization_category,
		usage_pattern = EXCLUDED.usage_pattern,
		optimization_score = EXCLUDED.optimization_score,
		recommended_action = EXCLUDED.recommended_action,
		recommended_instance_type = EXCLUDED.recommended_instance_type,
		confidence_score = EXCLUDED.confidence_score`

	_, err := db.Exec(fmt.Sprintf(query, windowHours),
		assessment.ResourceARN, windowHours, assessment.Category,
		assessment.UsagePattern, assessment.OptimizationScore,
		assessment.RecommendedAction, assessment.RecommendedInstanceType,
		assessment.ConfidenceScore)

	return err
}

func printAssessment(assessment Assessment) {
	fmt.Printf("  Resource: %s\n", assessment.ResourceARN)
	fmt.Printf("  Category: %s\n", assessment.Category)
	fmt.Printf("  Pattern: %s\n", assessment.UsagePattern)
	fmt.Printf("  Score: %.2f\n", assessment.OptimizationScore)
	fmt.Printf("  Action: %s\n", assessment.RecommendedAction)
	if assessment.RecommendedInstanceType != "" {
		fmt.Printf("  Recommended: %s\n", assessment.RecommendedInstanceType)
	}
	fmt.Println()
}

// Utility functions
func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func calculateVariance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	avg := calculateAverage(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - avg
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(values)-1))
}

func calculateIdlePercentage(values []float64, threshold float64) float64 {
	if len(values) == 0 {
		return 0
	}
	idleCount := 0
	for _, v := range values {
		if v < threshold {
			idleCount++
		}
	}
	return (float64(idleCount) / float64(len(values))) * 100
}