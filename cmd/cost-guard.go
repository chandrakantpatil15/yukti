package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Cost Guard - Automatic cost protection and monitoring
func main() {
	fmt.Println("💰 YUKTI FINOPS - COST GUARD")
	fmt.Println("===========================")

	// Configuration
	maxHourlyCost := getEnvFloat("MAX_HOURLY_COST", 1.0)  // $1/hour default
	maxDailyCost := getEnvFloat("MAX_DAILY_COST", 20.0)   // $20/day default
	checkInterval := getEnvDuration("CHECK_INTERVAL", 5*time.Minute)

	fmt.Printf("🛡️  Cost Protection Limits:\n")
	fmt.Printf("   Max Hourly Cost: $%.2f\n", maxHourlyCost)
	fmt.Printf("   Max Daily Cost:  $%.2f\n", maxDailyCost)
	fmt.Printf("   Check Interval:  %v\n", checkInterval)

	// AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("❌ AWS config failed:", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Start monitoring loop
	for {
		if err := checkCostLimits(ec2Client, maxHourlyCost, maxDailyCost); err != nil {
			log.Printf("❌ Cost check error: %v", err)
		}
		time.Sleep(checkInterval)
	}
}

func checkCostLimits(client *ec2.Client, maxHourly, maxDaily float64) error {
	// Get all test instances
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
		return fmt.Errorf("failed to describe instances: %w", err)
	}

	// Calculate current costs
	currentHourlyCost := 0.0
	instanceCount := 0
	var instanceIds []string

	costMap := map[string]float64{
		"t3.micro":   0.0104,
		"t3.small":   0.0208,
		"t3.medium":  0.0416,
		"t3.large":   0.0832,
		"m5.large":   0.096,
		"m5.xlarge":  0.192,
		"c5.large":   0.085,
		"r5.large":   0.126,
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instanceCount++
			instanceIds = append(instanceIds, *instance.InstanceId)
			
			if cost, exists := costMap[string(instance.InstanceType)]; exists {
				currentHourlyCost += cost
			}
		}
	}

	currentDailyCost := currentHourlyCost * 24

	// Log current status
	fmt.Printf("\n⏰ %s - Cost Check\n", time.Now().Format("15:04:05"))
	fmt.Printf("   Running Instances: %d\n", instanceCount)
	fmt.Printf("   Current Hourly Cost: $%.4f\n", currentHourlyCost)
	fmt.Printf("   Projected Daily Cost: $%.2f\n", currentDailyCost)

	// Check limits
	if currentHourlyCost > maxHourly {
		log.Printf("🚨 HOURLY COST LIMIT EXCEEDED: $%.4f > $%.2f", currentHourlyCost, maxHourly)
		return triggerEmergencyStop(client, instanceIds, "Hourly cost limit exceeded")
	}

	if currentDailyCost > maxDaily {
		log.Printf("🚨 DAILY COST LIMIT EXCEEDED: $%.2f > $%.2f", currentDailyCost, maxDaily)
		return triggerEmergencyStop(client, instanceIds, "Daily cost limit exceeded")
	}

	// Check for runaway instances (running > 4 hours)
	longRunningCount := 0
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if instance.LaunchTime != nil {
				uptime := time.Since(*instance.LaunchTime)
				if uptime > 4*time.Hour {
					longRunningCount++
					log.Printf("⚠️  Long-running instance: %s (uptime: %v)", *instance.InstanceId, uptime.Round(time.Minute))
				}
			}
		}
	}

	if longRunningCount > 0 {
		log.Printf("⚠️  %d instances running > 4 hours", longRunningCount)
	}

	if instanceCount == 0 {
		fmt.Printf("   ✅ No test instances running\n")
	} else {
		fmt.Printf("   ✅ Within cost limits\n")
	}

	return nil
}

func triggerEmergencyStop(client *ec2.Client, instanceIds []string, reason string) error {
	log.Printf("🚨 TRIGGERING EMERGENCY STOP: %s", reason)
	
	if len(instanceIds) == 0 {
		return nil
	}

	// Terminate all instances
	_, err := client.TerminateInstances(context.TODO(), &ec2.TerminateInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return fmt.Errorf("failed to terminate instances: %w", err)
	}

	log.Printf("🧹 EMERGENCY STOP: Terminated %d instances", len(instanceIds))
	
	// Send alert (in production, this would send email/Slack notification)
	fmt.Printf("\n🚨 COST GUARD ALERT 🚨\n")
	fmt.Printf("Reason: %s\n", reason)
	fmt.Printf("Action: Terminated %d instances\n", len(instanceIds))
	fmt.Printf("Time: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("Instances: %v\n", instanceIds)

	return nil
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}