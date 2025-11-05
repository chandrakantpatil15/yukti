package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type LogEntry struct {
	Timestamp    time.Time
	Message      string
	Level        string
	SourceIP     *string
	Hostname     *string
	LogSource    string
}

type ResourceMatch struct {
	InstanceID      string
	ConfidenceScore int
	HourlyCost      float64
	InstanceType    string
	Environment     string
}

func main() {
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

	// Example log entries from different sources
	sampleLogs := []LogEntry{
		{
			Timestamp: time.Now(),
			Message:   "Application started on instance i-0123456789abcdef0",
			Level:     "INFO",
			LogSource: "application",
		},
		{
			Timestamp: time.Now(),
			Message:   "High CPU usage detected",
			Level:     "WARN",
			SourceIP:  stringPtr("10.0.1.100"),
			LogSource: "cloudwatch",
		},
		{
			Timestamp: time.Now(),
			Message:   "Database connection established",
			Level:     "INFO",
			Hostname:  stringPtr("web-server-prod"),
			LogSource: "application",
		},
		{
			Timestamp: time.Now(),
			Message:   "Error processing request: timeout",
			Level:     "ERROR",
			SourceIP:  stringPtr("10.0.1.101"),
			LogSource: "nginx",
		},
	}

	fmt.Println("=== Log Correlation and Cost Attribution Demo ===\n")

	for i, logEntry := range sampleLogs {
		fmt.Printf("Log Entry #%d:\n", i+1)
		fmt.Printf("  Message: %s\n", logEntry.Message)
		fmt.Printf("  Level: %s\n", logEntry.Level)
		if logEntry.SourceIP != nil {
			fmt.Printf("  Source IP: %s\n", *logEntry.SourceIP)
		}
		if logEntry.Hostname != nil {
			fmt.Printf("  Hostname: %s\n", *logEntry.Hostname)
		}

		// Find matching resource
		match, err := findResourceFromLog(db, logEntry)
		if err != nil {
			fmt.Printf("  Error: %v\n\n", err)
			continue
		}

		if match != nil {
			fmt.Printf("  ✓ Matched Resource:\n")
			fmt.Printf("    Instance ID: %s\n", match.InstanceID)
			fmt.Printf("    Instance Type: %s\n", match.InstanceType)
			fmt.Printf("    Environment: %s\n", match.Environment)
			fmt.Printf("    Confidence: %d%%\n", match.ConfidenceScore)
			fmt.Printf("    Hourly Cost: $%.4f\n", match.HourlyCost)

			// Calculate cost attribution for this log entry
			costPerMinute := match.HourlyCost / 60
			fmt.Printf("    Cost per minute: $%.6f\n", costPerMinute)

			// Store log entry with resource correlation
			if err := storeLogEntry(db, logEntry, match); err != nil {
				fmt.Printf("    Error storing log: %v\n", err)
			} else {
				fmt.Printf("    ✓ Log entry stored with cost attribution\n")
			}
		} else {
			fmt.Printf("  ✗ No matching resource found\n")
		}
		fmt.Println()
	}

	// Show cost analysis summary
	showCostAnalysis(db)
}

func findResourceFromLog(db *sql.DB, logEntry LogEntry) (*ResourceMatch, error) {
	// Try different identification methods in order of confidence

	// Method 1: Direct instance ID in log message
	if instanceID := extractInstanceID(logEntry.Message); instanceID != "" {
		return findResourceByInstanceID(db, instanceID)
	}

	// Method 2: Match by IP address
	if logEntry.SourceIP != nil {
		if match, err := findResourceByIP(db, *logEntry.SourceIP); err == nil && match != nil {
			return match, nil
		}
	}

	// Method 3: Match by hostname
	if logEntry.Hostname != nil {
		if match, err := findResourceByHostname(db, *logEntry.Hostname); err == nil && match != nil {
			return match, nil
		}
	}

	// Method 4: Match by application/service name in log
	if match, err := findResourceByLogContent(db, logEntry.Message); err == nil && match != nil {
		return match, nil
	}

	return nil, nil
}

func extractInstanceID(message string) string {
	// Regex to match AWS instance IDs
	re := regexp.MustCompile(`i-[0-9a-f]{8,17}`)
	matches := re.FindStringSubmatch(message)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

func findResourceByInstanceID(db *sql.DB, instanceID string) (*ResourceMatch, error) {
	query := `
		SELECT r.instance_id, r.instance_type, r.environment, 
		       COALESCE(p.on_demand_price_usd, 0) as hourly_cost
		FROM yt_aws_resources r
		LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
		WHERE r.instance_id = $1 AND r.sync_status = 'active'`

	var match ResourceMatch
	err := db.QueryRow(query, instanceID).Scan(
		&match.InstanceID, &match.InstanceType, &match.Environment, &match.HourlyCost)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	match.ConfidenceScore = 95 // High confidence for direct instance ID match
	return &match, nil
}

func findResourceByIP(db *sql.DB, ip string) (*ResourceMatch, error) {
	query := `
		SELECT r.instance_id, r.instance_type, COALESCE(r.environment, '') as environment,
		       COALESCE(p.on_demand_price_usd, 0) as hourly_cost
		FROM yt_aws_resources r
		LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
		WHERE (r.private_ip = $1 OR r.public_ip = $1) AND r.sync_status = 'active'
		LIMIT 1`

	var match ResourceMatch
	err := db.QueryRow(query, ip).Scan(
		&match.InstanceID, &match.InstanceType, &match.Environment, &match.HourlyCost)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	match.ConfidenceScore = 85 // High confidence for IP match
	return &match, nil
}

func findResourceByHostname(db *sql.DB, hostname string) (*ResourceMatch, error) {
	query := `
		SELECT r.instance_id, r.instance_type, COALESCE(r.environment, '') as environment,
		       COALESCE(p.on_demand_price_usd, 0) as hourly_cost
		FROM yt_aws_resources r
		LEFT JOIN yt_resource_identifiers ri ON r.instance_id = ri.instance_id
		LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
		WHERE ri.hostname = $1 AND r.sync_status = 'active'
		LIMIT 1`

	var match ResourceMatch
	err := db.QueryRow(query, hostname).Scan(
		&match.InstanceID, &match.InstanceType, &match.Environment, &match.HourlyCost)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	match.ConfidenceScore = 75 // Medium confidence for hostname match
	return &match, nil
}

func findResourceByLogContent(db *sql.DB, message string) (*ResourceMatch, error) {
	// Look for application/service names in the log message
	query := `
		SELECT r.instance_id, r.instance_type, COALESCE(r.environment, '') as environment,
		       COALESCE(p.on_demand_price_usd, 0) as hourly_cost
		FROM yt_aws_resources r
		LEFT JOIN yt_resource_identifiers ri ON r.instance_id = ri.instance_id
		LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
		WHERE ri.application_name IS NOT NULL 
		AND $1 ILIKE '%' || ri.application_name || '%'
		AND r.sync_status = 'active'
		LIMIT 1`

	var match ResourceMatch
	err := db.QueryRow(query, message).Scan(
		&match.InstanceID, &match.InstanceType, &match.Environment, &match.HourlyCost)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	match.ConfidenceScore = 60 // Lower confidence for content-based match
	return &match, nil
}

func storeLogEntry(db *sql.DB, logEntry LogEntry, match *ResourceMatch) error {
	query := `
		INSERT INTO yt_log_entries 
		(timestamp, instance_id, log_source, message, level, source_ip, hostname, 
		 confidence_score, attributed_cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	// Calculate attributed cost (cost per minute)
	attributedCost := match.HourlyCost / 60

	_, err := db.Exec(query,
		logEntry.Timestamp, match.InstanceID, logEntry.LogSource,
		logEntry.Message, logEntry.Level, logEntry.SourceIP, logEntry.Hostname,
		match.ConfidenceScore, attributedCost)

	return err
}

func showCostAnalysis(db *sql.DB) {
	fmt.Println("=== Cost Analysis Summary ===")

	query := `
		SELECT 
			r.instance_id,
			r.instance_type,
			r.environment,
			COUNT(l.id) as log_count,
			SUM(l.attributed_cost) as total_attributed_cost,
			p.on_demand_price_usd as hourly_cost
		FROM yt_aws_resources r
		LEFT JOIN yt_log_entries l ON r.instance_id = l.instance_id
		LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
		WHERE r.sync_status = 'active'
		GROUP BY r.instance_id, r.instance_type, r.environment, p.on_demand_price_usd
		ORDER BY total_attributed_cost DESC NULLS LAST`

	rows, err := db.Query(query)
	if err != nil {
		fmt.Printf("Error querying cost analysis: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-20s %-15s %-12s %-10s %-15s %-12s\n",
		"Instance ID", "Type", "Environment", "Log Count", "Attributed Cost", "Hourly Cost")
	fmt.Println(strings.Repeat("-", 90))

	for rows.Next() {
		var instanceID, instanceType, environment string
		var logCount int
		var attributedCost, hourlyCost sql.NullFloat64

		err := rows.Scan(&instanceID, &instanceType, &environment,
			&logCount, &attributedCost, &hourlyCost)
		if err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			continue
		}

		fmt.Printf("%-20s %-15s %-12s %-10d $%-14.6f $%-11.4f\n",
			instanceID, instanceType, environment, logCount,
			attributedCost.Float64, hourlyCost.Float64)
	}
}

func stringPtr(s string) *string {
	return &s
}