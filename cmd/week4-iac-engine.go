package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"yukti/internal/iac"
)

// Week4IaCEngine demonstrates Infrastructure-as-Code recommendation generation
func main() {
	log.Println("🎯 Week 4: Infrastructure-as-Code Recommendation Engine")
	log.Println("======================================================")
	
	// Initialize IaC generators
	terraformGen := iac.NewTerraformGenerator("aws", "us-east-1")
	cloudFormationGen := iac.NewCloudFormationGenerator("us-east-1")
	multiCloudGen := iac.NewMultiCloudGenerator("us-east-1")
	
	// Generate sample optimization recommendations
	recommendations := generateSampleRecommendations()
	
	log.Printf("📋 Generated %d optimization recommendations", len(recommendations))
	
	// Generate IaC scripts for each recommendation
	var allScripts []*iac.IaCScript
	
	log.Println("🔧 Generating Terraform Scripts...")
	for _, rec := range recommendations[:3] { // First 3 for Terraform
		script := terraformGen.GenerateEC2Optimization(rec)
		allScripts = append(allScripts, script)
		log.Printf("✅ Generated Terraform script for %s (%s)", rec.ResourceID, rec.Action)
	}
	
	log.Println("☁️ Generating CloudFormation Templates...")
	for _, rec := range recommendations[3:6] { // Next 3 for CloudFormation
		script := cloudFormationGen.GenerateEC2Optimization(rec)
		allScripts = append(allScripts, script)
		log.Printf("✅ Generated CloudFormation template for %s (%s)", rec.ResourceID, rec.Action)
	}
	
	log.Println("🌐 Generating Multi-Cloud Scripts...")
	// Azure ARM templates
	for _, rec := range recommendations[6:8] { // Next 2 for Azure
		script := multiCloudGen.GenerateAzureOptimization(rec)
		allScripts = append(allScripts, script)
		log.Printf("✅ Generated Azure ARM template for %s (%s)", rec.ResourceID, rec.Action)
	}
	
	// GCP Deployment Manager templates
	for _, rec := range recommendations[8:10] { // Last 2 for GCP
		script := multiCloudGen.GenerateGCPOptimization(rec)
		allScripts = append(allScripts, script)
		log.Printf("✅ Generated GCP template for %s (%s)", rec.ResourceID, rec.Action)
	}
	
	// Generate comprehensive IaC report
	generateIaCReport(allScripts, recommendations)
	
	// Demonstrate script examples
	demonstrateScriptExamples(allScripts)
	
	log.Println("✅ Week 4 IaC Recommendation Engine Complete!")
}

// generateSampleRecommendations generates sample optimization recommendations
func generateSampleRecommendations() []*iac.OptimizationRecommendation {
	actions := []string{"downsize", "terminate", "schedule", "spot_conversion"}
	instanceTypes := []string{"t3.micro", "t3.small", "t3.medium", "m5.large", "c5.xlarge"}
	
	var recommendations []*iac.OptimizationRecommendation
	
	for i := 0; i < 10; i++ {
		action := actions[rand.Intn(len(actions))]
		instanceType := instanceTypes[rand.Intn(len(instanceTypes))]
		
		// Generate realistic savings based on action
		var savings float64
		switch action {
		case "downsize":
			savings = 15 + rand.Float64()*35 // $15-50/day
		case "terminate":
			savings = 25 + rand.Float64()*75 // $25-100/day
		case "schedule":
			savings = 10 + rand.Float64()*20 // $10-30/day
		case "spot_conversion":
			savings = 20 + rand.Float64()*40 // $20-60/day
		}
		
		rec := &iac.OptimizationRecommendation{
			ResourceID:              fmt.Sprintf("i-%s-%03d", instanceType, i),
			Action:                  action,
			RecommendedInstanceType: getRecommendedInstanceType(instanceType, action),
			EstimatedSavings:        savings,
			Confidence:              0.7 + rand.Float64()*0.3, // 70-100%
			Reasoning:               generateReasoning(action, savings),
		}
		
		recommendations = append(recommendations, rec)
	}
	
	return recommendations
}

// getRecommendedInstanceType returns recommended instance type based on action
func getRecommendedInstanceType(current, action string) string {
	if action != "downsize" {
		return current
	}
	
	downsizeMap := map[string]string{
		"m5.large":  "t3.medium",
		"c5.xlarge": "t3.large",
		"t3.medium": "t3.small",
		"t3.small":  "t3.micro",
		"t3.micro":  "t3.nano",
	}
	
	if recommended, exists := downsizeMap[current]; exists {
		return recommended
	}
	return current
}

// generateReasoning generates reasoning for the recommendation
func generateReasoning(action string, savings float64) string {
	switch action {
	case "downsize":
		return fmt.Sprintf("Low utilization detected. Downsizing can save $%.2f/day", savings)
	case "terminate":
		return fmt.Sprintf("Instance unused for 7+ days. Termination saves $%.2f/day", savings)
	case "schedule":
		return fmt.Sprintf("Non-production workload. Scheduling saves $%.2f/day", savings)
	case "spot_conversion":
		return fmt.Sprintf("Fault-tolerant workload. Spot instances save $%.2f/day", savings)
	default:
		return "Cost optimization opportunity identified"
	}
}

// generateIaCReport generates comprehensive IaC report
func generateIaCReport(scripts []*iac.IaCScript, recommendations []*iac.OptimizationRecommendation) {
	report := IaCReport{
		Timestamp:            time.Now(),
		Week:                 4,
		Phase:                "Infrastructure-as-Code Recommendation Engine",
		TotalRecommendations: len(recommendations),
		GeneratedScripts:     len(scripts),
		ScriptBreakdown:      make(map[string]int),
		ProviderBreakdown:    make(map[string]int),
		ActionBreakdown:      make(map[string]int),
		TotalSavings:         calculateTotalSavings(recommendations),
		Scripts:              scripts,
	}
	
	// Calculate breakdowns
	for _, script := range scripts {
		report.ScriptBreakdown[script.Type]++
		report.ProviderBreakdown[script.Provider]++
		report.ActionBreakdown[script.Action]++
	}
	
	// Convert to JSON for display
	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	
	log.Println("\n📋 Week 4 IaC Report:")
	log.Println("=====================")
	fmt.Println(string(reportJSON))
	
	// Summary metrics
	log.Println("\n📈 IaC Generation Summary:")
	log.Println("==========================")
	log.Printf("🔧 Total Recommendations: %d", report.TotalRecommendations)
	log.Printf("📜 Generated Scripts: %d", report.GeneratedScripts)
	log.Printf("💰 Total Potential Savings: $%.2f/month", report.TotalSavings*30)
	
	// Script type breakdown
	log.Println("\n📊 Script Type Breakdown:")
	log.Println("=========================")
	for scriptType, count := range report.ScriptBreakdown {
		log.Printf("• %s: %d scripts", scriptType, count)
	}
	
	// Provider breakdown
	log.Println("\n🌐 Provider Breakdown:")
	log.Println("======================")
	for provider, count := range report.ProviderBreakdown {
		log.Printf("• %s: %d scripts", provider, count)
	}
	
	// Action breakdown
	log.Println("\n⚡ Action Breakdown:")
	log.Println("===================")
	for action, count := range report.ActionBreakdown {
		log.Printf("• %s: %d scripts", action, count)
	}
	
	log.Println("\n🔍 Key Features:")
	log.Println("================")
	log.Println("• ✅ Read-only audit approach (no direct cloud modifications)")
	log.Println("• 🔧 Multi-format IaC generation (Terraform, CloudFormation, ARM, GCP)")
	log.Println("• 🛡️ Built-in rollback scripts for safety")
	log.Println("• 📋 Step-by-step deployment instructions")
	log.Println("• 💰 Accurate cost savings estimates")
	log.Println("• 🎯 Customer-controlled remediation process")
	
	log.Println("\n🔮 Week 5 Preview:")
	log.Println("==================")
	log.Println("• Advanced monitoring dashboards")
	log.Println("• Real-time cost alerting system")
	log.Println("• Budget tracking and forecasting")
	log.Println("• Multi-account cost allocation")
	log.Println("• Executive reporting and KPIs")
}

// demonstrateScriptExamples shows sample generated scripts
func demonstrateScriptExamples(scripts []*iac.IaCScript) {
	log.Println("\n📜 Sample Generated Scripts:")
	log.Println("============================")
	
	// Show first script of each type
	shownTypes := make(map[string]bool)
	
	for _, script := range scripts {
		if !shownTypes[script.Type] {
			log.Printf("\n🔧 %s Script Example (%s):", script.Type, script.Provider)
			log.Println("-----------------------------------")
			
			// Show first 10 lines of the script
			lines := splitLines(script.Code)
			maxLines := 10
			if len(lines) < maxLines {
				maxLines = len(lines)
			}
			
			for i := 0; i < maxLines; i++ {
				log.Printf("  %s", lines[i])
			}
			
			if len(lines) > maxLines {
				log.Printf("  ... (%d more lines)", len(lines)-maxLines)
			}
			
			log.Printf("\n💡 Instructions:")
			for i, instruction := range script.Instructions[:3] { // Show first 3 instructions
				log.Printf("  %d. %s", i+1, instruction)
			}
			log.Printf("  ... (%d total instructions)", len(script.Instructions))
			
			shownTypes[script.Type] = true
		}
	}
}

// splitLines splits text into lines
func splitLines(text string) []string {
	lines := []string{}
	current := ""
	
	for _, char := range text {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		lines = append(lines, current)
	}
	
	return lines
}

// calculateTotalSavings calculates total potential savings
func calculateTotalSavings(recommendations []*iac.OptimizationRecommendation) float64 {
	var total float64
	for _, rec := range recommendations {
		total += rec.EstimatedSavings
	}
	return total
}

// IaCReport represents the comprehensive IaC generation report
type IaCReport struct {
	Timestamp            time.Time                `json:"timestamp"`
	Week                 int                      `json:"week"`
	Phase                string                   `json:"phase"`
	TotalRecommendations int                      `json:"total_recommendations"`
	GeneratedScripts     int                      `json:"generated_scripts"`
	ScriptBreakdown      map[string]int           `json:"script_breakdown"`
	ProviderBreakdown    map[string]int           `json:"provider_breakdown"`
	ActionBreakdown      map[string]int           `json:"action_breakdown"`
	TotalSavings         float64                  `json:"total_savings"`
	Scripts              []*iac.IaCScript         `json:"scripts"`
}