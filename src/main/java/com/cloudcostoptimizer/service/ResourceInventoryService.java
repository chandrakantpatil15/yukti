package com.cloudcostoptimizer.service;

import com.cloudcostoptimizer.model.ResourceEntity;
import com.cloudcostoptimizer.repository.ResourceRepository;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

import java.util.*;
import java.util.concurrent.ThreadLocalRandom;

@Service
public class ResourceInventoryService {

    @Autowired
    private ResourceRepository resourceRepository;
    
    private final ObjectMapper objectMapper = new ObjectMapper();

    private final String[] INSTANCE_TYPES = {
        "t3.nano", "t3.micro", "t3.small", "t3.medium", "t3.large", "t3.xlarge",
        "m5.large", "m5.xlarge", "m5.2xlarge", "c5.large", "c5.xlarge", "r5.large"
    };

    private final String[] ENVIRONMENTS = {"production", "staging", "development", "testing"};
    private final String[] REGIONS = {"us-east-1", "us-west-2", "eu-west-1"};
    private final String[] PROJECTS = {"web-app", "api-service", "data-processing", "ml-pipeline"};
    private final String[] OWNERS = {"devops-team", "backend-team", "data-team", "qa-team"};

    public Flux<Map<String, Object>> generateInventory(int count) {
        return resourceRepository.count()
            .flatMapMany(existingCount -> {
                if (existingCount >= count) {
                    return getResourcesFromDb(count);
                } else {
                    return generateAndSaveResources(count);
                }
            });
    }
    
    public Flux<Map<String, Object>> getResourcesFromDb(int limit) {
        return resourceRepository.findAll()
            .take(limit)
            .map(this::entityToMap);
    }
    
    private Flux<Map<String, Object>> generateAndSaveResources(int count) {
        return Flux.range(0, count)
            .map(this::generateResource)
            .flatMap(this::saveResource)
            .map(this::entityToMap);
    }

    private Map<String, Object> generateResource(int index) {
        Random rand = ThreadLocalRandom.current();
        String instanceType = INSTANCE_TYPES[rand.nextInt(INSTANCE_TYPES.length)];
        String environment = ENVIRONMENTS[rand.nextInt(ENVIRONMENTS.length)];
        boolean isRunning = rand.nextDouble() > 0.15;
        double baseCost = getInstanceCost(instanceType);
        double monthlyCost = isRunning ? baseCost : 0.0;
        int cpuUtilization = isRunning ? rand.nextInt(95) + 5 : 0;
        
        Map<String, Object> resource = new HashMap<>();
        resource.put("resourceId", String.format("i-%016x", rand.nextLong() & Long.MAX_VALUE));
        resource.put("instanceType", instanceType);
        resource.put("status", isRunning ? "running" : "stopped");
        resource.put("region", REGIONS[rand.nextInt(REGIONS.length)]);
        resource.put("monthlyCost", Math.round(monthlyCost * 100.0) / 100.0);
        resource.put("cpuUtilization", cpuUtilization);
        resource.put("environment", environment);
        
        Map<String, String> tags = new HashMap<>();
        tags.put("Environment", environment);
        tags.put("Project", PROJECTS[rand.nextInt(PROJECTS.length)]);
        tags.put("Owner", OWNERS[rand.nextInt(OWNERS.length)]);
        resource.put("tags", tags);
        
        Map<String, Object> ebsVolume = new HashMap<>();
        ebsVolume.put("resourceType", "EBS Volume");
        ebsVolume.put("resourceId", String.format("vol-%016x", rand.nextLong() & Long.MAX_VALUE));
        ebsVolume.put("monthlyCost", 5 + rand.nextDouble() * 15);
        ebsVolume.put("status", "active");
        ebsVolume.put("tags", Map.of("VolumeType", "gp3"));
        resource.put("associatedResources", List.of(ebsVolume));
        
        Map<String, Object> compliance = new HashMap<>();
        compliance.put("score", 60 + rand.nextInt(40));
        compliance.put("lastAudit", "2024-01-15");
        compliance.put("issues", List.of());
        resource.put("securityCompliance", compliance);
        
        Map<String, Double> billing = new HashMap<>();
        billing.put("compute", monthlyCost * 0.6);
        billing.put("storage", monthlyCost * 0.2);
        billing.put("network", monthlyCost * 0.15);
        billing.put("associatedServices", monthlyCost * 0.05);
        billing.put("total", monthlyCost);
        resource.put("billingBreakdown", billing);
        
        return resource;
    }
    
    private Mono<ResourceEntity> saveResource(Map<String, Object> resourceMap) {
        try {
            @SuppressWarnings("unchecked")
            Map<String, String> tags = (Map<String, String>) resourceMap.get("tags");
            
            ResourceEntity entity = new ResourceEntity(
                (String) resourceMap.get("resourceId"),
                (String) resourceMap.get("instanceType"),
                (String) resourceMap.get("status"),
                (String) resourceMap.get("region"),
                (Double) resourceMap.get("monthlyCost"),
                (Integer) resourceMap.get("cpuUtilization"),
                (String) resourceMap.get("environment"),
                tags.get("Project"),
                tags.get("Owner"),
                tags.get("CostCenter"),
                tags.get("Schedule"),
                tags.get("Application"),
                objectMapper.writeValueAsString(resourceMap.get("tags")),
                objectMapper.writeValueAsString(resourceMap.get("associatedResources")),
                objectMapper.writeValueAsString(resourceMap.get("securityCompliance")),
                objectMapper.writeValueAsString(resourceMap.get("billingBreakdown"))
            );
            return resourceRepository.save(entity);
        } catch (JsonProcessingException e) {
            return Mono.error(e);
        }
    }
    
    public Mono<Map<String, Object>> calculateSummaryFromDb() {
        return resourceRepository.findAll()
            .collectList()
            .map(resources -> {
                double totalCost = resources.stream().mapToDouble(ResourceEntity::getMonthlyCost).sum();
                double avgUtilization = resources.stream().mapToInt(ResourceEntity::getCpuUtilization).average().orElse(0);
                long optimizationOpportunities = resources.stream().filter(r -> r.getCpuUtilization() < 30 || "stopped".equals(r.getStatus())).count();
                double potentialSavings = resources.stream().filter(r -> r.getCpuUtilization() < 30).mapToDouble(r -> r.getMonthlyCost() * 0.4).sum();
                
                Map<String, Object> summary = new HashMap<>();
                summary.put("totalResources", resources.size());
                summary.put("totalMonthlyCost", Math.round(totalCost * 100.0) / 100.0);
                summary.put("potentialSavings", Math.round(potentialSavings * 100.0) / 100.0);
                summary.put("optimizationOpportunities", optimizationOpportunities);
                summary.put("averageUtilization", Math.round(avgUtilization * 100.0) / 100.0);
                return summary;
            });
    }
    
    public Flux<Map<String, Object>> generateRecommendationsFromDb() {
        return resourceRepository.findAll()
            .filter(resource -> resource.getCpuUtilization() < 30 && "running".equals(resource.getStatus()))
            .map(resource -> {
                Map<String, Object> rec = new HashMap<>();
                rec.put("resourceId", resource.getResourceId());
                rec.put("type", "downsize");
                rec.put("from", resource.getInstanceType());
                rec.put("to", getOptimizedInstanceType(resource.getInstanceType()));
                rec.put("currentCost", resource.getMonthlyCost());
                rec.put("optimizedCost", resource.getMonthlyCost() * 0.6);
                rec.put("savings", resource.getMonthlyCost() * 0.4);
                rec.put("confidence", 85 + (30 - resource.getCpuUtilization()));
                return rec;
            });
    }
    
    private String getOptimizedInstanceType(String currentType) {
        return switch (currentType) {
            case "m5.xlarge" -> "t3.large";
            case "t3.large" -> "t3.medium";
            case "t3.medium" -> "t3.small";
            case "c5.xlarge" -> "c5.large";
            default -> "t3.small";
        };
    }
    
    private Map<String, Object> entityToMap(ResourceEntity entity) {
        try {
            Map<String, Object> map = new HashMap<>();
            map.put("resourceId", entity.getResourceId());
            map.put("instanceType", entity.getInstanceType());
            map.put("status", entity.getStatus());
            map.put("region", entity.getRegion());
            map.put("monthlyCost", entity.getMonthlyCost());
            map.put("cpuUtilization", entity.getCpuUtilization());
            map.put("environment", entity.getEnvironment());
            map.put("tags", objectMapper.readValue(entity.getTags(), Map.class));
            map.put("associatedResources", objectMapper.readValue(entity.getAssociatedResources(), List.class));
            map.put("securityCompliance", objectMapper.readValue(entity.getSecurityCompliance(), Map.class));
            map.put("billingBreakdown", objectMapper.readValue(entity.getBillingBreakdown(), Map.class));
            return map;
        } catch (JsonProcessingException e) {
            throw new RuntimeException(e);
        }
    }

    private double getInstanceCost(String instanceType) {
        return switch (instanceType) {
            case "t3.nano" -> 3.80;
            case "t3.micro" -> 7.59;
            case "t3.small" -> 15.18;
            case "t3.medium" -> 30.37;
            case "t3.large" -> 60.74;
            case "t3.xlarge" -> 121.49;
            case "m5.large" -> 70.08;
            case "m5.xlarge" -> 140.16;
            case "m5.2xlarge" -> 280.32;
            case "c5.large" -> 62.05;
            case "c5.xlarge" -> 124.10;
            case "r5.large" -> 91.25;
            default -> 30.37;
        };
    }
}