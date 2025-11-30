package engine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CloudProvider represents a cloud provider interface
type CloudProvider interface {
	GetName() string
	GetRegions() []string
	SyncResources(ctx context.Context) error
	GetCostData(ctx context.Context, startTime, endTime time.Time) (*CostData, error)
	OptimizeResources(ctx context.Context, resources []Resource) ([]Recommendation, error)
}

// MultiCloudEngine manages all cloud providers
type MultiCloudEngine struct {
	providers map[string]CloudProvider
	mu        sync.RWMutex
}

// CostData represents cost information across clouds
type CostData struct {
	Provider     string             `json:"provider"`
	TotalCost    float64           `json:"total_cost"`
	Currency     string            `json:"currency"`
	Period       string            `json:"period"`
	Services     map[string]float64 `json:"services"`
	Regions      map[string]float64 `json:"regions"`
	LastUpdated  time.Time         `json:"last_updated"`
}

// Resource represents a cloud resource
type Resource struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Provider     string            `json:"provider"`
	Region       string            `json:"region"`
	State        string            `json:"state"`
	Cost         float64           `json:"cost"`
	Tags         map[string]string `json:"tags"`
	Utilization  float64           `json:"utilization"`
	LastSeen     time.Time         `json:"last_seen"`
}

// Recommendation represents an optimization recommendation
type Recommendation struct {
	ID              string  `json:"id"`
	ResourceID      string  `json:"resource_id"`
	Type            string  `json:"type"`
	Description     string  `json:"description"`
	EstimatedSavings float64 `json:"estimated_savings"`
	Risk            string  `json:"risk"`
	Priority        int     `json:"priority"`
	Actions         []string `json:"actions"`
}

// NewMultiCloudEngine creates a new multi-cloud engine
func NewMultiCloudEngine() *MultiCloudEngine {
	return &MultiCloudEngine{
		providers: make(map[string]CloudProvider),
	}
}

// RegisterProvider registers a cloud provider
func (e *MultiCloudEngine) RegisterProvider(provider CloudProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.providers[provider.GetName()] = provider
}

// GetProviders returns all registered providers
func (e *MultiCloudEngine) GetProviders() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	providers := make([]string, 0, len(e.providers))
	for name := range e.providers {
		providers = append(providers, name)
	}
	return providers
}

// SyncAllResources syncs resources from all providers
func (e *MultiCloudEngine) SyncAllResources(ctx context.Context) error {
	e.mu.RLock()
	providers := make([]CloudProvider, 0, len(e.providers))
	for _, provider := range e.providers {
		providers = append(providers, provider)
	}
	e.mu.RUnlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(providers))

	for _, provider := range providers {
		wg.Add(1)
		go func(p CloudProvider) {
			defer wg.Done()
			if err := p.SyncResources(ctx); err != nil {
				errChan <- fmt.Errorf("failed to sync %s: %w", p.GetName(), err)
			}
		}(provider)
	}

	wg.Wait()
	close(errChan)

	// Collect any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("sync errors: %v", errors)
	}

	return nil
}

// GetCrossCloudCostAnalysis provides cost analysis across all clouds
func (e *MultiCloudEngine) GetCrossCloudCostAnalysis(ctx context.Context, startTime, endTime time.Time) (map[string]*CostData, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	results := make(map[string]*CostData)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, provider := range e.providers {
		wg.Add(1)
		go func(providerName string, p CloudProvider) {
			defer wg.Done()
			
			costData, err := p.GetCostData(ctx, startTime, endTime)
			if err != nil {
				// Log error but continue with other providers
				return
			}

			mu.Lock()
			results[providerName] = costData
			mu.Unlock()
		}(name, provider)
	}

	wg.Wait()
	return results, nil
}

// GenerateOptimizationRecommendations generates recommendations across all clouds
func (e *MultiCloudEngine) GenerateOptimizationRecommendations(ctx context.Context, resources []Resource) ([]Recommendation, error) {
	var allRecommendations []Recommendation
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Group resources by provider
	resourcesByProvider := make(map[string][]Resource)
	for _, resource := range resources {
		resourcesByProvider[resource.Provider] = append(resourcesByProvider[resource.Provider], resource)
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	for providerName, providerResources := range resourcesByProvider {
		if provider, exists := e.providers[providerName]; exists {
			wg.Add(1)
			go func(p CloudProvider, res []Resource) {
				defer wg.Done()
				
				recommendations, err := p.OptimizeResources(ctx, res)
				if err != nil {
					return
				}

				mu.Lock()
				allRecommendations = append(allRecommendations, recommendations...)
				mu.Unlock()
			}(provider, providerResources)
		}
	}

	wg.Wait()
	return allRecommendations, nil
}

// GetTotalCostAcrossProviders calculates total cost across all providers
func (e *MultiCloudEngine) GetTotalCostAcrossProviders(ctx context.Context, startTime, endTime time.Time) (float64, error) {
	costData, err := e.GetCrossCloudCostAnalysis(ctx, startTime, endTime)
	if err != nil {
		return 0, err
	}

	var totalCost float64
	for _, data := range costData {
		totalCost += data.TotalCost
	}

	return totalCost, nil
}