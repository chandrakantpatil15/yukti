package hiddencosts

// Compute Waste Detectors (7 additional patterns)

type LambdaMemoryDetector struct{}

func (d *LambdaMemoryDetector) Name() string { return "lambda_memory_overprovisioned" }
func (d *LambdaMemoryDetector) Category() Category { return CategoryCompute }

func (d *LambdaMemoryDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "lambda" {
			allocated := r.Metadata["memory_mb"].(float64)
			maxUsed := r.Metadata["max_memory_used_mb"].(float64)
			if maxUsed < allocated*0.6 {
				invocations := r.Metadata["monthly_invocations"].(float64)
				avgDuration := r.Metadata["avg_duration_ms"].(float64)
				currentCost := (allocated / 1024) * (avgDuration / 1000) * invocations * 0.0000166667
				optimizedMem := maxUsed * 1.2
				optimizedCost := (optimizedMem / 1024) * (avgDuration / 1000) * invocations * 0.0000166667
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "Lambda memory over-provisioned",
					Description:      "Max memory used <60% of allocated",
					ResourceARN:      r.ARN,
					EstimatedCost:    currentCost,
					EstimatedSavings: currentCost - optimizedCost,
					Confidence:       0.90,
					Recommendation:   "Right-size memory to 1.2x max observed usage",
				})
			}
		}
	}
	return findings, nil
}

type FargateOverprovisionedDetector struct{}

func (d *FargateOverprovisionedDetector) Name() string { return "fargate_overprovisioned" }
func (d *FargateOverprovisionedDetector) Category() Category { return CategoryCompute }

func (d *FargateOverprovisionedDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "ecs_task" {
			cpuUtil := r.Metadata["avg_cpu_utilization"].(float64)
			memUtil := r.Metadata["avg_memory_utilization"].(float64)
			if cpuUtil < 40 || memUtil < 40 {
				vCPU := r.Metadata["vcpu"].(float64)
				memGB := r.Metadata["memory_gb"].(float64)
				currentCost := (vCPU * 0.04048 * 730) + (memGB * 0.004445 * 730)
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "Fargate task over-provisioned",
					Description:      "CPU or memory utilization <40%",
					ResourceARN:      r.ARN,
					EstimatedCost:    currentCost,
					EstimatedSavings: currentCost * 0.4,
					Confidence:       0.85,
					Recommendation:   "Right-size task, use Fargate Spot for 70% savings",
				})
			}
		}
	}
	return findings, nil
}

type BeanstalkNonProdDetector struct{}

func (d *BeanstalkNonProdDetector) Name() string { return "beanstalk_nonprod_ha" }
func (d *BeanstalkNonProdDetector) Category() Category { return CategoryCompute }

func (d *BeanstalkNonProdDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "beanstalk" {
			env := r.Tags["environment"]
			if env == "dev" || env == "test" || env == "staging" {
				if ha, ok := r.Metadata["high_availability"].(bool); ok && ha {
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "Beanstalk HA enabled for non-prod",
						Description:      "Dev/test environments don't need load balancer + auto-scaling",
						ResourceARN:      r.ARN,
						EstimatedCost:    16.20,
						EstimatedSavings: 16.20,
						Confidence:       0.95,
						Recommendation:   "Use single-instance environment type",
					})
				}
			}
		}
	}
	return findings, nil
}

type EC2BurstableT2Detector struct{}

func (d *EC2BurstableT2Detector) Name() string { return "ec2_t2_vs_t3" }
func (d *EC2BurstableT2Detector) Category() Category { return CategoryCompute }

func (d *EC2BurstableT2Detector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "ec2" {
			instanceType := r.Metadata["instance_type"].(string)
			if len(instanceType) > 2 && instanceType[:2] == "t2" {
				// T3 is 10% cheaper and better performance
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityLow,
					Title:            "EC2 using T2 instead of T3",
					Description:      "T3 instances are 10% cheaper with better performance",
					ResourceARN:      r.ARN,
					EstimatedCost:    50.0,
					EstimatedSavings: 5.0,
					Confidence:       0.95,
					Recommendation:   "Migrate to T3 instance type",
				})
			}
		}
	}
	return findings, nil
}

type EC2PreviousGenDetector struct{}

func (d *EC2PreviousGenDetector) Name() string { return "ec2_previous_generation" }
func (d *EC2PreviousGenDetector) Category() Category { return CategoryCompute }

func (d *EC2PreviousGenDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	prevGen := map[string]bool{"m4": true, "c4": true, "r4": true, "t2": true}
	for _, r := range resources {
		if r.Type == "ec2" {
			instanceType := r.Metadata["instance_type"].(string)
			family := instanceType[:2]
			if prevGen[family] {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "EC2 using previous generation instance",
					Description:      "Current generation offers 20-40% better price/performance",
					ResourceARN:      r.ARN,
					EstimatedCost:    100.0,
					EstimatedSavings: 30.0,
					Confidence:       0.90,
					Recommendation:   "Migrate to current generation (m5, c5, r5, t3)",
				})
			}
		}
	}
	return findings, nil
}

type AutoScalingUnusedDetector struct{}

func (d *AutoScalingUnusedDetector) Name() string { return "autoscaling_unused" }
func (d *AutoScalingUnusedDetector) Category() Category { return CategoryCompute }

func (d *AutoScalingUnusedDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "autoscaling_group" {
			desired := r.Metadata["desired_capacity"].(int)
			min := r.Metadata["min_capacity"].(int)
			max := r.Metadata["max_capacity"].(int)
			if desired == min && min == max {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityLow,
					Title:            "Auto Scaling group with fixed capacity",
					Description:      "Min=Desired=Max, not actually auto-scaling",
					ResourceARN:      r.ARN,
					EstimatedCost:    0,
					EstimatedSavings: 0,
					Confidence:       1.0,
					Recommendation:   "Remove ASG and use fixed EC2 instances",
				})
			}
		}
	}
	return findings, nil
}

type SpotInstanceOpportunityDetector struct{}

func (d *SpotInstanceOpportunityDetector) Name() string { return "spot_instance_opportunity" }
func (d *SpotInstanceOpportunityDetector) Category() Category { return CategoryCompute }

func (d *SpotInstanceOpportunityDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "ec2" {
			// Safely check lifecycle
			lifecycleVal, ok := r.Metadata["lifecycle"]
			if !ok || lifecycleVal == nil {
				continue
			}
			lifecycle, ok := lifecycleVal.(string)
			if !ok || lifecycle != "on-demand" {
				continue
			}
			
			workload := r.Tags["workload"]
			if workload == "batch" || workload == "dev" || workload == "test" {
				// Safely get cost
				costVal, ok := r.Metadata["monthly_cost"]
				if !ok || costVal == nil {
					continue
				}
				cost, ok := costVal.(float64)
				if !ok {
					continue
				}
				
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "Fault-tolerant workload on On-Demand",
					Description:      "Batch/dev/test workloads can use Spot for 70% savings",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.70,
					Confidence:       0.80,
					Recommendation:   "Migrate to Spot instances",
				})
			}
		}
	}
	return findings, nil
}
