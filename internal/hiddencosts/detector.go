package hiddencosts

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Detector struct {
	db        *sql.DB
	detectors []DetectorFunc
}

type DetectorFunc interface {
	Name() string
	Category() Category
	Detect(resources []Resource) ([]Finding, error)
}

func NewDetector(db *sql.DB) *Detector {
	d := &Detector{db: db}
	d.registerDetectors()
	return d
}

func (d *Detector) registerDetectors() {
	d.detectors = []DetectorFunc{
		// Data Transfer (15 detectors)
		&CrossAZTransferDetector{},
		&NATGatewayDetector{},
		&CloudFrontOriginDetector{},
		&InterRegionReplicationDetector{},
		&VPCPeeringDetector{},
		&ELBCrossAZDetector{},
		&DirectConnectUnderutilizedDetector{},
		&S3TransferAccelerationDetector{},
		&CloudFrontFieldEncryptionDetector{},
		&GlobalAcceleratorDetector{},
		&PrivateLinkVsPeeringDetector{},
		&InternetVsCloudFrontDetector{},
		&CrossRegionSnapshotDetector{},
		&DataSyncVsS3TransferDetector{},
		&OutboundDataOptimizationDetector{},
		
		// Storage (16 detectors)
		&OrphanedSnapshotsDetector{},
		&S3IntelligentTieringDetector{},
		&RDSBackupRetentionDetector{},
		&GlacierRetrievalDetector{},
		&EFSLifecycleDetector{},
		&EBSUnusedVolumesDetector{},
		&S3VersioningExcessDetector{},
		&AMIUnusedDetector{},
		&GlacierDeepArchiveMisuseDetector{},
		&EBSgp2vsgp3Detector{},
		&S3ObjectLockRetentionDetector{},
		&FSxWindowsOverprovisionedDetector{},
		
		// Compute (16 detectors)
		&DetailedMonitoringDetector{},
		&LambdaMemoryDetector{},
		&FargateOverprovisionedDetector{},
		&BeanstalkNonProdDetector{},
		&EC2BurstableT2Detector{},
		&EC2PreviousGenDetector{},
		&AutoScalingUnusedDetector{},
		&SpotInstanceOpportunityDetector{},
		&SavingsPlansUnderutilizedDetector{},
		&ReservedInstanceWasteDetector{},
		&DedicatedHostsUnderutilizedDetector{},
		&BatchVsSpotDetector{},
		
		// Database (12 detectors)
		&RDSMultiAZNonProdDetector{},
		&DynamoDBOnDemandDetector{},
		&RDSStorageAutoScalingDetector{},
		&ElastiCacheReservedUnusedDetector{},
		&AuroraServerlessV1vsV2Detector{},
		&RDSProxyUnnecessaryDetector{},
		&DocumentDBVsMongoDBDetector{},
		&NeptuneOverprovisionedDetector{},
		
		// Networking (6 detectors)
		&IdleLoadBalancerDetector{},
		&UnattachedEIPDetector{},
		&IdleVPNDetector{},
		&TransitGatewayUnderutilizedDetector{},
		&CloudFrontVsS3DirectDetector{},
		
		// Managed Services (4 detectors)
		&ManagedPrometheusDetector{},
		&TransferFamilyDetector{},
		&AWSBackupPremiumDetector{},
		
		// Serverless (4 detectors)
		&APIGatewayRESTDetector{},
		&StepFunctionsDetector{},
		&EventBridgeVsSNSDetector{},
		
		// Container (2 detectors)
		&ECRScanningDetector{},
		&EKSControlPlaneDetector{},
		
		// Kubernetes (12 detectors)
		&K8sPodCPUOverprovisionDetector{},
		&K8sPodMemoryOverprovisionDetector{},
		&K8sNodeOverprovisionDetector{},
		&K8sSpotOpportunityDetector{},
		&K8sClusterAutoscalerDetector{},
		&K8sPVCWasteDetector{},
		&K8sLoadBalancerWasteDetector{},
		&K8sHPAMisconfigDetector{},
		&K8sNamespaceQuotaDetector{},
		&K8sDaemonSetCostDetector{},
		&K8sGPUIdleDetector{},
		&K8sFargateOveruseDetector{},
		
		// End-of-Life (5 detectors) - CRITICAL
		&EC2EOLDetector{},
		&RDSEOLDetector{},
		&LambdaEOLRuntimeDetector{},
		&EKSEOLVersionDetector{},
		&ElastiCacheEOLVersionDetector{},
	}
}

func (d *Detector) RunDetection(ctx context.Context, tenantID string, resources []Resource) ([]Finding, error) {
	var allFindings []Finding

	for _, detector := range d.detectors {
		findings, err := detector.Detect(resources)
		if err != nil {
			continue
		}
		for i := range findings {
			findings[i].ID = uuid.New().String()
			findings[i].TenantID = tenantID
			findings[i].DetectedAt = time.Now()
		}
		allFindings = append(allFindings, findings...)
	}

	if err := d.storeFindings(ctx, allFindings); err != nil {
		return nil, err
	}

	return allFindings, nil
}

func (d *Detector) storeFindings(ctx context.Context, findings []Finding) error {
	query := `
		INSERT INTO yt_hidden_cost_findings 
		(id, tenant_id, detector_name, category, severity, title, description, 
		 resource_arn, estimated_cost, estimated_savings, confidence, 
		 recommendation, detected_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (tenant_id, resource_arn, detector_name) 
		DO UPDATE SET estimated_cost = $9, estimated_savings = $10, detected_at = $13
	`

	for _, f := range findings {
		_, err := d.db.ExecContext(ctx, query,
			f.ID, f.TenantID, f.DetectorName, f.Category, f.Severity,
			f.Title, f.Description, f.ResourceARN, f.EstimatedCost,
			f.EstimatedSavings, f.Confidence, f.Recommendation, f.DetectedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// Sample Detectors

type CrossAZTransferDetector struct{}

func (d *CrossAZTransferDetector) Name() string { return "cross_az_data_transfer" }
func (d *CrossAZTransferDetector) Category() Category { return CategoryDataTransfer }

func (d *CrossAZTransferDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "rds" {
			if multiAZ, ok := r.Metadata["multi_az"].(bool); ok && multiAZ {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "RDS Multi-AZ cross-AZ data transfer costs",
					Description:      "RDS Multi-AZ incurs cross-AZ data transfer charges",
					ResourceARN:      r.ARN,
					EstimatedCost:    500.0,
					EstimatedSavings: 450.0,
					Confidence:       0.95,
					Recommendation:   "Co-locate read replicas in same AZ as application",
					RemediationSteps: []string{"Create read replica in app AZ", "Update connection string"},
				})
			}
		}
	}
	return findings, nil
}

type NATGatewayDetector struct{}

func (d *NATGatewayDetector) Name() string { return "nat_gateway_waste" }
func (d *NATGatewayDetector) Category() Category { return CategoryDataTransfer }

func (d *NATGatewayDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "nat_gateway" {
			if dataProcessed, ok := r.Metadata["data_processed_gb"].(float64); ok && dataProcessed > 1000 {
				savings := (dataProcessed * 0.045) - 7.20
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityHigh,
					Title:            "NAT Gateway processing high data volume",
					Description:      "Use VPC endpoints for S3/DynamoDB to save 84%",
					ResourceARN:      r.ARN,
					EstimatedCost:    dataProcessed * 0.045,
					EstimatedSavings: savings,
					Confidence:       0.90,
					Recommendation:   "Create VPC gateway endpoints for S3 and DynamoDB",
				})
			}
		}
	}
	return findings, nil
}

type OrphanedSnapshotsDetector struct{}

func (d *OrphanedSnapshotsDetector) Name() string { return "orphaned_snapshots" }
func (d *OrphanedSnapshotsDetector) Category() Category { return CategoryStorage }

func (d *OrphanedSnapshotsDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "ebs_snapshot" {
			if age, ok := r.Metadata["age_days"].(float64); ok && age > 90 {
				if orphaned, ok := r.Metadata["orphaned"].(bool); ok && orphaned {
					size := r.Metadata["size_gb"].(float64)
					cost := size * 0.05
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityMedium,
						Title:            "Orphaned EBS snapshot older than 90 days",
						Description:      "Snapshot has no associated volume or AMI",
						ResourceARN:      r.ARN,
						EstimatedCost:    cost,
						EstimatedSavings: cost,
						Confidence:       0.98,
						Recommendation:   "Delete orphaned snapshot",
					})
				}
			}
		}
	}
	return findings, nil
}

type DetailedMonitoringDetector struct{}

func (d *DetailedMonitoringDetector) Name() string { return "detailed_monitoring_waste" }
func (d *DetailedMonitoringDetector) Category() Category { return CategoryCompute }

func (d *DetailedMonitoringDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "ec2" {
			if detailed, ok := r.Metadata["detailed_monitoring"].(bool); ok && detailed {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityLow,
					Title:            "EC2 detailed monitoring enabled",
					Description:      "Costs $2.10/month per instance for 1-min metrics",
					ResourceARN:      r.ARN,
					EstimatedCost:    2.10,
					EstimatedSavings: 2.10,
					Confidence:       1.0,
					Recommendation:   "Disable detailed monitoring, use 5-min free metrics",
				})
			}
		}
	}
	return findings, nil
}

type IdleLoadBalancerDetector struct{}

func (d *IdleLoadBalancerDetector) Name() string { return "idle_load_balancer" }
func (d *IdleLoadBalancerDetector) Category() Category { return CategoryNetworking }

func (d *IdleLoadBalancerDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	for _, r := range resources {
		if r.Type == "alb" || r.Type == "nlb" {
			if requests, ok := r.Metadata["requests_per_hour"].(float64); ok && requests < 100 {
				findings = append(findings, Finding{
					DetectorName:     d.Name(),
					Category:         d.Category(),
					Severity:         SeverityMedium,
					Title:            "Idle load balancer with minimal traffic",
					Description:      "Load balancer costs $16.20/month with <100 requests/hour",
					ResourceARN:      r.ARN,
					EstimatedCost:    16.20,
					EstimatedSavings: 16.20,
					Confidence:       0.85,
					Recommendation:   "Delete idle load balancer or consolidate apps",
				})
			}
		}
	}
	return findings, nil
}
