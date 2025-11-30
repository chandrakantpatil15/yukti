package hiddencosts

// Data Transfer Cost Detectors (10 additional patterns)

type CloudFrontOriginDetector struct{}

func (d *CloudFrontOriginDetector) Name() string { return "cloudfront_origin_waste" }
func (d *CloudFrontOriginDetector) Category() Category { return CategoryDataTransfer }

func (d *CloudFrontOriginDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "cloudfront" {
			if hitRatio, ok := r.Metadata["cache_hit_ratio"].(float64); ok && hitRatio < 0.80 {
				cost := r.Metadata["origin_fetches_gb"].(float64) * 0.085
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "CloudFront cache hit ratio below 80%",
					Description:      "Low cache hit ratio causes excessive origin fetches",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.5,
					Confidence:       0.90,
					Recommendation:   "Increase TTL, enable compression, optimize cache policies",
				})
			}
		}
	}
	return findings, nil
}

type InterRegionReplicationDetector struct{}

func (d *InterRegionReplicationDetector) Name() string { return "inter_region_replication" }
func (d *InterRegionReplicationDetector) Category() Category { return CategoryDataTransfer }

func (d *InterRegionReplicationDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "s3_bucket" {
			if replication, ok := r.Metadata["cross_region_replication"].(bool); ok && replication {
				dataGB := r.Metadata["replication_data_gb"].(float64)
				cost := dataGB * 0.02
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "S3 cross-region replication costs",
					Description:      "Cross-region replication costs $0.02/GB",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.3,
					Confidence:       0.95,
					Recommendation:   "Use same-region replication or compress before transfer",
				})
			}
		}
	}
	return findings, nil
}

type VPCPeeringDetector struct{}

func (d *VPCPeeringDetector) Name() string { return "vpc_peering_costs" }
func (d *VPCPeeringDetector) Category() Category { return CategoryDataTransfer }

func (d *VPCPeeringDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "vpc_peering" {
			dataGB := r.Metadata["data_transfer_gb"].(float64)
			cost := dataGB * 0.01
			if vpcCount, ok := r.Metadata["vpc_count"].(int); ok && vpcCount > 5 {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "VPC peering with >5 VPCs",
					Description:      "Transit Gateway may be cheaper for hub-spoke topology",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.4,
					Confidence:       0.85,
					Recommendation:   "Consolidate VPCs or use Transit Gateway",
				})
			}
		}
	}
	return findings, nil
}

type ELBCrossAZDetector struct{}

func (d *ELBCrossAZDetector) Name() string { return "elb_cross_az_transfer" }
func (d *ELBCrossAZDetector) Category() Category { return CategoryDataTransfer }

func (d *ELBCrossAZDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "alb" || r.Type == "elb" {
			if crossAZ, ok := r.Metadata["cross_az_enabled"].(bool); ok && crossAZ {
				dataGB := r.Metadata["cross_az_data_gb"].(float64)
				cost := dataGB * 0.01
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "Load balancer cross-AZ data transfer",
					Description:      "Cross-AZ load balancing incurs $0.01/GB charges",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.8,
					Confidence:       0.90,
					Recommendation:   "Co-locate targets in same AZ as clients",
				})
			}
		}
	}
	return findings, nil
}

type DirectConnectUnderutilizedDetector struct{}

func (d *DirectConnectUnderutilizedDetector) Name() string { return "direct_connect_underutilized" }
func (d *DirectConnectUnderutilizedDetector) Category() Category { return CategoryDataTransfer }

func (d *DirectConnectUnderutilizedDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "direct_connect" {
			utilization := r.Metadata["utilization_percent"].(float64)
			if utilization < 30 {
				capacity := r.Metadata["capacity_gbps"].(float64)
				cost := capacity * 0.30 * 730 // $0.30/hour
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "Direct Connect underutilized (<30%)",
					Description:      "Paying for unused Direct Connect capacity",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.7,
					Confidence:       0.95,
					Recommendation:   "Downgrade capacity or use VPN for low-volume traffic",
				})
			}
		}
	}
	return findings, nil
}
