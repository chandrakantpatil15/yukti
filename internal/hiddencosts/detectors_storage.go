package hiddencosts

// Storage Lifecycle Detectors (7 additional patterns)

type S3IntelligentTieringDetector struct{}

func (d *S3IntelligentTieringDetector) Name() string { return "s3_intelligent_tiering_overhead" }
func (d *S3IntelligentTieringDetector) Category() Category { return CategoryStorage }

func (d *S3IntelligentTieringDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "s3_bucket" {
			if tier, ok := r.Metadata["storage_class"].(string); ok && tier == "INTELLIGENT_TIERING" {
				objectCount := r.Metadata["object_count"].(float64)
				avgSize := r.Metadata["avg_object_size_kb"].(float64)
				if avgSize < 128 {
					cost := (objectCount / 1000) * 0.0025
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "S3 Intelligent-Tiering for small objects",
						Description:      "Monitoring fee not cost-effective for <128KB objects",
						ResourceARN:      r.ARN,
						EstimatedCost:    cost,
						EstimatedSavings: cost,
						Confidence:       0.95,
						Recommendation:   "Use lifecycle rules or standard storage",
					})
				}
			}
		}
	}
	return findings, nil
}

type RDSBackupRetentionDetector struct{}

func (d *RDSBackupRetentionDetector) Name() string { return "rds_backup_retention" }
func (d *RDSBackupRetentionDetector) Category() Category { return CategoryStorage }

func (d *RDSBackupRetentionDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "rds" {
			retention := r.Metadata["backup_retention_days"].(int)
			if retention > 7 {
				dbSize := r.Metadata["allocated_storage_gb"].(float64)
				excessBackups := float64(retention-7) * dbSize
				cost := excessBackups * 0.095
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "RDS backup retention >7 days",
					Description:      "Additional backups cost $0.095/GB-month",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.8,
					Confidence:       0.90,
					Recommendation:   "Reduce retention to 7 days, use S3 for long-term",
				})
			}
		}
	}
	return findings, nil
}

type GlacierRetrievalDetector struct{}

func (d *GlacierRetrievalDetector) Name() string { return "glacier_retrieval_risk" }
func (d *GlacierRetrievalDetector) Category() Category { return CategoryStorage }

func (d *GlacierRetrievalDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "s3_bucket" {
			if lifecycle, ok := r.Metadata["lifecycle_glacier"].(bool); ok && lifecycle {
				accessFreq := r.Metadata["access_frequency_days"].(float64)
				if accessFreq < 90 {
					dataGB := r.Metadata["glacier_data_gb"].(float64)
					retrievalCost := dataGB * 0.01
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityHigh,
						Title:            "Glacier data accessed frequently",
						Description:      "Frequent retrieval costs negate Glacier savings",
						ResourceARN:      r.ARN,
						EstimatedCost:    retrievalCost,
						EstimatedSavings: retrievalCost * 0.9,
						Confidence:       0.85,
						Recommendation:   "Use S3 Glacier Instant Retrieval or Standard-IA",
					})
				}
			}
		}
	}
	return findings, nil
}

type EFSLifecycleDetector struct{}

func (d *EFSLifecycleDetector) Name() string { return "efs_lifecycle_not_enabled" }
func (d *EFSLifecycleDetector) Category() Category { return CategoryStorage }

func (d *EFSLifecycleDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "efs" {
			if lifecycle, ok := r.Metadata["lifecycle_enabled"].(bool); ok && !lifecycle {
				sizeGB := r.Metadata["size_gb"].(float64)
				currentCost := sizeGB * 0.30
				iaCost := sizeGB * 0.5 * 0.025 // Assume 50% moves to IA
				savings := currentCost - iaCost - (sizeGB * 0.5 * 0.30)
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "EFS lifecycle policy not enabled",
					Description:      "Save 92% by moving infrequent files to IA storage",
					ResourceARN:      r.ARN,
					EstimatedCost:    currentCost,
					EstimatedSavings: savings,
					Confidence:       0.90,
					Recommendation:   "Enable 30-day lifecycle policy to IA",
				})
			}
		}
	}
	return findings, nil
}

type EBSUnusedVolumesDetector struct{}

func (d *EBSUnusedVolumesDetector) Name() string { return "ebs_unused_volumes" }
func (d *EBSUnusedVolumesDetector) Category() Category { return CategoryStorage }

func (d *EBSUnusedVolumesDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "ebs_volume" {
			if attached, ok := r.Metadata["attached"].(bool); ok && !attached {
				sizeGB := r.Metadata["size_gb"].(float64)
				cost := sizeGB * 0.10
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "Unattached EBS volume",
					Description:      "Paying for unused storage",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost,
					Confidence:       0.98,
					Recommendation:   "Delete volume or create snapshot and delete",
				})
			}
		}
	}
	return findings, nil
}

type S3VersioningExcessDetector struct{}

func (d *S3VersioningExcessDetector) Name() string { return "s3_versioning_excess" }
func (d *S3VersioningExcessDetector) Category() Category { return CategoryStorage }

func (d *S3VersioningExcessDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "s3_bucket" {
			if versioning, ok := r.Metadata["versioning_enabled"].(bool); ok && versioning {
				versionCount := r.Metadata["version_count"].(float64)
				currentCount := r.Metadata["current_count"].(float64)
				if versionCount > currentCount*5 {
					oldVersionsGB := r.Metadata["old_versions_gb"].(float64)
					cost := oldVersionsGB * 0.023
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "S3 versioning with excessive old versions",
						Description:      "Old versions exceed current objects by 5x",
						ResourceARN:      r.ARN,
						EstimatedCost:    cost,
						EstimatedSavings: cost * 0.8,
						Confidence:       0.90,
						Recommendation:   "Add lifecycle rule to delete old versions after 90 days",
					})
				}
			}
		}
	}
	return findings, nil
}

type AMIUnusedDetector struct{}

func (d *AMIUnusedDetector) Name() string { return "ami_unused" }
func (d *AMIUnusedDetector) Category() Category { return CategoryStorage }

func (d *AMIUnusedDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "ami" {
			ageDays := r.Metadata["age_days"].(float64)
			lastUsed := r.Metadata["last_used_days"].(float64)
			if ageDays > 180 && lastUsed > 90 {
				snapshotSizeGB := r.Metadata["snapshot_size_gb"].(float64)
				cost := snapshotSizeGB * 0.05
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityLow,
					Title:            "Unused AMI older than 180 days",
					Description:      "AMI not used in 90 days, snapshots still charged",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost,
					Confidence:       0.85,
					Recommendation:   "Deregister AMI and delete associated snapshots",
				})
			}
		}
	}
	return findings, nil
}
