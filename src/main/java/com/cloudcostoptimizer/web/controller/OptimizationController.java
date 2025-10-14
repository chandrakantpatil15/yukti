package com.cloudcostoptimizer.web.controller;

import com.cloudcostoptimizer.core.plugin.OptimizationResult;
import com.cloudcostoptimizer.core.plugin.PluginManager;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RestController;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

@RestController
public class OptimizationController {
    
    private final PluginManager pluginManager;
    
    public OptimizationController(PluginManager pluginManager) {
        this.pluginManager = pluginManager;
    }
    
    @GetMapping("/optimize")
    public Flux<OptimizationResult> optimizeAll() {
        return pluginManager.analyzeAll();
    }
    
    @GetMapping("/optimize/{service}")
    public Mono<OptimizationResult> optimizeService(@PathVariable String service) {
        return pluginManager.analyzeService(service);
    }
}