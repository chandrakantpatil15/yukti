package optimization

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DependencyAnalyzer analyzes cross-service dependencies for optimization
type DependencyAnalyzer struct {
	dependencies map[string][]string
	graph        *DependencyGraph
}

// NewDependencyAnalyzer creates a new dependency analyzer
func NewDependencyAnalyzer() *DependencyAnalyzer {
	return &DependencyAnalyzer{
		dependencies: make(map[string][]string),
		graph:        NewDependencyGraph(),
	}
}

// DependencyGraph represents resource dependency relationships
type DependencyGraph struct {
	nodes map[string]*ResourceNode
	edges map[string][]string
}

// ResourceNode represents a resource in the dependency graph
type ResourceNode struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Cost         float64                `json:"cost"`
	Utilization  float64                `json:"utilization"`
	Dependencies []string               `json:"dependencies"`
	Dependents   []string               `json:"dependents"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*ResourceNode),
		edges: make(map[string][]string),
	}
}

// AddResource adds a resource to the dependency graph
func (dg *DependencyGraph) AddResource(resource *ResourceNode) {
	dg.nodes[resource.ID] = resource
	if dg.edges[resource.ID] == nil {
		dg.edges[resource.ID] = make([]string, 0)
	}
}

// AddDependency adds a dependency relationship
func (dg *DependencyGraph) AddDependency(from, to string) {
	dg.edges[from] = append(dg.edges[from], to)
	
	// Update node dependencies
	if fromNode, exists := dg.nodes[from]; exists {
		fromNode.Dependencies = append(fromNode.Dependencies, to)
	}
	if toNode, exists := dg.nodes[to]; exists {
		toNode.Dependents = append(toNode.Dependents, from)
	}
}

// AnalyzeDependencies discovers and analyzes resource dependencies
func (da *DependencyAnalyzer) AnalyzeDependencies(ctx context.Context, resources []ResourceMetric) (*DependencyAnalysis, error) {
	// Build dependency graph
	for _, resource := range resources {
		node := &ResourceNode{
			ID:          resource.ResourceID,
			Type:        resource.InstanceType,
			Cost:        resource.Cost,
			Utilization: (resource.CPUUtil + resource.MemoryUtil + resource.NetworkIO) / 3.0,
			Metadata:    make(map[string]interface{}),
		}
		da.graph.AddResource(node)
	}
	
	// Discover dependencies based on naming patterns and network relationships
	da.discoverDependencies(resources)
	
	// Analyze dependency chains
	chains := da.findDependencyChains()
	
	// Calculate impact scores
	impactScores := da.calculateImpactScores()
	
	// Generate optimization recommendations
	recommendations := da.generateDependencyRecommendations(chains, impactScores)
	
	return &DependencyAnalysis{
		TotalResources:    len(resources),
		DependencyChains:  chains,
		ImpactScores:      impactScores,
		Recommendations:   recommendations,
		AnalyzedAt:        time.Now(),
	}, nil
}

// discoverDependencies discovers dependencies using heuristics
func (da *DependencyAnalyzer) discoverDependencies(resources []ResourceMetric) {
	// Group resources by naming patterns
	groups := make(map[string][]string)
	
	for _, resource := range resources {
		// Extract base name (remove instance numbers, etc.)
		baseName := da.extractBaseName(resource.ResourceID)
		groups[baseName] = append(groups[baseName], resource.ResourceID)
	}
	
	// Create dependencies within groups (assume related resources depend on each other)
	for _, group := range groups {
		if len(group) > 1 {
			// Create chain dependencies (each depends on the previous)
			for i := 1; i < len(group); i++ {
				da.graph.AddDependency(group[i], group[i-1])
			}
		}
	}
	
	// Add common dependency patterns
	da.addCommonDependencyPatterns(resources)
}

// extractBaseName extracts base name from resource ID
func (da *DependencyAnalyzer) extractBaseName(resourceID string) string {
	// Remove common suffixes and numbers
	name := resourceID
	name = strings.ReplaceAll(name, "-prod", "")
	name = strings.ReplaceAll(name, "-staging", "")
	name = strings.ReplaceAll(name, "-dev", "")
	
	// Remove trailing numbers
	for len(name) > 0 && name[len(name)-1] >= '0' && name[len(name)-1] <= '9' {
		name = name[:len(name)-1]
	}
	
	return strings.TrimSuffix(name, "-")
}

// addCommonDependencyPatterns adds common AWS dependency patterns
func (da *DependencyAnalyzer) addCommonDependencyPatterns(resources []ResourceMetric) {
	var webServers, databases, loadBalancers []string
	
	// Categorize resources by type/name patterns
	for _, resource := range resources {
		id := strings.ToLower(resource.ResourceID)
		
		if strings.Contains(id, "web") || strings.Contains(id, "app") {
			webServers = append(webServers, resource.ResourceID)
		} else if strings.Contains(id, "db") || strings.Contains(id, "database") {
			databases = append(databases, resource.ResourceID)
		} else if strings.Contains(id, "lb") || strings.Contains(id, "balancer") {
			loadBalancers = append(loadBalancers, resource.ResourceID)
		}
	}
	
	// Create typical web application dependencies
	// Load Balancer -> Web Servers -> Databases
	for _, lb := range loadBalancers {
		for _, web := range webServers {
			da.graph.AddDependency(lb, web)
		}
	}
	
	for _, web := range webServers {
		for _, db := range databases {
			da.graph.AddDependency(web, db)
		}
	}
}

// findDependencyChains finds dependency chains in the graph
func (da *DependencyAnalyzer) findDependencyChains() []DependencyChain {
	var chains []DependencyChain
	visited := make(map[string]bool)
	
	// Find chains starting from nodes with no dependents (root nodes)
	for nodeID, node := range da.graph.nodes {
		if len(node.Dependents) == 0 && !visited[nodeID] {
			chain := da.buildChain(nodeID, visited)
			if len(chain.Resources) > 1 {
				chains = append(chains, chain)
			}
		}
	}
	
	return chains
}

// buildChain builds a dependency chain starting from a node
func (da *DependencyAnalyzer) buildChain(startID string, visited map[string]bool) DependencyChain {
	var resources []string
	var totalCost float64
	
	current := startID
	for current != "" && !visited[current] {
		visited[current] = true
		resources = append(resources, current)
		
		if node, exists := da.graph.nodes[current]; exists {
			totalCost += node.Cost
		}
		
		// Move to next dependency (first one if multiple)
		if deps := da.graph.edges[current]; len(deps) > 0 {
			current = deps[0]
		} else {
			current = ""
		}
	}
	
	return DependencyChain{
		ID:        fmt.Sprintf("chain-%s", startID),
		Resources: resources,
		TotalCost: totalCost,
		Length:    len(resources),
	}
}

// calculateImpactScores calculates impact scores for each resource
func (da *DependencyAnalyzer) calculateImpactScores() map[string]float64 {
	scores := make(map[string]float64)
	
	for nodeID, node := range da.graph.nodes {
		// Base score from cost and utilization
		baseScore := node.Cost * (1.0 + node.Utilization/100.0)
		
		// Multiply by number of dependents (resources that depend on this one)
		dependentMultiplier := 1.0 + float64(len(node.Dependents))*0.5
		
		// Add dependency depth bonus
		depthBonus := float64(len(node.Dependencies)) * 0.2
		
		scores[nodeID] = baseScore * dependentMultiplier + depthBonus
	}
	
	return scores
}

// generateDependencyRecommendations generates optimization recommendations
func (da *DependencyAnalyzer) generateDependencyRecommendations(chains []DependencyChain, impactScores map[string]float64) []DependencyRecommendation {
	var recommendations []DependencyRecommendation
	
	// Analyze each chain for optimization opportunities
	for _, chain := range chains {
		if chain.Length > 2 && chain.TotalCost > 100 { // Focus on significant chains
			
			// Find bottleneck (highest impact score in chain)
			var bottleneck string
			var maxScore float64
			
			for _, resourceID := range chain.Resources {
				if score := impactScores[resourceID]; score > maxScore {
					maxScore = score
					bottleneck = resourceID
				}
			}
			
			// Generate recommendation based on bottleneck analysis
			if bottleneck != "" {
				rec := DependencyRecommendation{
					ChainID:     chain.ID,
					Type:        "bottleneck_optimization",
					Priority:    da.calculatePriority(maxScore),
					Description: fmt.Sprintf("Optimize bottleneck resource %s in dependency chain", bottleneck),
					Resources:   []string{bottleneck},
					EstimatedSavings: chain.TotalCost * 0.2, // 20% potential savings
					ImpactScore: maxScore,
					GeneratedAt: time.Now(),
				}
				recommendations = append(recommendations, rec)
			}
		}
		
		// Check for consolidation opportunities
		if chain.Length > 3 {
			rec := DependencyRecommendation{
				ChainID:     chain.ID,
				Type:        "consolidation",
				Priority:    "medium",
				Description: fmt.Sprintf("Consider consolidating %d resources in dependency chain", chain.Length),
				Resources:   chain.Resources,
				EstimatedSavings: chain.TotalCost * 0.15, // 15% consolidation savings
				ImpactScore: chain.TotalCost,
				GeneratedAt: time.Now(),
			}
			recommendations = append(recommendations, rec)
		}
	}
	
	return recommendations
}

// calculatePriority calculates recommendation priority
func (da *DependencyAnalyzer) calculatePriority(impactScore float64) string {
	if impactScore > 1000 {
		return "critical"
	} else if impactScore > 500 {
		return "high"
	} else if impactScore > 100 {
		return "medium"
	}
	return "low"
}

// DependencyAnalysis represents the result of dependency analysis
type DependencyAnalysis struct {
	TotalResources   int                        `json:"total_resources"`
	DependencyChains []DependencyChain          `json:"dependency_chains"`
	ImpactScores     map[string]float64         `json:"impact_scores"`
	Recommendations  []DependencyRecommendation `json:"recommendations"`
	AnalyzedAt       time.Time                  `json:"analyzed_at"`
}

// DependencyChain represents a chain of dependent resources
type DependencyChain struct {
	ID        string   `json:"id"`
	Resources []string `json:"resources"`
	TotalCost float64  `json:"total_cost"`
	Length    int      `json:"length"`
}

// DependencyRecommendation represents a dependency-based optimization recommendation
type DependencyRecommendation struct {
	ChainID          string    `json:"chain_id"`
	Type             string    `json:"type"`
	Priority         string    `json:"priority"`
	Description      string    `json:"description"`
	Resources        []string  `json:"resources"`
	EstimatedSavings float64   `json:"estimated_savings"`
	ImpactScore      float64   `json:"impact_score"`
	GeneratedAt      time.Time `json:"generated_at"`
}