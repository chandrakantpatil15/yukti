package hiddencosts

// Advanced Database Detectors (4 additional patterns)

type AuroraServerlessV1vsV2Detector struct{}

func (d *AuroraServerlessV1vsV2Detector) Name() string { return "aurora_serverless_v1_vs_v2" }
func (d *AuroraServerlessV1vsV2Detector) Category() Category { return CategoryDatabase }

func (d *AuroraServerlessV1vsV2Detector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "aurora_serverless" {
			version := r.Metadata["version"].(string)
			if version == "v1" {
				acu := r.Metadata["avg_acu"].(float64)
				v1Cost := acu * 0.06 * 730
				v2Cost := acu * 0.12 * 730 * 0.5 // v2 scales faster, uses less time
				if v2Cost < v1Cost {
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "Aurora Serverless v1 (v2 is more cost-effective)",
						Description:      "v2 has instant scaling and better price/performance",
						ResourceARN:      r.ARN,
						EstimatedCost:    v1Cost,
						EstimatedSavings: v1Cost - v2Cost,
						Confidence:       0.85,
						Recommendation:   "Migrate to Aurora Serverless v2",
					})
				}
			}
		}
	}
	return findings, nil
}

type RDSProxyUnnecessaryDetector struct{}

func (d *RDSProxyUnnecessaryDetector) Name() string { return "rds_proxy_unnecessary" }
func (d *RDSProxyUnnecessaryDetector) Category() Category { return CategoryDatabase }

func (d *RDSProxyUnnecessaryDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "rds_proxy" {
			connections := r.Metadata["avg_connections"].(float64)
			if connections < 100 {
				vCPU := r.Metadata["vcpu"].(float64)
				cost := vCPU * 0.015 * 730
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "RDS Proxy with <100 connections",
					Description:      "RDS Proxy costs $0.015/vCPU/hour, not justified for low connections",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.9,
					Confidence:       0.90,
					Recommendation:   "Remove RDS Proxy, use direct connections",
				})
			}
		}
	}
	return findings, nil
}

type DocumentDBVsMongoDBDetector struct{}

func (d *DocumentDBVsMongoDBDetector) Name() string { return "documentdb_vs_mongodb_cost" }
func (d *DocumentDBVsMongoDBDetector) Category() Category { return CategoryDatabase }

func (d *DocumentDBVsMongoDBDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "documentdb" {
			_ = r.Metadata["instance_class"].(string) // instanceClass
			instanceCount := r.Metadata["instance_count"].(int)
			_ = r.Metadata["storage_gb"].(float64) // storageGB
			
			// DocumentDB pricing
			docDBCost := r.Metadata["monthly_cost"].(float64)
			
			// Rough MongoDB Atlas equivalent
			atlasEquivalent := float64(instanceCount) * 200 // Approximate
			
			if docDBCost > atlasEquivalent*1.5 {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "DocumentDB significantly more expensive than MongoDB Atlas",
					Description:      "Consider MongoDB Atlas for potential cost savings",
					ResourceARN:      r.ARN,
					EstimatedCost:    docDBCost,
					EstimatedSavings: docDBCost - atlasEquivalent,
					Confidence:       0.70,
					Recommendation:   "Evaluate MongoDB Atlas pricing for your workload",
				})
			}
		}
	}
	return findings, nil
}

type NeptuneOverprovisionedDetector struct{}

func (d *NeptuneOverprovisionedDetector) Name() string { return "neptune_overprovisioned" }
func (d *NeptuneOverprovisionedDetector) Category() Category { return CategoryDatabase }

func (d *NeptuneOverprovisionedDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "neptune" {
			storageGB := r.Metadata["storage_gb"].(float64)
			vertexCount := r.Metadata["vertex_count"].(float64)
			edgeCount := r.Metadata["edge_count"].(float64)
			
			// Rough estimate: 1M vertices + edges = ~10GB
			estimatedStorageNeeded := (vertexCount + edgeCount) / 100000
			
			if storageGB > estimatedStorageNeeded*2 {
				instanceCost := r.Metadata["instance_cost"].(float64)
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "Neptune over-provisioned for graph size",
					Description:      "Instance size larger than needed for current graph",
					ResourceARN:      r.ARN,
					EstimatedCost:    instanceCost,
					EstimatedSavings: instanceCost * 0.4,
					Confidence:       0.75,
					Recommendation:   "Right-size Neptune instance based on graph size",
				})
			}
		}
	}
	return findings, nil
}
