package com.cloudcostoptimizer.core.plugin;

import java.math.BigDecimal;
import java.util.List;

public record OptimizationResult(
    String serviceName,
    BigDecimal currentCost,
    BigDecimal potentialSavings,
    List<String> recommendations
) {}