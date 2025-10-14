package com.cloudcostoptimizer;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class CloudCostOptimizerApplication {
    public static void main(String[] args) {
        System.setProperty("spring.threads.virtual.enabled", "true");
        SpringApplication.run(CloudCostOptimizerApplication.class, args);
    }
}