package aws

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"golang.org/x/time/rate"
)

type EnterpriseAWSClient struct {
	cfg           aws.Config
	ec2Client     *ec2.Client
	ceClient      *costexplorer.Client
	rateLimiter   *rate.Limiter
	connPool      *ConnectionPool
	mu            sync.RWMutex
}

type ConnectionPool struct {
	maxConnections int
	activeConns    int
	mu             sync.Mutex
}

func NewEnterpriseAWSClient(ctx context.Context) (*EnterpriseAWSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Rate limiter: 100 requests per second with burst of 200
	limiter := rate.NewLimiter(rate.Limit(100), 200)

	return &EnterpriseAWSClient{
		cfg:         cfg,
		ec2Client:   ec2.NewFromConfig(cfg),
		ceClient:    costexplorer.NewFromConfig(cfg),
		rateLimiter: limiter,
		connPool: &ConnectionPool{
			maxConnections: 50,
			activeConns:    0,
		},
	}, nil
}

func (c *EnterpriseAWSClient) acquireConnection(ctx context.Context) error {
	c.connPool.mu.Lock()
	defer c.connPool.mu.Unlock()

	if c.connPool.activeConns >= c.connPool.maxConnections {
		return fmt.Errorf("connection pool exhausted")
	}

	c.connPool.activeConns++
	return nil
}

func (c *EnterpriseAWSClient) releaseConnection() {
	c.connPool.mu.Lock()
	defer c.connPool.mu.Unlock()

	if c.connPool.activeConns > 0 {
		c.connPool.activeConns--
	}
}

func (c *EnterpriseAWSClient) DescribeInstancesWithRateLimit(ctx context.Context, input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Acquire connection from pool
	if err := c.acquireConnection(ctx); err != nil {
		return nil, err
	}
	defer c.releaseConnection()

	// Execute with retry logic
	var output *ec2.DescribeInstancesOutput
	var err error

	for attempt := 0; attempt < 3; attempt++ {
		output, err = c.ec2Client.DescribeInstances(ctx, input)
		if err == nil {
			break
		}

		// Exponential backoff
		backoff := time.Duration(attempt+1) * time.Second
		time.Sleep(backoff)
	}

	return output, err
}

func (c *EnterpriseAWSClient) GetCostAndUsageWithRateLimit(ctx context.Context, input *costexplorer.GetCostAndUsageInput) (*costexplorer.GetCostAndUsageOutput, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Acquire connection from pool
	if err := c.acquireConnection(ctx); err != nil {
		return nil, err
	}
	defer c.releaseConnection()

	// Execute with retry logic
	var output *costexplorer.GetCostAndUsageOutput
	var err error

	for attempt := 0; attempt < 3; attempt++ {
		output, err = c.ceClient.GetCostAndUsage(ctx, input)
		if err == nil {
			break
		}

		// Exponential backoff
		backoff := time.Duration(attempt+1) * time.Second
		time.Sleep(backoff)
	}

	return output, err
}

func (c *EnterpriseAWSClient) BatchDescribeInstances(ctx context.Context, customerIDs []string) (map[string]*ec2.DescribeInstancesOutput, error) {
	results := make(map[string]*ec2.DescribeInstancesOutput)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Process customers in parallel with controlled concurrency
	semaphore := make(chan struct{}, 10) // Max 10 concurrent requests

	for _, customerID := range customerIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			input := &ec2.DescribeInstancesInput{}
			output, err := c.DescribeInstancesWithRateLimit(ctx, input)
			
			mu.Lock()
			if err == nil {
				results[id] = output
			}
			mu.Unlock()
		}(customerID)
	}

	wg.Wait()
	return results, nil
}

func (c *EnterpriseAWSClient) GetConnectionPoolStats() (int, int) {
	c.connPool.mu.Lock()
	defer c.connPool.mu.Unlock()
	
	return c.connPool.activeConns, c.connPool.maxConnections
}