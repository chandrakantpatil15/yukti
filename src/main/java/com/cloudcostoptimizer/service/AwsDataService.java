package com.cloudcostoptimizer.service;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.cloudwatch.CloudWatchClient;
import software.amazon.awssdk.services.cloudwatch.model.*;
import software.amazon.awssdk.services.costexplorer.CostExplorerClient;
import software.amazon.awssdk.services.costexplorer.model.*;
import software.amazon.awssdk.services.ec2.Ec2Client;
import software.amazon.awssdk.services.ec2.model.*;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.*;
import java.util.stream.Collectors;

@Service
public class AwsDataService {

    private final Ec2Client ec2Client;
    private final CostExplorerClient costExplorerClient;
    private final CloudWatchClient cloudWatchClient;

    @Value("${aws.region:us-east-1}")
    private String awsRegion;

    public AwsDataService() {
        this.ec2Client = Ec2Client.builder().region(Region.US_EAST_1).build();
        this.costExplorerClient = CostExplorerClient.builder().region(Region.US_EAST_1).build();
        this.cloudWatchClient = CloudWatchClient.builder().region(Region.US_EAST_1).build();
    }

    public Flux<Map<String, Object>> fetchLiveEc2Instances() {
        return Mono.fromCallable(() -> {
            DescribeInstancesResponse response = ec2Client.describeInstances();
            return response.reservations().stream()
                .flatMap(reservation -> reservation.instances().stream())
                .map(this::mapInstanceToResource)
                .collect(Collectors.toList());
        })
        .flatMapMany(Flux::fromIterable)
        .onErrorResume(e -> {
            System.err.println("Error fetching EC2 instances: " + e.getMessage());
            return Flux.empty();
        });
    }

    public Mono<Map<String, Object>> fetchCostData(String instanceId) {
        return Mono.fromCallable(() -> {
            Instant endDate = Instant.now();
            Instant startDate = endDate.minus(30, ChronoUnit.DAYS);
            
            GetCostAndUsageRequest request = GetCostAndUsageRequest.builder()
                .timePeriod(DateInterval.builder()
                    .start(startDate.toString().substring(0, 10))
                    .end(endDate.toString().substring(0, 10))
                    .build())
                .granularity(Granularity.MONTHLY)
                .metrics("BlendedCost")
                .groupBy(GroupDefinition.builder()
                    .type(GroupDefinitionType.DIMENSION)
                    .key("SERVICE")
                    .build())
                .build();
                
            GetCostAndUsageResponse response = costExplorerClient.getCostAndUsage(request);
            return mapCostData(response);
        })
        .onErrorResume(e -> {
            System.err.println("Error fetching cost data: " + e.getMessage());
            return Mono.just(Map.of("monthlyCost", 0.0, "error", e.getMessage()));
        });
    }

    public Mono<Map<String, Object>> fetchCloudWatchMetrics(String instanceId) {
        return Mono.fromCallable(() -> {
            Instant endTime = Instant.now();
            Instant startTime = endTime.minus(1, ChronoUnit.HOURS);
            
            GetMetricStatisticsRequest request = GetMetricStatisticsRequest.builder()
                .namespace("AWS/EC2")
                .metricName("CPUUtilization")
                .dimensions(Dimension.builder()
                    .name("InstanceId")
                    .value(instanceId)
                    .build())
                .startTime(startTime)
                .endTime(endTime)
                .period(300)
                .statistics(Statistic.AVERAGE)
                .build();
                
            GetMetricStatisticsResponse response = cloudWatchClient.getMetricStatistics(request);
            return mapMetrics(response);
        })
        .onErrorResume(e -> {
            System.err.println("Error fetching CloudWatch metrics: " + e.getMessage());
            return Mono.just(Map.of("cpuUtilization", 0.0, "error", e.getMessage()));
        });
    }

    private Map<String, Object> mapInstanceToResource(Instance instance) {
        Map<String, String> tags = instance.tags().stream()
            .collect(Collectors.toMap(Tag::key, Tag::value));
            
        return Map.of(
            "resourceId", instance.instanceId(),
            "instanceType", instance.instanceType().toString(),
            "status", instance.state().name().toString(),
            "region", awsRegion,
            "environment", tags.getOrDefault("Environment", "unknown"),
            "launchTime", instance.launchTime().toString(),
            "tags", tags,
            "associatedResources", mapAssociatedResources(instance),
            "securityCompliance", mapSecurityCompliance(instance),
            "billingBreakdown", Map.of("total", 0.0) // Will be populated by cost data
        );
    }

    private List<Map<String, Object>> mapAssociatedResources(Instance instance) {
        List<Map<String, Object>> resources = new ArrayList<>();
        
        // EBS Volumes
        instance.blockDeviceMappings().forEach(mapping -> {
            if (mapping.ebs() != null) {
                resources.add(Map.of(
                    "resourceType", "EBS Volume",
                    "resourceId", mapping.ebs().volumeId(),
                    "status", "active",
                    "tags", Map.of("VolumeType", mapping.ebs().volumeType().toString())
                ));
            }
        });
        
        // Network Interfaces
        instance.networkInterfaces().forEach(ni -> {
            resources.add(Map.of(
                "resourceType", "Network Interface",
                "resourceId", ni.networkInterfaceId(),
                "status", "active",
                "tags", Map.of("Type", "primary")
            ));
        });
        
        return resources;
    }

    private Map<String, Object> mapSecurityCompliance(Instance instance) {
        int score = 75; // Base score
        List<Map<String, Object>> issues = new ArrayList<>();
        
        // Check security groups
        instance.securityGroups().forEach(sg -> {
            if (sg.groupName().contains("default")) {
                score -= 10;
                issues.add(Map.of(
                    "severity", "medium",
                    "category", "Network",
                    "description", "Using default security group",
                    "recommendation", "Use custom security groups"
                ));
            }
        });
        
        return Map.of(
            "score", Math.max(score, 0),
            "lastAudit", Instant.now().toString().substring(0, 10),
            "issues", issues
        );
    }

    private Map<String, Object> mapCostData(GetCostAndUsageResponse response) {
        double totalCost = response.resultsByTime().stream()
            .flatMap(result -> result.groups().stream())
            .mapToDouble(group -> Double.parseDouble(group.metrics().get("BlendedCost").amount()))
            .sum();
            
        return Map.of(
            "monthlyCost", Math.round(totalCost * 100.0) / 100.0,
            "currency", "USD",
            "period", "30-days"
        );
    }

    private Map<String, Object> mapMetrics(GetMetricStatisticsResponse response) {
        double avgCpu = response.datapoints().stream()
            .mapToDouble(Datapoint::average)
            .average()
            .orElse(0.0);
            
        return Map.of(
            "cpuUtilization", Math.round(avgCpu * 100.0) / 100.0,
            "dataPoints", response.datapoints().size(),
            "period", "1-hour"
        );
    }
}