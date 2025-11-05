package main

import (
	"context"
	"log"
	"os"

	"github.com/cloudcostoptimizer/yukti/internal/config"
	"github.com/cloudcostoptimizer/yukti/internal/services"
)

func main() {
	log.Println("🚀 Starting AWS Pricing Import...")

	// Load database connection
	db, err := config.SetupDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize AWS Pricing Service
	pricingService, err := services.NewAWSPricingService(db)
	if err != nil {
		log.Fatalf("Failed to initialize AWS Pricing Service: %v", err)
	}

	ctx := context.Background()

	// Get region from command line or default to us-east-1
	region := "us-east-1"
	if len(os.Args) > 1 {
		region = os.Args[1]
	}

	log.Printf("Importing pricing data for region: %s", region)

	// Import EC2 pricing data
	if err := pricingService.ImportEC2Pricing(ctx, region); err != nil {
		log.Fatalf("Failed to import pricing data: %v", err)
	}

	log.Println("✅ AWS Pricing import completed successfully!")

	// Show imported data
	log.Println("📊 Imported pricing summary:")
	
	// You can add a query here to show the imported data
	var count int64
	db.Model(&struct{}{}).Table("aws_pricings").Count(&count)
	log.Printf("Total pricing records: %d", count)
}