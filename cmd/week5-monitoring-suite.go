package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"yukti/internal/monitoring"
)

// Week5MonitoringSuite demonstrates advanced monitoring and reporting features
func main() {
	log.Println("🎯 Week 5: Advanced Monitoring Suite (Premium Features)")
	log.Println("=======================================================")
	
	ctx := context.Background()
	
	// Initialize services
	dataSource := NewMockDataSource()
	cache := NewMockCache()
	notifier := NewMockNotifier()
	
	dashboardSvc := monitoring.NewDashboardService(dataSource, cache)
	alertingSvc := monitoring.NewAlertingService(notifier, dataSource)
	budgetSvc := monitoring.NewBudgetService(dataSource)
	reportingSvc := monitoring.NewReportingService(dataSource)
	
	accountID := "acc-12345"
	
	// Demonstrate Real-time Dashboard
	log.Println("📊 Generating Real-time Cost Dashboard...")
	dashboard, _ := dashboardSvc.GetRealTimeCostDashboard(ctx, accountID)
	displayDashboard(dashboard)
	
	// Demonstrate Cost Trend Chart
	log.Println("📈 Generating Cost Trend Chart...")
	trendChart, _ := dashboardSvc.GetCostTrendChart(ctx, accountID, 30)
	displayTrendChart(trendChart)
	
	// Demonstrate Custom Alerting
	log.Println("🚨 Setting up Custom Alerts...")
	setupAlerts(alertingSvc, accountID)
	alerts, _ := alertingSvc.EvaluateAlerts(ctx, accountID)
	displayAlerts(alerts)
	
	// Demonstrate Budget Tracking
	log.Println("💰 Creating and Tracking Budgets...")
	setupBudgets(budgetSvc, accountID)
	budgets, _ := budgetSvc.GetAllBudgets(ctx, accountID)
	displayBudgets(budgets)
	
	// Demonstrate Budget Forecast
	log.Println("🔮 Generating Budget Forecast...")
	if len(budgets) > 0 {
		forecast, _ := budgetSvc.GetBudgetForecast(ctx, budgets[0].BudgetID, 30)
		displayForecast(forecast)
	}
	
	// Demonstrate Executive Reporting
	log.Println("📋 Generating Executive Report...")
	execReport, _ := reportingSvc.GenerateExecutiveReport(ctx, accountID)
	displayExecutiveReport(execReport)
	
	// Demonstrate Cost Allocation
	log.Println("📊 Generating Cost Allocation Report...")
	allocationReport, _ := reportingSvc.GenerateCostAllocationReport(ctx, accountID)
	displayAllocationReport(allocationReport)
	
	// Demonstrate Chargeback Report
	log.Println("💳 Generating Chargeback Report...")
	chargebackReport, _ := reportingSvc.GenerateChargebackReport(ctx, accountID)
	displayChargebackReport(chargebackReport)
	
	// Generate comprehensive summary
	generateWeek5Summary()
	
	log.Println("✅ Week 5 Advanced Monitoring Suite Complete!")
}

// displayDashboard displays dashboard data
func displayDashboard(dashboard *monitoring.CostDashboard) {
	log.Println("\n📊 Real-time Cost Dashboard:")
	log.Println("============================")
	log.Printf("Current Month Cost: $%.2f", dashboard.CurrentMonthCost)
	log.Printf("Previous Month Cost: $%.2f", dashboard.PreviousMonthCost)
	log.Printf("Cost Trend: %.2f%%", dashboard.CostTrend)
	log.Printf("Daily Average: $%.2f", dashboard.DailyAverage)
	log.Printf("Forecasted Month End: $%.2f", dashboard.ForecastedMonthEnd)
	
	log.Println("\nTop 5 Services by Cost:")
	for i, svc := range dashboard.ServiceBreakdown {
		if i >= 5 {
			break
		}
		log.Printf("  %d. %s: $%.2f", i+1, svc.ServiceName, svc.Cost)
	}
	
	log.Println("\nTop 5 Regions by Cost:")
	for i, region := range dashboard.RegionBreakdown {
		if i >= 5 {
			break
		}
		log.Printf("  %d. %s: $%.2f (%.1f%%)", i+1, region.Region, region.Cost, region.Percentage)
	}
}

// displayTrendChart displays trend chart data
func displayTrendChart(chart *monitoring.TrendChart) {
	log.Println("\n📈 Cost Trend Chart (Last 30 Days):")
	log.Println("====================================")
	log.Printf("Average: $%.2f", chart.Average)
	log.Printf("Min: $%.2f", chart.Min)
	log.Printf("Max: $%.2f", chart.Max)
	log.Printf("Total: $%.2f", chart.Total)
	
	log.Println("\nSample Data Points (First 5 days):")
	for i, point := range chart.DataPoints {
		if i >= 5 {
			break
		}
		log.Printf("  Day %d: $%.2f", i+1, point.Value)
	}
}

// setupAlerts sets up sample alert rules
func setupAlerts(svc *monitoring.AlertingService, accountID string) {
	rules := []*monitoring.AlertRule{
		{
			AccountID: accountID,
			Name:      "Daily Cost Threshold",
			Type:      "daily_cost",
			Threshold: 1000.00,
			Operator:  "greater_than",
			Severity:  "warning",
		},
		{
			AccountID: accountID,
			Name:      "Monthly Budget Alert",
			Type:      "monthly_cost",
			Threshold: 25000.00,
			Operator:  "greater_than",
			Severity:  "critical",
		},
		{
			AccountID:   accountID,
			Name:        "EC2 Cost Spike",
			Type:        "service_cost",
			ServiceName: "EC2",
			Threshold:   5000.00,
			Operator:    "greater_than",
			Severity:    "warning",
		},
	}
	
	for _, rule := range rules {
		svc.CreateAlertRule(rule)
	}
	
	log.Printf("✅ Created %d alert rules", len(rules))
}

// displayAlerts displays triggered alerts
func displayAlerts(alerts []*monitoring.Alert) {
	log.Println("\n🚨 Triggered Alerts:")
	log.Println("====================")
	
	if len(alerts) == 0 {
		log.Println("No alerts triggered")
		return
	}
	
	for i, alert := range alerts {
		log.Printf("%d. [%s] %s", i+1, alert.Severity, alert.RuleName)
		log.Printf("   Current: $%.2f | Threshold: $%.2f", alert.CurrentValue, alert.Threshold)
		log.Printf("   Message: %s", alert.Message)
	}
}

// setupBudgets sets up sample budgets
func setupBudgets(svc *monitoring.BudgetService, accountID string) {
	budgets := []*monitoring.Budget{
		{
			AccountID: accountID,
			Name:      "Monthly Cloud Budget",
			Amount:    30000.00,
			Period:    "monthly",
			StartDate: time.Now().AddDate(0, 0, -time.Now().Day()+1),
			EndDate:   time.Now().AddDate(0, 1, -time.Now().Day()),
		},
		{
			AccountID: accountID,
			Name:      "Development Environment Budget",
			Amount:    5000.00,
			Period:    "monthly",
			StartDate: time.Now().AddDate(0, 0, -time.Now().Day()+1),
			EndDate:   time.Now().AddDate(0, 1, -time.Now().Day()),
		},
	}
	
	for _, budget := range budgets {
		svc.CreateBudget(budget)
	}
	
	log.Printf("✅ Created %d budgets", len(budgets))
}

// displayBudgets displays budget status
func displayBudgets(budgets []*monitoring.BudgetStatus) {
	log.Println("\n💰 Budget Status:")
	log.Println("=================")
	
	for i, budget := range budgets {
		log.Printf("\n%d. %s", i+1, budget.BudgetName)
		log.Printf("   Budget: $%.2f", budget.BudgetAmount)
		log.Printf("   Current Spend: $%.2f (%.1f%% used)", budget.CurrentSpend, budget.PercentUsed)
		log.Printf("   Remaining: $%.2f", budget.RemainingBudget)
		log.Printf("   Forecasted: $%.2f", budget.ForecastedSpend)
		log.Printf("   Status: %s [%s]", budget.Status, budget.AlertLevel)
		
		if budget.ForecastedOverrun > 0 {
			log.Printf("   ⚠️  Forecasted Overrun: $%.2f", budget.ForecastedOverrun)
		}
	}
}

// displayForecast displays budget forecast
func displayForecast(forecast *monitoring.BudgetForecast) {
	log.Println("\n🔮 Budget Forecast (Next 30 Days):")
	log.Println("===================================")
	log.Printf("Budget Amount: $%.2f", forecast.BudgetAmount)
	log.Printf("Current Spend: $%.2f", forecast.CurrentSpend)
	log.Printf("Daily Average: $%.2f", forecast.DailyAverage)
	
	log.Println("\nProjections (First 10 days):")
	for i, proj := range forecast.Projections {
		if i >= 10 {
			break
		}
		status := "✅"
		if proj.OverBudget {
			status = "❌"
		}
		log.Printf("  Day %d: $%.2f (Remaining: $%.2f) %s", proj.Day, proj.ProjectedSpend, proj.BudgetRemaining, status)
	}
}

// displayExecutiveReport displays executive report
func displayExecutiveReport(report *monitoring.ExecutiveReport) {
	log.Println("\n📋 Executive Report:")
	log.Println("====================")
	log.Printf("Period: %s", report.Period)
	log.Printf("Current Month: $%.2f", report.CurrentMonthCost)
	log.Printf("Previous Month: $%.2f", report.PreviousMonthCost)
	log.Printf("Change: %.2f%%", report.MonthOverMonthChange)
	
	log.Println("\nKey Performance Indicators:")
	for i, kpi := range report.KPIs {
		log.Printf("  %d. %s: %.2f %s [%s]", i+1, kpi.Name, kpi.Value, kpi.Unit, kpi.Status)
	}
	
	log.Println("\nOptimization Opportunities:")
	log.Printf("  Total: %d opportunities", report.OptimizationOpportunities.TotalOpportunities)
	log.Printf("  Potential Savings: $%.2f", report.OptimizationOpportunities.PotentialSavings)
	log.Printf("  High Priority: %d", report.OptimizationOpportunities.HighPriority)
}

// displayAllocationReport displays cost allocation report
func displayAllocationReport(report *monitoring.CostAllocationReport) {
	log.Println("\n📊 Cost Allocation Report:")
	log.Println("==========================")
	log.Printf("Total Cost: $%.2f", report.TotalCost)
	
	log.Println("\nBy Environment:")
	for _, item := range report.ByEnvironment {
		log.Printf("  %s: $%.2f (%.1f%%)", item.Name, item.Cost, item.Percentage)
	}
	
	log.Println("\nBy Project:")
	for _, item := range report.ByProject {
		log.Printf("  %s: $%.2f (%.1f%%)", item.Name, item.Cost, item.Percentage)
	}
}

// displayChargebackReport displays chargeback report
func displayChargebackReport(report *monitoring.ChargebackReport) {
	log.Println("\n💳 Chargeback Report:")
	log.Println("=====================")
	log.Printf("Total Chargeback: $%.2f", report.TotalChargeback)
	
	log.Println("\nBy Team:")
	for i, cb := range report.Chargebacks {
		log.Printf("  %d. %s: $%.2f (%d resources)", i+1, cb.Team, cb.Amount, cb.Resources)
	}
}

// generateWeek5Summary generates comprehensive Week 5 summary
func generateWeek5Summary() {
	summary := Week5Summary{
		Week:                 5,
		Phase:                "Advanced Monitoring Suite (Premium Features)",
		Status:               "COMPLETE",
		FeaturesImplemented:  7,
		PremiumCapabilities:  []string{
			"Real-time cost dashboards",
			"Custom alerting system",
			"Budget tracking and forecasting",
			"Executive reporting",
			"Cost allocation reports",
			"Chargeback/showback reports",
			"Multi-account support",
		},
		BusinessValue:        "Premium subscription tier with advanced monitoring",
		TargetCustomers:      "Enterprise customers requiring advanced cost visibility",
		MonthlySubscription:  99.00,
		CompletedAt:          time.Now(),
	}
	
	summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
	
	log.Println("\n📋 Week 5 Summary:")
	log.Println("==================")
	fmt.Println(string(summaryJSON))
	
	log.Println("\n🎯 Premium Features Delivered:")
	log.Println("==============================")
	for i, feature := range summary.PremiumCapabilities {
		log.Printf("%d. %s", i+1, feature)
	}
	
	log.Println("\n💰 Subscription Model:")
	log.Println("======================")
	log.Printf("Professional Tier: $%.2f/month", summary.MonthlySubscription)
	log.Println("Includes:")
	log.Println("  • Real-time monitoring dashboards")
	log.Println("  • Custom cost alerts")
	log.Println("  • Budget tracking")
	log.Println("  • Executive reports")
	log.Println("  • Multi-account support")
	
	log.Println("\n🔮 Week 6 Preview:")
	log.Println("==================")
	log.Println("• Compliance & governance auditing")
	log.Println("• AWS Well-Architected Framework checks")
	log.Println("• Security best practices compliance")
	log.Println("• Resource tagging compliance")
	log.Println("• Policy enforcement engine")
}

// Week5Summary represents Week 5 completion summary
type Week5Summary struct {
	Week                int       `json:"week"`
	Phase               string    `json:"phase"`
	Status              string    `json:"status"`
	FeaturesImplemented int       `json:"features_implemented"`
	PremiumCapabilities []string  `json:"premium_capabilities"`
	BusinessValue       string    `json:"business_value"`
	TargetCustomers     string    `json:"target_customers"`
	MonthlySubscription float64   `json:"monthly_subscription"`
	CompletedAt         time.Time `json:"completed_at"`
}

// Mock implementations
type MockDataSource struct{}

func NewMockDataSource() *MockDataSource {
	return &MockDataSource{}
}

func (m *MockDataSource) GetCurrentMonthCost(ctx context.Context, accountID string) float64 {
	return 25000.00 + rand.Float64()*5000
}

func (m *MockDataSource) GetPreviousMonthCost(ctx context.Context, accountID string) float64 {
	return 22000.00 + rand.Float64()*3000
}

func (m *MockDataSource) GetServiceBreakdown(ctx context.Context, accountID string) []monitoring.ServiceCost {
	return []monitoring.ServiceCost{
		{ServiceName: "EC2", Cost: 8000.00, Trend: 5.2},
		{ServiceName: "RDS", Cost: 6000.00, Trend: -2.1},
		{ServiceName: "S3", Cost: 4000.00, Trend: 1.5},
		{ServiceName: "Lambda", Cost: 3000.00, Trend: 8.3},
		{ServiceName: "CloudFront", Cost: 2500.00, Trend: 3.7},
	}
}

func (m *MockDataSource) GetRegionBreakdown(ctx context.Context, accountID string) []monitoring.RegionCost {
	return []monitoring.RegionCost{
		{Region: "us-east-1", Cost: 12000.00, Percentage: 48},
		{Region: "us-west-2", Cost: 8000.00, Percentage: 32},
		{Region: "eu-west-1", Cost: 5000.00, Percentage: 20},
	}
}

func (m *MockDataSource) GetTopCostResources(ctx context.Context, accountID string, limit int) []monitoring.ResourceCost {
	return []monitoring.ResourceCost{
		{ResourceID: "i-abc123", ResourceType: "EC2", Cost: 500.00, Region: "us-east-1"},
		{ResourceID: "db-xyz789", ResourceType: "RDS", Cost: 450.00, Region: "us-west-2"},
		{ResourceID: "bucket-data", ResourceType: "S3", Cost: 300.00, Region: "us-east-1"},
	}
}

func (m *MockDataSource) GetDailyCosts(ctx context.Context, accountID string, days int) []monitoring.DataPoint {
	points := make([]monitoring.DataPoint, days)
	baseDate := time.Now().AddDate(0, 0, -days)
	
	for i := 0; i < days; i++ {
		points[i] = monitoring.DataPoint{
			Date:  baseDate.AddDate(0, 0, i),
			Value: 800.00 + rand.Float64()*200,
			Label: fmt.Sprintf("Day %d", i+1),
		}
	}
	
	return points
}

type MockCache struct {
	data map[string]interface{}
}

func NewMockCache() *MockCache {
	return &MockCache{data: make(map[string]interface{})}
}

func (m *MockCache) Get(key string) interface{} {
	return m.data[key]
}

func (m *MockCache) Set(key string, value interface{}, ttl time.Duration) {
	m.data[key] = value
}

func (m *MockCache) Delete(key string) {
	delete(m.data, key)
}

type MockNotifier struct{}

func NewMockNotifier() *MockNotifier {
	return &MockNotifier{}
}

func (m *MockNotifier) SendAlert(ctx context.Context, alert *monitoring.Alert) error {
	log.Printf("📧 Alert sent: [%s] %s", alert.Severity, alert.Message)
	return nil
}