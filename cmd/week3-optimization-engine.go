package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"yukti/internal/optimization"
)

// Week3OptimizationEngine demonstrates advanced optimization algorithms
func main() {
	log.Println("🎯 Week 3: Advanced Optimization Algorithms")
	log.Println("==========================================")
	
	ctx := context.Background()
	
	// Initialize optimization engines
	mlEngine := optimization.NewMLEngine()
	depAnalyzer := optimization.NewDependencyAnalyzer()
	realtimeMonitor := optimization.NewRealtimeMonitor()
	
	// Generate sample data for demonstration
	historicalData := generateSampleData(100)
	
	log.Println("🤖 Training ML Models...")
	if err := mlEngine.TrainModels(ctx, historicalData); err != nil {
		log.Fatalf("Failed to train ML models: %v", err)
	}
	
	log.Println("🔗 Analyzing Dependencies...")
	depAnalysis, err := depAnalyzer.AnalyzeDependencies(ctx, historicalData)
	if err != nil {
		log.Fatalf("Failed to analyze dependencies: %v", err)
	}
	
	log.Println("📊 Starting Real-time Monitoring...")
	if err := realtimeMonitor.Start(ctx); err != nil {
		log.Fatalf("Failed to start real-time monitor: %v", err)
	}
	defer realtimeMonitor.Stop()
	
	// Set up monitoring thresholds
	setupMonitoringThresholds(realtimeMonitor)
	
	// Generate ML-based recommendations
	log.Println("🧠 Generating ML Recommendations...")
	mlRecommendations := generateMLRecommendations(ctx, mlEngine, historicalData[:10])
	
	// Detect anomalies
	log.Println("🚨 Detecting Cost Anomalies...")
	anomalies, err := mlEngine.DetectAnomalies(ctx, historicalData)
	if err != nil {
		log.Fatalf("Failed to detect anomalies: %v", err)
	}
	
	// Test real-time monitoring
	log.Println("⚡ Testing Real-time Monitoring...")
	testRealtimeMonitoring(realtimeMonitor, historicalData[:5])
	
	// Generate comprehensive report
	generateOptimizationReport(mlRecommendations, depAnalysis, anomalies)
	
	log.Println("✅ Week 3 Advanced Optimization Complete!")
}

// generateSampleData generates sample historical data for demonstration
func generateSampleData(count int) []optimization.ResourceMetric {
	var data []optimization.ResourceMetric
	
	instanceTypes := []string{"t3.micro", "t3.small", "t3.medium", "m5.large", "c5.xlarge"}
	baseTime := time.Now().AddDate(0, 0, -30) // 30 days ago
	
	for i := 0; i < count; i++ {
		instanceType := instanceTypes[rand.Intn(len(instanceTypes))]
		
		// Generate realistic utilization patterns
		var cpuUtil, memUtil, netIO, cost float64
		
		switch instanceType {
		case "t3.micro":
			cpuUtil = 10 + rand.Float64()*30  // 10-40%
			memUtil = 20 + rand.Float64()*40  // 20-60%
			netIO = 5 + rand.Float64()*15     // 5-20%
			cost = 8.5 + rand.Float64()*5     // $8.5-13.5/day
		case "t3.small":
			cpuUtil = 15 + rand.Float64()*35  // 15-50%
			memUtil = 25 + rand.Float64()*45  // 25-70%
			netIO = 8 + rand.Float64()*20     // 8-28%
			cost = 17 + rand.Float64()*8      // $17-25/day
		case "m5.large":
			cpuUtil = 30 + rand.Float64()*50  // 30-80%
			memUtil = 40 + rand.Float64()*50  // 40-90%
			netIO = 15 + rand.Float64()*35    // 15-50%
			cost = 70 + rand.Float64()*30     // $70-100/day
		default:
			cpuUtil = 20 + rand.Float64()*60
			memUtil = 30 + rand.Float64()*60
			netIO = 10 + rand.Float64()*30
			cost = 50 + rand.Float64()*50
		}
		
		metric := optimization.ResourceMetric{
			ResourceID:   fmt.Sprintf("i-%s-%03d", instanceType, i),
			Timestamp:    baseTime.Add(time.Duration(i) * time.Hour),
			CPUUtil:      cpuUtil,
			MemoryUtil:   memUtil,
			NetworkIO:    netIO,
			Cost:         cost,
			InstanceType: instanceType,
		}
		
		data = append(data, metric)
	}
	
	return data
}

// setupMonitoringThresholds sets up alert thresholds
func setupMonitoringThresholds(monitor *optimization.RealtimeMonitor) {
	// Set up email subscriber
	emailSub := &optimization.EmailSubscriber{Email: "admin@company.com"}
	monitor.Subscribe(emailSub)
	
	// Set up Slack subscriber
	slackSub := &optimization.SlackSubscriber{
		WebhookURL: "https://hooks.slack.com/webhook",
		Channel:    "#cost-alerts",
	}
	monitor.Subscribe(slackSub)
	
	// Set thresholds for different instance types
	thresholds := map[string]optimization.AlertThreshold{
		"t3.micro": {
			DailyCost:     15.0,  // Alert if daily cost > $15
			MonthlyCost:   400.0, // Alert if monthly projection > $400
			PercentChange: 50.0,  // Alert if 50% increase
			Enabled:       true,
		},
		"m5.large": {
			DailyCost:     120.0, // Alert if daily cost > $120
			MonthlyCost:   3000.0, // Alert if monthly projection > $3000
			PercentChange: 30.0,  // Alert if 30% increase
			Enabled:       true,
		},
	}
	
	for resourceType, threshold := range thresholds {
		monitor.SetThreshold(resourceType, threshold)
	}
}

// generateMLRecommendations generates ML-based optimization recommendations
func generateMLRecommendations(ctx context.Context, mlEngine *optimization.MLEngine, samples []optimization.ResourceMetric) []*optimization.OptimizationRecommendation {
	var recommendations []*optimization.OptimizationRecommendation
	
	for _, sample := range samples {
		rec, err := mlEngine.PredictOptimalSize(ctx, sample.ResourceID, sample)
		if err != nil {
			log.Printf("Failed to generate recommendation for %s: %v", sample.ResourceID, err)
			continue
		}
		recommendations = append(recommendations, rec)
	}
	
	return recommendations
}

// testRealtimeMonitoring tests the real-time monitoring system
func testRealtimeMonitoring(monitor *optimization.RealtimeMonitor, samples []optimization.ResourceMetric) {
	for _, sample := range samples {
		// Simulate cost spike for demonstration
		if sample.InstanceType == "t3.micro" {
			sample.Cost = 20.0 // Spike above threshold
		}
		
		monitor.CheckCost(sample)
		time.Sleep(100 * time.Millisecond) // Small delay for demonstration
	}
}

// generateOptimizationReport generates a comprehensive optimization report
func generateOptimizationReport(mlRecs []*optimization.OptimizationRecommendation, depAnalysis *optimization.DependencyAnalysis, anomalies []optimization.CostAnomaly) {
	report := OptimizationReport{
		Timestamp:           time.Now(),
		Week:                3,
		Phase:               "Advanced Optimization Algorithms",
		MLRecommendations:   mlRecs,
		DependencyAnalysis:  depAnalysis,
		CostAnomalies:       anomalies,
		TotalSavings:        calculateTotalSavings(mlRecs, depAnalysis),
		OptimizationScore:   calculateOptimizationScore(mlRecs, anomalies),
	}
	
	// Convert to JSON for display
	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	
	log.Println("\n📋 Week 3 Optimization Report:")
	log.Println("==============================")
	fmt.Println(string(reportJSON))
	
	// Summary metrics
	log.Println("\n📈 Optimization Summary:")
	log.Println("========================")
	log.Printf("🤖 ML Recommendations: %d", len(mlRecs))
	log.Printf("🔗 Dependency Chains: %d", len(depAnalysis.DependencyChains))
	log.Printf("🚨 Cost Anomalies: %d", len(anomalies))
	log.Printf("💰 Total Potential Savings: $%.2f", report.TotalSavings)
	log.Printf("📊 Optimization Score: %.1f/100", report.OptimizationScore)
	
	// Key insights
	log.Println("\n🔍 Key Insights:")
	log.Println("================")
	
	// ML insights
	highConfidenceRecs := 0
	for _, rec := range mlRecs {
		if rec.Confidence > 0.8 {
			highConfidenceRecs++
		}
	}
	log.Printf("• %d high-confidence ML recommendations (>80%% confidence)", highConfidenceRecs)
	
	// Dependency insights
	if len(depAnalysis.DependencyChains) > 0 {
		longestChain := 0
		for _, chain := range depAnalysis.DependencyChains {
			if chain.Length > longestChain {
				longestChain = chain.Length
			}
		}
		log.Printf("• Longest dependency chain: %d resources", longestChain)
	}
	
	// Anomaly insights
	criticalAnomalies := 0
	for _, anomaly := range anomalies {
		if anomaly.Severity == "critical" {
			criticalAnomalies++
		}
	}
	log.Printf("• %d critical cost anomalies detected", criticalAnomalies)
	
	log.Println("\n🔮 Week 4 Preview:")
	log.Println("==================")
	log.Println("• Automated remediation workflows")
	log.Println("• Policy-based governance")
	log.Println("• Advanced scheduling algorithms")
	log.Println("• Multi-cloud cost optimization")
}

// calculateTotalSavings calculates total potential savings
func calculateTotalSavings(mlRecs []*optimization.OptimizationRecommendation, depAnalysis *optimization.DependencyAnalysis) float64 {
	var total float64
	
	// Add ML recommendation savings
	for _, rec := range mlRecs {
		if rec.EstimatedSavings > 0 {
			total += rec.EstimatedSavings
		}
	}
	
	// Add dependency optimization savings
	for _, rec := range depAnalysis.Recommendations {
		total += rec.EstimatedSavings
	}
	
	return total
}

// calculateOptimizationScore calculates overall optimization score
func calculateOptimizationScore(mlRecs []*optimization.OptimizationRecommendation, anomalies []optimization.CostAnomaly) float64 {
	score := 50.0 // Base score
	
	// Add points for recommendations
	score += float64(len(mlRecs)) * 2.0
	
	// Subtract points for anomalies
	for _, anomaly := range anomalies {
		switch anomaly.Severity {
		case "critical":
			score -= 10.0
		case "high":
			score -= 5.0
		case "medium":
			score -= 2.0
		case "low":
			score -= 1.0
		}
	}
	
	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}
	
	return score
}

// OptimizationReport represents the comprehensive optimization report
type OptimizationReport struct {
	Timestamp          time.Time                                    `json:"timestamp"`
	Week               int                                          `json:"week"`
	Phase              string                                       `json:"phase"`
	MLRecommendations  []*optimization.OptimizationRecommendation  `json:"ml_recommendations"`
	DependencyAnalysis *optimization.DependencyAnalysis            `json:"dependency_analysis"`
	CostAnomalies      []optimization.CostAnomaly                  `json:"cost_anomalies"`
	TotalSavings       float64                                      `json:"total_savings"`
	OptimizationScore  float64                                      `json:"optimization_score"`
}