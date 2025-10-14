package com.cloudcostoptimizer.repository;

import com.cloudcostoptimizer.model.ResourceEntity;
import org.springframework.data.r2dbc.repository.R2dbcRepository;
import org.springframework.stereotype.Repository;
import reactor.core.publisher.Flux;

@Repository
public interface ResourceRepository extends R2dbcRepository<ResourceEntity, Long> {
    Flux<ResourceEntity> findByEnvironment(String environment);
    Flux<ResourceEntity> findByStatus(String status);
    Flux<ResourceEntity> findByProject(String project);
    Flux<ResourceEntity> findByOwner(String owner);
    Flux<ResourceEntity> findByCostCenter(String costCenter);
    Flux<ResourceEntity> findByEnvironmentAndProject(String environment, String project);
}