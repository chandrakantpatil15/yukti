package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	_ "github.com/lib/pq"
)

type HealthStatus struct {
	Status        string            `json:"status"`
	Timestamp     string            `json:"timestamp"`
	Version       string            `json:"version"`
	Components    map[string]string `json:"components"`
	TestInstances int               `json:"test_instances"`
	TotalCost     float64           `json:"estimated_hourly_cost"`
	Uptime        string            `json:"uptime"`
}

type KillSwitchStatus struct {
	Enabled   bool   `json:"enabled"`
	Reason    string `json:"reason"`
	Timestamp string `json:"timestamp"`
}

var (
	startTime    = time.Now()
	killSwitch   = false
	killReason   = ""
	killTime     = ""
)

func main() {
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/kill-switch", killSwitchHandler)
	http.HandleFunc("/emergency-stop", emergencyStopHandler)

	port := os.Getenv("HEALTH_PORT")
	if port == "" {
		port = "8081"
	}

	fmt.Printf("🏥 Health Monitor & Kill Switch running on port %s\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /health        - System health check")
	fmt.Println("  POST /kill-switch   - Enable/disable kill switch")
	fmt.Println("  POST /emergency-stop - Emergency stop all test instances")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	health := HealthStatus{
		Status:     "healthy",
		Timestamp:  time.Now().Format(time.RFC3339),
		Version:    "1.0.0",
		Components: make(map[string]string),
		Uptime:     time.Since(startTime).String(),
	}

	// Check kill switch
	if killSwitch {
		health.Status = "kill-switch-enabled"
		health.Components["kill_switch"] = fmt.Sprintf("ENABLED: %s (since %s)", killReason, killTime)
	}

	// Check database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/yukti?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		health.Status = "unhealthy"
		health.Components["database"] = "connection_failed"
	} else {
		defer db.Close()
		if err := db.Ping(); err != nil {
			health.Status = "unhealthy"
			health.Components["database"] = "ping_failed"
		} else {
			health.Components["database"] = "healthy"
		}
	}

	// Check AWS connectivity
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		health.Components["aws"] = "config_failed"
	} else {
		ec2Client := ec2.NewFromConfig(cfg)
		
		// Count test instances and calculate cost
		instances, cost := getTestInstancesInfo(ec2Client)
		health.TestInstances = instances
		health.TotalCost = cost
		
		if instances >= 0 {
			health.Components["aws"] = "healthy"
		} else {
			health.Components["aws"] = "api_failed"
		}
	}

	// Set overall status
	if health.Status == "healthy" {
		for _, status := range health.Components {
			if status != "healthy" {
				health.Status = "degraded"
				break
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if health.Status == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else if health.Status == "kill-switch-enabled" {
		w.WriteHeader(http.StatusLocked)
	}
	
	json.NewEncoder(w).Encode(health)
}

func killSwitchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		status := KillSwitchStatus{
			Enabled:   killSwitch,
			Reason:    killReason,
			Timestamp: killTime,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Enable bool   `json:"enable"`
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	killSwitch = req.Enable
	killReason = req.Reason
	killTime = time.Now().Format(time.RFC3339)

	if killSwitch {
		log.Printf("🚨 KILL SWITCH ENABLED: %s", killReason)
		
		// Auto-cleanup test instances when kill switch is enabled
		go func() {
			if err := emergencyCleanup(); err != nil {
				log.Printf("❌ Emergency cleanup failed: %v", err)
			} else {
				log.Println("✅ Emergency cleanup completed")
			}
		}()
	} else {
		log.Println("✅ Kill switch disabled")
	}

	status := KillSwitchStatus{
		Enabled:   killSwitch,
		Reason:    killReason,
		Timestamp: killTime,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func emergencyStopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("🚨 EMERGENCY STOP TRIGGERED")

	// Enable kill switch
	killSwitch = true
	killReason = "Emergency stop triggered"
	killTime = time.Now().Format(time.RFC3339)

	// Cleanup instances
	if err := emergencyCleanup(); err != nil {
		log.Printf("❌ Emergency cleanup failed: %v", err)
		http.Error(w, fmt.Sprintf("Emergency cleanup failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":    "emergency_stop_completed",
		"timestamp": time.Now().Format(time.RFC3339),
		"message":   "All test instances terminated",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getTestInstancesInfo(client *ec2.Client) (int, float64) {
	result, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   &[]string{"tag:Project"}[0],
				Values: []string{"yukti-finops"},
			},
			{
				Name:   &[]string{"instance-state-name"}[0],
				Values: []string{"running", "pending"},
			},
		},
	})
	if err != nil {
		return -1, 0
	}

	count := 0
	totalCost := 0.0
	
	// Rough cost estimates (per hour)
	costMap := map[string]float64{
		"t3.micro":  0.0104,
		"t3.small":  0.0208,
		"t3.medium": 0.0416,
		"m5.large":  0.096,
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			count++
			if cost, exists := costMap[string(instance.InstanceType)]; exists {
				totalCost += cost
			}
		}
	}

	return count, totalCost
}

func emergencyCleanup() error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("AWS config failed: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Get test instances
	result, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   &[]string{"tag:Project"}[0],
				Values: []string{"yukti-finops"},
			},
			{
				Name:   &[]string{"instance-state-name"}[0],
				Values: []string{"running", "pending"},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to describe instances: %w", err)
	}

	var instanceIds []string
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instanceIds = append(instanceIds, *instance.InstanceId)
		}
	}

	if len(instanceIds) == 0 {
		log.Println("ℹ️  No test instances to cleanup")
		return nil
	}

	// Terminate instances
	_, err = ec2Client.TerminateInstances(context.TODO(), &ec2.TerminateInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return fmt.Errorf("failed to terminate instances: %w", err)
	}

	log.Printf("🧹 Terminated %d test instances: %v", len(instanceIds), instanceIds)
	return nil
}