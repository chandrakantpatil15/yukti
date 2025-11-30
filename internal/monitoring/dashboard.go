package monitoring

import (
	"context"
	"time"
)

// DashboardService provides real-time cost monitoring dashboards
type DashboardService struct {
	dataSource DataSource
	cache      CacheService
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(ds DataSource, cache CacheService) *DashboardService {
	return &DashboardService{
		dataSource: ds,
		cache:      cache,
	}
}

// GetRealTimeCostDashboard returns real-time cost dashboard data
func (ds *DashboardService) GetRealTimeCostDashboard(ctx context.Context, accountID string) (*CostDashboard, error) {
	// Check cache first
	if cached := ds.cache.Get("dashboard:" + accountID); cached != nil {
		return cached.(*CostDashboard), nil
	}

	dashboard := &CostDashboard{
		AccountID:   accountID,
		GeneratedAt: time.Now(),
		Period:      "last-30-days",
	}

	// Fetch current month costs
	dashboard.CurrentMonthCost = ds.dataSource.GetCurrentMonthCost(ctx, accountID)
	dashboard.PreviousMonthCost = ds.dataSource.GetPreviousMonthCost(ctx, accountID)
	dashboard.CostTrend = calculateTrend(dashboard.CurrentMonthCost, dashboard.PreviousMonthCost)

	// Fetch cost breakdown by service
	dashboard.ServiceBreakdown = ds.dataSource.GetServiceBreakdown(ctx, accountID)

	// Fetch cost breakdown by region
	dashboard.RegionBreakdown = ds.dataSource.GetRegionBreakdown(ctx, accountID)

	// Fetch top cost resources
	dashboard.TopResources = ds.dataSource.GetTopCostResources(ctx, accountID, 10)

	// Calculate daily average
	dashboard.DailyAverage = dashboard.CurrentMonthCost / float64(time.Now().Day())

	// Forecast end of month
	daysInMonth := 30.0
	dashboard.ForecastedMonthEnd = dashboard.DailyAverage * daysInMonth

	// Cache for 5 minutes
	ds.cache.Set("dashboard:"+accountID, dashboard, 5*time.Minute)

	return dashboard, nil
}

// GetCostTrendChart returns cost trend data for charting
func (ds *DashboardService) GetCostTrendChart(ctx context.Context, accountID string, days int) (*TrendChart, error) {
	dataPoints := ds.dataSource.GetDailyCosts(ctx, accountID, days)

	chart := &TrendChart{
		Title:      "Cost Trend",
		Period:     days,
		DataPoints: dataPoints,
		GeneratedAt: time.Now(),
	}

	// Calculate statistics
	if len(dataPoints) > 0 {
		var total, min, max float64
		min = dataPoints[0].Value
		max = dataPoints[0].Value

		for _, point := range dataPoints {
			total += point.Value
			if point.Value < min {
				min = point.Value
			}
			if point.Value > max {
				max = point.Value
			}
		}

		chart.Average = total / float64(len(dataPoints))
		chart.Min = min
		chart.Max = max
		chart.Total = total
	}

	return chart, nil
}

// GetServiceCostBreakdown returns detailed service cost breakdown
func (ds *DashboardService) GetServiceCostBreakdown(ctx context.Context, accountID string) (*ServiceBreakdown, error) {
	services := ds.dataSource.GetServiceBreakdown(ctx, accountID)

	breakdown := &ServiceBreakdown{
		AccountID:   accountID,
		Services:    services,
		TotalCost:   0,
		GeneratedAt: time.Now(),
	}

	// Calculate total and percentages
	for _, service := range services {
		breakdown.TotalCost += service.Cost
	}

	for i := range breakdown.Services {
		if breakdown.TotalCost > 0 {
			breakdown.Services[i].Percentage = (breakdown.Services[i].Cost / breakdown.TotalCost) * 100
		}
	}

	return breakdown, nil
}

// calculateTrend calculates cost trend percentage
func calculateTrend(current, previous float64) float64 {
	if previous == 0 {
		return 0
	}
	return ((current - previous) / previous) * 100
}

// CostDashboard represents real-time cost dashboard
type CostDashboard struct {
	AccountID          string                 `json:"account_id"`
	CurrentMonthCost   float64                `json:"current_month_cost"`
	PreviousMonthCost  float64                `json:"previous_month_cost"`
	CostTrend          float64                `json:"cost_trend_percent"`
	DailyAverage       float64                `json:"daily_average"`
	ForecastedMonthEnd float64                `json:"forecasted_month_end"`
	ServiceBreakdown   []ServiceCost          `json:"service_breakdown"`
	RegionBreakdown    []RegionCost           `json:"region_breakdown"`
	TopResources       []ResourceCost         `json:"top_resources"`
	Period             string                 `json:"period"`
	GeneratedAt        time.Time              `json:"generated_at"`
}

// TrendChart represents cost trend chart data
type TrendChart struct {
	Title       string       `json:"title"`
	Period      int          `json:"period_days"`
	DataPoints  []DataPoint  `json:"data_points"`
	Average     float64      `json:"average"`
	Min         float64      `json:"min"`
	Max         float64      `json:"max"`
	Total       float64      `json:"total"`
	GeneratedAt time.Time    `json:"generated_at"`
}

// DataPoint represents a single data point in a chart
type DataPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
	Label string    `json:"label"`
}

// ServiceBreakdown represents service cost breakdown
type ServiceBreakdown struct {
	AccountID   string        `json:"account_id"`
	Services    []ServiceCost `json:"services"`
	TotalCost   float64       `json:"total_cost"`
	GeneratedAt time.Time     `json:"generated_at"`
}

// ServiceCost represents cost for a specific service
type ServiceCost struct {
	ServiceName string  `json:"service_name"`
	Cost        float64 `json:"cost"`
	Percentage  float64 `json:"percentage"`
	Trend       float64 `json:"trend_percent"`
}

// RegionCost represents cost for a specific region
type RegionCost struct {
	Region     string  `json:"region"`
	Cost       float64 `json:"cost"`
	Percentage float64 `json:"percentage"`
}

// ResourceCost represents cost for a specific resource
type ResourceCost struct {
	ResourceID   string  `json:"resource_id"`
	ResourceType string  `json:"resource_type"`
	Cost         float64 `json:"cost"`
	Region       string  `json:"region"`
}

// DataSource interface for fetching cost data
type DataSource interface {
	GetCurrentMonthCost(ctx context.Context, accountID string) float64
	GetPreviousMonthCost(ctx context.Context, accountID string) float64
	GetServiceBreakdown(ctx context.Context, accountID string) []ServiceCost
	GetRegionBreakdown(ctx context.Context, accountID string) []RegionCost
	GetTopCostResources(ctx context.Context, accountID string, limit int) []ResourceCost
	GetDailyCosts(ctx context.Context, accountID string, days int) []DataPoint
}

// CacheService interface for caching dashboard data
type CacheService interface {
	Get(key string) interface{}
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
}