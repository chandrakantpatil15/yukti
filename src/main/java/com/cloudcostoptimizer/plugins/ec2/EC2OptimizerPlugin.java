package com.cloudcostoptimizer.plugins.ec2;

import com.cloudcostoptimizer.core.plugin.CostOptimizerPlugin;
import com.cloudcostoptimizer.core.plugin.OptimizationResult;
import org.springframework.stereotype.Component;
import reactor.core.publisher.Mono;

import java.math.BigDecimal;
import java.util.List;

@Component
public class EC2OptimizerPlugin implements CostOptimizerPlugin {
    
    @Override
    public String getServiceName() {
        return "EC2";
    }
    
    @Override
    public Mono<OptimizationResult> analyze() {
        // TODO: Implement actual EC2 cost analysis
        return Mono.just(new OptimizationResult(
            "EC2",
            new BigDecimal("1000.00"),
            new BigDecimal("200.00"),
            List.of("Right-size underutilized instances", "Use Spot instances for non-critical workloads")
        ));
    }
    
    @Override
    public boolean isEnabled() {
        return true;
    }
}