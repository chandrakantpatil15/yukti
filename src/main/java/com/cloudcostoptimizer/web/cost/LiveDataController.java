package com.cloudcostoptimizer.web.cost;

import com.cloudcostoptimizer.service.AwsDataService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

import java.time.Duration;
import java.util.Map;

@RestController
@RequestMapping("/api/live")
public class LiveDataController {

    @Autowired
    private AwsDataService awsDataService;

    @GetMapping("/instances")
    public Flux<Map<String, Object>> getLiveInstances() {
        return awsDataService.fetchLiveEc2Instances();
    }

    @GetMapping("/instances/{instanceId}/cost")
    public Mono<Map<String, Object>> getInstanceCost(@PathVariable String instanceId) {
        return awsDataService.fetchCostData(instanceId);
    }

    @GetMapping("/instances/{instanceId}/metrics")
    public Mono<Map<String, Object>> getInstanceMetrics(@PathVariable String instanceId) {
        return awsDataService.fetchCloudWatchMetrics(instanceId);
    }

    @GetMapping("/instances/{instanceId}/realtime")
    public Flux<Map<String, Object>> getRealtimeMetrics(@PathVariable String instanceId) {
        return Flux.interval(Duration.ofSeconds(30))
            .flatMap(tick -> awsDataService.fetchCloudWatchMetrics(instanceId))
            .take(Duration.ofMinutes(10));
    }

    @GetMapping("/cost/summary")
    public Mono<Map<String, Object>> getLiveCostSummary() {
        return awsDataService.fetchLiveEc2Instances()
            .flatMap(instance -> awsDataService.fetchCostData((String) instance.get("resourceId")))
            .reduce(Map.of("totalCost", 0.0, "instanceCount", 0), (acc, cost) -> {
                double currentTotal = (Double) acc.get("totalCost");
                int currentCount = (Integer) acc.get("instanceCount");
                double instanceCost = (Double) cost.getOrDefault("monthlyCost", 0.0);
                
                return Map.of(
                    "totalCost", currentTotal + instanceCost,
                    "instanceCount", currentCount + 1
                );
            });
    }
}