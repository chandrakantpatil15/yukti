package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type CloudWatchClient struct {
	client *cloudwatch.Client
}

type MetricData struct {
	MetricName string                 `json:"metric_name"`
	Namespace  string                 `json:"namespace"`
	Dimensions map[string]string      `json:"dimensions"`
	Values     []MetricDataPoint      `json:"values"`
}

type MetricDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
}

func NewCloudWatchClient(cfg aws.Config) *CloudWatchClient {
	return &CloudWatchClient{
		client: cloudwatch.NewFromConfig(cfg),
	}
}

// GetEC2Metrics fetches CPU, Network, and Disk metrics for EC2 instance
func (c *CloudWatchClient) GetEC2Metrics(ctx context.Context, instanceID string) ([]MetricData, error) {
	if c == nil || c.client == nil {
		return nil, nil
	}
	
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour) // Last 24 hours

	metrics := []struct {
		name      string
		namespace string
		statistic string
		unit      string
	}{
		{"CPUUtilization", "AWS/EC2", "Average", "Percent"},
		{"NetworkIn", "AWS/EC2", "Sum", "Bytes"},
		{"NetworkOut", "AWS/EC2", "Sum", "Bytes"},
		{"DiskReadBytes", "AWS/EBS", "Sum", "Bytes"},
		{"DiskWriteBytes", "AWS/EBS", "Sum", "Bytes"},
	}

	var results []MetricData
	for _, metric := range metrics {
		data, err := c.getMetricStatistics(ctx, metric.name, metric.namespace, instanceID, startTime, endTime, metric.statistic)
		if err != nil {
			continue // Skip failed metrics
		}
		
		var dataPoints []MetricDataPoint
		for _, dp := range data {
			if dp.Timestamp != nil && dp.Average != nil {
				dataPoints = append(dataPoints, MetricDataPoint{
					Timestamp: *dp.Timestamp,
					Value:     *dp.Average,
					Unit:      metric.unit,
				})
			}
		}

		results = append(results, MetricData{
			MetricName: metric.name,
			Namespace:  metric.namespace,
			Dimensions: map[string]string{"InstanceId": instanceID},
			Values:     dataPoints,
		})
	}

	return results, nil
}

// GetRDSMetrics fetches database performance metrics
func (c *CloudWatchClient) GetRDSMetrics(ctx context.Context, dbInstanceID string) ([]MetricData, error) {
	if c == nil || c.client == nil {
		return nil, nil
	}
	
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	metrics := []struct {
		name      string
		statistic string
		unit      string
	}{
		{"CPUUtilization", "Average", "Percent"},
		{"DatabaseConnections", "Average", "Count"},
		{"ReadLatency", "Average", "Seconds"},
		{"WriteLatency", "Average", "Seconds"},
		{"ReadIOPS", "Average", "Count/Second"},
		{"WriteIOPS", "Average", "Count/Second"},
	}

	var results []MetricData
	for _, metric := range metrics {
		data, err := c.getRDSMetricStatistics(ctx, metric.name, dbInstanceID, startTime, endTime, metric.statistic)
		if err != nil {
			continue
		}
		
		var dataPoints []MetricDataPoint
		for _, dp := range data {
			if dp.Timestamp != nil && dp.Average != nil {
				dataPoints = append(dataPoints, MetricDataPoint{
					Timestamp: *dp.Timestamp,
					Value:     *dp.Average,
					Unit:      metric.unit,
				})
			}
		}

		results = append(results, MetricData{
			MetricName: metric.name,
			Namespace:  "AWS/RDS",
			Dimensions: map[string]string{"DBInstanceIdentifier": dbInstanceID},
			Values:     dataPoints,
		})
	}

	return results, nil
}

// GetS3Metrics fetches bucket size and request metrics
func (c *CloudWatchClient) GetS3Metrics(ctx context.Context, bucketName string) ([]MetricData, error) {
	if c == nil || c.client == nil {
		return nil, nil
	}
	
	endTime := time.Now()
	startTime := endTime.Add(-7 * 24 * time.Hour) // S3 metrics are daily

	metrics := []struct {
		name      string
		statistic string
		unit      string
	}{
		{"BucketSizeBytes", "Average", "Bytes"},
		{"NumberOfObjects", "Average", "Count"},
	}

	var results []MetricData
	for _, metric := range metrics {
		data, err := c.getS3MetricStatistics(ctx, metric.name, bucketName, startTime, endTime, metric.statistic)
		if err != nil {
			continue
		}
		
		var dataPoints []MetricDataPoint
		for _, dp := range data {
			if dp.Timestamp != nil && dp.Average != nil {
				dataPoints = append(dataPoints, MetricDataPoint{
					Timestamp: *dp.Timestamp,
					Value:     *dp.Average,
					Unit:      metric.unit,
				})
			}
		}

		results = append(results, MetricData{
			MetricName: metric.name,
			Namespace:  "AWS/S3",
			Dimensions: map[string]string{"BucketName": bucketName},
			Values:     dataPoints,
		})
	}

	return results, nil
}

func (c *CloudWatchClient) getMetricStatistics(ctx context.Context, metricName, namespace, instanceID string, startTime, endTime time.Time, statistic string) ([]types.Datapoint, error) {
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String(metricName),
		Dimensions: []types.Dimension{
			{
				Name:  aws.String("InstanceId"),
				Value: aws.String(instanceID),
			},
		},
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(endTime),
		Period:     aws.Int32(3600), // 1 hour intervals
		Statistics: []types.Statistic{types.Statistic(statistic)},
	}

	result, err := c.client.GetMetricStatistics(ctx, input)
	if err != nil {
		return nil, err
	}

	return result.Datapoints, nil
}

func (c *CloudWatchClient) getRDSMetricStatistics(ctx context.Context, metricName, dbInstanceID string, startTime, endTime time.Time, statistic string) ([]types.Datapoint, error) {
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/RDS"),
		MetricName: aws.String(metricName),
		Dimensions: []types.Dimension{
			{
				Name:  aws.String("DBInstanceIdentifier"),
				Value: aws.String(dbInstanceID),
			},
		},
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(endTime),
		Period:     aws.Int32(3600),
		Statistics: []types.Statistic{types.Statistic(statistic)},
	}

	result, err := c.client.GetMetricStatistics(ctx, input)
	if err != nil {
		return nil, err
	}

	return result.Datapoints, nil
}

func (c *CloudWatchClient) getS3MetricStatistics(ctx context.Context, metricName, bucketName string, startTime, endTime time.Time, statistic string) ([]types.Datapoint, error) {
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/S3"),
		MetricName: aws.String(metricName),
		Dimensions: []types.Dimension{
			{
				Name:  aws.String("BucketName"),
				Value: aws.String(bucketName),
			},
			{
				Name:  aws.String("StorageType"),
				Value: aws.String("StandardStorage"),
			},
		},
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(endTime),
		Period:     aws.Int32(86400), // Daily for S3
		Statistics: []types.Statistic{types.Statistic(statistic)},
	}

	result, err := c.client.GetMetricStatistics(ctx, input)
	if err != nil {
		return nil, err
	}

	return result.Datapoints, nil
}