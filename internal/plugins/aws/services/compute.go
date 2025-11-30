package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/batch"
)

// ComputeService handles all AWS compute services
type ComputeService struct {
	ec2Client    *ec2.Client
	lambdaClient *lambda.Client
	ecsClient    *ecs.Client
	eksClient    *eks.Client
	batchClient  *batch.Client
	region       string
}

// NewComputeService creates a new compute service
func NewComputeService(cfg aws.Config, region string) *ComputeService {
	regionalCfg := cfg.Copy()
	regionalCfg.Region = region
	
	return &ComputeService{
		ec2Client:    ec2.NewFromConfig(regionalCfg),
		lambdaClient: lambda.NewFromConfig(regionalCfg),
		ecsClient:    ecs.NewFromConfig(regionalCfg),
		eksClient:    eks.NewFromConfig(regionalCfg),
		batchClient:  batch.NewFromConfig(regionalCfg),
		region:       region,
	}
}

// SyncEC2Instances syncs EC2 instances
func (s *ComputeService) SyncEC2Instances(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe EC2 instances: %w", err)
	}
	
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			resource := Resource{
				ID:       aws.ToString(instance.InstanceId),
				Type:     "ec2",
				Provider: "aws",
				Region:   s.region,
				State:    string(instance.State.Name),
				Tags:     extractTags(instance.Tags),
				LastSeen: time.Now(),
			}
			resources = append(resources, resource)
		}
	}
	
	return resources, nil
}

// SyncLambdaFunctions syncs Lambda functions
func (s *ComputeService) SyncLambdaFunctions(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.lambdaClient.ListFunctions(ctx, &lambda.ListFunctionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Lambda functions: %w", err)
	}
	
	for _, function := range result.Functions {
		resource := Resource{
			ID:       aws.ToString(function.FunctionName),
			Type:     "lambda",
			Provider: "aws",
			Region:   s.region,
			State:    string(function.State),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncECSClusters syncs ECS clusters
func (s *ComputeService) SyncECSClusters(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.ecsClient.ListClusters(ctx, &ecs.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ECS clusters: %w", err)
	}
	
	for _, clusterArn := range result.ClusterArns {
		resource := Resource{
			ID:       clusterArn,
			Type:     "ecs-cluster",
			Provider: "aws",
			Region:   s.region,
			State:    "active",
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncEKSClusters syncs EKS clusters
func (s *ComputeService) SyncEKSClusters(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.eksClient.ListClusters(ctx, &eks.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list EKS clusters: %w", err)
	}
	
	for _, clusterName := range result.Clusters {
		resource := Resource{
			ID:       clusterName,
			Type:     "eks-cluster",
			Provider: "aws",
			Region:   s.region,
			State:    "active",
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncBatchJobs syncs AWS Batch jobs
func (s *ComputeService) SyncBatchJobs(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.batchClient.DescribeJobQueues(ctx, &batch.DescribeJobQueuesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe Batch job queues: %w", err)
	}
	
	for _, queue := range result.JobQueues {
		resource := Resource{
			ID:       aws.ToString(queue.JobQueueName),
			Type:     "batch-queue",
			Provider: "aws",
			Region:   s.region,
			State:    string(queue.State),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// extractTags converts EC2 tags to map
func extractTags(tags []ec2types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}

// Resource represents a cloud resource
type Resource struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Provider    string            `json:"provider"`
	Region      string            `json:"region"`
	State       string            `json:"state"`
	Cost        float64           `json:"cost"`
	Tags        map[string]string `json:"tags"`
	Utilization float64           `json:"utilization"`
	LastSeen    time.Time         `json:"last_seen"`
}