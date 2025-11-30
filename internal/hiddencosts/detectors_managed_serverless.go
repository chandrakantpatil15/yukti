package hiddencosts

// Managed Services & Serverless Detectors (4 additional patterns)

type ManagedPrometheusDetector struct{}

func (d *ManagedPrometheusDetector) Name() string { return "managed_prometheus_cost" }
func (d *ManagedPrometheusDetector) Category() Category { return CategoryManagedServices }

func (d *ManagedPrometheusDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "amp" {
			metrics := r.Metadata["metric_count"].(float64)
			if metrics < 300 {
				cost := metrics*0.30 + r.Metadata["data_ingested_gb"].(float64)*0.10
				selfHostedCost := 80.0 // t3.large + EBS
				if cost > selfHostedCost {
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "Managed Prometheus for <300 metrics",
						Description:      "Self-hosted is cheaper for small deployments",
						ResourceARN:      r.ARN,
						EstimatedCost:    cost,
						EstimatedSavings: cost - selfHostedCost,
						Confidence:       0.85,
						Recommendation:   "Self-host Prometheus on EC2",
					})
				}
			}
		}
	}
	return findings, nil
}

type TransferFamilyDetector struct{}

func (d *TransferFamilyDetector) Name() string { return "transfer_family_waste" }
func (d *TransferFamilyDetector) Category() Category { return CategoryManagedServices }

func (d *TransferFamilyDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "transfer_server" {
			hoursUsed := r.Metadata["hours_used_monthly"].(float64)
			if hoursUsed < 10 {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "AWS Transfer Family used <10 hours/month",
					Description:      "Costs $216/month, use Lambda + S3 for infrequent transfers",
					ResourceARN:      r.ARN,
					EstimatedCost:    216.0,
					EstimatedSavings: 200.0,
					Confidence:       0.90,
					Recommendation:   "Use Lambda + S3 for infrequent transfers",
				})
			}
		}
	}
	return findings, nil
}

type APIGatewayRESTDetector struct{}

func (d *APIGatewayRESTDetector) Name() string { return "api_gateway_rest_vs_http" }
func (d *APIGatewayRESTDetector) Category() Category { return CategoryServerless }

func (d *APIGatewayRESTDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "api_gateway" {
			apiType := r.Metadata["api_type"].(string)
			if apiType == "REST" {
				requests := r.Metadata["monthly_requests"].(float64)
				restCost := (requests / 1000000) * 3.50
				httpCost := (requests / 1000000) * 1.00
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "API Gateway REST API vs HTTP API",
					Description:      "HTTP API is 71% cheaper ($1 vs $3.50 per million)",
					ResourceARN:      r.ARN,
					EstimatedCost:    restCost,
					EstimatedSavings: restCost - httpCost,
					Confidence:       0.85,
					Recommendation:   "Migrate to HTTP API unless REST-specific features needed",
				})
			}
		}
	}
	return findings, nil
}

type StepFunctionsDetector struct{}

func (d *StepFunctionsDetector) Name() string { return "step_functions_express_opportunity" }
func (d *StepFunctionsDetector) Category() Category { return CategoryServerless }

func (d *StepFunctionsDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "step_function" {
			sfType := r.Metadata["type"].(string)
			if sfType == "STANDARD" {
				avgTransitions := r.Metadata["avg_transitions_per_execution"].(float64)
				if avgTransitions > 40 {
					executions := r.Metadata["monthly_executions"].(float64)
					standardCost := (executions * avgTransitions / 1000) * 0.025
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "Step Functions Standard with >40 transitions",
						Description:      "Express is cheaper for high-volume, short-duration workflows",
						ResourceARN:      r.ARN,
						EstimatedCost:    standardCost,
						EstimatedSavings: standardCost * 0.5,
						Confidence:       0.80,
						Recommendation:   "Use Express for high-volume, short-duration (<5 min) workflows",
					})
				}
			}
		}
	}
	return findings, nil
}
