package com.cloudcostoptimizer.web;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import reactor.core.publisher.Mono;

import java.time.Instant;
import java.util.Map;

@RestController
@RequestMapping("/actuator")
public class InfoController {

    @GetMapping("/info")
    public Mono<Map<String, Object>> getInfo() {
        return Mono.fromCallable(() -> {
            var runtime = Runtime.getRuntime();
            var props = System.getProperties();
            
            return Map.of(
                "app", Map.of(
                    "name", "Yukti Cloud Cost Optimizer",
                    "version", "1.0.0",
                    "description", "Reactive Spring Boot application for AWS cost optimization"
                ),
                "java", Map.of(
                    "version", props.getProperty("java.version"),
                    "vendor", props.getProperty("java.vendor"),
                    "runtime", props.getProperty("java.runtime.name")
                ),
                "system", Map.of(
                    "os", props.getProperty("os.name"),
                    "arch", props.getProperty("os.arch"),
                    "processors", runtime.availableProcessors(),
                    "maxMemory", runtime.maxMemory() / 1024 / 1024 + " MB"
                ),
                "build", Map.of(
                    "time", Instant.now().toString(),
                    "group", "com.cloudcostoptimizer",
                    "artifact", "yukti"
                )
            );
        });
    }
}