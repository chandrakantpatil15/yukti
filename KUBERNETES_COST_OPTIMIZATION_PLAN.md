# Kubernetes Cost Optimization - Implementation Plan

## Overview
Add comprehensive Kubernetes/EKS cost optimization to Yukti platform as a new detection category.

---

## Current Coverage (3 detectors)
✅ **EKS Control Plane Waste** - Detects underutilized clusters (<10 nodes)  
✅ **EKS EOL Versions** - Detects unsupported Kubernetes versions  
✅ **Fargate Over-Provisioning** - Detects over-provisioned ECS/Fargate tasks  

---

## Missing Coverage (12 detectors to add)

### Category: Kubernetes Cost Optimization

#### 1. Pod-Level Cost Allocation
**Detector**: `PodCostAllocationDetector`
- Track cost per pod, namespace, team, application
- Allocate node costs based on CPU/memory requests
- Identify most expensive pods/namespaces
- **Savings**: Visibility for chargeback/showback
- **Priority**: High (enterprise requirement)

#### 2. Pod CPU/Memory Right-Sizing
**Detector**: `PodResourceRightSizingDetector`
- Compare requests vs actual usage (from metrics-server)
- Detect over-provisioned pods (usage <50% of requests)
- Detect under-provisioned pods (throttling, OOMKilled)
- **Savings**: $1,000 - $5,000/month per cluster
- **Priority**: Critical (biggest waste in K8s)

#### 3. Node Over-Provisioning
**Detector**: `NodeOverProvisioningDetector`
- Calculate node utilization (allocated/capacity)
- Detect nodes with <40% utilization
- Recommend node consolidation or downsizing
- **Savings**: $500 - $3,000/month per cluster
- **Priority**: High

#### 4. Spot vs On-Demand Nodes
**Detector**: `NodeSpotOpportunityDetector`
- Identify fault-tolerant workloads (stateless, batch)
- Recommend Spot instances for 70% savings
- Check if Spot interruption handling configured
- **Savings**: $2,000 - $10,000/month per cluster
- **Priority**: Critical (huge savings)

#### 5. Cluster Autoscaler Efficiency
**Detector**: `ClusterAutoscalerWasteDetector`
- Detect frequent scale-up/scale-down cycles
- Identify nodes added but immediately removed
- Check for pending pods due to misconfiguration
- **Savings**: $300 - $1,500/month per cluster
- **Priority**: Medium

#### 6. Persistent Volume Waste
**Detector**: `PersistentVolumeWasteDetector`
- Detect unattached PVCs (no pod using them)
- Detect over-provisioned PVCs (usage <50%)
- Identify expensive storage classes (io2 vs gp3)
- **Savings**: $200 - $2,000/month per cluster
- **Priority**: Medium

#### 7. Unused LoadBalancer Services
**Detector**: `UnusedLoadBalancerDetector`
- Detect LoadBalancer services with no traffic
- Detect multiple LBs that could be consolidated (Ingress)
- Each ALB costs $16.20/month + data charges
- **Savings**: $200 - $1,000/month per cluster
- **Priority**: Medium

#### 8. Ingress Controller Waste
**Detector**: `IngressControllerWasteDetector`
- Detect multiple ingress controllers (consolidate)
- Recommend AWS Load Balancer Controller vs NGINX
- Check for unused ingress rules
- **Savings**: $100 - $500/month per cluster
- **Priority**: Low

#### 9. DaemonSet Over-Provisioning
**Detector**: `DaemonSetOverProvisioningDetector`
- Detect DaemonSets with excessive CPU/memory requests
- DaemonSets run on every node (multiplied waste)
- Common culprits: logging, monitoring agents
- **Savings**: $500 - $2,000/month per cluster
- **Priority**: High

#### 10. Namespace Resource Quotas Missing
**Detector**: `NamespaceQuotaMissingDetector`
- Detect namespaces without resource quotas
- Risk of resource exhaustion, cost overruns
- Recommend quotas based on historical usage
- **Savings**: Prevention (avoid runaway costs)
- **Priority**: Medium

#### 11. HPA (Horizontal Pod Autoscaler) Misconfiguration
**Detector**: `HPAMisconfigurationDetector`
- Detect HPA with min=max (not actually autoscaling)
- Detect HPA thrashing (frequent scale up/down)
- Detect HPA with wrong metrics (CPU vs custom)
- **Savings**: $200 - $1,000/month per cluster
- **Priority**: Medium

#### 12. Cluster Multi-Tenancy Waste
**Detector**: `ClusterMultiTenancyWasteDetector`
- Detect multiple small clusters (consolidate with namespaces)
- Each EKS cluster costs $73/month control plane
- Recommend single cluster with namespace isolation
- **Savings**: $73/month per eliminated cluster
- **Priority**: High

---

## Implementation Architecture

### Data Collection
```go
// Collect Kubernetes metrics via:
// 1. EKS API (cluster info, node groups)
// 2. Kubernetes API (pods, services, PVCs, namespaces)
// 3. Metrics Server (CPU/memory usage)
// 4. Prometheus (detailed metrics if available)

type KubernetesCollector struct {
    eksClient        *eks.Client
    k8sClient        *kubernetes.Clientset
    metricsClient    *metrics.Clientset
    prometheusClient *prometheus.Client
}

func (c *KubernetesCollector) CollectClusterData(clusterName string) (*ClusterData, error) {
    // Collect nodes, pods, services, PVCs, namespaces
    // Collect metrics (CPU, memory, network)
    // Calculate costs based on node types
    return clusterData, nil
}
```

### Cost Calculation
```go
// Calculate pod-level costs
func CalculatePodCost(pod *v1.Pod, nodeCost float64, nodeCapacity v1.ResourceList) float64 {
    // Cost = (pod CPU request / node CPU capacity) * node cost
    //      + (pod memory request / node memory capacity) * node cost
    
    cpuRequest := pod.Spec.Containers[0].Resources.Requests.Cpu().MilliValue()
    memRequest := pod.Spec.Containers[0].Resources.Requests.Memory().Value()
    
    nodeCPU := nodeCapacity.Cpu().MilliValue()
    nodeMem := nodeCapacity.Memory().Value()
    
    cpuCost := (float64(cpuRequest) / float64(nodeCPU)) * nodeCost
    memCost := (float64(memRequest) / float64(nodeMem)) * nodeCost
    
    return cpuCost + memCost
}
```

### Detector Example
```go
type PodResourceRightSizingDetector struct{}

func (d *PodResourceRightSizingDetector) Detect(clusterData *ClusterData) ([]Finding, error) {
    var findings []Finding
    
    for _, pod := range clusterData.Pods {
        // Get actual usage from metrics
        usage := clusterData.Metrics[pod.Name]
        
        // Compare to requests
        cpuUtil := usage.CPU / pod.Spec.Resources.Requests.CPU
        memUtil := usage.Memory / pod.Spec.Resources.Requests.Memory
        
        if cpuUtil < 0.5 || memUtil < 0.5 {
            // Over-provisioned
            currentCost := CalculatePodCost(pod, nodeCost, nodeCapacity)
            optimizedCost := currentCost * 0.6 // Right-size to 1.2x actual usage
            
            findings = append(findings, Finding{
                DetectorName:     "pod_resource_rightsizing",
                Category:         CategoryKubernetes,
                Severity:         SeverityHigh,
                Title:            "Pod over-provisioned (usage <50% of requests)",
                Description:      fmt.Sprintf("Pod %s using %d%% CPU, %d%% memory", pod.Name, cpuUtil*100, memUtil*100),
                ResourceARN:      pod.ARN,
                EstimatedCost:    currentCost,
                EstimatedSavings: currentCost - optimizedCost,
                Confidence:       0.90,
                Recommendation:   "Reduce CPU/memory requests to 1.2x actual usage",
            })
        }
    }
    
    return findings, nil
}
```

---

## Integration Points

### 1. AWS EKS API
- List clusters
- Describe node groups
- Get cluster version
- Get node group instance types

### 2. Kubernetes API
- List nodes, pods, services, PVCs, namespaces
- Get resource requests/limits
- Get pod status, events
- Get HPA configurations

### 3. Metrics Server
- Get node CPU/memory usage
- Get pod CPU/memory usage
- Historical metrics (if available)

### 4. Prometheus (Optional)
- Detailed metrics (network, disk I/O)
- Custom metrics for HPA
- Long-term historical data

### 5. AWS Cost Explorer
- Get actual EKS costs
- Get node costs by instance type
- Get EBS volume costs

---

## Database Schema

```sql
-- Kubernetes cluster metadata
CREATE TABLE yt_kubernetes_clusters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES yt_tenants(id),
    cluster_name VARCHAR(255) NOT NULL,
    cluster_arn VARCHAR(500) NOT NULL,
    region VARCHAR(50) NOT NULL,
    version VARCHAR(20) NOT NULL,
    node_count INT NOT NULL,
    pod_count INT NOT NULL,
    control_plane_cost DECIMAL(10,2),
    node_cost DECIMAL(10,2),
    total_cost DECIMAL(10,2),
    last_synced_at TIMESTAMP NOT NULL,
    UNIQUE(tenant_id, cluster_arn)
);

-- Pod-level cost allocation
CREATE TABLE yt_kubernetes_pod_costs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cluster_id UUID NOT NULL REFERENCES yt_kubernetes_clusters(id),
    namespace VARCHAR(255) NOT NULL,
    pod_name VARCHAR(255) NOT NULL,
    cpu_request_millicores INT,
    memory_request_mb INT,
    cpu_usage_millicores INT,
    memory_usage_mb INT,
    monthly_cost DECIMAL(10,2),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pod_costs_cluster ON yt_kubernetes_pod_costs(cluster_id);
CREATE INDEX idx_pod_costs_namespace ON yt_kubernetes_pod_costs(namespace);
```

---

## API Endpoints

### List Kubernetes Clusters
```http
GET /api/v1/kubernetes/clusters
Authorization: Bearer <jwt_token>

Response:
{
  "clusters": [
    {
      "id": "cluster-123",
      "name": "prod-eks-cluster",
      "region": "us-east-1",
      "version": "1.28",
      "node_count": 15,
      "pod_count": 120,
      "monthly_cost": 3250.00,
      "optimization_potential": 1200.00
    }
  ]
}
```

### Get Pod-Level Costs
```http
GET /api/v1/kubernetes/clusters/{id}/pods
Authorization: Bearer <jwt_token>

Response:
{
  "pods": [
    {
      "namespace": "production",
      "name": "api-server-abc123",
      "cpu_request": 500,
      "cpu_usage": 200,
      "memory_request": 1024,
      "memory_usage": 512,
      "monthly_cost": 45.00,
      "optimization_potential": 22.50,
      "recommendation": "Reduce CPU to 300m, memory to 700Mi"
    }
  ],
  "total_cost": 2850.00,
  "total_savings": 1150.00
}
```

### Get Kubernetes Findings
```http
GET /api/v1/kubernetes/findings
Authorization: Bearer <jwt_token>

Response:
{
  "findings": [
    {
      "detector": "pod_resource_rightsizing",
      "severity": "High",
      "title": "120 pods over-provisioned",
      "estimated_savings": 1150.00,
      "affected_resources": 120
    }
  ]
}
```

---

## UI Components

### Kubernetes Dashboard
```jsx
function KubernetesDashboard() {
  return (
    <div>
      <ClusterSummary clusters={clusters} />
      <CostByNamespace data={namespaceCosts} />
      <TopExpensivePods pods={expensivePods} />
      <OptimizationOpportunities findings={findings} />
    </div>
  );
}
```

### Pod Cost Breakdown
```jsx
function PodCostBreakdown({ pod }) {
  return (
    <div>
      <ResourceUsage 
        cpuRequest={pod.cpu_request}
        cpuUsage={pod.cpu_usage}
        memoryRequest={pod.memory_request}
        memoryUsage={pod.memory_usage}
      />
      <CostImpact 
        current={pod.monthly_cost}
        optimized={pod.monthly_cost - pod.optimization_potential}
      />
      <Recommendation text={pod.recommendation} />
    </div>
  );
}
```

---

## Pricing Strategy

### Feature Availability by Tier
- **FREE**: 0 Kubernetes detectors
- **PROFESSIONAL**: 5 detectors (basic pod/node optimization)
- **ENTERPRISE**: 12 detectors (full Kubernetes optimization)
- **FINANCIAL**: 12 detectors + custom policies

### Upsell Messaging
```
"Your EKS clusters are costing $15,000/month with $6,500 in waste.

Upgrade to ENTERPRISE to unlock:
✅ Pod-level cost allocation
✅ CPU/memory right-sizing (save $3,200/month)
✅ Spot instance recommendations (save $2,800/month)
✅ Namespace chargeback/showback

ROI: 13x (pay $499/month, save $6,500/month)"
```

---

## Implementation Timeline

### Phase 1: Data Collection (1 week)
- EKS API integration
- Kubernetes API integration
- Metrics Server integration
- Database schema

### Phase 2: Core Detectors (1 week)
- Pod right-sizing (most impactful)
- Node over-provisioning
- Spot opportunity
- PVC waste

### Phase 3: Advanced Detectors (1 week)
- Pod cost allocation
- Cluster autoscaler efficiency
- LoadBalancer waste
- DaemonSet optimization

### Phase 4: UI & API (1 week)
- Kubernetes dashboard
- Pod cost breakdown
- API endpoints
- Documentation

**Total: 4 weeks for complete Kubernetes optimization**

---

## Success Metrics

### Product Metrics
- **Detection Accuracy**: >85% (K8s metrics can be noisy)
- **Cost Allocation Accuracy**: >95% (critical for chargeback)
- **Savings Realization**: 50% of findings acted upon (K8s changes are complex)

### Business Metrics
- **Average Savings**: $3,000 - $8,000/month per EKS cluster
- **Upsell Rate**: 40% of customers with EKS upgrade to ENTERPRISE
- **Customer Satisfaction**: NPS >60 for K8s features

---

## Competitive Advantage

### Yukti vs Competitors (Kubernetes)
| Feature | Yukti | Kubecost | Cast.ai | CloudHealth |
|---------|-------|----------|---------|-------------|
| **Pod-level costs** | ✅ | ✅ | ✅ | ❌ |
| **Right-sizing** | ✅ | ✅ | ✅ | ⚠️ |
| **Spot recommendations** | ✅ | ✅ | ✅ | ❌ |
| **Multi-cloud** | ✅ | ✅ | ✅ | ✅ |
| **Integrated with AWS** | ✅ | ❌ | ❌ | ✅ |
| **Pricing** | $499/mo | $449/mo | $500/mo | $4K/mo |

**Advantage**: Integrated AWS + Kubernetes optimization in single platform

---

## Risk Mitigation

### Technical Risks
- **Metrics accuracy**: Validate against actual costs
- **API rate limits**: Cache data, batch requests
- **Cluster access**: Require read-only RBAC

### Business Risks
- **Complexity**: K8s optimization is complex, need good UX
- **Competition**: Kubecost is strong competitor, need differentiation
- **Adoption**: Customers may not have K8s expertise

### Mitigation Strategies
- Start with simple detectors (pod right-sizing)
- Provide clear, actionable recommendations
- Offer managed service for complex optimizations
- Integrate with existing K8s tools (Prometheus, Grafana)

---

**Status**: Planned for Phase 2.5 (after ML Enhancement)  
**Priority**: High (enterprise requirement)  
**Estimated Effort**: 4 weeks  
**Expected Impact**: $3K-$8K/month savings per cluster
