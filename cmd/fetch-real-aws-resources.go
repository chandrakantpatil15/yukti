package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🚀 FETCHING REAL AWS EC2 RESOURCES")
	fmt.Println("Account: 144403604430 (User: Shruti)")
	fmt.Println(strings.Repeat("=", 50))

	// Database connection
	dsn := "host=localhost user=yukti password=yukti123 dbname=yukti_finops port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Database failed: %v", err)
	}

	// Check if sync is needed (every hour)
	var lastSync time.Time
	db.Raw("SELECT COALESCE(MAX(last_synced), '1970-01-01') FROM yt_aws_resources WHERE sync_status = 'active'").Scan(&lastSync)
	
	if time.Since(lastSync) < time.Hour {
		fmt.Printf("✅ Resources synced recently: %v ago\n", time.Since(lastSync).Round(time.Minute))
		fmt.Println("⏭️  Skipping sync (runs every hour)")
		showCurrentResources(db)
		return
	}

	fmt.Println("🔄 Syncing real AWS EC2 resources...")

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("❌ AWS config failed: %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	fmt.Printf("📍 Region: %s\n", cfg.Region)

	// Get ALL EC2 instances from your AWS account
	fmt.Println("🔍 Fetching ALL EC2 instances from your AWS account...")
	
	result, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		log.Fatalf("❌ Failed to get EC2 instances: %v", err)
	}

	totalInstances := 0
	syncedInstances := 0

	// Mark existing resources as potentially deleted
	db.Exec("UPDATE yt_aws_resources SET sync_status = 'checking' WHERE sync_status = 'active'")

	// Process all reservations and instances
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			totalInstances++
			
			// Extract real instance data
			instanceID := *instance.InstanceId
			instanceType := string(instance.InstanceType)
			state := string(instance.State.Name)
			region := cfg.Region
			az := ""
			if instance.Placement != nil && instance.Placement.AvailabilityZone != nil {
				az = *instance.Placement.AvailabilityZone
			}

			// Extract tags
			tags := make(map[string]string)
			environment := "unknown"
			projectName := ""
			owner := ""
			
			for _, tag := range instance.Tags {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
					
					// Extract common organizational tags
					switch strings.ToLower(*tag.Key) {
					case "environment", "env":
						environment = *tag.Value
					case "project", "project-name":
						projectName = *tag.Value
					case "owner", "created-by":
						owner = *tag.Value
					}
				}
			}

			// Convert tags to JSON
			tagsJSON, _ := json.Marshal(tags)

			// Insert/Update real resource
			sql := `
			INSERT INTO yt_aws_resources 
			(instance_id, instance_type, region, availability_zone, state, platform, architecture, 
			 launch_time, environment, project_name, owner, tags, last_synced, sync_status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), 'active')
			ON CONFLICT (instance_id) 
			DO UPDATE SET 
				instance_type = EXCLUDED.instance_type,
				region = EXCLUDED.region,
				availability_zone = EXCLUDED.availability_zone,
				state = EXCLUDED.state,
				platform = EXCLUDED.platform,
				architecture = EXCLUDED.architecture,
				launch_time = EXCLUDED.launch_time,
				environment = EXCLUDED.environment,
				project_name = EXCLUDED.project_name,
				owner = EXCLUDED.owner,
				tags = EXCLUDED.tags,
				last_synced = NOW(),
				sync_status = 'active'
			`

			platform := "linux"
			if instance.Platform != "" {
				platform = string(instance.Platform)
			}

			architecture := "x86_64"
			if instance.Architecture != "" {
				architecture = string(instance.Architecture)
			}

			err = db.Exec(sql,
				instanceID,
				instanceType,
				region,
				az,
				state,
				platform,
				architecture,
				instance.LaunchTime,
				environment,
				projectName,
				owner,
				string(tagsJSON),
			).Error

			if err != nil {
				fmt.Printf("❌ Failed to save %s: %v\n", instanceID, err)
				continue
			}

			fmt.Printf("✅ %s (%s) - %s [%s]\n", instanceID, instanceType, state, environment)
			syncedInstances++
		}
	}

	// Mark resources not found in AWS as deleted
	db.Exec("UPDATE yt_aws_resources SET sync_status = 'deleted' WHERE sync_status = 'checking'")

	fmt.Printf("\n🎉 SYNC COMPLETE!\n")
	fmt.Printf("📊 Total instances found in AWS: %d\n", totalInstances)
	fmt.Printf("✅ Successfully synced: %d\n", syncedInstances)

	if totalInstances == 0 {
		fmt.Println("\n⚠️  No EC2 instances found in your AWS account")
		fmt.Println("💡 To test with real data:")
		fmt.Println("   1. Launch some EC2 instances in AWS Console")
		fmt.Println("   2. Run this sync again")
		fmt.Println("   3. Add tags like 'Environment=prod' for better organization")
	}

	showCurrentResources(db)
}

func showCurrentResources(db *gorm.DB) {
	var total int64
	db.Raw("SELECT COUNT(*) FROM yt_aws_resources WHERE sync_status = 'active'").Scan(&total)
	fmt.Printf("\n📊 Total active resources: %d\n", total)

	if total == 0 {
		return
	}

	// Show resources by state
	var stateCount []map[string]interface{}
	db.Raw("SELECT state, COUNT(*) as count FROM yt_aws_resources WHERE sync_status = 'active' GROUP BY state ORDER BY count DESC").Scan(&stateCount)

	fmt.Println("\n🔄 Resources by State:")
	for _, s := range stateCount {
		fmt.Printf("  %s: %v instances\n", s["state"], s["count"])
	}

	// Show resources by environment
	var envCount []map[string]interface{}
	db.Raw("SELECT environment, COUNT(*) as count FROM yt_aws_resources WHERE sync_status = 'active' GROUP BY environment ORDER BY count DESC").Scan(&envCount)

	fmt.Println("\n🏷️  Resources by Environment:")
	for _, e := range envCount {
		fmt.Printf("  %s: %v instances\n", e["environment"], e["count"])
	}

	// Show sample resources with pricing
	var resources []map[string]interface{}
	db.Raw(`
		SELECT r.instance_id, r.instance_type, r.state, r.environment, 
		       COALESCE(p.on_demand_price_usd, 0) as hourly_cost,
		       COALESCE(p.on_demand_price_usd * 24 * 30, 0) as monthly_cost
		FROM yt_aws_resources r
		LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
		WHERE r.sync_status = 'active'
		ORDER BY monthly_cost DESC
		LIMIT 10
	`).Scan(&resources)

	if len(resources) > 0 {
		fmt.Println("\n💰 Resources with Cost Estimates:")
		fmt.Println("Instance ID          | Type         | State    | Env     | Monthly Cost")
		fmt.Println(strings.Repeat("-", 75))
		for _, r := range resources {
			fmt.Printf("%-19s | %-12s | %-8s | %-7s | $%.2f\n",
				r["instance_id"],
				r["instance_type"],
				r["state"],
				r["environment"],
				r["monthly_cost"])
		}
	}

	fmt.Println("\n✅ Real AWS resources synced to yt_aws_resources table!")
	fmt.Println("🔗 Table: yt_aws_resources")
	fmt.Println("⏰ Next sync: 1 hour from now")
}