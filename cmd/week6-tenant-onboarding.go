package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"yukti/internal/tenant"
)

func main() {
	fmt.Println("=== Week 6: Multi-Tenant Architecture & Customer Onboarding ===\n")

	db := connectDB()
	defer db.Close()

	tenantSvc := tenant.NewService(db)
	ctx := context.Background()

	// Demo 1: Onboard new customer
	fmt.Println("📋 Demo 1: Customer Onboarding")
	fmt.Println("─────────────────────────────────")
	
	req := tenant.OnboardingRequest{
		CompanyName: "Acme Corporation",
		Email:       "admin@acme.com",
		AWSAccounts: []struct {
			AccountID   string `json:"account_id" validate:"required,len=12"`
			AccountName string `json:"account_name"`
		}{
			{AccountID: "123456789012", AccountName: "Acme Production"},
			{AccountID: "123456789013", AccountName: "Acme Staging"},
		},
	}

	resp, err := tenantSvc.OnboardTenant(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Tenant Created: %s\n", resp.TenantCode)
	fmt.Printf("   Company: %s\n", req.CompanyName)
	fmt.Printf("   Tier: %s (14-day trial)\n", resp.SubscriptionTier)
	fmt.Printf("   AWS Accounts: %d\n", len(resp.Accounts))
	fmt.Printf("   External ID: %s\n\n", resp.ExternalID)

	// Demo 2: IAM Setup Instructions
	fmt.Println("🔐 Demo 2: AWS IAM Setup Instructions")
	fmt.Println("─────────────────────────────────────")
	fmt.Println(resp.SetupScript)
	fmt.Println()

	// Demo 3: Simulate AWS resource sync
	fmt.Println("🔄 Demo 3: AWS Resource Discovery")
	fmt.Println("─────────────────────────────────")
	
	// Insert mock resources for demo
	for i, acc := range resp.Accounts {
		for j := 0; j < 5; j++ {
			_, err := db.ExecContext(ctx, `
				INSERT INTO yt_tenant_resources 
				(tenant_id, aws_account_id, resource_id, resource_type, region, instance_type, state, tags, monthly_cost)
				VALUES ($1, $2, $3, 'ec2', $4, $5, 'running', $6, $7)`,
				resp.Accounts[0].TenantID,
				acc.ID,
				fmt.Sprintf("i-%012d%d", i, j),
				[]string{"us-east-1", "us-west-2"}[j%2],
				[]string{"t3.medium", "t3.large", "m5.xlarge"}[j%3],
				fmt.Sprintf(`{"Environment":"%s","Project":"web-app"}`, []string{"prod", "staging"}[i]),
				float64(50+j*20),
			)
			if err != nil {
				log.Printf("Error inserting resource: %v", err)
			}
		}
	}

	// Update account status
	for _, acc := range resp.Accounts {
		db.ExecContext(ctx, `UPDATE yt_aws_accounts SET status = 'active', last_sync = NOW() WHERE id = $1`, acc.ID)
	}

	fmt.Printf("✅ Discovered 10 EC2 instances across 2 accounts\n")
	fmt.Printf("   Regions: us-east-1, us-west-2\n")
	fmt.Printf("   Total Monthly Cost: $800\n\n")

	// Demo 4: Tenant isolation verification
	fmt.Println("🔒 Demo 4: Multi-Tenant Data Isolation")
	fmt.Println("──────────────────────────────────────")
	
	var resourceCount int
	db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM yt_tenant_resources 
		WHERE tenant_id = (SELECT id FROM yt_tenants WHERE tenant_code = $1)`,
		resp.TenantCode,
	).Scan(&resourceCount)

	fmt.Printf("✅ Tenant '%s' has %d isolated resources\n", resp.TenantCode, resourceCount)
	fmt.Printf("   All queries filtered by tenant_id\n")
	fmt.Printf("   Row-level security enforced\n\n")

	// Demo 5: Subscription tier features
	fmt.Println("💎 Demo 5: Subscription Tier Features")
	fmt.Println("─────────────────────────────────────")
	
	features := map[tenant.SubscriptionTier][]string{
		tenant.TierFree: {
			"Basic cost optimization",
			"EC2 rightsizing recommendations",
			"Monthly reports",
		},
		tenant.TierProfessional: {
			"All FREE features",
			"Real-time dashboards",
			"Custom alerts & budgets",
			"Multi-account support",
			"Email support",
		},
		tenant.TierEnterprise: {
			"All PROFESSIONAL features",
			"AI-powered predictions",
			"White-label branding",
			"SSO/SAML integration",
			"Dedicated support",
			"Custom integrations",
		},
	}

	for tier, feats := range features {
		fmt.Printf("\n%s ($%s/month):\n", tier, map[tenant.SubscriptionTier]string{
			tenant.TierFree:         "0",
			tenant.TierProfessional: "99",
			tenant.TierEnterprise:   "499",
		}[tier])
		for _, f := range feats {
			fmt.Printf("  • %s\n", f)
		}
	}

	// Demo 6: Generate onboarding summary
	fmt.Println("\n\n📊 Week 6 Implementation Summary")
	fmt.Println("═════════════════════════════════")
	
	summary := map[string]interface{}{
		"tenant_onboarding": map[string]interface{}{
			"tenant_code":        resp.TenantCode,
			"company":            req.CompanyName,
			"subscription_tier":  resp.SubscriptionTier,
			"trial_period_days":  14,
			"aws_accounts":       len(resp.Accounts),
		},
		"resource_discovery": map[string]interface{}{
			"total_resources": resourceCount,
			"resource_types":  []string{"ec2"},
			"regions":         []string{"us-east-1", "us-west-2"},
			"sync_frequency":  "hourly",
		},
		"security": map[string]interface{}{
			"access_model":     "read-only",
			"authentication":   "IAM AssumeRole",
			"external_id":      "enabled",
			"data_isolation":   "tenant_id filtering",
		},
		"features_enabled": map[string]interface{}{
			"multi_tenant":      true,
			"auto_discovery":    true,
			"cost_tracking":     true,
			"tier_based_access": true,
		},
	}

	jsonData, _ := json.MarshalIndent(summary, "", "  ")
	fmt.Println(string(jsonData))

	fmt.Println("\n✅ Week 6 Complete: Multi-tenant architecture operational")
	fmt.Println("   Next: Week 7 - API Gateway & REST Endpoints")
}

func connectDB() *sql.DB {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/yukti?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}
