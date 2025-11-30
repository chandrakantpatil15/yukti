package hiddencosts

// Networking Cost Detectors (3 additional patterns)

type UnattachedEIPDetector struct{}

func (d *UnattachedEIPDetector) Name() string { return "unattached_eip" }
func (d *UnattachedEIPDetector) Category() Category { return CategoryNetworking }

func (d *UnattachedEIPDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "eip" {
			if attached, ok := r.Metadata["attached"].(bool); ok && !attached {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityLow,
					Title:            "Unattached Elastic IP",
					Description:      "Unattached EIPs cost $3.60/month",
					ResourceARN:      r.ARN,
					EstimatedCost:    3.60,
					EstimatedSavings: 3.60,
					Confidence:       1.0,
					Recommendation:   "Release unattached EIP",
				})
			}
		}
	}
	return findings, nil
}

type IdleVPNDetector struct{}

func (d *IdleVPNDetector) Name() string { return "idle_vpn_connection" }
func (d *IdleVPNDetector) Category() Category { return CategoryNetworking }

func (d *IdleVPNDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "vpn_connection" {
			dataGB := r.Metadata["data_transfer_gb_monthly"].(float64)
			if dataGB < 1 {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "Idle VPN connection",
					Description:      "VPN costs $36/month with <1GB data transfer",
					ResourceARN:      r.ARN,
					EstimatedCost:    36.0,
					EstimatedSavings: 36.0,
					Confidence:       0.90,
					Recommendation:   "Delete idle VPN, use AWS Client VPN for temporary access",
				})
			}
		}
	}
	return findings, nil
}

type TransitGatewayUnderutilizedDetector struct{}

func (d *TransitGatewayUnderutilizedDetector) Name() string { return "transit_gateway_underutilized" }
func (d *TransitGatewayUnderutilizedDetector) Category() Category { return CategoryNetworking }

func (d *TransitGatewayUnderutilizedDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "transit_gateway" {
			attachments := r.Metadata["attachment_count"].(int)
			if attachments < 3 {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "Transit Gateway with <3 attachments",
					Description:      "TGW costs $36/month + $0.02/GB, not cost-effective for <3 VPCs",
					ResourceARN:      r.ARN,
					EstimatedCost:    36.0,
					EstimatedSavings: 20.0,
					Confidence:       0.85,
					Recommendation:   "Use VPC peering for simple topologies",
				})
			}
		}
	}
	return findings, nil
}
