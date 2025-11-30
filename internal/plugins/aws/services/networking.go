package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

// NetworkingService handles all AWS networking services
type NetworkingService struct {
	ec2Client        *ec2.Client
	cloudfrontClient *cloudfront.Client
	route53Client    *route53.Client
	region           string
}

// NewNetworkingService creates a new networking service
func NewNetworkingService(cfg aws.Config, region string) *NetworkingService {
	regionalCfg := cfg.Copy()
	regionalCfg.Region = region
	
	return &NetworkingService{
		ec2Client:        ec2.NewFromConfig(regionalCfg),
		cloudfrontClient: cloudfront.NewFromConfig(regionalCfg),
		route53Client:    route53.NewFromConfig(regionalCfg),
		region:           region,
	}
}

// SyncVPCs syncs VPCs
func (s *NetworkingService) SyncVPCs(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %w", err)
	}
	
	for _, vpc := range result.Vpcs {
		resource := Resource{
			ID:       aws.ToString(vpc.VpcId),
			Type:     "vpc",
			Provider: "aws",
			Region:   s.region,
			State:    string(vpc.State),
			Tags:     extractVPCTags(vpc.Tags),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncLoadBalancers syncs Load Balancers (placeholder for ELB integration)
func (s *NetworkingService) SyncLoadBalancers(ctx context.Context) ([]Resource, error) {
	// TODO: Implement ELB/ALB/NLB sync when ELBv2 dependency is resolved
	return []Resource{}, nil
}

// SyncCloudFrontDistributions syncs CloudFront distributions
func (s *NetworkingService) SyncCloudFrontDistributions(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.cloudfrontClient.ListDistributions(ctx, &cloudfront.ListDistributionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CloudFront distributions: %w", err)
	}
	
	if result.DistributionList != nil {
		for _, dist := range result.DistributionList.Items {
			resource := Resource{
				ID:       aws.ToString(dist.Id),
				Type:     "cloudfront-distribution",
				Provider: "aws",
				Region:   "global",
				State:    aws.ToString(dist.Status),
				LastSeen: time.Now(),
			}
			resources = append(resources, resource)
		}
	}
	
	return resources, nil
}

// SyncRoute53HostedZones syncs Route53 hosted zones
func (s *NetworkingService) SyncRoute53HostedZones(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.route53Client.ListHostedZones(ctx, &route53.ListHostedZonesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Route53 hosted zones: %w", err)
	}
	
	for _, zone := range result.HostedZones {
		resource := Resource{
			ID:       aws.ToString(zone.Id),
			Type:     "route53-zone",
			Provider: "aws",
			Region:   "global",
			State:    "active",
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// SyncNATGateways syncs NAT Gateways
func (s *NetworkingService) SyncNATGateways(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	
	result, err := s.ec2Client.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe NAT gateways: %w", err)
	}
	
	for _, natGw := range result.NatGateways {
		resource := Resource{
			ID:       aws.ToString(natGw.NatGatewayId),
			Type:     "nat-gateway",
			Provider: "aws",
			Region:   s.region,
			State:    string(natGw.State),
			Tags:     extractNATTags(natGw.Tags),
			LastSeen: time.Now(),
		}
		resources = append(resources, resource)
	}
	
	return resources, nil
}

// extractVPCTags converts VPC tags to map
func extractVPCTags(tags []ec2types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}

// extractNATTags converts NAT Gateway tags to map
func extractNATTags(tags []ec2types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}