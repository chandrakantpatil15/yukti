package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
)

// DatabaseService handles all AWS database services
type DatabaseService struct {
	rdsClient         *rds.Client
	dynamoClient      *dynamodb.Client
	redshiftClient    *redshift.Client
	elasticacheClient *elasticache.Client
	region            string
}

// NewDatabaseService creates a new database service
func NewDatabaseService(cfg aws.Config, region string) *DatabaseService {
	regionalCfg := cfg.Copy()
	regionalCfg.Region = region
	
	return &DatabaseService{
		rdsClient:         rds.NewFromConfig(regionalCfg),
		dynamoClient:      dynamodb.NewFromConfig(regionalCfg),
		redshiftClient:    redshift.NewFromConfig(regionalCfg),
		elasticacheClient: elasticache.NewFromConfig(regionalCfg),
		region:            region,
	}
}

// SyncRDSInstances syncs RDS instances
func (s *DatabaseService) SyncRDSInstances(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.rdsClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe RDS instances: %w", err)
	}
	
	for _, instance := range result.DBInstances {
		resource := Resource{
			ID:       aws.ToString(instance.DBInstanceIdentifier),
			Type:     "rds-instance",
			Provider: "aws",
			Region:   s.region,
			State:    aws.ToString(instance.DBInstanceStatus),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncDynamoDBTables syncs DynamoDB tables
func (s *DatabaseService) SyncDynamoDBTables(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.dynamoClient.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list DynamoDB tables: %w", err)
	}
	
	for _, tableName := range result.TableNames {
		resource := Resource{
			ID:       tableName,
			Type:     "dynamodb-table",
			Provider: "aws",
			Region:   s.region,
			State:    "active",
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncRedshiftClusters syncs Redshift clusters
func (s *DatabaseService) SyncRedshiftClusters(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.redshiftClient.DescribeClusters(ctx, &redshift.DescribeClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe Redshift clusters: %w", err)
	}
	
	for _, cluster := range result.Clusters {
		resource := Resource{
			ID:       aws.ToString(cluster.ClusterIdentifier),
			Type:     "redshift-cluster",
			Provider: "aws",
			Region:   s.region,
			State:    aws.ToString(cluster.ClusterStatus),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncElastiCacheClusters syncs ElastiCache clusters
func (s *DatabaseService) SyncElastiCacheClusters(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.elasticacheClient.DescribeCacheClusters(ctx, &elasticache.DescribeCacheClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe ElastiCache clusters: %w", err)
	}
	
	for _, cluster := range result.CacheClusters {
		resource := Resource{
			ID:       aws.ToString(cluster.CacheClusterId),
			Type:     "elasticache-cluster",
			Provider: "aws",
			Region:   s.region,
			State:    aws.ToString(cluster.CacheClusterStatus),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}