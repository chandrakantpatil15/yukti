package assessment

import "time"

// Config holds assessment configuration
type Config struct {
	TenantID                     string
	UnderutilizedCPUThreshold    float64
	UnderutilizedMemoryThreshold float64
	OverutilizedCPUThreshold     float64
	OverutilizedMemoryThreshold  float64
	BatchHighThreshold           float64
	BatchLowThreshold            float64
	BatchVarianceThreshold       float64
	WindowDays                   int
}

// ResourceMetrics holds collected metrics for a resource
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
	WindowStart    time.Time
	WindowEnd      time.Time
}

// Assessment holds the assessment result for a resource
type Assessment struct {
	ResourceARN              string
	InstanceID               string
	AssessmentTimestamp      time.Time
	WindowHours              int
	WindowStart              time.Time
	WindowEnd                time.Time
	UtilizationCategory      string
	UsagePattern            string
	PeakHours               string
	IdlePercentage          float64
	OptimizationScore       float64
	RecommendedAction       string
	RecommendedInstanceType string
	PotentialMonthlySavings float64
	CurrentHourlyCost       float64
	ProjectedHourlyCost     float64
	ConfidenceScore         int
	DataPointsAnalyzed      int
	EngineVersion           string
}

// Categories for utilization classification
const (
	CategoryUnderutilized = "underutilized"
	CategoryOverutilized  = "overutilized"
	CategoryIntermittent  = "intermittent"
	CategoryBatch         = "batch"
	CategoryIdle          = "idle"
	CategoryNormal        = "normal"
)

// Usage patterns
const (
	PatternSteady    = "steady"
	PatternBursty    = "bursty"
	PatternScheduled = "scheduled"
	PatternBatch     = "batch"
	PatternVariable  = "variable"
)

// Recommended actions
const (
	ActionDownsize   = "downsize"
	ActionUpsize     = "upsize"
	ActionSpot       = "spot"
	ActionBurstable  = "burstable"
	ActionTerminate  = "terminate"
	ActionSchedule   = "schedule"
	ActionMonitor    = "monitor"
)