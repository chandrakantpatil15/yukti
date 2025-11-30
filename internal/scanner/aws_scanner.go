package scanner

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"yukti/internal/aws"
	"yukti/internal/hiddencosts"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/lib/pq"
)

type AWSScanner struct {
	db *sql.DB
}

func NewAWSScanner(db *sql.DB) *AWSScanner {
	return &AWSScanner{db: db}
}

// ScanTenant scans a tenant's AWS account and runs all detectors
func (s *AWSScanner) ScanTenant(ctx context.Context, tenantID int) error {
	log.Printf("[Scanner] ========== STARTING AWS SCAN ===========")
	log.Printf("[Scanner] Tenant ID: %d", tenantID)
	log.Printf("[Scanner] Timestamp: %s", time.Now().Format(time.RFC3339))

	// Get AWS connection details including regions
	var accountID, roleARN, externalID string
	var verified bool
	var lastVerified sql.NullTime
	var regions []string
	tenantIDStr := fmt.Sprintf("%d", tenantID)
	
	log.Printf("[Scanner] Querying AWS connection for tenant: %s", tenantIDStr)
	err := s.db.QueryRowContext(ctx, `
		SELECT account_id, role_arn, external_id, verified, last_verified_at, regions
		FROM yt_aws_connections
		WHERE tenant_id = $1
		LIMIT 1
	`, tenantIDStr).Scan(&accountID, &roleARN, &externalID, &verified, &lastVerified, pq.Array(&regions))

	if err == sql.ErrNoRows {
		log.Printf("[Scanner] ERROR: No AWS connection found for tenant %d", tenantID)
		log.Printf("[Scanner] TROUBLESHOOTING: Customer needs to complete onboarding first")
		return fmt.Errorf("no AWS connection configured for tenant %d", tenantID)
	}

	if err != nil {
		log.Printf("[Scanner] ERROR: Database query failed: %v", err)
		return fmt.Errorf("failed to get AWS connection: %w", err)
	}

	log.Printf("[Scanner] AWS Connection Details:")
	log.Printf("[Scanner]   Account ID: %s", accountID)
	log.Printf("[Scanner]   Role ARN: %s", roleARN)
	log.Printf("[Scanner]   External ID: %s", externalID)
	log.Printf("[Scanner]   Regions: %v", regions)
	log.Printf("[Scanner]   Verified: %t", verified)
	if lastVerified.Valid {
		log.Printf("[Scanner]   Last Verified: %s", lastVerified.Time.Format(time.RFC3339))
	} else {
		log.Printf("[Scanner]   Last Verified: Never")
	}

	// Use all AWS regions for comprehensive scanning
	allAWSRegions := []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1", "eu-north-1",
		"ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ap-northeast-2", "ap-south-1",
		"ca-central-1", "sa-east-1",
	}
	
	// Use configured regions if available, otherwise scan all regions
	if len(regions) == 0 {
		regions = allAWSRegions
		log.Printf("[Scanner] No regions configured, scanning ALL AWS regions: %d total", len(regions))
	} else {
		log.Printf("[Scanner] Using configured regions: %v", regions)
	}

	if !verified {
		log.Printf("[Scanner] ERROR: AWS connection not verified")
		log.Printf("[Scanner] TROUBLESHOOTING: Customer needs to verify IAM role setup")
		return fmt.Errorf("AWS connection not verified for tenant %d", tenantID)
	}

	// Scan all regions (configured or all AWS regions)
	var allResources []hiddencosts.Resource
	log.Printf("[Scanner] ========== MULTI-REGION SCAN STARTING ===========")
	log.Printf("[Scanner] Total regions to scan: %d", len(regions))
	
	for i, region := range regions {
		log.Printf("[Scanner] ========== SCANNING REGION %d/%d: %s ===========", i+1, len(regions), region)
		
		// Assume IAM role for this region
		log.Printf("[Scanner] Assuming IAM role for region: %s", region)
		awsCfg, err := s.assumeRoleWithRegion(ctx, roleARN, externalID, fmt.Sprintf("yukti-scan-%d", tenantID), region)
		if err != nil {
			log.Printf("[Scanner] ERROR: Failed to assume role for region %s: %v", region, err)
			log.Printf("[Scanner] TROUBLESHOOTING: Check IAM role trust policy and permissions")
			continue // Try next region
		}
		log.Printf("[Scanner] ✓ Successfully assumed IAM role for region: %s", region)

		// Fetch AWS resources from this region
		log.Printf("[Scanner] Fetching AWS resources from region: %s", region)
		resources, err := s.fetchResources(ctx, awsCfg)
		if err != nil {
			log.Printf("[Scanner] ERROR: Failed to fetch resources from region %s: %v", region, err)
			continue // Try next region
		}
		
		allResources = append(allResources, resources...)
		if len(resources) > 0 {
			log.Printf("[Scanner] ✓ Region %s scan complete: %d resources found", region, len(resources))
		} else {
			log.Printf("[Scanner] ○ Region %s scan complete: no resources found", region)
		}
	}
	
	resources := allResources

	log.Printf("[Scanner] ========== MULTI-REGION SCAN SUMMARY ===========")
	log.Printf("[Scanner] ✓ Successfully scanned %d regions", len(regions))
	log.Printf("[Scanner] ✓ Total resources found: %d", len(resources))
	
	resourceCount := make(map[string]int)
	for _, resource := range resources {
		resourceCount[resource.Type]++
	}
	for resourceType, count := range resourceCount {
		log.Printf("[Scanner]   - %s: %d", resourceType, count)
	}

	if len(resources) == 0 {
		log.Printf("[Scanner] WARNING: No resources found across all scanned regions")
		log.Printf("[Scanner] TROUBLESHOOTING: Check if resources exist in your AWS account")
	}

	// Store discovered resources in database
	log.Printf("[Scanner] Storing discovered resources in database...")
	err = s.storeResources(ctx, tenantIDStr, accountID, resources)
	if err != nil {
		log.Printf("[Scanner] ERROR: Failed to store resources: %v", err)
		// Continue with detectors even if storage fails
	} else {
		log.Printf("[Scanner] ✓ Successfully stored %d resources", len(resources))
	}

	// Run detectors
	log.Printf("[Scanner] Running cost optimization detectors...")
	detector := hiddencosts.NewDetector(s.db)
	findings, err := detector.RunDetection(ctx, tenantIDStr, resources)
	if err != nil {
		log.Printf("[Scanner] ERROR: Failed to run detectors: %v", err)
		return fmt.Errorf("failed to run detectors: %w", err)
	}

	log.Printf("[Scanner] ✓ Detectors completed successfully")
	log.Printf("[Scanner] Found %d cost optimization opportunities", len(findings))
	log.Printf("[Scanner] ========== SCAN COMPLETED ===========")
	return nil
}

// assumeRoleWithRegion assumes the customer's IAM role for a specific region
func (s *AWSScanner) assumeRoleWithRegion(ctx context.Context, roleARN, externalID, sessionName, region string) (awssdk.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return awssdk.Config{}, err
	}

	stsClient := sts.NewFromConfig(cfg)
	creds := stscreds.NewAssumeRoleProvider(stsClient, roleARN, func(o *stscreds.AssumeRoleOptions) {
		o.ExternalID = awssdk.String(externalID)
		o.RoleSessionName = sessionName
	})

	assumedCfg, err := config.LoadDefaultConfig(ctx, 
		config.WithCredentialsProvider(creds),
		config.WithRegion(region),
	)
	if err != nil {
		return awssdk.Config{}, err
	}

	return assumedCfg, nil
}

// storeResources stores discovered resources in the database
func (s *AWSScanner) storeResources(ctx context.Context, tenantID, accountID string, resources []hiddencosts.Resource) error {
	// First, get or create AWS account record
	var awsAccountDBID int
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO yt_aws_accounts (account_id, tenant_id, account_name, role_arn, external_id, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (account_id, tenant_id) DO UPDATE SET account_name = EXCLUDED.account_name
		RETURNING id
	`, accountID, tenantID, fmt.Sprintf("AWS Account %s", accountID), "arn:aws:iam::" + accountID + ":role/YuktiReadOnlyRole", "yukti-" + tenantID).Scan(&awsAccountDBID)
	
	if err != nil {
		return fmt.Errorf("failed to create/get AWS account record: %w", err)
	}

	// Clear existing resources for this tenant (fresh scan)
	_, err = s.db.ExecContext(ctx, `DELETE FROM yt_tenant_resources WHERE tenant_id = $1`, tenantID)
	if err != nil {
		log.Printf("[Scanner] Warning: Failed to clear existing resources: %v", err)
	}

	// Store each resource
	for _, resource := range resources {
		// Extract region from ARN or metadata
		region := "unknown"
		if regionVal, ok := resource.Metadata["region"]; ok {
			if regionStr, ok := regionVal.(string); ok {
				region = regionStr
			}
		}

		// Extract instance type
		instanceType := ""
		if typeVal, ok := resource.Metadata["instance_type"]; ok {
			if typeStr, ok := typeVal.(string); ok {
				instanceType = typeStr
			}
		}

		// Extract state
		state := "unknown"
		if stateVal, ok := resource.Metadata["state"]; ok {
			if stateStr, ok := stateVal.(string); ok {
				state = stateStr
			}
		}

		// Extract resource ID from metadata or ARN
		resourceID := resource.ARN
		if idVal, ok := resource.Metadata["instance_id"]; ok {
			if idStr, ok := idVal.(string); ok {
				resourceID = idStr
			}
		}
		if idVal, ok := resource.Metadata["db_instance_id"]; ok {
			if idStr, ok := idVal.(string); ok {
				resourceID = idStr
			}
		}
		if idVal, ok := resource.Metadata["bucket_name"]; ok {
			if idStr, ok := idVal.(string); ok {
				resourceID = idStr
			}
		}

		// Extract and serialize tags
		tagsJSON := []byte("{}")
		if tagsVal, ok := resource.Metadata["tags"]; ok {
			if tagsMap, ok := tagsVal.(map[string]string); ok && len(tagsMap) > 0 {
				tagsJSON, _ = json.Marshal(tagsMap)
			}
		}
		
		// Serialize metadata as JSON
		metadataJSON, _ := json.Marshal(resource.Metadata)

		// Estimate monthly cost (placeholder)
		monthlyCost := 0.0
		if resource.Type == "ec2" && instanceType != "" {
			// Simple cost estimation based on instance type
			switch {
			case strings.Contains(instanceType, "t3.large"):
				monthlyCost = 60.0
			case strings.Contains(instanceType, "t3.medium"):
				monthlyCost = 30.0
			case strings.Contains(instanceType, "t3.small"):
				monthlyCost = 15.0
			default:
				monthlyCost = 50.0
			}
		}

		// Insert resource
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO yt_tenant_resources 
			(tenant_id, aws_account_id, resource_id, resource_type, region, instance_type, state, tags, metadata, monthly_cost, last_synced)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		`, tenantID, awsAccountDBID, resourceID, resource.Type, region, instanceType, state, tagsJSON, metadataJSON, monthlyCost)
		
		if err != nil {
			log.Printf("[Scanner] Warning: Failed to store resource %s: %v", resourceID, err)
			continue
		}
		
		log.Printf("[Scanner] ✅ Stored resource: %s (%s) in %s", resourceID, resource.Type, region)
	}

	log.Printf("[Scanner] 💾 DATABASE STORAGE COMPLETE: %d resources stored for tenant %s", len(resources), tenantID)
	return nil
}

// fetchResources fetches AWS resources from customer account
func (s *AWSScanner) fetchResources(ctx context.Context, cfg awssdk.Config) ([]hiddencosts.Resource, error) {
	var resources []hiddencosts.Resource

	ec2Resources, err := s.fetchEC2Instances(ctx, cfg)
	if err != nil {
		log.Printf("[Scanner] Warning: Failed to fetch EC2: %v", err)
	} else {
		resources = append(resources, ec2Resources...)
	}

	rdsResources, err := s.fetchRDSInstances(ctx, cfg)
	if err != nil {
		log.Printf("[Scanner] Warning: Failed to fetch RDS: %v", err)
	} else {
		resources = append(resources, rdsResources...)
	}

	s3Resources, err := s.fetchS3Buckets(ctx, cfg)
	if err != nil {
		log.Printf("[Scanner] Warning: Failed to fetch S3: %v", err)
	} else {
		resources = append(resources, s3Resources...)
	}

	return resources, nil
}

func (s *AWSScanner) fetchEC2Instances(ctx context.Context, cfg awssdk.Config) ([]hiddencosts.Resource, error) {
	log.Printf("[Scanner] Fetching EC2 instances from region: %s", cfg.Region)
	client := ec2.NewFromConfig(cfg)
	result, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		log.Printf("[Scanner] ERROR: EC2 DescribeInstances failed: %v", err)
		return nil, err
	}

	var resources []hiddencosts.Resource
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Build complete metadata from ALL AWS fields
			metadata := map[string]interface{}{
				"instance_id":   *instance.InstanceId,
				"instance_type": string(instance.InstanceType),
				"state":         string(instance.State.Name),
				"region":        cfg.Region,
			}
			
			if instance.Placement != nil {
				if instance.Placement.AvailabilityZone != nil {
					metadata["availability_zone"] = *instance.Placement.AvailabilityZone
				}
				metadata["tenancy"] = string(instance.Placement.Tenancy)
			}
			if instance.LaunchTime != nil {
				metadata["launch_time"] = instance.LaunchTime.Format(time.RFC3339)
			}
			if len(instance.Platform) > 0 {
				metadata["platform"] = string(instance.Platform)
			} else {
				metadata["platform"] = "linux"
			}
			if instance.VpcId != nil {
				metadata["vpc_id"] = *instance.VpcId
			}
			if instance.SubnetId != nil {
				metadata["subnet_id"] = *instance.SubnetId
			}
			if instance.PrivateIpAddress != nil {
				metadata["private_ip"] = *instance.PrivateIpAddress
			}
			if instance.PublicIpAddress != nil {
				metadata["public_ip"] = *instance.PublicIpAddress
			}
			if instance.ImageId != nil {
				metadata["ami_id"] = *instance.ImageId
			}
			if instance.KeyName != nil {
				metadata["key_name"] = *instance.KeyName
			}
			if len(instance.Architecture) > 0 {
				metadata["architecture"] = string(instance.Architecture)
			}
			if len(instance.RootDeviceType) > 0 {
				metadata["root_device_type"] = string(instance.RootDeviceType)
			}
			if instance.EbsOptimized != nil {
				metadata["ebs_optimized"] = *instance.EbsOptimized
			}
			if instance.Monitoring != nil {
				metadata["detailed_monitoring"] = string(instance.Monitoring.State) == "enabled"
			}
			if instance.IamInstanceProfile != nil && instance.IamInstanceProfile.Arn != nil {
				metadata["iam_instance_profile"] = *instance.IamInstanceProfile.Arn
			}
			if instance.PrivateDnsName != nil {
				metadata["private_dns"] = *instance.PrivateDnsName
			}
			if instance.PublicDnsName != nil {
				metadata["public_dns"] = *instance.PublicDnsName
			}
			if instance.StateTransitionReason != nil {
				metadata["state_transition_reason"] = *instance.StateTransitionReason
			}
			if len(instance.SecurityGroups) > 0 {
				sgs := []string{}
				for _, sg := range instance.SecurityGroups {
					if sg.GroupId != nil {
						sgs = append(sgs, *sg.GroupId)
					}
				}
				metadata["security_groups"] = sgs
			}
			if len(instance.NetworkInterfaces) > 0 {
				metadata["network_interface_count"] = len(instance.NetworkInterfaces)
			}
			if len(instance.BlockDeviceMappings) > 0 {
				metadata["block_device_count"] = len(instance.BlockDeviceMappings)
			}
			
			// Extract ALL tags
			tags := make(map[string]string)
			for _, tag := range instance.Tags {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}
			metadata["tags"] = tags
			
			// Fetch CloudWatch metrics for running instances (optional - don't fail scan if metrics unavailable)
			if string(instance.State.Name) == "running" {
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("[Scanner] Warning: CloudWatch metrics fetch failed for %s: %v", *instance.InstanceId, r)
						}
					}()
					
					cloudwatchClient := aws.NewCloudWatchClient(cfg)
					if cloudwatchClient != nil {
						metrics, err := cloudwatchClient.GetEC2Metrics(ctx, *instance.InstanceId)
						if err == nil && len(metrics) > 0 {
							for _, metric := range metrics {
								if len(metric.Values) > 0 {
									var sum, count float64
									for _, v := range metric.Values {
										sum += v.Value
										count++
									}
									if count > 0 {
										metadata["avg_"+strings.ToLower(metric.MetricName)] = sum / count
									}
								}
							}
							log.Printf("[Scanner] EC2: %s - CloudWatch metrics fetched", *instance.InstanceId)
						}
					}
				}()
			}
			
			log.Printf("[Scanner] EC2: %s (%s) - %d tags", *instance.InstanceId, string(instance.InstanceType), len(tags))
			
			resources = append(resources, hiddencosts.Resource{
				ARN:  fmt.Sprintf("arn:aws:ec2:%s::instance/%s", cfg.Region, *instance.InstanceId),
				Type: "ec2",
				Metadata: metadata,
			})
		}
	}

	log.Printf("[Scanner] ✓ Found %d EC2 instances", len(resources))
	return resources, nil
}

func (s *AWSScanner) fetchRDSInstances(ctx context.Context, cfg awssdk.Config) ([]hiddencosts.Resource, error) {
	log.Printf("[Scanner] Fetching RDS instances from region: %s", cfg.Region)
	client := rds.NewFromConfig(cfg)
	result, err := client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		log.Printf("[Scanner] ERROR: RDS DescribeDBInstances failed: %v", err)
		return nil, err
	}

	var resources []hiddencosts.Resource
	for _, instance := range result.DBInstances {
		metadata := map[string]interface{}{
			"db_instance_id": *instance.DBInstanceIdentifier,
			"engine":         *instance.Engine,
			"region":         cfg.Region,
		}
		
		if instance.DBInstanceClass != nil {
			metadata["instance_class"] = *instance.DBInstanceClass
		}
		if instance.EngineVersion != nil {
			metadata["engine_version"] = *instance.EngineVersion
		}
		if instance.DBInstanceStatus != nil {
			metadata["status"] = *instance.DBInstanceStatus
		}
		metadata["multi_az"] = instance.MultiAZ
		if instance.AvailabilityZone != nil {
			metadata["availability_zone"] = *instance.AvailabilityZone
		}
		metadata["allocated_storage_gb"] = instance.AllocatedStorage
		if instance.StorageType != nil {
			metadata["storage_type"] = *instance.StorageType
		}
		if instance.Iops != nil {
			metadata["iops"] = *instance.Iops
		}
		metadata["storage_encrypted"] = instance.StorageEncrypted
		metadata["publicly_accessible"] = instance.PubliclyAccessible
		if instance.Endpoint != nil {
			if instance.Endpoint.Address != nil {
				metadata["endpoint_address"] = *instance.Endpoint.Address
			}
			metadata["endpoint_port"] = instance.Endpoint.Port
		}
		if instance.MasterUsername != nil {
			metadata["master_username"] = *instance.MasterUsername
		}
		if instance.DBName != nil {
			metadata["database_name"] = *instance.DBName
		}
		metadata["backup_retention_days"] = instance.BackupRetentionPeriod
		if instance.PreferredBackupWindow != nil {
			metadata["backup_window"] = *instance.PreferredBackupWindow
		}
		if instance.PreferredMaintenanceWindow != nil {
			metadata["maintenance_window"] = *instance.PreferredMaintenanceWindow
		}
		if instance.InstanceCreateTime != nil {
			metadata["create_time"] = instance.InstanceCreateTime.Format(time.RFC3339)
		}
		if len(instance.VpcSecurityGroups) > 0 {
			sgs := []string{}
			for _, sg := range instance.VpcSecurityGroups {
				if sg.VpcSecurityGroupId != nil {
					sgs = append(sgs, *sg.VpcSecurityGroupId)
				}
			}
			metadata["security_groups"] = sgs
		}
		if len(instance.DBSubnetGroup.Subnets) > 0 {
			metadata["subnet_count"] = len(instance.DBSubnetGroup.Subnets)
		}
		
		tags := make(map[string]string)
		for _, tag := range instance.TagList {
			if tag.Key != nil && tag.Value != nil {
				tags[*tag.Key] = *tag.Value
			}
		}
		metadata["tags"] = tags
		
		// Fetch CloudWatch metrics for available instances (optional - don't fail scan if metrics unavailable)
		if instance.DBInstanceStatus != nil && *instance.DBInstanceStatus == "available" {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[Scanner] Warning: CloudWatch metrics fetch failed for %s: %v", *instance.DBInstanceIdentifier, r)
					}
				}()
				
				cloudwatchClient := aws.NewCloudWatchClient(cfg)
				if cloudwatchClient != nil {
					metrics, err := cloudwatchClient.GetRDSMetrics(ctx, *instance.DBInstanceIdentifier)
					if err == nil && len(metrics) > 0 {
						for _, metric := range metrics {
							if len(metric.Values) > 0 {
								var sum, count float64
								for _, v := range metric.Values {
									sum += v.Value
									count++
								}
								if count > 0 {
									metadata["avg_"+strings.ToLower(metric.MetricName)] = sum / count
								}
							}
						}
						log.Printf("[Scanner] RDS: %s - CloudWatch metrics fetched", *instance.DBInstanceIdentifier)
					}
				}
			}()
		}
		
		log.Printf("[Scanner] RDS: %s (%s) - %d tags", *instance.DBInstanceIdentifier, *instance.Engine, len(tags))
		
		resources = append(resources, hiddencosts.Resource{
			ARN:  *instance.DBInstanceArn,
			Type: "rds",
			Metadata: metadata,
		})
	}

	log.Printf("[Scanner] ✓ Found %d RDS instances", len(resources))
	return resources, nil
}

func (s *AWSScanner) fetchS3Buckets(ctx context.Context, cfg awssdk.Config) ([]hiddencosts.Resource, error) {
	log.Printf("[Scanner] Fetching S3 buckets (global service)")
	client := s3.NewFromConfig(cfg)
	result, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Printf("[Scanner] ERROR: S3 ListBuckets failed: %v", err)
		return nil, err
	}

	var resources []hiddencosts.Resource
	for _, bucket := range result.Buckets {
		metadata := map[string]interface{}{
			"bucket_name": *bucket.Name,
			"region":      cfg.Region,
		}
		
		if bucket.CreationDate != nil {
			metadata["creation_date"] = bucket.CreationDate.Format(time.RFC3339)
		}
		
		// Get bucket location
		location, err := client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{Bucket: bucket.Name})
		if err == nil && location.LocationConstraint != "" {
			metadata["bucket_region"] = string(location.LocationConstraint)
		}
		
		// Get bucket versioning
		versioning, err := client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{Bucket: bucket.Name})
		if err == nil {
			metadata["versioning"] = string(versioning.Status)
		}
		
		// Get bucket encryption
		encryption, err := client.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{Bucket: bucket.Name})
		if err == nil && len(encryption.ServerSideEncryptionConfiguration.Rules) > 0 {
			metadata["encryption_enabled"] = true
		} else {
			metadata["encryption_enabled"] = false
		}
		
		// Get bucket tags
		tagsResult, err := client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{Bucket: bucket.Name})
		tags := make(map[string]string)
		if err == nil {
			for _, tag := range tagsResult.TagSet {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}
		}
		metadata["tags"] = tags
		
		// Fetch CloudWatch metrics for S3 buckets (optional - don't fail scan if metrics unavailable)
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Scanner] Warning: CloudWatch metrics fetch failed for %s: %v", *bucket.Name, r)
				}
			}()
			
			cloudwatchClient := aws.NewCloudWatchClient(cfg)
			if cloudwatchClient != nil {
				metrics, err := cloudwatchClient.GetS3Metrics(ctx, *bucket.Name)
				if err == nil && len(metrics) > 0 {
					for _, metric := range metrics {
						if len(metric.Values) > 0 {
							latestValue := metric.Values[len(metric.Values)-1].Value
							metadata[strings.ToLower(metric.MetricName)] = latestValue
						}
					}
					log.Printf("[Scanner] S3: %s - CloudWatch metrics fetched", *bucket.Name)
				}
			}
		}()
		
		log.Printf("[Scanner] S3: %s - %d tags", *bucket.Name, len(tags))
		
		resources = append(resources, hiddencosts.Resource{
			ARN:  fmt.Sprintf("arn:aws:s3:::%s", *bucket.Name),
			Type: "s3",
			Metadata: metadata,
		})
	}

	log.Printf("[Scanner] ✓ Found %d S3 buckets", len(resources))
	return resources, nil
}


