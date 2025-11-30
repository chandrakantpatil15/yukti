package hiddencosts

// Advanced Compute Detectors (4 additional patterns)

type SavingsPlansUnderutilizedDetector struct{}

func (d *SavingsPlansUnderutilizedDetector) Name() string { return "savings_plans_underutilized" }
func (d *SavingsPlansUnderutilizedDetector) Category() Category { return CategoryCompute }

func (d *SavingsPlansUnderutilizedDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "savings_plan" {
			utilization := r.Metadata["utilization_percent"].(float64)
			if utilization < 80 {
				commitment := r.Metadata["hourly_commitment"].(float64)
				wastedHours := (100 - utilization) / 100 * 730
				cost := commitment * wastedHours
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "Savings Plan utilization <80%",
					Description:      "Paying for unused commitment",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost,
					Confidence:       0.95,
					Recommendation:   "Reduce commitment or increase usage to match",
				})
			}
		}
	}
	return findings, nil
}

type ReservedInstanceWasteDetector struct{}

func (d *ReservedInstanceWasteDetector) Name() string { return "reserved_instance_waste" }
func (d *ReservedInstanceWasteDetector) Category() Category { return CategoryCompute }

func (d *ReservedInstanceWasteDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "reserved_instance" {
			if unused, ok := r.Metadata["unused"].(bool); ok && unused {
				monthlyCost := r.Metadata["monthly_cost"].(float64)
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityCritical,
					Title:            "Unused Reserved Instance",
					Description:      "Paying for RI with no matching instance running",
					ResourceARN:      r.ARN,
					EstimatedCost:    monthlyCost,
					EstimatedSavings: monthlyCost,
					Confidence:       1.0,
					Recommendation:   "Modify RI or sell on Reserved Instance Marketplace",
				})
			}
		}
	}
	return findings, nil
}

type DedicatedHostsUnderutilizedDetector struct{}

func (d *DedicatedHostsUnderutilizedDetector) Name() string { return "dedicated_hosts_underutilized" }
func (d *DedicatedHostsUnderutilizedDetector) Category() Category { return CategoryCompute }

func (d *DedicatedHostsUnderutilizedDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "dedicated_host" {
			capacity := r.Metadata["instance_capacity"].(int)
			running := r.Metadata["running_instances"].(int)
			if running < capacity/2 {
				hostCost := r.Metadata["hourly_cost"].(float64) * 730
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "Dedicated Host <50% utilized",
					Description:      "Paying for full host with half capacity unused",
					ResourceARN:      r.ARN,
					EstimatedCost:    hostCost,
					EstimatedSavings: hostCost * 0.5,
					Confidence:       0.90,
					Recommendation:   "Consolidate instances or release host",
				})
			}
		}
	}
	return findings, nil
}

type BatchVsSpotDetector struct{}

func (d *BatchVsSpotDetector) Name() string { return "batch_vs_spot_cost" }
func (d *BatchVsSpotDetector) Category() Category { return CategoryCompute }

func (d *BatchVsSpotDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "batch_compute_environment" {
			_ = r.Metadata["instance_type"].(string) // instanceType
			if lifecycle, ok := r.Metadata["lifecycle"].(string); ok && lifecycle == "ON_DEMAND" {
				cost := r.Metadata["monthly_cost"].(float64)
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "AWS Batch using On-Demand instances",
					Description:      "Batch workloads are fault-tolerant, use Spot for 70% savings",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.70,
					Confidence:       0.95,
					Recommendation:   "Configure Batch to use Spot instances",
				})
			}
		}
	}
	return findings, nil
}
