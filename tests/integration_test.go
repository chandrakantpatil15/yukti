package tests

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/yukti?sslmode=disable"
	}

	var err error
	testDB, err = sql.Open("postgres", dbURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}
}

func teardown() {
	if testDB != nil {
		testDB.Close()
	}
}

func TestDatabaseConnectivity(t *testing.T) {
	t.Log("🔍 Testing database connectivity...")
	
	err := testDB.Ping()
	if err != nil {
		t.Fatalf("❌ Database ping failed: %v", err)
	}
	
	t.Log("✅ Database connectivity test passed")
}

func TestRealResourceData(t *testing.T) {
	t.Log("🔍 Testing real resource data integrity...")
	
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM yt_aws_resources WHERE sync_status = 'active'").Scan(&count)
	if err != nil {
		t.Fatalf("❌ Failed to query resources: %v", err)
	}
	
	if count == 0 {
		t.Error("❌ No active resources found - run sync-resources first")
	} else {
		t.Logf("✅ Found %d active resources", count)
	}
}

func TestPricingDataIntegrity(t *testing.T) {
	t.Log("🔍 Testing pricing data integrity...")
	
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM yt_aws_pricing WHERE on_demand_price_usd > 0").Scan(&count)
	if err != nil {
		t.Fatalf("❌ Failed to query pricing: %v", err)
	}
	
	if count < 1000 {
		t.Errorf("❌ Insufficient pricing data: %d records (expected >1000)", count)
	} else {
		t.Logf("✅ Pricing data integrity verified: %d records", count)
	}
}

func TestAssessmentClassification(t *testing.T) {
	t.Log("🔍 Testing assessment classification accuracy...")
	
	testCases := []struct {
		name     string
		avgCPU   float64
		maxCPU   float64
		avgMem   float64
		variance float64
		expected string
	}{
		{"Idle Resource", 2.0, 5.0, 8.0, 1.0, "idle"},
		{"Underutilized", 15.0, 25.0, 20.0, 5.0, "underutilized"},
		{"Batch Workload", 25.0, 95.0, 40.0, 30.0, "batch"},
		{"Overutilized", 85.0, 95.0, 88.0, 8.0, "overutilized"},
	}
	
	correctCount := 0
	for _, tc := range testCases {
		result := classifyUtilization(tc.avgCPU, tc.maxCPU, tc.avgMem, tc.variance)
		if result == tc.expected {
			correctCount++
			t.Logf("✅ %s: %s (correct)", tc.name, result)
		} else {
			t.Errorf("❌ %s: got %s, expected %s", tc.name, result, tc.expected)
		}
	}
	
	accuracy := float64(correctCount) / float64(len(testCases)) * 100
	if accuracy < 90.0 {
		t.Errorf("❌ Classification accuracy too low: %.1f%% (expected ≥90%%)", accuracy)
	} else {
		t.Logf("✅ Classification accuracy: %.1f%%", accuracy)
	}
}

func TestAPIEndpoints(t *testing.T) {
	t.Log("🔍 Testing API endpoints...")
	
	resp, err := http.Get("http://localhost:8080/api/v1/health")
	if err != nil {
		t.Skip("⚠️ API server not running, skipping API tests")
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("❌ Health check failed: status %d", resp.StatusCode)
	} else {
		t.Log("✅ Health endpoint working")
	}
}

func TestCostCalculations(t *testing.T) {
	t.Log("🔍 Testing cost calculations...")
	
	query := `
		SELECT r.instance_type, p.on_demand_price_usd,
		       (p.on_demand_price_usd * 24 * 30) as monthly_cost
		FROM yt_aws_resources r
		JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
		WHERE r.sync_status = 'active' AND r.state = 'running'
		LIMIT 5`
	
	rows, err := testDB.Query(query)
	if err != nil {
		t.Fatalf("❌ Cost calculation query failed: %v", err)
	}
	defer rows.Close()
	
	totalCost := 0.0
	resourceCount := 0
	
	for rows.Next() {
		var instanceType string
		var hourlyCost, monthlyCost float64
		
		err := rows.Scan(&instanceType, &hourlyCost, &monthlyCost)
		if err != nil {
			t.Errorf("❌ Failed to scan cost data: %v", err)
			continue
		}
		
		if hourlyCost <= 0 || monthlyCost <= 0 {
			t.Errorf("❌ Invalid cost calculation for %s: hourly=%.4f, monthly=%.2f", 
				instanceType, hourlyCost, monthlyCost)
		}
		
		totalCost += monthlyCost
		resourceCount++
		t.Logf("💰 %s: $%.4f/hour → $%.2f/month", instanceType, hourlyCost, monthlyCost)
	}
	
	if resourceCount == 0 {
		t.Error("❌ No cost data found - check pricing integration")
	} else {
		t.Logf("✅ Cost calculations verified for %d resources (Total: $%.2f/month)", 
			resourceCount, totalCost)
	}
}

func classifyUtilization(avgCPU, maxCPU, avgMemory, cpuVariance float64) string {
	if maxCPU > 80 && avgCPU < 30 && cpuVariance > 25 {
		return "batch"
	}
	if avgCPU > 80 || avgMemory > 80 {
		return "overutilized"
	}
	if avgCPU < 20 && avgMemory < 25 {
		return "underutilized"
	}
	if cpuVariance > 20 {
		return "intermittent"
	}
	if avgCPU < 5 && avgMemory < 10 {
		return "idle"
	}
	return "normal"
}