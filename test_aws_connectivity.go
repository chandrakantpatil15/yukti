package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func main() {
	ctx := context.Background()

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to load AWS config: %v\n", err)
		os.Exit(1)
	}

	// Test STS GetCallerIdentity
	stsClient := sts.NewFromConfig(cfg)
	result, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		fmt.Printf("❌ AWS connectivity failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ AWS Connectivity Successful!")
	fmt.Printf("Account ID: %s\n", *result.Account)
	fmt.Printf("User ARN: %s\n", *result.Arn)
	fmt.Printf("User ID: %s\n", *result.UserId)
}
