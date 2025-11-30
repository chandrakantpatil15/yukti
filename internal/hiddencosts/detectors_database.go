package hiddencosts

// Database Inefficiency Detectors (4 additional patterns)

type RDSMultiAZNonProdDetector struct{}

func (d *RDSMultiAZNonProdDetector) Name() string { return "rds_multiaz_nonprod" }
func (d *RDSMultiAZNonProdDetector) Category() Category { return CategoryDatabase }

func (d *RDSMultiAZNonProdDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "rds" {
			env := r.Tags["environment"]
			multiAZ := r.Metadata["multi_az"].(bool)
			if (env == "dev" || env == "test" || env == "staging") && multiAZ {
				cost := r.Metadata["monthly_cost"].(float64)
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "RDS Multi-AZ for non-production",
					Description:      "Multi-AZ doubles cost, not needed for dev/test",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.5,
					Confidence:       0.95,
					Recommendation:   "Disable Multi-AZ, use snapshots for recovery",
				})
			}
		}
	}
	return findings, nil
}

type DynamoDBOnDemandDetector struct{}

func (d *DynamoDBOnDemandDetector) Name() string { return "dynamodb_ondemand_predictable" }
func (d *DynamoDBOnDemandDetector) Category() Category { return CategoryDatabase }

func (d *DynamoDBOnDemandDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "dynamodb" {
			billingMode := r.Metadata["billing_mode"].(string)
			if billingMode == "PAY_PER_REQUEST" {
				cv := r.Metadata["traffic_coefficient_variation"].(float64)
				if cv < 0.30 {
					reads := r.Metadata["monthly_reads"].(float64)
					writes := r.Metadata["monthly_writes"].(float64)
					onDemandCost := (writes/1000000)*1.25 + (reads/1000000)*0.25
					provisionedCost := ((writes/3600/1000000)*0.00065 + (reads/3600/1000000)*0.00013) * 730
					if provisionedCost < onDemandCost {
						findings = append(findings, Finding{
							DetectorName:     d.Name(),
							Category:         d.Category(),
							Severity:         SeverityHigh,
							Title:            "DynamoDB On-Demand for predictable traffic",
							Description:      "Traffic is predictable (CV<30%), provisioned is cheaper",
							ResourceARN:      r.ARN,
							EstimatedCost:    onDemandCost,
							EstimatedSavings: onDemandCost - provisionedCost,
							Confidence:       0.90,
							Recommendation:   "Switch to provisioned capacity with auto-scaling",
						})
					}
				}
			}
		}
	}
	return findings, nil
}

type RDSStorageAutoScalingDetector struct{}

func (d *RDSStorageAutoScalingDetector) Name() string { return "rds_storage_autoscaling_runaway" }
func (d *RDSStorageAutoScalingDetector) Category() Category { return CategoryDatabase }

func (d *RDSStorageAutoScalingDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "rds" {
			allocated := r.Metadata["allocated_storage_gb"].(float64)
			used := r.Metadata["used_storage_gb"].(float64)
			if used < allocated*0.6 {
				growthRate := r.Metadata["storage_growth_30d_percent"].(float64)
				if growthRate > 50 {
					wastedGB := allocated - used
					cost := wastedGB * 0.115
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "RDS storage auto-scaling runaway",
						Description:      "Storage increased >50% in 30 days, usage <60%",
						ResourceARN:      r.ARN,
						EstimatedCost:    cost,
						EstimatedSavings: cost,
						Confidence:       0.85,
						Recommendation:   "Disable auto-scaling, restore from snapshot to smaller volume",
					})
				}
			}
		}
	}
	return findings, nil
}

type ElastiCacheReservedUnusedDetector struct{}

func (d *ElastiCacheReservedUnusedDetector) Name() string { return "elasticache_reserved_unused" }
func (d *ElastiCacheReservedUnusedDetector) Category() Category { return CategoryDatabase }

func (d *ElastiCacheReservedUnusedDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "elasticache_reservation" {
			if unused, ok := r.Metadata["unused"].(bool); ok && unused {
				cost := r.Metadata["annual_cost"].(float64) / 12
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "ElastiCache reserved node not used",
					Description:      "Paying for unused reservation (no refunds)",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost,
					Confidence:       1.0,
					Recommendation:   "Modify reservation or sell on Reserved Instance Marketplace",
				})
			}
		}
	}
	return findings, nil
}
