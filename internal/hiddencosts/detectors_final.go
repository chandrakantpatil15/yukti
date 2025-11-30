package hiddencosts

// Final Detectors (3 additional patterns)

type CloudFrontVsS3DirectDetector struct{}

func (d *CloudFrontVsS3DirectDetector) Name() string { return "cloudfront_vs_s3_direct" }
func (d *CloudFrontVsS3DirectDetector) Category() Category { return CategoryNetworking }

func (d *CloudFrontVsS3DirectDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "s3_bucket" {
			if cloudfront, ok := r.Metadata["cloudfront_enabled"].(bool); ok && !cloudfront {
				outboundGB := r.Metadata["outbound_data_gb"].(float64)
				if outboundGB > 1000 {
					s3DirectCost := outboundGB * 0.09
					cloudFrontCost := outboundGB * 0.085
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "S3 direct access for high-volume content",
						Description:      "CloudFront is 5% cheaper and provides caching + CDN",
						ResourceARN:      r.ARN,
						EstimatedCost:    s3DirectCost,
						EstimatedSavings: s3DirectCost - cloudFrontCost,
						Confidence:       0.90,
						Recommendation:   "Use CloudFront distribution for high-volume public content",
					})
				}
			}
		}
	}
	return findings, nil
}

type AWSBackupPremiumDetector struct{}

func (d *AWSBackupPremiumDetector) Name() string { return "aws_backup_premium" }
func (d *AWSBackupPremiumDetector) Category() Category { return CategoryManagedServices }

func (d *AWSBackupPremiumDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "backup_plan" {
			resourceCount := r.Metadata["resource_count"].(int)
			if resourceCount < 10 {
				backupGB := r.Metadata["backup_storage_gb"].(float64)
				awsBackupCost := backupGB * 0.05
				nativeSnapshotCost := backupGB * 0.05 * 0.9 // Slightly cheaper
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityLow,
					Title:            "AWS Backup for <10 resources",
					Description:      "Native snapshots are simpler and slightly cheaper for small deployments",
					ResourceARN:      r.ARN,
					EstimatedCost:    awsBackupCost,
					EstimatedSavings: awsBackupCost - nativeSnapshotCost,
					Confidence:       0.80,
					Recommendation:   "Use native EBS/RDS snapshots for simple backup needs",
				})
			}
		}
	}
	return findings, nil
}

type EventBridgeVsSNSDetector struct{}

func (d *EventBridgeVsSNSDetector) Name() string { return "eventbridge_vs_sns_cost" }
func (d *EventBridgeVsSNSDetector) Category() Category { return CategoryServerless }

func (d *EventBridgeVsSNSDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "eventbridge_bus" {
			events := r.Metadata["monthly_events"].(float64)
			targets := r.Metadata["target_count"].(int)
			
			if targets == 1 && events > 1000000 {
				eventBridgeCost := (events / 1000000) * 1.00
				snsCost := (events / 1000000) * 0.50
				
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityLow,
					Title:            "EventBridge with single target",
					Description:      "SNS is 50% cheaper for simple pub/sub with one subscriber",
					ResourceARN:      r.ARN,
					EstimatedCost:    eventBridgeCost,
					EstimatedSavings: eventBridgeCost - snsCost,
					Confidence:       0.85,
					Recommendation:   "Use SNS for simple pub/sub, EventBridge for complex routing",
				})
			}
		}
	}
	return findings, nil
}
