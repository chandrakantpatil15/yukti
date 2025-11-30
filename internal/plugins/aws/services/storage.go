package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
)

// StorageService handles all AWS storage services
type StorageService struct {
	s3Client  *s3.Client
	ec2Client *ec2.Client
	efsClient *efs.Client
	fsxClient *fsx.Client
	region    string
}

// NewStorageService creates a new storage service
func NewStorageService(cfg aws.Config, region string) *StorageService {
	regionalCfg := cfg.Copy()
	regionalCfg.Region = region
	
	return &StorageService{
		s3Client:  s3.NewFromConfig(regionalCfg),
		ec2Client: ec2.NewFromConfig(regionalCfg),
		efsClient: efs.NewFromConfig(regionalCfg),
		fsxClient: fsx.NewFromConfig(regionalCfg),
		region:    region,
	}
}

// SyncS3Buckets syncs S3 buckets
func (s *StorageService) SyncS3Buckets(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 buckets: %w", err)
	}
	
	for _, bucket := range result.Buckets {
		resource := Resource{
			ID:       aws.ToString(bucket.Name),
			Type:     "s3-bucket",
			Provider: "aws",
			Region:   s.region,
			State:    "active",
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncEBSVolumes syncs EBS volumes
func (s *StorageService) SyncEBSVolumes(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.ec2Client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe EBS volumes: %w", err)
	}
	
	for _, volume := range result.Volumes {
		resource := Resource{
			ID:       aws.ToString(volume.VolumeId),
			Type:     "ebs-volume",
			Provider: "aws",
			Region:   s.region,
			State:    string(volume.State),
			Tags:     extractEBSTags(volume.Tags),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncEFSFileSystems syncs EFS file systems
func (s *StorageService) SyncEFSFileSystems(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.efsClient.DescribeFileSystems(ctx, &efs.DescribeFileSystemsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe EFS file systems: %w", err)
	}
	
	for _, fs := range result.FileSystems {
		resource := Resource{
			ID:       aws.ToString(fs.FileSystemId),
			Type:     "efs-filesystem",
			Provider: "aws",
			Region:   s.region,
			State:    string(fs.LifeCycleState),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncFSxFileSystems syncs FSx file systems
func (s *StorageService) SyncFSxFileSystems(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.fsxClient.DescribeFileSystems(ctx, &fsx.DescribeFileSystemsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe FSx file systems: %w", err)
	}
	
	for _, fs := range result.FileSystems {
		resource := Resource{
			ID:       aws.ToString(fs.FileSystemId),
			Type:     "fsx-filesystem",
			Provider: "aws",
			Region:   s.region,
			State:    string(fs.Lifecycle),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// extractEBSTags converts EBS tags to map
func extractEBSTags(tags []ec2types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}