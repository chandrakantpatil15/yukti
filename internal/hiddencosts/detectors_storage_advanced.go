package hiddencosts

// Advanced Storage Detectors (4 additional patterns)

type GlacierDeepArchiveMisuseDetector struct{}

func (d *GlacierDeepArchiveMisuseDetector) Name() string { return "glacier_deep_archive_misuse" }
func (d *GlacierDeepArchiveMisuseDetector) Category() Category { return CategoryStorage }

func (d *GlacierDeepArchiveMisuseDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "s3_bucket" {
			if storageClass, ok := r.Metadata["storage_class"].(string); ok && storageClass == "DEEP_ARCHIVE" {
				accessFreqDays := r.Metadata["access_frequency_days"].(float64)
				if accessFreqDays < 180 {
					dataGB := r.Metadata["deep_archive_data_gb"].(float64)
					retrievalCost := dataGB * 0.02
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityHigh,
						Title:            "Glacier Deep Archive accessed within 180 days",
						Description:      "Deep Archive has 12-hour retrieval time and high retrieval costs",
						ResourceARN:      r.ARN,
						EstimatedCost:    retrievalCost,
						EstimatedSavings: retrievalCost * 0.8,
						Confidence:       0.90,
						Recommendation:   "Use Glacier Flexible Retrieval for data accessed within 180 days",
					})
				}
			}
		}
	}
	return findings, nil
}

type EBSgp2vsgp3Detector struct{}

func (d *EBSgp2vsgp3Detector) Name() string { return "ebs_gp2_vs_gp3" }
func (d *EBSgp2vsgp3Detector) Category() Category { return CategoryStorage }

func (d *EBSgp2vsgp3Detector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "ebs_volume" {
			volumeType := r.Metadata["volume_type"].(string)
			if volumeType == "gp2" {
				sizeGB := r.Metadata["size_gb"].(float64)
				gp2Cost := sizeGB * 0.10
				gp3Cost := sizeGB * 0.08
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "EBS gp2 volume (gp3 is 20% cheaper)",
					Description:      "gp3 offers same performance at 20% lower cost",
					ResourceARN:      r.ARN,
					EstimatedCost:    gp2Cost,
					EstimatedSavings: gp2Cost - gp3Cost,
					Confidence:       0.98,
					Recommendation:   "Migrate to gp3 volume type",
				})
			}
		}
	}
	return findings, nil
}

type S3ObjectLockRetentionDetector struct{}

func (d *S3ObjectLockRetentionDetector) Name() string { return "s3_object_lock_unnecessary" }
func (d *S3ObjectLockRetentionDetector) Category() Category { return CategoryStorage }

func (d *S3ObjectLockRetentionDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "s3_bucket" {
			if objectLock, ok := r.Metadata["object_lock_enabled"].(bool); ok && objectLock {
				retentionDays := r.Metadata["retention_days"].(int)
				if retentionDays > 2555 { // 7 years
					dataGB := r.Metadata["locked_data_gb"].(float64)
					cost := dataGB * 0.023 * float64(retentionDays) / 30
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "S3 Object Lock retention >7 years",
						Description:      "Extremely long retention may not be necessary",
						ResourceARN:      r.ARN,
						EstimatedCost:    cost,
						EstimatedSavings: cost * 0.3,
						Confidence:       0.75,
						Recommendation:   "Review compliance requirements, reduce retention if possible",
					})
				}
			}
		}
	}
	return findings, nil
}

type FSxWindowsOverprovisionedDetector struct{}

func (d *FSxWindowsOverprovisionedDetector) Name() string { return "fsx_windows_overprovisioned" }
func (d *FSxWindowsOverprovisionedDetector) Category() Category { return CategoryStorage }

func (d *FSxWindowsOverprovisionedDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "fsx_windows" {
			capacityGB := r.Metadata["storage_capacity_gb"].(float64)
			usedGB := r.Metadata["used_storage_gb"].(float64)
			if usedGB < capacityGB*0.5 {
				throughput := r.Metadata["throughput_mbps"].(float64)
				cost := (capacityGB * 0.13) + (throughput * 2.20)
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "FSx for Windows over-provisioned",
					Description:      "Storage usage <50% of capacity",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.4,
					Confidence:       0.85,
					Recommendation:   "Reduce storage capacity to match actual usage",
				})
			}
		}
	}
	return findings, nil
}
