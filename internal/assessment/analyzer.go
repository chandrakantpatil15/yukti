package assessment

import "math"

// Analyzer processes metrics and generates assessments
type Analyzer struct {
	config Config
}

// NewAnalyzer creates a new analyzer with configuration
func NewAnalyzer(config Config) *Analyzer {
	return &Analyzer{config: config}
}

// Analyze processes resource metrics and generates assessment
func (a *Analyzer) Analyze(metrics *ResourceMetrics) Assessment {
	assessment := Assessment{
		ResourceARN:         metrics.ResourceARN,
		InstanceID:          metrics.InstanceID,
		WindowHours:         24, // Default window
		WindowStart:         metrics.WindowStart,
		WindowEnd:           metrics.WindowEnd,
		IdlePercentage:      metrics.IdlePercentage,
		DataPointsAnalyzed:  metrics.DataPoints,
		EngineVersion:       "1.0",
		ConfidenceScore:     a.calculateConfidence(metrics),
	}

	// Classify utilization category
	assessment.UtilizationCategory = a.classifyUtilization(metrics)

	// Determine usage pattern
	assessment.UsagePattern = a.determineUsagePattern(metrics)

	// Calculate optimization score
	assessment.OptimizationScore = a.calculateOptimizationScore(assessment.UtilizationCategory, metrics)

	// Generate recommendations
	assessment.RecommendedAction, assessment.RecommendedInstanceType = 
		a.generateRecommendations(assessment.UtilizationCategory, metrics)

	return assessment
}

// classifyUtilization determines the utilization category
func (a *Analyzer) classifyUtilization(metrics *ResourceMetrics) string {
	// Batch workload pattern
	if metrics.MaxCPU > a.config.BatchHighThreshold && 
	   metrics.AvgCPU < a.config.BatchLowThreshold && 
	   metrics.CPUVariance > a.config.BatchVarianceThreshold {
		return CategoryBatch
	}
	
	// Overutilized
	if metrics.AvgCPU > a.config.OverutilizedCPUThreshold || 
	   metrics.AvgMemory > a.config.OverutilizedMemoryThreshold {
		return CategoryOverutilized
	}
	
	// Underutilized
	if metrics.AvgCPU < a.config.UnderutilizedCPUThreshold && 
	   metrics.AvgMemory < a.config.UnderutilizedMemoryThreshold {
		return CategoryUnderutilized
	}
	
	// Intermittent/Bursty
	if metrics.CPUVariance > 20 {
		return CategoryIntermittent
	}
	
	// Idle
	if metrics.AvgCPU < 5 && metrics.AvgMemory < 10 {
		return CategoryIdle
	}
	
	return CategoryNormal
}

// determineUsagePattern analyzes usage patterns
func (a *Analyzer) determineUsagePattern(metrics *ResourceMetrics) string {
	if metrics.IdlePercentage > 70 {
		return PatternScheduled
	}
	if metrics.CPUVariance > 25 {
		return PatternBursty
	}
	if metrics.MaxCPU > 80 && metrics.AvgCPU < 30 {
		return PatternBatch
	}
	if metrics.CPUVariance < 15 {
		return PatternSteady
	}
	return PatternVariable
}

// calculateOptimizationScore computes optimization potential
func (a *Analyzer) calculateOptimizationScore(category string, metrics *ResourceMetrics) float64 {
	baseScore := a.getCategoryBaseScore(category)
	
	// Adjust based on actual utilization
	utilizationFactor := 1.0
	if metrics.AvgCPU < 10 {
		utilizationFactor = 1.2 // Very low utilization increases score
	} else if metrics.AvgCPU > 70 {
		utilizationFactor = 0.8 // High utilization decreases score
	}
	
	// Adjust based on data confidence
	confidenceFactor := math.Min(1.0, float64(metrics.DataPoints)/100.0)
	
	score := baseScore * utilizationFactor * confidenceFactor
	return math.Min(1.0, math.Max(0.0, score))
}

// getCategoryBaseScore returns base optimization score for category
func (a *Analyzer) getCategoryBaseScore(category string) float64 {
	switch category {
	case CategoryIdle:
		return 0.95
	case CategoryUnderutilized:
		return 0.85
	case CategoryBatch:
		return 0.75
	case CategoryIntermittent:
		return 0.65
	case CategoryOverutilized:
		return 0.30
	default:
		return 0.50
	}
}

// generateRecommendations creates action recommendations
func (a *Analyzer) generateRecommendations(category string, metrics *ResourceMetrics) (string, string) {
	switch category {
	case CategoryIdle:
		return ActionTerminate, ""
	case CategoryUnderutilized:
		return ActionDownsize, a.suggestSmallerInstance(metrics.InstanceType)
	case CategoryOverutilized:
		return ActionUpsize, a.suggestLargerInstance(metrics.InstanceType)
	case CategoryBatch:
		return ActionSpot, metrics.InstanceType
	case CategoryIntermittent:
		return ActionBurstable, a.suggestBurstableInstance(metrics.InstanceType)
	default:
		return ActionMonitor, metrics.InstanceType
	}
}

// calculateConfidence determines confidence score based on data quality
func (a *Analyzer) calculateConfidence(metrics *ResourceMetrics) int {
	baseConfidence := 85
	
	// Adjust based on data points
	if metrics.DataPoints < 50 {
		baseConfidence -= 20
	} else if metrics.DataPoints > 200 {
		baseConfidence += 10
	}
	
	// Adjust based on time window
	windowHours := int(metrics.WindowEnd.Sub(metrics.WindowStart).Hours())
	if windowHours < 12 {
		baseConfidence -= 15
	} else if windowHours > 48 {
		baseConfidence += 5
	}
	
	return int(math.Min(100, math.Max(0, float64(baseConfidence))))
}

// Instance type suggestion helpers
func (a *Analyzer) suggestSmallerInstance(currentType string) string {
	downsizeMap := map[string]string{
		"m5.large":   "m5.medium",
		"m5.xlarge":  "m5.large",
		"m5.2xlarge": "m5.xlarge",
		"m5.4xlarge": "m5.2xlarge",
		"t3.medium":  "t3.small",
		"t3.large":   "t3.medium",
		"t3.xlarge":  "t3.large",
	}
	if smaller, exists := downsizeMap[currentType]; exists {
		return smaller
	}
	return currentType
}

func (a *Analyzer) suggestLargerInstance(currentType string) string {
	upsizeMap := map[string]string{
		"m5.medium":  "m5.large",
		"m5.large":   "m5.xlarge",
		"m5.xlarge":  "m5.2xlarge",
		"m5.2xlarge": "m5.4xlarge",
		"t3.small":   "t3.medium",
		"t3.medium":  "t3.large",
		"t3.large":   "t3.xlarge",
	}
	if larger, exists := upsizeMap[currentType]; exists {
		return larger
	}
	return currentType
}

func (a *Analyzer) suggestBurstableInstance(currentType string) string {
	burstableMap := map[string]string{
		"m5.medium":  "t3.medium",
		"m5.large":   "t3.large",
		"m5.xlarge":  "t3.xlarge",
		"m5.2xlarge": "t3.2xlarge",
	}
	if burstable, exists := burstableMap[currentType]; exists {
		return burstable
	}
	return currentType
}