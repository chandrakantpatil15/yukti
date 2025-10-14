package com.cloudcostoptimizer.core.plugin;

import reactor.core.publisher.Mono;

public interface CostOptimizerPlugin {
    String getServiceName();
    Mono<OptimizationResult> analyze();
    boolean isEnabled();
}