package com.cloudcostoptimizer.web;

import io.micrometer.core.instrument.MeterRegistry;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import reactor.core.publisher.Mono;

import java.lang.management.ManagementFactory;
import java.lang.management.MemoryMXBean;
import java.time.Instant;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentLinkedQueue;

@RestController
@RequestMapping("/api/metrics")
public class MetricsController {

    private final MeterRegistry meterRegistry;
    private final MemoryMXBean memoryBean = ManagementFactory.getMemoryMXBean();
    private final com.sun.management.OperatingSystemMXBean osBean = (com.sun.management.OperatingSystemMXBean) ManagementFactory.getOperatingSystemMXBean();
    private final ConcurrentLinkedQueue<Map<String, Object>> metricsHistory = new ConcurrentLinkedQueue<>();

    public MetricsController(MeterRegistry meterRegistry) {
        this.meterRegistry = meterRegistry;
    }

    @GetMapping("/current")
    public Mono<Map<String, Object>> getCurrentMetrics() {
        return Mono.fromCallable(() -> {
            var heapMemory = memoryBean.getHeapMemoryUsage();
            var nonHeapMemory = memoryBean.getNonHeapMemoryUsage();
            var timestamp = Instant.now().toEpochMilli();
            
            var metrics = Map.of(
                "timestamp", timestamp,
                "jvm", Map.of(
                    "heapUsed", heapMemory.getUsed() / 1024 / 1024, // MB
                    "heapMax", heapMemory.getMax() / 1024 / 1024,
                    "nonHeapUsed", nonHeapMemory.getUsed() / 1024 / 1024,
                    "threads", Thread.activeCount(),
                    "heapUtilization", (double) heapMemory.getUsed() / heapMemory.getMax() * 100
                ),
                "system", Map.of(
                    "cpuUsage", Math.max(0, osBean.getProcessCpuLoad() * 100),
                    "availableProcessors", osBean.getAvailableProcessors()
                ),
                "http", Map.of(
                    "totalRequests", (long) meterRegistry.counter("http.server.requests").count()
                )
            );
            
            // Store for history (keep last 1000 points)
            metricsHistory.offer(metrics);
            if (metricsHistory.size() > 1000) {
                metricsHistory.poll();
            }
            
            return metrics;
        });
    }

    @GetMapping("/history")
    public Mono<List<Map<String, Object>>> getMetricsHistory(
            @RequestParam(defaultValue = "300") long seconds) {
        return Mono.fromCallable(() -> {
            var cutoff = Instant.now().minusSeconds(seconds).toEpochMilli();
            return metricsHistory.stream()
                .filter(m -> (Long) m.get("timestamp") >= cutoff)
                .toList();
        });
    }
}