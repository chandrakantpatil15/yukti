package hiddencosts

// Kubernetes Cost Optimization Detectors (12 patterns)

type K8sPodCPUOverprovisionDetector struct{}

func (d *K8sPodCPUOverprovisionDetector) Name() string { return "k8s_pod_cpu_overprovisioned" }
func (d *K8sPodCPUOverprovisionDetector) Category() Category { return CategoryKubernetes }

func (d *K8sPodCPUOverprovisionDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "k8s_pod" {
			cpuRequest := r.Metadata["cpu_request_cores"].(float64)
			cpuUsageAvg := r.Metadata["cpu_usage_avg_cores"].(float64)
			if cpuUsageAvg < cpuRequest*0.3 {
				wastedCPU := cpuRequest - cpuUsageAvg
				costPerCore := 30.0 // ~$30/vCPU/month
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "Pod CPU request >3x actual usage",
					Description:      "Pod using <30% of requested CPU, wasting node capacity",
					ResourceARN:      r.ARN,
					EstimatedCost:    cpuRequest * costPerCore,
					EstimatedSavings: wastedCPU * costPerCore,
					Confidence:       0.90,
					Recommendation:   "Right-size CPU request to match actual usage + 20% buffer",
				})
			}
		}
	}
	return findings, nil
}

type K8sPodMemoryOverprovisionDetector struct{}

func (d *K8sPodMemoryOverprovisionDetector) Name() string { return "k8s_pod_memory_overprovisioned" }
func (d *K8sPodMemoryOverprovisionDetector) Category() Category { return CategoryKubernetes }

func (d *K8sPodMemoryOverprovisionDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "k8s_pod" {
			memRequest := r.Metadata["memory_request_gb"].(float64)
			memUsageAvg := r.Metadata["memory_usage_avg_gb"].(float64)
			if memUsageAvg < memRequest*0.3 {
				wastedMem := memRequest - memUsageAvg
				costPerGB := 4.0 // ~$4/GB/month
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "Pod memory request >3x actual usage",
					Description:      "Pod using <30% of requested memory, wasting node capacity",
					ResourceARN:      r.ARN,
					EstimatedCost:    memRequest * costPerGB,
					EstimatedSavings: wastedMem * costPerGB,
					Confidence:       0.90,
					Recommendation:   "Right-size memory request to match actual usage + 20% buffer",
				})
			}
		}
	}
	return findings, nil
}

type K8sNodeOverprovisionDetector struct{}

func (d *K8sNodeOverprovisionDetector) Name() string { return "k8s_node_overprovisioned" }
func (d *K8sNodeOverprovisionDetector) Category() Category { return CategoryKubernetes }

func (d *K8sNodeOverprovisionDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "k8s_node" {
			allocatable := r.Metadata["allocatable_cpu_cores"].(float64)
			requested := r.Metadata["requested_cpu_cores"].(float64)
			if requested < allocatable*0.4 {
				instanceType := r.Metadata["instance_type"].(string)
				monthlyCost := r.Metadata["monthly_cost"].(float64)
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "Node CPU allocation <40%",
					Description:      "Node underutilized, consolidate workloads or downsize",
					ResourceARN:      r.ARN,
					EstimatedCost:    monthlyCost,
					EstimatedSavings: monthlyCost * 0.5,
					Confidence:       0.85,
					Recommendation:   "Consolidate pods or use smaller instance type: " + instanceType,
				})
			}
		}
	}
	return findings, nil
}

type K8sSpotOpportunityDetector struct{}

func (d *K8sSpotOpportunityDetector) Name() string { return "k8s_spot_opportunity" }
func (d *K8sSpotOpportunityDetector) Category() Category { return CategoryKubernetes }

func (d *K8sSpotOpportunityDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "k8s_node_group" {
			capacityType := r.Metadata["capacity_type"].(string)
			if capacityType == "ON_DEMAND" {
				workloadType := r.Metadata["workload_type"].(string)
				if workloadType == "stateless" || workloadType == "batch" {
					monthlyCost := r.Metadata["monthly_cost"].(float64)
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityHigh,
						Title:            "Stateless workload on On-Demand nodes",
						Description:      "Spot instances save 70% for fault-tolerant workloads",
						ResourceARN:      r.ARN,
						EstimatedCost:    monthlyCost,
						EstimatedSavings: monthlyCost * 0.70,
						Confidence:       0.85,
						Recommendation:   "Use Spot instances with proper pod disruption budgets",
					})
				}
			}
		}
	}
	return findings, nil
}

type K8sClusterAutoscalerDetector struct{}

func (d *K8sClusterAutoscalerDetector) Name() string { return "k8s_cluster_autoscaler_inefficient" }
func (d *K8sClusterAutoscalerDetector) Category() Category { return CategoryKubernetes }

func (d *K8sClusterAutoscalerDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "k8s_cluster" {
			scaleDownDelay := r.Metadata["scale_down_delay_seconds"].(float64)
			if scaleDownDelay > 600 {
				avgIdleNodes := r.Metadata["avg_idle_nodes"].(float64)
				costPerNode := r.Metadata["cost_per_node"].(float64)
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "Cluster Autoscaler scale-down delay >10 min",
					Description:      "Long delay keeps idle nodes running, wasting cost",
					ResourceARN:      r.ARN,
					EstimatedCost:    avgIdleNodes * costPerNode,
					EstimatedSavings: avgIdleNodes * costPerNode * 0.6,
					Confidence:       0.80,
					Recommendation:   "Reduce scale-down-delay to 300s (5 min)",
				})
			}
		}
	}
	return findings, nil
}

type K8sPVCWasteDetector struct{}

func (d *K8sPVCWasteDetector) Name() string { return "k8s_pvc_waste" }
func (d *K8sPVCWasteDetector) Category() Category { return CategoryStorage }

func (d *K8sPVCWasteDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "k8s_pvc" {
			status := r.Metadata["status"].(string)
			if status == "Released" || status == "Available" {
				sizeGB := r.Metadata["size_gb"].(float64)
				storageClass := r.Metadata["storage_class"].(string)
				costPerGB := 0.10
				if storageClass == "gp3" {
					costPerGB = 0.08
				}
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "Unused PVC in Released/Available state",
					Description:      "PVC not bound to pod but still incurring EBS costs",
					ResourceARN:      r.ARN,
					EstimatedCost:    sizeGB * costPerGB,
					EstimatedSavings: sizeGB * costPerGB,
					Confidence:       0.95,
					Recommendation:   "Delete unused PVC or reclaim policy to Delete",
				})
			}
		}
	}
	return findings, nil
}

type K8sLoadBalancerWasteDetector struct{}

func (d *K8sLoadBalancerWasteDetector) Name() string { return "k8s_loadbalancer_waste" }
func (d *K8sLoadBalancerWasteDetector) Category() Category { return CategoryNetworking }

func (d *K8sLoadBalancerWasteDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "k8s_service" {
			serviceType := r.Metadata["type"].(string)
			if serviceType == "LoadBalancer" {
				lbType := r.Metadata["lb_type"].(string)
				if lbType == "classic" {
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "Classic Load Balancer per service",
						Description:      "Each LoadBalancer service costs $16/month, use Ingress",
						ResourceARN:      r.ARN,
						EstimatedCost:    16.0,
						EstimatedSavings: 14.0,
						Confidence:       0.90,
						Recommendation:   "Use Ingress controller (ALB/NGINX) to share one LB",
					})
				}
			}
		}
	}
	return findings, nil
}

type K8sHPAMisconfigDetector struct{}

func (d *K8sHPAMisconfigDetector) Name() string { return "k8s_hpa_misconfig" }
func (d *K8sHPAMisconfigDetector) Category() Category { return CategoryKubernetes }

func (d *K8sHPAMisconfigDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "k8s_hpa" {
			minReplicas := r.Metadata["min_replicas"].(float64)
			avgReplicas := r.Metadata["avg_replicas"].(float64)
			if avgReplicas < minReplicas*0.6 {
				podCost := r.Metadata["cost_per_pod"].(float64)
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "HPA min replicas too high",
					Description:      "Average replicas <60% of minimum, wasting capacity",
					ResourceARN:      r.ARN,
					EstimatedCost:    minReplicas * podCost,
					EstimatedSavings: (minReplicas - avgReplicas) * podCost,
					Confidence:       0.85,
					Recommendation:   "Lower min replicas to match actual demand",
				})
			}
		}
	}
	return findings, nil
}

type K8sNamespaceQuotaDetector struct{}

func (d *K8sNamespaceQuotaDetector) Name() string { return "k8s_namespace_no_quota" }
func (d *K8sNamespaceQuotaDetector) Category() Category { return CategoryKubernetes }

func (d *K8sNamespaceQuotaDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "k8s_namespace" {
			hasQuota := r.Metadata["has_resource_quota"].(bool)
			if !hasQuota {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityLow,
					Title:            "Namespace without ResourceQuota",
					Description:      "No limits on resource consumption, risk of cost overrun",
					ResourceARN:      r.ARN,
					EstimatedCost:    0,
					EstimatedSavings: 0,
					Confidence:       0.75,
					Recommendation:   "Set ResourceQuota to prevent unbounded resource usage",
				})
			}
		}
	}
	return findings, nil
}

type K8sDaemonSetCostDetector struct{}

func (d *K8sDaemonSetCostDetector) Name() string { return "k8s_daemonset_cost" }
func (d *K8sDaemonSetCostDetector) Category() Category { return CategoryKubernetes }

func (d *K8sDaemonSetCostDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "k8s_daemonset" {
			cpuRequest := r.Metadata["cpu_request_per_pod"].(float64)
			nodeCount := r.Metadata["node_count"].(float64)
			if cpuRequest*nodeCount > 10 {
				totalCost := cpuRequest * nodeCount * 30.0
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "DaemonSet consuming >10 cores cluster-wide",
					Description:      "DaemonSet overhead scales with nodes, review necessity",
					ResourceARN:      r.ARN,
					EstimatedCost:    totalCost,
					EstimatedSavings: totalCost * 0.3,
					Confidence:       0.75,
					Recommendation:   "Evaluate if DaemonSet is necessary or use sidecar pattern",
				})
			}
		}
	}
	return findings, nil
}

type K8sGPUIdleDetector struct{}

func (d *K8sGPUIdleDetector) Name() string { return "k8s_gpu_idle" }
func (d *K8sGPUIdleDetector) Category() Category { return CategoryCompute }

func (d *K8sGPUIdleDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "k8s_node" {
			gpuCount := r.Metadata["gpu_count"].(float64)
			if gpuCount > 0 {
				gpuUtil := r.Metadata["gpu_utilization_avg"].(float64)
				if gpuUtil < 30 {
					costPerGPU := 500.0 // ~$500/GPU/month
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityCritical,
						Title:            "GPU node with <30% utilization",
						Description:      "GPU instances are expensive, ensure high utilization",
						ResourceARN:      r.ARN,
						EstimatedCost:    gpuCount * costPerGPU,
						EstimatedSavings: gpuCount * costPerGPU * 0.7,
						Confidence:       0.90,
						Recommendation:   "Use GPU time-slicing or move to Spot instances",
					})
				}
			}
		}
	}
	return findings, nil
}

type K8sFargateOveruseDetector struct{}

func (d *K8sFargateOveruseDetector) Name() string { return "k8s_fargate_overuse" }
func (d *K8sFargateOveruseDetector) Category() Category { return CategoryKubernetes }

func (d *K8sFargateOveruseDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "k8s_fargate_profile" {
			podCount := r.Metadata["avg_pod_count"].(float64)
			if podCount > 20 {
				fargateCost := r.Metadata["monthly_cost"].(float64)
				ec2Cost := fargateCost * 0.6 // EC2 ~40% cheaper
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "Fargate profile with >20 pods",
					Description:      "Fargate costs 40% more than EC2 for steady workloads",
					ResourceARN:      r.ARN,
					EstimatedCost:    fargateCost,
					EstimatedSavings: fargateCost - ec2Cost,
					Confidence:       0.85,
					Recommendation:   "Use EC2 node groups for steady-state workloads",
				})
			}
		}
	}
	return findings, nil
}
