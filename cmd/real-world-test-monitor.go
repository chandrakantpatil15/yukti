package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	_ "github.com/lib/pq"
)

// Real-world test monitoring and data collection
func main() {
	fmt.Println("📊 YUKTI FINOPS - REAL WORLD TEST MONITORING")
	fmt.Println("=" * 50)

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/yukti?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("❌ Database connection failed:", err)
	}
	defer db.Close()

	// AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("❌ AWS config failed:", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	cwClient := cloudwatch.NewFromConfig(cfg)

	// Monitor test instances
	for {
		fmt.Printf("\n🔍 Monitoring cycle at %s\n", time.Now().Format("15:04:05"))
		
		if err := monitorTestInstances(db, ec2Client, cwClient); err != nil {
			log.Printf("❌ Monitoring error: %v", err)
		}

		// Wait 5 minutes before next monitoring cycle
		time.Sleep(5 * time.Minute)
	}
}

func monitorTestInstances(db *sql.DB, ec2Client *ec2.Client, cwClient *cloudwatch.Client) error {
	// Get test instances
	instances, err := getTestInstances(ec2Client)
	if err != nil {
		return fmt.Errorf("failed to get test instances: %w", err)
	}

	if len(instances) == 0 {
		fmt.Println("⚠️  No test instances found. Run setup script first.")
		return nil
	}

	fmt.Printf("📋 Monitoring %d test instances\n", len(instances))

	// Collect metrics for each instance
	for _, instance := range instances {
		fmt.Printf("  🔍 %s (%s)...", *instance.InstanceId, *instance.InstanceType)
		
		metrics, err := collectInstanceMetrics(cwClient, *instance.InstanceId)
		if err != nil {
			fmt.Printf(" ❌ Error: %v\n", err)
			continue
		}

		// Store metrics in database
		if err := storeMetrics(db, *instance.InstanceId, metrics); err != nil {
			fmt.Printf(" ❌ Storage error: %v\n", err)
			continue
		}

		// Classify workload pattern
		pattern := classifyWorkloadPattern(metrics)
		fmt.Printf(" ✅ %s (CPU: %.1f%%, Pattern: %s)\n", 
			*instance.InstanceId, metrics.AvgCPU, pattern)
	}

	// Show summary
	showMonitoringSummary(db)
	return nil
}

func getTestInstances(client *ec2.Client) ([]ec2.Instance, error) {
	result, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		Filters: []ec2.Filter{
			{
				Name:   aws.String("tag:Project"),
				Values: []string{"yukti-finops"},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var instances []ec2.Instance
	for _, reservation := range result.Reservations {
		instances = append(instances, reservation.Instances...)
	}

	return instances, nil
}

type InstanceMetrics struct {
	InstanceID   string
	Timestamp    time.Time
	AvgCPU       float64
	MaxCPU       float64
	AvgMemory    float64
	MaxMemory    float64
	CPUVariance  float64
	DataPoints   int
}

func collectInstanceMetrics(client *cloudwatch.Client, instanceID string) (*InstanceMetrics, error) {
	endTime := time.Now()
	startTime := endTime.Add(-15 * time.Minute) // Last 15 minutes

	// Get CPU metrics
	cpuMetrics, err := getMetricData(client, "AWS/EC2", "CPUUtilization", instanceID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU metrics: %w", err)
	}

	if len(cpuMetrics) == 0 {
		return nil, fmt.Errorf("no CPU metrics available")
	}

	// Calculate statistics
	avgCPU := calculateAverage(cpuMetrics)
	maxCPU := calculateMax(cpuMetrics)
	cpuVariance := calculateVariance(cpuMetrics)

	// Estimate memory (CloudWatch doesn't provide memory by default)
	avgMemory := avgCPU * 0.8 // Rough correlation
	maxMemory := maxCPU * 0.8

	return &InstanceMetrics{
		InstanceID:  instanceID,
		Timestamp:   endTime,
		AvgCPU:      avgCPU,
		MaxCPU:      maxCPU,
		AvgMemory:   avgMemory,
		MaxMemory:   maxMemory,
		CPUVariance: cpuVariance,
		DataPoints:  len(cpuMetrics),
	}, nil
}

func getMetricData(client *cloudwatch.Client, namespace, metricName, instanceID string, 
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

	result, err := client.GetMetricStatistics(context.TODO(), input)
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

func storeMetrics(db *sql.DB, instanceID string, metrics *InstanceMetrics) error {
	query := `
		INSERT INTO yt_resource_metrics 
		(instance_id, timestamp, cpu_utilization, memory_utilization, 
		 workload_pattern, metric_source)
		VALUES ($1, $2, $3, $4, $5, 'real-world-test')
		ON CONFLICT (instance_id, timestamp) DO UPDATE SET
		cpu_utilization = EXCLUDED.cpu_utilization,
		memory_utilization = EXCLUDED.memory_utilization,
		workload_pattern = EXCLUDED.workload_pattern`

	pattern := classifyWorkloadPattern(metrics)

	_, err := db.Exec(query, instanceID, metrics.Timestamp, 
		metrics.AvgCPU, metrics.AvgMemory, pattern)
	
	return err
}

func classifyWorkloadPattern(metrics *InstanceMetrics) string {
	// Batch workload: High spikes then low usage
	if metrics.MaxCPU > 80 && metrics.AvgCPU < 30 && metrics.CPUVariance > 25 {
		return "batch"
	}
	
	// High variance = bursty
	if metrics.CPUVariance > 20 {
		return "bursty"
	}
	
	// Steady workload
	if metrics.CPUVariance < 15 && metrics.AvgCPU > 20 {
		return "steady"
	}
	
	// Idle
	if metrics.AvgCPU < 10 {
		return "idle"
	}
	
	return "variable"
}

func showMonitoringSummary(db *sql.DB) {
	fmt.Println("\n📊 MONITORING SUMMARY")
	fmt.Println("-" * 30)

	// Pattern distribution
	query := `
		SELECT workload_pattern, COUNT(*) as count,
		       AVG(cpu_utilization) as avg_cpu,
		       MAX(cpu_utilization) as max_cpu
		FROM yt_resource_metrics 
		WHERE timestamp > NOW() - INTERVAL '1 hour'
		AND metric_source = 'real-world-test'
		GROUP BY workload_pattern
		ORDER BY count DESC`

	rows, err := db.Query(query)
	if err != nil {
		fmt.Printf("❌ Query error: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-12s %-8s %-10s %-10s\n", "Pattern", "Count", "Avg CPU", "Max CPU")
	fmt.Println("-" * 45)

	for rows.Next() {
		var pattern string
		var count int
		var avgCPU, maxCPU float64

		err := rows.Scan(&pattern, &count, &avgCPU, &maxCPU)
		if err != nil {
			continue
		}

		fmt.Printf("%-12s %-8d %-10.1f %-10.1f\n", pattern, count, avgCPU, maxCPU)
	}

	// Total metrics collected
	var totalMetrics int
	db.QueryRow(`
		SELECT COUNT(*) FROM yt_resource_metrics 
		WHERE timestamp > NOW() - INTERVAL '1 hour'
		AND metric_source = 'real-world-test'`).Scan(&totalMetrics)

	fmt.Printf("\n📈 Total metrics collected (last hour): %d\n", totalMetrics)
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
	return sumSquares / float64(len(values)-1)
}