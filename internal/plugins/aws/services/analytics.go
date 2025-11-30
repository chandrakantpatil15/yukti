package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/aws/aws-sdk-go-v2/service/athena"
)

// AnalyticsService handles all AWS analytics services
type AnalyticsService struct {
	emrClient     *emr.Client
	kinesisClient *kinesis.Client
	glueClient    *glue.Client
	athenaClient  *athena.Client
	region        string
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(cfg aws.Config, region string) *AnalyticsService {
	regionalCfg := cfg.Copy()
	regionalCfg.Region = region
	
	return &AnalyticsService{
		emrClient:     emr.NewFromConfig(regionalCfg),
		kinesisClient: kinesis.NewFromConfig(regionalCfg),
		glueClient:    glue.NewFromConfig(regionalCfg),
		athenaClient:  athena.NewFromConfig(regionalCfg),
		region:        region,
	}
}

// SyncEMRClusters syncs EMR clusters
func (s *AnalyticsService) SyncEMRClusters(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.emrClient.ListClusters(ctx, &emr.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list EMR clusters: %w", err)
	}
	
	for _, cluster := range result.Clusters {
		resource := Resource{
			ID:       aws.ToString(cluster.Id),
			Type:     "emr-cluster",
			Provider: "aws",
			Region:   s.region,
			State:    string(cluster.Status.State),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncKinesisStreams syncs Kinesis streams
func (s *AnalyticsService) SyncKinesisStreams(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.kinesisClient.ListStreams(ctx, &kinesis.ListStreamsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Kinesis streams: %w", err)
	}
	
	for _, streamName := range result.StreamNames {
		resource := Resource{
			ID:       streamName,
			Type:     "kinesis-stream",
			Provider: "aws",
			Region:   s.region,
			State:    "active",
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncGlueCrawlers syncs Glue crawlers
func (s *AnalyticsService) SyncGlueCrawlers(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.glueClient.GetCrawlers(ctx, &glue.GetCrawlersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Glue crawlers: %w", err)
	}
	
	for _, crawler := range result.Crawlers {
		resource := Resource{
			ID:       aws.ToString(crawler.Name),
			Type:     "glue-crawler",
			Provider: "aws",
			Region:   s.region,
			State:    string(crawler.State),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncAthenaWorkGroups syncs Athena work groups
func (s *AnalyticsService) SyncAthenaWorkGroups(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.athenaClient.ListWorkGroups(ctx, &athena.ListWorkGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Athena work groups: %w", err)
	}
	
	for _, workGroup := range result.WorkGroups {
		resource := Resource{
			ID:       aws.ToString(workGroup.Name),
			Type:     "athena-workgroup",
			Provider: "aws",
			Region:   s.region,
			State:    string(workGroup.State),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}