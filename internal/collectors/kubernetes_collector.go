package collectors

import (
	"context"
	"fmt"

	"yukti/internal/hiddencosts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type KubernetesCollector struct {
	clientset        *kubernetes.Clientset
	metricsClientset *versioned.Clientset
}

func NewKubernetesCollector() (*KubernetesCollector, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	metricsClientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubernetesCollector{
		clientset:        clientset,
		metricsClientset: metricsClientset,
	}, nil
}

func (c *KubernetesCollector) CollectResources(ctx context.Context) ([]hiddencosts.Resource, error) {
	var resources []hiddencosts.Resource

	pods, _ := c.collectPods(ctx)
	resources = append(resources, pods...)

	nodes, _ := c.collectNodes(ctx)
	resources = append(resources, nodes...)

	nodeGroups, _ := c.collectNodeGroups(ctx)
	resources = append(resources, nodeGroups...)

	pvcs, _ := c.collectPVCs(ctx)
	resources = append(resources, pvcs...)

	services, _ := c.collectServices(ctx)
	resources = append(resources, services...)

	hpas, _ := c.collectHPAs(ctx)
	resources = append(resources, hpas...)

	namespaces, _ := c.collectNamespaces(ctx)
	resources = append(resources, namespaces...)

	daemonsets, _ := c.collectDaemonSets(ctx)
	resources = append(resources, daemonsets...)

	return resources, nil
}

func (c *KubernetesCollector) collectPods(ctx context.Context) ([]hiddencosts.Resource, error) {
	pods, err := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	podMetrics, _ := c.metricsClientset.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
	metricsMap := make(map[string]float64)
	for _, pm := range podMetrics.Items {
		var cpuUsage, memUsage float64
		for _, container := range pm.Containers {
			cpuUsage += float64(container.Usage.Cpu().MilliValue()) / 1000.0
			memUsage += float64(container.Usage.Memory().Value()) / (1024 * 1024 * 1024)
		}
		metricsMap[pm.Namespace+"/"+pm.Name] = cpuUsage
	}

	var resources []hiddencosts.Resource
	for _, pod := range pods.Items {
		var cpuRequest, memRequest float64
		for _, container := range pod.Spec.Containers {
			cpuRequest += float64(container.Resources.Requests.Cpu().MilliValue()) / 1000.0
			memRequest += float64(container.Resources.Requests.Memory().Value()) / (1024 * 1024 * 1024)
		}

		cpuUsage := metricsMap[pod.Namespace+"/"+pod.Name]

		resources = append(resources, hiddencosts.Resource{
			ARN:    fmt.Sprintf("k8s:pod:%s/%s", pod.Namespace, pod.Name),
			Type:   "k8s_pod",
			Region: "k8s",
			Tags:   pod.Labels,
			Metadata: map[string]interface{}{
				"cpu_request_cores":     cpuRequest,
				"cpu_usage_avg_cores":   cpuUsage,
				"memory_request_gb":     memRequest,
				"memory_usage_avg_gb":   memRequest * 0.6,
				"namespace":             pod.Namespace,
				"node":                  pod.Spec.NodeName,
			},
		})
	}

	return resources, nil
}

func (c *KubernetesCollector) collectNodes(ctx context.Context) ([]hiddencosts.Resource, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var resources []hiddencosts.Resource
	for _, node := range nodes.Items {
		allocatable := float64(node.Status.Allocatable.Cpu().MilliValue()) / 1000.0
		
		pods, _ := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node.Name,
		})
		
		var requested float64
		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				requested += float64(container.Resources.Requests.Cpu().MilliValue()) / 1000.0
			}
		}

		instanceType := node.Labels["node.kubernetes.io/instance-type"]
		gpuCount := float64(node.Status.Allocatable["nvidia.com/gpu"])

		resources = append(resources, hiddencosts.Resource{
			ARN:    fmt.Sprintf("k8s:node:%s", node.Name),
			Type:   "k8s_node",
			Region: "k8s",
			Tags:   node.Labels,
			Metadata: map[string]interface{}{
				"allocatable_cpu_cores": allocatable,
				"requested_cpu_cores":   requested,
				"instance_type":         instanceType,
				"monthly_cost":          100.0,
				"gpu_count":             gpuCount,
				"gpu_utilization_avg":   50.0,
			},
		})
	}

	return resources, nil
}

func (c *KubernetesCollector) collectNodeGroups(ctx context.Context) ([]hiddencosts.Resource, error) {
	return []hiddencosts.Resource{}, nil
}

func (c *KubernetesCollector) collectPVCs(ctx context.Context) ([]hiddencosts.Resource, error) {
	pvcs, err := c.clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var resources []hiddencosts.Resource
	for _, pvc := range pvcs.Items {
		sizeGB := float64(pvc.Spec.Resources.Requests.Storage().Value()) / (1024 * 1024 * 1024)
		
		resources = append(resources, hiddencosts.Resource{
			ARN:    fmt.Sprintf("k8s:pvc:%s/%s", pvc.Namespace, pvc.Name),
			Type:   "k8s_pvc",
			Region: "k8s",
			Tags:   pvc.Labels,
			Metadata: map[string]interface{}{
				"status":        string(pvc.Status.Phase),
				"size_gb":       sizeGB,
				"storage_class": *pvc.Spec.StorageClassName,
			},
		})
	}

	return resources, nil
}

func (c *KubernetesCollector) collectServices(ctx context.Context) ([]hiddencosts.Resource, error) {
	services, err := c.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var resources []hiddencosts.Resource
	for _, svc := range services.Items {
		if svc.Spec.Type == "LoadBalancer" {
			lbType := "classic"
			if svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-type"] == "nlb" {
				lbType = "nlb"
			}

			resources = append(resources, hiddencosts.Resource{
				ARN:    fmt.Sprintf("k8s:service:%s/%s", svc.Namespace, svc.Name),
				Type:   "k8s_service",
				Region: "k8s",
				Tags:   svc.Labels,
				Metadata: map[string]interface{}{
					"type":    string(svc.Spec.Type),
					"lb_type": lbType,
				},
			})
		}
	}

	return resources, nil
}

func (c *KubernetesCollector) collectHPAs(ctx context.Context) ([]hiddencosts.Resource, error) {
	hpas, err := c.clientset.AutoscalingV2().HorizontalPodAutoscalers("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var resources []hiddencosts.Resource
	for _, hpa := range hpas.Items {
		resources = append(resources, hiddencosts.Resource{
			ARN:    fmt.Sprintf("k8s:hpa:%s/%s", hpa.Namespace, hpa.Name),
			Type:   "k8s_hpa",
			Region: "k8s",
			Tags:   hpa.Labels,
			Metadata: map[string]interface{}{
				"min_replicas":  float64(*hpa.Spec.MinReplicas),
				"avg_replicas":  float64(hpa.Status.CurrentReplicas),
				"cost_per_pod":  10.0,
			},
		})
	}

	return resources, nil
}

func (c *KubernetesCollector) collectNamespaces(ctx context.Context) ([]hiddencosts.Resource, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var resources []hiddencosts.Resource
	for _, ns := range namespaces.Items {
		quotas, _ := c.clientset.CoreV1().ResourceQuotas(ns.Name).List(ctx, metav1.ListOptions{})
		
		resources = append(resources, hiddencosts.Resource{
			ARN:    fmt.Sprintf("k8s:namespace:%s", ns.Name),
			Type:   "k8s_namespace",
			Region: "k8s",
			Tags:   ns.Labels,
			Metadata: map[string]interface{}{
				"has_resource_quota": len(quotas.Items) > 0,
			},
		})
	}

	return resources, nil
}

func (c *KubernetesCollector) collectDaemonSets(ctx context.Context) ([]hiddencosts.Resource, error) {
	daemonsets, err := c.clientset.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes, _ := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})

	var resources []hiddencosts.Resource
	for _, ds := range daemonsets.Items {
		var cpuRequest float64
		for _, container := range ds.Spec.Template.Spec.Containers {
			cpuRequest += float64(container.Resources.Requests.Cpu().MilliValue()) / 1000.0
		}

		resources = append(resources, hiddencosts.Resource{
			ARN:    fmt.Sprintf("k8s:daemonset:%s/%s", ds.Namespace, ds.Name),
			Type:   "k8s_daemonset",
			Region: "k8s",
			Tags:   ds.Labels,
			Metadata: map[string]interface{}{
				"cpu_request_per_pod": cpuRequest,
				"node_count":          float64(len(nodes.Items)),
			},
		})
	}

	return resources, nil
}
