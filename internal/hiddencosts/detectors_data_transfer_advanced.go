package hiddencosts

// Advanced Data Transfer Detectors (8 additional patterns)

type S3TransferAccelerationDetector struct{}

func (d *S3TransferAccelerationDetector) Name() string { return "s3_transfer_acceleration_waste" }
func (d *S3TransferAccelerationDetector) Category() Category { return CategoryDataTransfer }

func (d *S3TransferAccelerationDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "s3_bucket" {
			if accel, ok := r.Metadata["transfer_acceleration_enabled"].(bool); ok && accel {
				dataGB := r.Metadata["upload_data_gb"].(float64)
				distance := r.Metadata["avg_distance_km"].(float64)
				if distance < 1000 || dataGB < 100 {
					accelerationCost := dataGB * 0.04
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "S3 Transfer Acceleration for nearby/low-volume uploads",
						Description:      "Transfer Acceleration costs $0.04/GB, not worth it for <1000km or <100GB",
						ResourceARN:      r.ARN,
						EstimatedCost:    accelerationCost,
						EstimatedSavings: accelerationCost * 0.9,
						Confidence:       0.90,
						Recommendation:   "Disable Transfer Acceleration for nearby regions or low-volume uploads",
					})
				}
			}
		}
	}
	return findings, nil
}

type CloudFrontFieldEncryptionDetector struct{}

func (d *CloudFrontFieldEncryptionDetector) Name() string { return "cloudfront_field_encryption_overhead" }
func (d *CloudFrontFieldEncryptionDetector) Category() Category { return CategoryDataTransfer }

func (d *CloudFrontFieldEncryptionDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "cloudfront" {
			if fieldEnc, ok := r.Metadata["field_level_encryption"].(bool); ok && fieldEnc {
				requests := r.Metadata["monthly_requests"].(float64)
				cost := (requests / 10000) * 0.02
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityLow,
					Title:            "CloudFront field-level encryption overhead",
					Description:      "Field-level encryption costs $0.02 per 10,000 requests",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.5,
					Confidence:       0.85,
					Recommendation:   "Use application-level encryption if field-level not required",
				})
			}
		}
	}
	return findings, nil
}

type GlobalAcceleratorDetector struct{}

func (d *GlobalAcceleratorDetector) Name() string { return "global_accelerator_underutilized" }
func (d *GlobalAcceleratorDetector) Category() Category { return CategoryDataTransfer }

func (d *GlobalAcceleratorDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "global_accelerator" {
			dataGB := r.Metadata["data_processed_gb"].(float64)
			if dataGB < 100 {
				cost := 0.025 * 730 // $0.025/hour
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "Global Accelerator processing <100GB/month",
					Description:      "Fixed cost of $18.25/month not justified for low traffic",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.8,
					Confidence:       0.90,
					Recommendation:   "Use CloudFront or direct routing for low-volume traffic",
				})
			}
		}
	}
	return findings, nil
}

type PrivateLinkVsPeeringDetector struct{}

func (d *PrivateLinkVsPeeringDetector) Name() string { return "privatelink_vs_peering_cost" }
func (d *PrivateLinkVsPeeringDetector) Category() Category { return CategoryDataTransfer }

func (d *PrivateLinkVsPeeringDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "vpc_endpoint" {
			endpointType := r.Metadata["endpoint_type"].(string)
			if endpointType == "Interface" {
				dataGB := r.Metadata["data_processed_gb"].(float64)
				privateLinkCost := (0.01 * 730) + (dataGB * 0.01)
				peeringCost := dataGB * 0.01
				if privateLinkCost > peeringCost*1.5 {
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "PrivateLink more expensive than VPC Peering",
						Description:      "PrivateLink has $7.30/month fixed cost + data charges",
						ResourceARN:      r.ARN,
						EstimatedCost:    privateLinkCost,
						EstimatedSavings: privateLinkCost - peeringCost,
						Confidence:       0.85,
						Recommendation:   "Use VPC Peering if both VPCs in same account/region",
					})
				}
			}
		}
	}
	return findings, nil
}

type InternetVsCloudFrontDetector struct{}

func (d *InternetVsCloudFrontDetector) Name() string { return "internet_vs_cloudfront_cost" }
func (d *InternetVsCloudFrontDetector) Category() Category { return CategoryDataTransfer }

func (d *InternetVsCloudFrontDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "s3_bucket" {
			if public, ok := r.Metadata["public_access"].(bool); ok && public {
				dataGB := r.Metadata["outbound_data_gb"].(float64)
				if dataGB > 1000 {
					directCost := dataGB * 0.09
					cloudFrontCost := dataGB * 0.085
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "S3 direct internet transfer >1TB/month",
						Description:      "CloudFront is 5% cheaper and provides caching",
						ResourceARN:      r.ARN,
						EstimatedCost:    directCost,
						EstimatedSavings: directCost - cloudFrontCost,
						Confidence:       0.90,
						Recommendation:   "Use CloudFront for high-volume public content",
					})
				}
			}
		}
	}
	return findings, nil
}

type CrossRegionSnapshotDetector struct{}

func (d *CrossRegionSnapshotDetector) Name() string { return "cross_region_snapshot_copy" }
func (d *CrossRegionSnapshotDetector) Category() Category { return CategoryDataTransfer }

func (d *CrossRegionSnapshotDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "ebs_snapshot" {
			if crossRegion, ok := r.Metadata["cross_region_copy"].(bool); ok && crossRegion {
				sizeGB := r.Metadata["size_gb"].(float64)
				cost := sizeGB * 0.02
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "Cross-region EBS snapshot copy",
					Description:      "Cross-region copy costs $0.02/GB",
					ResourceARN:      r.ARN,
					EstimatedCost:    cost,
					EstimatedSavings: cost * 0.3,
					Confidence:       0.95,
					Recommendation:   "Use same-region snapshots unless DR required",
				})
			}
		}
	}
	return findings, nil
}

type DataSyncVsS3TransferDetector struct{}

func (d *DataSyncVsS3TransferDetector) Name() string { return "datasync_vs_s3_transfer" }
func (d *DataSyncVsS3TransferDetector) Category() Category { return CategoryDataTransfer }

func (d *DataSyncVsS3TransferDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "datasync_task" {
			dataGB := r.Metadata["data_transferred_gb"].(float64)
			frequency := r.Metadata["frequency_hours"].(float64)
			if frequency > 24 && dataGB < 100 {
				dataSyncCost := dataGB * 0.0125
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityLow,
					Title:            "DataSync for infrequent, low-volume transfers",
					Description:      "DataSync costs $0.0125/GB, use S3 CLI for simple transfers",
					ResourceARN:      r.ARN,
					EstimatedCost:    dataSyncCost,
					EstimatedSavings: dataSyncCost * 0.5,
					Confidence:       0.80,
					Recommendation:   "Use S3 CLI or SDK for simple, infrequent transfers",
				})
			}
		}
	}
	return findings, nil
}

type OutboundDataOptimizationDetector struct{}

func (d *OutboundDataOptimizationDetector) Name() string { return "outbound_data_optimization" }
func (d *OutboundDataOptimizationDetector) Category() Category { return CategoryDataTransfer }

func (d *OutboundDataOptimizationDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "ec2" || r.Type == "alb" {
			// Safely check if outbound_data_gb exists and is not nil
			if outboundData, ok := r.Metadata["outbound_data_gb"]; ok && outboundData != nil {
				if outboundGB, ok := outboundData.(float64); ok && outboundGB > 1000 {
					if compressed, ok := r.Metadata["compression_enabled"].(bool); ok && !compressed {
						cost := outboundGB * 0.09
						findings = append(findings, Finding{
							DetectorName:     d.Name(),
							Category:         d.Category(),
							Severity:         SeverityMedium,
							Title:            "High outbound traffic without compression",
							Description:      "Compression can reduce data transfer by 60-80%",
							ResourceARN:      r.ARN,
							EstimatedCost:    cost,
							EstimatedSavings: cost * 0.7,
							Confidence:       0.85,
							Recommendation:   "Enable gzip/brotli compression for text content",
						})
					}
				}
			}
		}
	}
	return findings, nil
}
