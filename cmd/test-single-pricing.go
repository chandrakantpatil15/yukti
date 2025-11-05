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
	fmt.Println("🔍 TESTING SINGLE INSTANCE PRICING")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("❌ AWS config failed: %v", err)
	}

	client := pricing.NewFromConfig(cfg)

	// Test with a known instance type
	instanceType := "m5.large"
	region := "US East (N. Virginia)"

	fmt.Printf("Testing: %s in %s\n", instanceType, region)

	filters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("instanceType"), Value: aws.String(instanceType)},
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String(region)},
		{Type: types.FilterTypeTermMatch, Field: aws.String("tenancy"), Value: aws.String("Shared")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("operating-system"), Value: aws.String("Linux")},
	}

	result, err := client.GetProducts(context.TODO(), &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters:     filters,
		MaxResults:  aws.Int32(1),
	})

	if err != nil {
		log.Fatalf("❌ API call failed: %v", err)
	}

	fmt.Printf("✅ Found %d products\n", len(result.PriceList))

	if len(result.PriceList) > 0 {
		var product map[string]interface{}
		if err := json.Unmarshal([]byte(result.PriceList[0]), &product); err == nil {
			// Show structure
			if terms, ok := product["terms"].(map[string]interface{}); ok {
				if onDemand, ok := terms["OnDemand"].(map[string]interface{}); ok {
					fmt.Printf("✅ OnDemand terms found: %d\n", len(onDemand))
					
					for termKey, term := range onDemand {
						fmt.Printf("Term: %s\n", termKey)
						if termMap, ok := term.(map[string]interface{}); ok {
							if priceDimensions, exists := termMap["priceDimensions"].(map[string]interface{}); exists {
								for dimKey, dimension := range priceDimensions {
									fmt.Printf("  Dimension: %s\n", dimKey)
									if dimMap, ok := dimension.(map[string]interface{}); ok {
										if pricePerUnit, exists := dimMap["pricePerUnit"].(map[string]interface{}); exists {
											if usdPrice, exists := pricePerUnit["USD"].(string); exists {
												fmt.Printf("  💰 Price: $%s/hour\n", usdPrice)
											}
										}
									}
								}
							}
						}
						break // Just show first term
					}
				}
			}
		}
	}
}