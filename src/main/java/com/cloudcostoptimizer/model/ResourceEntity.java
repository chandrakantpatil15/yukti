package com.cloudcostoptimizer.model;

import org.springframework.data.annotation.Id;
import org.springframework.data.relational.core.mapping.Table;

@Table("resources")
public class ResourceEntity {
    @Id
    private Long id;
    private String resourceId;
    private String instanceType;
    private String status;
    private String region;
    private Double monthlyCost;
    private Integer cpuUtilization;
    private String environment;
    private String project;
    private String owner;
    private String costCenter;
    private String backupSchedule;
    private String application;
    private String tags;
    private String associatedResources;
    private String securityCompliance;
    private String billingBreakdown;

    public ResourceEntity() {}

    public ResourceEntity(String resourceId, String instanceType, String status, String region,
                         Double monthlyCost, Integer cpuUtilization, String environment,
                         String project, String owner, String costCenter, String backupSchedule,
                         String application, String tags, String associatedResources, 
                         String securityCompliance, String billingBreakdown) {
        this.resourceId = resourceId;
        this.instanceType = instanceType;
        this.status = status;
        this.region = region;
        this.monthlyCost = monthlyCost;
        this.cpuUtilization = cpuUtilization;
        this.environment = environment;
        this.project = project;
        this.owner = owner;
        this.costCenter = costCenter;
        this.backupSchedule = backupSchedule;
        this.application = application;
        this.tags = tags;
        this.associatedResources = associatedResources;
        this.securityCompliance = securityCompliance;
        this.billingBreakdown = billingBreakdown;
    }

    public Long getId() { return id; }
    public void setId(Long id) { this.id = id; }
    public String getResourceId() { return resourceId; }
    public void setResourceId(String resourceId) { this.resourceId = resourceId; }
    public String getInstanceType() { return instanceType; }
    public void setInstanceType(String instanceType) { this.instanceType = instanceType; }
    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }
    public String getRegion() { return region; }
    public void setRegion(String region) { this.region = region; }
    public Double getMonthlyCost() { return monthlyCost; }
    public void setMonthlyCost(Double monthlyCost) { this.monthlyCost = monthlyCost; }
    public Integer getCpuUtilization() { return cpuUtilization; }
    public void setCpuUtilization(Integer cpuUtilization) { this.cpuUtilization = cpuUtilization; }
    public String getEnvironment() { return environment; }
    public void setEnvironment(String environment) { this.environment = environment; }
    public String getProject() { return project; }
    public void setProject(String project) { this.project = project; }
    public String getOwner() { return owner; }
    public void setOwner(String owner) { this.owner = owner; }
    public String getCostCenter() { return costCenter; }
    public void setCostCenter(String costCenter) { this.costCenter = costCenter; }
    public String getBackupSchedule() { return backupSchedule; }
    public void setBackupSchedule(String backupSchedule) { this.backupSchedule = backupSchedule; }
    public String getApplication() { return application; }
    public void setApplication(String application) { this.application = application; }
    public String getTags() { return tags; }
    public void setTags(String tags) { this.tags = tags; }
    public String getAssociatedResources() { return associatedResources; }
    public void setAssociatedResources(String associatedResources) { this.associatedResources = associatedResources; }
    public String getSecurityCompliance() { return securityCompliance; }
    public void setSecurityCompliance(String securityCompliance) { this.securityCompliance = securityCompliance; }
    public String getBillingBreakdown() { return billingBreakdown; }
    public void setBillingBreakdown(String billingBreakdown) { this.billingBreakdown = billingBreakdown; }
}