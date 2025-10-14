package com.cloudcostoptimizer.web.cost;

import com.cloudcostoptimizer.service.ResourceInventoryService;
import com.cloudcostoptimizer.service.AwsDataService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/cost")
public class CostOptimizationController {

    @Autowired
    private ResourceInventoryService inventoryService;
    
    @Autowired
    private AwsDataService awsDataService;

    @GetMapping("/resources/live")
    public Flux<Map<String, Object>> getLiveResources(@RequestParam(defaultValue = "100") int limit) {
        // Use dummy data for now - will switch to AWS when infrastructure is ready
        return inventoryService.generateInventory(limit);
    }

    @GetMapping("/resources")
    public Flux<Map<String, Object>> getResources() {
        return inventoryService.getResourcesFromDb(100);
    }

    @GetMapping("/recommendations")
    public Flux<Map<String, Object>> getRecommendations() {
        return inventoryService.generateRecommendationsFromDb();
    }

    @GetMapping("/summary")
    public Mono<Map<String, Object>> getCostSummary() {
        return inventoryService.calculateSummaryFromDb();
    }

    @GetMapping("/resources/bulk")
    public Mono<List<Map<String, Object>>> getBulkResources(@RequestParam(defaultValue = "10000") int count) {
        return inventoryService.generateInventory(count).collectList();
    }
}