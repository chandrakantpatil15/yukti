package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type RemediationAPI struct {
	// Add dependencies here
}

type RemediationRequest struct {
	CustomerID     string            `json:"customer_id"`
	ResourceID     string            `json:"resource_id"`
	ActionType     string            `json:"action_type"`
	Parameters     map[string]string `json:"parameters"`
	TemplateFormat string            `json:"template_format"` // "terraform", "terragrunt", or "cloudformation"
	Environment    string            `json:"environment"`     // "dev", "staging", "prod"
}

type RemediationResponse struct {
	RequestID       string            `json:"request_id"`
	CustomerID      string            `json:"customer_id"`
	ResourceID      string            `json:"resource_id"`
	ActionType      string            `json:"action_type"`
	TemplateFormat  string            `json:"template_format"`
	Environment     string            `json:"environment"`
	Template        string            `json:"template,omitempty"`
	TerragruntHCL   string            `json:"terragrunt_hcl,omitempty"`
	MainTF          string            `json:"main_tf,omitempty"`
	Files           map[string]string `json:"files,omitempty"`
	EstimatedSavings float64           `json:"estimated_savings"`
	Risk            string            `json:"risk"`
	Instructions    string            `json:"instructions"`
	CreatedAt       time.Time         `json:"created_at"`
}

func NewRemediationAPI() *RemediationAPI {
	return &RemediationAPI{}
}

func (api *RemediationAPI) GenerateRemediation(w http.ResponseWriter, r *http.Request) {
	var req RemediationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.CustomerID == "" || req.ResourceID == "" || req.ActionType == "" {
		http.Error(w, "customer_id, resource_id, and action_type are required", http.StatusBadRequest)
		return
	}

	// Set default template format
	if req.TemplateFormat == "" {
		req.TemplateFormat = "terragrunt" // Default to Terragrunt for enterprise customers
	}
	
	// Set default environment
	if req.Environment == "" {
		req.Environment = "dev"
	}

	// Generate remediation template
	response, err := api.generateRemediationTemplate(req)
	if err != nil {
		log.Printf("Error generating remediation: %v", err)
		http.Error(w, "Failed to generate remediation template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *RemediationAPI) generateRemediationTemplate(req RemediationRequest) (*RemediationResponse, error) {
	var template string
	var estimatedSavings float64
	var risk string
	var instructions string

	var template, terragruntHCL, mainTF string
	var files map[string]string
	
	switch req.ActionType {
	case "rightsizing":
		if req.TemplateFormat == "terragrunt" {
			terragruntHCL, mainTF = api.generateTerragruntRightSizing(req)
			files = api.generateTerragruntStructure(req)
		} else {
			template = api.generateRightSizingTemplate(req)
		}
		estimatedSavings = 120.50 // Mock calculation
		risk = "low"
		instructions = api.getTerragruntInstructions(req.TemplateFormat, req.Environment)

	case "termination":
		template = api.generateTerminationTemplate(req)
		estimatedSavings = 85.75 // Mock calculation
		risk = "medium"
		instructions = "1. BACKUP ALL DATA FIRST\n2. Verify no dependencies\n3. Consider stopping instead of terminating for testing"

	case "reserved_instances":
		template = api.generateReservedInstanceTemplate(req)
		estimatedSavings = 200.00 // Mock calculation
		risk = "low"
		instructions = "1. Review usage patterns\n2. Purchase through AWS Console\n3. Monitor utilization after purchase"

	default:
		return nil, fmt.Errorf("unsupported action type: %s", req.ActionType)
	}

	return &RemediationResponse{
		RequestID:       fmt.Sprintf("rem_%d", time.Now().Unix()),
		CustomerID:      req.CustomerID,
		ResourceID:      req.ResourceID,
		ActionType:      req.ActionType,
		TemplateFormat:  req.TemplateFormat,
		Environment:     req.Environment,
		Template:        template,
		TerragruntHCL:   terragruntHCL,
		MainTF:          mainTF,
		Files:           files,
		EstimatedSavings: estimatedSavings,
		Risk:            risk,
		Instructions:    instructions,
		CreatedAt:       time.Now(),
	}, nil
}

func (api *RemediationAPI) generateRightSizingTemplate(req RemediationRequest) string {
	if req.TemplateFormat == "cloudformation" {
		return api.generateCFRightSizing(req)
	}

	// Terraform template
	currentType := req.Parameters["current_type"]
	recommendedType := req.Parameters["recommended_type"]
	if currentType == "" {
		currentType = "m5.large"
	}
	if recommendedType == "" {
		recommendedType = "m5.medium"
	}

	template := `# Yukti FinOps - EC2 Right-sizing
# Customer: %s
# Resource: %s
# Action: %s -> %s

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Right-size EC2 instance
resource "aws_instance" "optimized" {
  instance_type = "%s"
  
  # Copy from existing instance
  ami                    = data.aws_instance.current.ami
  key_name              = data.aws_instance.current.key_name
  vpc_security_group_ids = data.aws_instance.current.vpc_security_group_ids
  subnet_id             = data.aws_instance.current.subnet_id
  
  tags = merge(
    data.aws_instance.current.tags,
    {
      "OptimizedBy" = "Yukti-FinOps"
      "OriginalType" = "%s"
      "OptimizedDate" = "%s"
    }
  )
  
  lifecycle {
    create_before_destroy = true
  }
}

data "aws_instance" "current" {
  instance_id = "%s"
}

output "new_instance_id" {
  value = aws_instance.optimized.id
}

output "estimated_monthly_savings" {
  value = "$120.50"
}
`

	return fmt.Sprintf(template,
		req.CustomerID,
		req.ResourceID,
		currentType,
		recommendedType,
		recommendedType,
		currentType,
		time.Now().Format("2006-01-02"),
		req.ResourceID,
	)
}

func (api *RemediationAPI) generateTerminationTemplate(req RemediationRequest) string {
	if req.TemplateFormat == "cloudformation" {
		return api.generateCFTermination(req)
	}

	template := `# Yukti FinOps - EC2 Termination
# Customer: %s
# Resource: %s
# WARNING: This will permanently delete the instance

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Terminate unused instance
resource "null_resource" "terminate" {
  provisioner "local-exec" {
    command = "aws ec2 terminate-instances --instance-ids %s"
  }
  
  provisioner "local-exec" {
    command = "echo 'Instance %s terminated - Estimated savings: $85.75/month'"
  }
}

output "action_completed" {
  value = "Instance termination initiated"
}

output "estimated_monthly_savings" {
  value = "$85.75"
}

# CRITICAL: Review before applying!
# 1. Backup all data
# 2. Verify no dependencies
# 3. Consider stopping first for testing
`

	return fmt.Sprintf(template,
		req.CustomerID,
		req.ResourceID,
		req.ResourceID,
		req.ResourceID,
	)
}

func (api *RemediationAPI) generateReservedInstanceTemplate(req RemediationRequest) string {
	instanceType := req.Parameters["instance_type"]
	quantity := req.Parameters["quantity"]
	if instanceType == "" {
		instanceType = "m5.large"
	}
	if quantity == "" {
		quantity = "3"
	}

	template := `# Yukti FinOps - Reserved Instance Recommendation
# Customer: %s
# Instance Type: %s
# Quantity: %s

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Reserved Instance recommendation (informational)
resource "null_resource" "ri_recommendation" {
  provisioner "local-exec" {
    command = <<-EOT
      echo "Reserved Instance Recommendation:"
      echo "Instance Type: %s"
      echo "Quantity: %s"
      echo "Term: 1 year, No Upfront"
      echo "Estimated Monthly Savings: $200.00"
      echo ""
      echo "Purchase at: https://console.aws.amazon.com/ec2/v2/home#ReservedInstances"
    EOT
  }
}

output "recommendation" {
  value = {
    instance_type = "%s"
    quantity = %s
    estimated_monthly_savings = "$200.00"
    purchase_url = "https://console.aws.amazon.com/ec2/v2/home#ReservedInstances"
  }
}
`

	return fmt.Sprintf(template,
		req.CustomerID,
		instanceType,
		quantity,
		instanceType,
		quantity,
		instanceType,
		quantity,
	)
}

func (api *RemediationAPI) generateCFRightSizing(req RemediationRequest) string {
	return `AWSTemplateFormatVersion: '2010-09-09'
Description: 'Yukti FinOps - EC2 Right-sizing'

Parameters:
  InstanceId:
    Type: String
    Default: '` + req.ResourceID + `'
  NewInstanceType:
    Type: String
    Default: 'm5.medium'

Resources:
  OptimizedInstance:
    Type: AWS::EC2::Instance
    Properties:
      InstanceType: !Ref NewInstanceType
      Tags:
        - Key: OptimizedBy
          Value: Yukti-FinOps

Outputs:
  EstimatedSavings:
    Value: '$120.50/month'`
}

func (api *RemediationAPI) generateCFTermination(req RemediationRequest) string {
	return `AWSTemplateFormatVersion: '2010-09-09'
Description: 'Yukti FinOps - EC2 Termination'

Parameters:
  InstanceId:
    Type: String
    Default: '` + req.ResourceID + `'

Resources:
  TerminationFunction:
    Type: AWS::Lambda::Function
    Properties:
      Runtime: python3.9
      Handler: index.handler
      Code:
        ZipFile: |
          import boto3
          def handler(event, context):
              ec2 = boto3.client('ec2')
              return ec2.terminate_instances(InstanceIds=[event['InstanceId']])

Outputs:
  EstimatedSavings:
    Value: '$85.75/month'`
}

func (api *RemediationAPI) ListRemediations(w http.ResponseWriter, r *http.Request) {
	// Available remediation types with template format support
	remediations := []map[string]interface{}{
		{
			"type":        "rightsizing",
			"name":        "EC2 Instance Right-sizing",
			"description": "Optimize instance types based on utilization",
			"risk":        "low",
			"formats":     []string{"terraform", "terragrunt", "cloudformation"},
		},
		{
			"type":        "termination",
			"name":        "Terminate Unused Instances",
			"description": "Remove instances with no activity",
			"risk":        "medium",
			"formats":     []string{"terraform", "terragrunt", "cloudformation"},
		},
		{
			"type":        "reserved_instances",
			"name":        "Reserved Instance Recommendations",
			"description": "Purchase RIs for consistent workloads",
			"risk":        "low",
			"formats":     []string{"terraform", "terragrunt"},
		},
	}

	templateFormats := []map[string]interface{}{
		{
			"format":      "terragrunt",
			"name":        "Terragrunt (Recommended)",
			"description": "Enterprise-grade with multi-environment support",
			"features":    []string{"Multi-environment", "DRY configuration", "Remote state", "Dependency management"},
		},
		{
			"format":      "terraform",
			"name":        "Terraform",
			"description": "Standard Terraform templates",
			"features":    []string{"Infrastructure as Code", "Resource management"},
		},
		{
			"format":      "cloudformation",
			"name":        "CloudFormation",
			"description": "AWS native templates",
			"features":    []string{"AWS native", "Stack management"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"remediations":     remediations,
		"template_formats": templateFormats,
		"environments":     []string{"dev", "staging", "prod"},
	})
}

func main() {
	api := NewRemediationAPI()

	r := mux.NewRouter()

	// API endpoints
	r.HandleFunc("/api/remediation/generate", api.GenerateRemediation).Methods("POST")
	r.HandleFunc("/api/remediation/types", api.ListRemediations).Methods("GET")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}).Methods("GET")

	// CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	server := &http.Server{
		Addr:         ":8087",
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down remediation API...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Remediation API starting on port 8087...")
	log.Printf("Endpoints:")
	log.Printf("  POST /api/remediation/generate - Generate remediation templates")
	log.Printf("  GET  /api/remediation/types - List available remediation types")
	log.Printf("  GET  /health - Health check")
	log.Printf("")
	log.Printf("Supported formats: terraform, terragrunt (default), cloudformation")
	log.Printf("Supported environments: dev, staging, prod")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}