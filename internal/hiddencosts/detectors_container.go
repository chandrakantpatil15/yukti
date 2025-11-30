package hiddencosts

// Container Waste Detectors (2 additional patterns)

type ECRScanningDetector struct{}

func (d *ECRScanningDetector) Name() string { return "ecr_enhanced_scanning_waste" }
func (d *ECRScanningDetector) Category() Category { return CategoryContainer }

func (d *ECRScanningDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "ecr_repository" {
			if enhanced, ok := r.Metadata["enhanced_scanning"].(bool); ok && enhanced {
				env := r.Tags["environment"]
				if env == "dev" || env == "test" {
					imageCount := r.Metadata["image_count"].(float64)
					cost := imageCount * 0.09
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityLow,
						Title:            "ECR enhanced scanning for non-prod",
						Description:      "Enhanced scanning costs $0.09/image, use basic for dev/test",
						ResourceARN:      r.ARN,
						EstimatedCost:    cost,
						EstimatedSavings: cost,
						Confidence:       0.95,
						Recommendation:   "Use basic scanning for non-prod, enhanced for production only",
					})
				}
			}
		}
	}
	return findings, nil
}

type EKSControlPlaneDetector struct{}

func (d *EKSControlPlaneDetector) Name() string { return "eks_control_plane_waste" }
func (d *EKSControlPlaneDetector) Category() Category { return CategoryContainer }

func (d *EKSControlPlaneDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "eks_cluster" {
			nodeCount := r.Metadata["node_count"].(int)
			env := r.Tags["environment"]
			if (env == "dev" || env == "test") && nodeCount < 10 {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "EKS cluster for dev/test with <10 nodes",
					Description:      "Control plane costs $73/month, overhead too high for small clusters",
					ResourceARN:      r.ARN,
					EstimatedCost:    73.0,
					EstimatedSavings: 73.0,
					Confidence:       0.85,
					Recommendation:   "Consolidate dev/test into single cluster, use namespaces for isolation",
				})
			}
		}
	}
	return findings, nil
}
