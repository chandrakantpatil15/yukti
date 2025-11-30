package iac

import (
	"fmt"
	"strings"
	"time"
)

// MultiCloudGenerator generates IaC scripts for multiple cloud providers
type MultiCloudGenerator struct {
	terraformGen     *TerraformGenerator
	cloudFormationGen *CloudFormationGenerator
}

// NewMultiCloudGenerator creates a new multi-cloud IaC generator
func NewMultiCloudGenerator(region string) *MultiCloudGenerator {
	return &MultiCloudGenerator{
		terraformGen:     NewTerraformGenerator("aws", region),
		cloudFormationGen: NewCloudFormationGenerator(region),
	}
}

// GenerateAzureOptimization generates Azure ARM templates
func (mcg *MultiCloudGenerator) GenerateAzureOptimization(recommendation *OptimizationRecommendation) *IaCScript {
	var armTemplate strings.Builder
	
	armTemplate.WriteString(fmt.Sprintf(`{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "metadata": {
    "description": "Yukti FinOps - Azure VM Optimization",
    "generated": "%s",
    "action": "%s",
    "estimatedSavings": "%.2f USD/month"
  },
  "parameters": {
    "vmName": {
      "type": "string",
      "defaultValue": "%s",
      "metadata": {
        "description": "Name of the VM to optimize"
      }
    },
    "resourceGroupName": {
      "type": "string",
      "metadata": {
        "description": "Resource group containing the VM"
      }
    }
  },
  "variables": {
    "location": "[resourceGroup().location]"
  },
  "resources": [
`, time.Now().Format("2006-01-02 15:04:05"), recommendation.Action, recommendation.EstimatedSavings*30, recommendation.ResourceID))

	switch recommendation.Action {
	case "downsize":
		armTemplate.WriteString(mcg.generateAzureDownsizeTemplate(recommendation))
	case "schedule":
		armTemplate.WriteString(mcg.generateAzureScheduleTemplate(recommendation))
	case "deallocate":
		armTemplate.WriteString(mcg.generateAzureDeallocateTemplate(recommendation))
	}

	armTemplate.WriteString(`
  ],
  "outputs": {
    "optimizationResult": {
      "type": "string",
      "value": "Azure VM optimization completed"
    }
  }
}`)

	return &IaCScript{
		ID:               fmt.Sprintf("azure-%s-%d", recommendation.ResourceID, time.Now().Unix()),
		Type:             "arm-template",
		Provider:         "azure",
		Action:           recommendation.Action,
		ResourceID:       recommendation.ResourceID,
		Code:             armTemplate.String(),
		RollbackCode:     mcg.generateAzureRollbackTemplate(recommendation),
		EstimatedSavings: recommendation.EstimatedSavings,
		GeneratedAt:      time.Now(),
		Instructions:     mcg.generateAzureInstructions(recommendation),
	}
}

// GenerateGCPOptimization generates GCP Deployment Manager templates
func (mcg *MultiCloudGenerator) GenerateGCPOptimization(recommendation *OptimizationRecommendation) *IaCScript {
	var gcpTemplate strings.Builder
	
	gcpTemplate.WriteString(fmt.Sprintf(`# Yukti FinOps - GCP Compute Engine Optimization
# Generated: %s
# Action: %s
# Estimated Savings: $%.2f/month

resources:
`, time.Now().Format("2006-01-02 15:04:05"), recommendation.Action, recommendation.EstimatedSavings*30))

	switch recommendation.Action {
	case "downsize":
		gcpTemplate.WriteString(mcg.generateGCPDownsizeTemplate(recommendation))
	case "schedule":
		gcpTemplate.WriteString(mcg.generateGCPScheduleTemplate(recommendation))
	case "preemptible":
		gcpTemplate.WriteString(mcg.generateGCPPreemptibleTemplate(recommendation))
	}

	return &IaCScript{
		ID:               fmt.Sprintf("gcp-%s-%d", recommendation.ResourceID, time.Now().Unix()),
		Type:             "deployment-manager",
		Provider:         "gcp",
		Action:           recommendation.Action,
		ResourceID:       recommendation.ResourceID,
		Code:             gcpTemplate.String(),
		RollbackCode:     mcg.generateGCPRollbackTemplate(recommendation),
		EstimatedSavings: recommendation.EstimatedSavings,
		GeneratedAt:      time.Now(),
		Instructions:     mcg.generateGCPInstructions(recommendation),
	}
}

// generateAzureDownsizeTemplate generates Azure VM downsizing template
func (mcg *MultiCloudGenerator) generateAzureDownsizeTemplate(rec *OptimizationRecommendation) string {
	return fmt.Sprintf(`    {
      "type": "Microsoft.Compute/virtualMachines",
      "apiVersion": "2021-03-01",
      "name": "[parameters('vmName')]",
      "location": "[variables('location')]",
      "properties": {
        "hardwareProfile": {
          "vmSize": "%s"
        },
        "storageProfile": {
          "osDisk": {
            "createOption": "Attach"
          }
        },
        "networkProfile": {
          "networkInterfaces": [
            {
              "id": "[resourceId('Microsoft.Network/networkInterfaces', concat(parameters('vmName'), '-nic'))]"
            }
          ]
        }
      },
      "tags": {
        "OptimizedBy": "Yukti-FinOps",
        "OriginalSize": "Standard_D2s_v3",
        "EstimatedMonthlySavings": "%.2f"
      }
    }`, rec.RecommendedInstanceType, rec.EstimatedSavings*30)
}

// generateAzureScheduleTemplate generates Azure VM scheduling template
func (mcg *MultiCloudGenerator) generateAzureScheduleTemplate(rec *OptimizationRecommendation) string {
	return `    {
      "type": "Microsoft.Automation/automationAccounts",
      "apiVersion": "2020-01-13-preview",
      "name": "[concat('automation-', parameters('vmName'))]",
      "location": "[variables('location')]",
      "properties": {
        "sku": {
          "name": "Basic"
        }
      }
    },
    {
      "type": "Microsoft.Automation/automationAccounts/runbooks",
      "apiVersion": "2020-01-13-preview",
      "name": "[concat('automation-', parameters('vmName'), '/vm-scheduler')]",
      "dependsOn": [
        "[resourceId('Microsoft.Automation/automationAccounts', concat('automation-', parameters('vmName')))]"
      ],
      "properties": {
        "runbookType": "PowerShell",
        "logProgress": false,
        "logVerbose": false,
        "description": "VM scheduling runbook",
        "publishContentLink": {
          "uri": "https://raw.githubusercontent.com/yukti-finops/scripts/main/azure-vm-scheduler.ps1"
        }
      }
    }`
}

// generateAzureDeallocateTemplate generates Azure VM deallocation template
func (mcg *MultiCloudGenerator) generateAzureDeallocateTemplate(rec *OptimizationRecommendation) string {
	return `    {
      "type": "Microsoft.Compute/virtualMachines",
      "apiVersion": "2021-03-01",
      "name": "[parameters('vmName')]",
      "location": "[variables('location')]",
      "properties": {
        "extended": {
          "instanceView": {
            "powerState": {
              "code": "PowerState/deallocated"
            }
          }
        }
      },
      "tags": {
        "DeallocatedBy": "Yukti-FinOps",
        "DeallocatedDate": "[utcNow()]"
      }
    }`
}

// generateGCPDownsizeTemplate generates GCP VM downsizing template
func (mcg *MultiCloudGenerator) generateGCPDownsizeTemplate(rec *OptimizationRecommendation) string {
	return fmt.Sprintf(`- name: optimized-instance
  type: compute.v1.instance
  properties:
    zone: us-central1-a
    machineType: zones/us-central1-a/machineTypes/%s
    disks:
    - deviceName: boot
      type: PERSISTENT
      boot: true
      autoDelete: true
      initializeParams:
        sourceImage: projects/debian-cloud/global/images/family/debian-11
    networkInterfaces:
    - network: global/networks/default
      accessConfigs:
      - name: External NAT
        type: ONE_TO_ONE_NAT
    metadata:
      items:
      - key: optimized-by
        value: yukti-finops
      - key: estimated-monthly-savings
        value: "%.2f"
    labels:
      optimized: "true"
      original-machine-type: "n1-standard-2"`, rec.RecommendedInstanceType, rec.EstimatedSavings*30)
}

// generateGCPScheduleTemplate generates GCP VM scheduling template
func (mcg *MultiCloudGenerator) generateGCPScheduleTemplate(rec *OptimizationRecommendation) string {
	return `- name: vm-scheduler-function
  type: gcp-types/cloudfunctions-v1:projects.locations.functions
  properties:
    location: us-central1
    function: vm-scheduler
    sourceArchiveUrl: gs://yukti-finops-functions/vm-scheduler.zip
    entryPoint: scheduleVM
    runtime: python39
    trigger:
      httpsTrigger: {}
    environmentVariables:
      INSTANCE_NAME: $(ref.optimized-instance.name)
      PROJECT_ID: $(ref.optimized-instance.project)
      ZONE: $(ref.optimized-instance.zone)

- name: start-schedule
  type: gcp-types/cloudscheduler-v1:projects.locations.jobs
  properties:
    location: us-central1
    job: start-vm-job
    schedule: "0 8 * * *"  # 8 AM daily
    timeZone: "UTC"
    httpTarget:
      uri: $(ref.vm-scheduler-function.httpsTrigger.url)
      httpMethod: POST
      body: '{"action": "start"}'

- name: stop-schedule
  type: gcp-types/cloudscheduler-v1:projects.locations.jobs
  properties:
    location: us-central1
    job: stop-vm-job
    schedule: "0 18 * * *"  # 6 PM daily
    timeZone: "UTC"
    httpTarget:
      uri: $(ref.vm-scheduler-function.httpsTrigger.url)
      httpMethod: POST
      body: '{"action": "stop"}'`
}

// generateGCPPreemptibleTemplate generates GCP preemptible instance template
func (mcg *MultiCloudGenerator) generateGCPPreemptibleTemplate(rec *OptimizationRecommendation) string {
	return `- name: preemptible-instance
  type: compute.v1.instance
  properties:
    zone: us-central1-a
    machineType: zones/us-central1-a/machineTypes/n1-standard-1
    scheduling:
      preemptible: true
    disks:
    - deviceName: boot
      type: PERSISTENT
      boot: true
      autoDelete: true
      initializeParams:
        sourceImage: projects/debian-cloud/global/images/family/debian-11
    networkInterfaces:
    - network: global/networks/default
      accessConfigs:
      - name: External NAT
        type: ONE_TO_ONE_NAT
    metadata:
      items:
      - key: converted-by
        value: yukti-finops
      - key: preemptible
        value: "true"
    labels:
      preemptible: "true"
      cost-optimized: "true"`
}

// generateAzureRollbackTemplate generates Azure rollback template
func (mcg *MultiCloudGenerator) generateAzureRollbackTemplate(rec *OptimizationRecommendation) string {
	return fmt.Sprintf(`{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "metadata": {
    "description": "ROLLBACK TEMPLATE - Yukti FinOps Azure",
    "resourceId": "%s"
  },
  "parameters": {
    "vmName": {
      "type": "string",
      "defaultValue": "%s"
    }
  },
  "resources": [
    {
      "type": "Microsoft.Resources/deploymentScripts",
      "apiVersion": "2020-10-01",
      "name": "rollback-script",
      "location": "[resourceGroup().location]",
      "kind": "AzurePowerShell",
      "properties": {
        "azPowerShellVersion": "5.0",
        "scriptContent": "# Rollback logic here",
        "timeout": "PT30M",
        "cleanupPreference": "OnSuccess",
        "retentionInterval": "P1D"
      }
    }
  ]
}`, rec.ResourceID, rec.ResourceID)
}

// generateGCPRollbackTemplate generates GCP rollback template
func (mcg *MultiCloudGenerator) generateGCPRollbackTemplate(rec *OptimizationRecommendation) string {
	return fmt.Sprintf(`# ROLLBACK TEMPLATE - Yukti FinOps GCP
# Resource ID: %s

resources:
- name: rollback-function
  type: gcp-types/cloudfunctions-v1:projects.locations.functions
  properties:
    location: us-central1
    function: rollback-optimization
    sourceArchiveUrl: gs://yukti-finops-functions/rollback.zip
    entryPoint: rollbackOptimization
    runtime: python39
    trigger:
      httpsTrigger: {}
    environmentVariables:
      RESOURCE_ID: "%s"
      ROLLBACK_TYPE: "%s"`, rec.ResourceID, rec.ResourceID, rec.Action)
}

// generateAzureInstructions generates Azure deployment instructions
func (mcg *MultiCloudGenerator) generateAzureInstructions(rec *OptimizationRecommendation) []string {
	return []string{
		"1. Review the ARM template carefully",
		"2. Ensure you have Azure CLI installed and configured",
		"3. Deploy using: az deployment group create --resource-group <rg> --template-file template.json",
		"4. Or use Azure Portal for deployment",
		"5. Monitor the deployment progress",
		"6. Verify the optimization was applied correctly",
		"7. Keep the rollback template for emergency restoration",
		fmt.Sprintf("8. Expected savings: $%.2f per month", rec.EstimatedSavings*30),
	}
}

// generateGCPInstructions generates GCP deployment instructions
func (mcg *MultiCloudGenerator) generateGCPInstructions(rec *OptimizationRecommendation) []string {
	return []string{
		"1. Review the Deployment Manager template carefully",
		"2. Ensure you have gcloud CLI installed and configured",
		"3. Deploy using: gcloud deployment-manager deployments create <name> --config template.yaml",
		"4. Or use GCP Console for deployment",
		"5. Monitor the deployment progress",
		"6. Verify the optimization was applied correctly",
		"7. Keep the rollback template for emergency restoration",
		fmt.Sprintf("8. Expected savings: $%.2f per month", rec.EstimatedSavings*30),
	}
}