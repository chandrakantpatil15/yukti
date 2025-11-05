package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func main() {
	fmt.Println("🚨 EMERGENCY COST CLEANUP")
	fmt.Println("========================")

	// AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("❌ AWS config failed:", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Get ALL running instances
	result, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		log.Fatal("❌ Failed to describe instances:", err)
	}

	var longRunningInstances []string
	totalCost := 0.0

	costMap := map[string]float64{
		"t3.micro":  0.0104,
		"t3.small":  0.0208,
		"t3.medium": 0.0416,
		"t3.large":  0.0832,
		"m5.large":  0.096,
	}

	fmt.Println("\n📊 INSTANCE ANALYSIS:")
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if instance.State.Name != "running" {
				continue
			}

			uptime := time.Since(*instance.LaunchTime)
			hourlyCost := costMap[string(instance.InstanceType)]
			totalCostSoFar := uptime.Hours() * hourlyCost

			fmt.Printf("🖥️  %s (%s)\n", *instance.InstanceId, instance.InstanceType)
			fmt.Printf("   Uptime: %v\n", uptime.Round(time.Hour))
			fmt.Printf("   Cost so far: $%.2f\n", totalCostSoFar)

			// Flag instances running > 1 hour as potentially wasteful
			if uptime > 1*time.Hour {
				longRunningInstances = append(longRunningInstances, *instance.InstanceId)
				totalCost += totalCostSoFar
				fmt.Printf("   ⚠️  LONG RUNNING - Consider termination\n")
			}
			fmt.Println()
		}
	}

	if len(longRunningInstances) == 0 {
		fmt.Println("✅ No long-running instances found")
		return
	}

	fmt.Printf("🚨 COST ALERT:\n")
	fmt.Printf("   Long-running instances: %d\n", len(longRunningInstances))
	fmt.Printf("   Total wasted cost: $%.2f\n", totalCost)
	fmt.Printf("   Instances: %v\n", longRunningInstances)

	fmt.Print("\n❓ Terminate these instances? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if response == "y" || response == "Y" {
		fmt.Println("\n🧹 TERMINATING INSTANCES...")
		
		_, err := ec2Client.TerminateInstances(context.TODO(), &ec2.TerminateInstancesInput{
			InstanceIds: longRunningInstances,
		})
		if err != nil {
			log.Fatal("❌ Failed to terminate instances:", err)
		}

		fmt.Printf("✅ Terminated %d instances\n", len(longRunningInstances))
		fmt.Printf("💰 Prevented further cost: $%.2f/hour\n", totalCost/time.Since(*result.Reservations[0].Instances[0].LaunchTime).Hours())
	} else {
		fmt.Println("⚠️  Instances left running - costs will continue")
	}
}