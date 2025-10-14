package com.cloudcostoptimizer.core.plugin;

import org.springframework.stereotype.Component;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

import java.util.List;

@Component
public class PluginManager {
    
    private final List<CostOptimizerPlugin> plugins;
    
    public PluginManager(List<CostOptimizerPlugin> plugins) {
        this.plugins = plugins;
    }
    
    public Flux<OptimizationResult> analyzeAll() {
        return Flux.fromIterable(plugins)
            .filter(CostOptimizerPlugin::isEnabled)
            .flatMap(CostOptimizerPlugin::analyze);
    }
    
    public Mono<OptimizationResult> analyzeService(String serviceName) {
        return Flux.fromIterable(plugins)
            .filter(plugin -> plugin.getServiceName().equals(serviceName))
            .filter(CostOptimizerPlugin::isEnabled)
            .next()
            .flatMap(CostOptimizerPlugin::analyze);
    }
}