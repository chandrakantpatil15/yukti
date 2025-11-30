package optimization

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// MLEngine provides machine learning-based cost optimization
type MLEngine struct {
	historicalData []ResourceMetric
	models         map[string]*PredictionModel
}

// NewMLEngine creates a new ML optimization engine
func NewMLEngine() *MLEngine {
	return &MLEngine{
		models: make(map[string]*PredictionModel),
	}
}

// ResourceMetric represents historical resource usage data
type ResourceMetric struct {
	ResourceID   string    `json:"resource_id"`
	Timestamp    time.Time `json:"timestamp"`
	CPUUtil      float64   `json:"cpu_util"`
	MemoryUtil   float64   `json:"memory_util"`
	NetworkIO    float64   `json:"network_io"`
	Cost         float64   `json:"cost"`
	InstanceType string    `json:"instance_type"`
}

// PredictionModel represents a trained ML model
type PredictionModel struct {
	ResourceType string    `json:"resource_type"`
	Accuracy     float64   `json:"accuracy"`
	LastTrained  time.Time `json:"last_trained"`
	Weights      []float64 `json:"weights"`
}

// TrainModels trains ML models on historical data
func (ml *MLEngine) TrainModels(ctx context.Context, data []ResourceMetric) error {
	ml.historicalData = data
	
	// Group data by resource type
	typeGroups := make(map[string][]ResourceMetric)
	for _, metric := range data {
		typeGroups[metric.InstanceType] = append(typeGroups[metric.InstanceType], metric)
	}
	
	// Train model for each resource type
	for resourceType, metrics := range typeGroups {
		if len(metrics) < 10 { // Need minimum data points
			continue
		}
		
		model := ml.trainLinearRegression(metrics)
		model.ResourceType = resourceType
		model.LastTrained = time.Now()
		
		ml.models[resourceType] = model
	}
	
	return nil
}

// trainLinearRegression trains a simple linear regression model
func (ml *MLEngine) trainLinearRegression(data []ResourceMetric) *PredictionModel {
	n := len(data)
	if n == 0 {
		return &PredictionModel{Accuracy: 0.0}
	}
	
	// Simple linear regression: cost = w0 + w1*cpu + w2*memory + w3*network
	weights := make([]float64, 4)
	
	// Calculate means
	var meanCost, meanCPU, meanMemory, meanNetwork float64
	for _, metric := range data {
		meanCost += metric.Cost
		meanCPU += metric.CPUUtil
		meanMemory += metric.MemoryUtil
		meanNetwork += metric.NetworkIO
	}
	meanCost /= float64(n)
	meanCPU /= float64(n)
	meanMemory /= float64(n)
	meanNetwork /= float64(n)
	
	// Calculate weights using simplified least squares
	var sumXY, sumX2 float64
	for _, metric := range data {
		x := (metric.CPUUtil + metric.MemoryUtil + metric.NetworkIO) / 3.0
		y := metric.Cost
		sumXY += x * y
		sumX2 += x * x
	}
	
	if sumX2 > 0 {
		weights[1] = sumXY / sumX2
		weights[0] = meanCost - weights[1]*((meanCPU+meanMemory+meanNetwork)/3.0)
	}
	
	// Calculate accuracy (R-squared approximation)
	accuracy := ml.calculateAccuracy(data, weights)
	
	return &PredictionModel{
		Weights:  weights,
		Accuracy: accuracy,
	}
}

// calculateAccuracy calculates model accuracy
func (ml *MLEngine) calculateAccuracy(data []ResourceMetric, weights []float64) float64 {
	if len(data) == 0 {
		return 0.0
	}
	
	var totalError, totalVariance float64
	var meanCost float64
	
	// Calculate mean cost
	for _, metric := range data {
		meanCost += metric.Cost
	}
	meanCost /= float64(len(data))
	
	// Calculate errors and variance
	for _, metric := range data {
		predicted := ml.predictCost(metric, weights)
		error := metric.Cost - predicted
		totalError += error * error
		
		variance := metric.Cost - meanCost
		totalVariance += variance * variance
	}
	
	if totalVariance == 0 {
		return 1.0
	}
	
	rSquared := 1.0 - (totalError / totalVariance)
	return math.Max(0.0, math.Min(1.0, rSquared))
}

// predictCost predicts cost using trained model
func (ml *MLEngine) predictCost(metric ResourceMetric, weights []float64) float64 {
	if len(weights) < 2 {
		return metric.Cost
	}
	
	avgUtil := (metric.CPUUtil + metric.MemoryUtil + metric.NetworkIO) / 3.0
	return weights[0] + weights[1]*avgUtil
}

// PredictOptimalSize predicts optimal instance size
func (ml *MLEngine) PredictOptimalSize(ctx context.Context, resourceID string, currentMetrics ResourceMetric) (*OptimizationRecommendation, error) {
	model, exists := ml.models[currentMetrics.InstanceType]
	if !exists {
		return ml.generateRuleBasedRecommendation(currentMetrics), nil
	}
	
	// Predict cost for current configuration
	currentCost := ml.predictCost(currentMetrics, model.Weights)
	
	// Generate recommendations based on utilization patterns
	recommendations := ml.generateMLRecommendations(currentMetrics, currentCost, model)
	
	return recommendations, nil
}

// generateMLRecommendations generates ML-based recommendations
func (ml *MLEngine) generateMLRecommendations(metrics ResourceMetric, currentCost float64, model *PredictionModel) *OptimizationRecommendation {
	avgUtil := (metrics.CPUUtil + metrics.MemoryUtil + metrics.NetworkIO) / 3.0
	
	var action string
	var estimatedSavings float64
	var confidence float64 = model.Accuracy
	
	if avgUtil < 20 {
		action = "downsize"
		estimatedSavings = currentCost * 0.4 // 40% savings
	} else if avgUtil > 80 {
		action = "upsize"
		estimatedSavings = -currentCost * 0.3 // 30% cost increase for better performance
	} else {
		action = "maintain"
		estimatedSavings = 0
	}
	
	return &OptimizationRecommendation{
		ResourceID:       metrics.ResourceID,
		CurrentCost:      currentCost,
		RecommendedAction: action,
		EstimatedSavings: estimatedSavings,
		Confidence:       confidence,
		Reasoning:        fmt.Sprintf("ML model prediction based on %.1f%% average utilization", avgUtil),
		ModelAccuracy:    model.Accuracy,
		GeneratedAt:      time.Now(),
	}
}

// generateRuleBasedRecommendation fallback for when ML model is not available
func (ml *MLEngine) generateRuleBasedRecommendation(metrics ResourceMetric) *OptimizationRecommendation {
	avgUtil := (metrics.CPUUtil + metrics.MemoryUtil + metrics.NetworkIO) / 3.0
	
	var action string
	var estimatedSavings float64
	
	if avgUtil < 15 {
		action = "terminate"
		estimatedSavings = metrics.Cost
	} else if avgUtil < 30 {
		action = "downsize"
		estimatedSavings = metrics.Cost * 0.3
	} else if avgUtil > 85 {
		action = "upsize"
		estimatedSavings = -metrics.Cost * 0.2
	} else {
		action = "maintain"
		estimatedSavings = 0
	}
	
	return &OptimizationRecommendation{
		ResourceID:       metrics.ResourceID,
		CurrentCost:      metrics.Cost,
		RecommendedAction: action,
		EstimatedSavings: estimatedSavings,
		Confidence:       0.7, // Rule-based confidence
		Reasoning:        fmt.Sprintf("Rule-based recommendation for %.1f%% utilization", avgUtil),
		ModelAccuracy:    0.0,
		GeneratedAt:      time.Now(),
	}
}

// DetectAnomalies detects cost anomalies using ML
func (ml *MLEngine) DetectAnomalies(ctx context.Context, metrics []ResourceMetric) ([]CostAnomaly, error) {
	var anomalies []CostAnomaly
	
	if len(metrics) < 7 { // Need at least a week of data
		return anomalies, nil
	}
	
	// Sort by timestamp
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Timestamp.Before(metrics[j].Timestamp)
	})
	
	// Calculate moving average and detect outliers
	windowSize := 7
	for i := windowSize; i < len(metrics); i++ {
		current := metrics[i]
		
		// Calculate average of previous window
		var avgCost float64
		for j := i - windowSize; j < i; j++ {
			avgCost += metrics[j].Cost
		}
		avgCost /= float64(windowSize)
		
		// Detect anomaly if cost is significantly higher
		threshold := avgCost * 1.5 // 50% increase threshold
		if current.Cost > threshold {
			anomaly := CostAnomaly{
				ResourceID:    current.ResourceID,
				Timestamp:     current.Timestamp,
				ActualCost:    current.Cost,
				ExpectedCost:  avgCost,
				Deviation:     ((current.Cost - avgCost) / avgCost) * 100,
				Severity:      ml.calculateSeverity(current.Cost, avgCost),
				DetectedAt:    time.Now(),
			}
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies, nil
}

// calculateSeverity calculates anomaly severity
func (ml *MLEngine) calculateSeverity(actual, expected float64) string {
	deviation := ((actual - expected) / expected) * 100
	
	if deviation > 200 {
		return "critical"
	} else if deviation > 100 {
		return "high"
	} else if deviation > 50 {
		return "medium"
	}
	return "low"
}

// OptimizationRecommendation represents an ML-generated recommendation
type OptimizationRecommendation struct {
	ResourceID        string    `json:"resource_id"`
	CurrentCost       float64   `json:"current_cost"`
	RecommendedAction string    `json:"recommended_action"`
	EstimatedSavings  float64   `json:"estimated_savings"`
	Confidence        float64   `json:"confidence"`
	Reasoning         string    `json:"reasoning"`
	ModelAccuracy     float64   `json:"model_accuracy"`
	GeneratedAt       time.Time `json:"generated_at"`
}

// CostAnomaly represents a detected cost anomaly
type CostAnomaly struct {
	ResourceID   string    `json:"resource_id"`
	Timestamp    time.Time `json:"timestamp"`
	ActualCost   float64   `json:"actual_cost"`
	ExpectedCost float64   `json:"expected_cost"`
	Deviation    float64   `json:"deviation_percent"`
	Severity     string    `json:"severity"`
	DetectedAt   time.Time `json:"detected_at"`
}