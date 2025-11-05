package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
)

func main() {
	fmt.Println("🔍 TESTING AWS PRICING API ACCESS")

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("❌ AWS config failed: %v", err)
	}

	client := pricing.NewFromConfig(cfg)

	// Test 1: Get all available services
	fmt.Println("\n1. Testing AWS Pricing API connectivity...")
	services, err := client.DescribeServices(context.TODO(), &pricing.DescribeServicesInput{
		MaxResults: aws.Int32(5),
	})
	if err != nil {
		log.Fatalf("❌ AWS Pricing API failed: %v", err)
	}

	fmt.Printf("✅ Connected! Found %d services\n", len(services.Services))

	// Test 2: Get EC2 service attributes
	fmt.Println("\n2. Getting EC2 service attributes...")
	attrs, err := client.GetAttributeValues(context.TODO(), &pricing.GetAttributeValuesInput{
		ServiceCode:   aws.String("AmazonEC2"),
		AttributeName: aws.String("instanceType"),
		MaxResults:    aws.Int32(10),
	})
	if err != nil {
		log.Printf("❌ Error getting attributes: %v", err)
	} else {
		fmt.Printf("✅ Found %d instance types\n", len(attrs.AttributeValues))
		for i, attr := range attrs.AttributeValues {
			if i < 5 {
				fmt.Printf("  - %s\n", *attr.Value)
			}
		}
	}

	// Test 3: Try simplified filters
	fmt.Println("\n3. Testing simplified EC2 pricing query...")
	
	filters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("ServiceCode"), Value: aws.String("AmazonEC2")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("instanceType"), Value: aws.String("m5.large")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String("US East (N. Virginia)")},
	}

	result, err := client.GetProducts(context.TODO(), &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters:     filters,
		MaxResults:  aws.Int32(1),
	})

	if err != nil {
		log.Printf("❌ GetProducts failed: %v", err)
	} else {
		fmt.Printf("✅ Found %d products\n", len(result.PriceList))
		
		if len(result.PriceList) > 0 {
			// Parse and show the structure
			var product map[string]interface{}
			if err := json.Unmarshal([]byte(result.PriceList[0]), &product); err == nil {
				fmt.Println("\n📋 Product structure:")
				if terms, ok := product["terms"].(map[string]interface{}); ok {
					for termType := range terms {
						fmt.Printf("  - %s\n", termType)
					}
				}
			}
		}
	}

	fmt.Println("\n✅ AWS Pricing API test completed!")
}