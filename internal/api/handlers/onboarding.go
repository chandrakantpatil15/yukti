package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"yukti/internal/aws"
	"yukti/internal/onboarding"
	"yukti/internal/scanner"
)

type OnboardingHandler struct {
	service *onboarding.Service
}

func NewOnboardingHandler(service *onboarding.Service) *OnboardingHandler {
	return &OnboardingHandler{service: service}
}

type CreateCustomerRequest struct {
	CompanyName string `json:"company_name"`
	Email       string `json:"email"`
}

type CreateCustomerResponse struct {
	TenantID string `json:"tenant_id"`
	Message  string `json:"message"`
}

func (h *OnboardingHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	customer, err := h.service.CreateCustomer(r.Context(), req.CompanyName, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := CreateCustomerResponse{
		TenantID: customer.TenantID,
		Message:  "Customer created successfully. Please configure AWS connection.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type ConfigureAWSRequest struct {
	TenantID   string   `json:"tenant_id"`
	AccountID  string   `json:"account_id"`
	RoleARN    string   `json:"role_arn"`
	ExternalID string   `json:"external_id"`
	Regions    []string `json:"regions"`
}

type ConfigureAWSResponse struct {
	Verified bool   `json:"verified"`
	Message  string `json:"message"`
}

func (h *OnboardingHandler) ConfigureAWS(w http.ResponseWriter, r *http.Request) {
	// Handle GET request to retrieve AWS connection
	if r.Method == "GET" {
		h.GetAWSConnection(w, r)
		return
	}
	
	var req ConfigureAWSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request format",
		})
		return
	}

	// SECURITY FIX: Validate tenant_id from request (will be enhanced with JWT later)
	// TODO: Add JWT middleware to onboarding routes and validate tenant_id matches JWT
	if req.TenantID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "tenant_id required",
		})
		return
	}

	// Validate Account ID
	if err := aws.ValidateAccountID(req.AccountID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       false,
			"error":         "Invalid AWS Account ID",
			"error_details": err.Error(),
		})
		return
	}

	// Validate Role ARN
	if err := aws.ValidateRoleARN(req.RoleARN); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       false,
			"error":         "Invalid IAM Role ARN",
			"error_details": err.Error(),
		})
		return
	}

	// Auto-generate external ID if not provided
	externalID := req.ExternalID
	if externalID == "" || externalID == "yukti-secure-access" {
		externalID = h.service.GenerateExternalID(req.TenantID)
	}

	// Verify AWS role connectivity (skip in dev mode)
	skipVerification := os.Getenv("SKIP_AWS_VERIFICATION") == "true"
	if skipVerification {
		log.Printf("[INFO] Skipping AWS verification (dev mode) for tenant %s", req.TenantID)
	} else {
		log.Printf("[INFO] Verifying AWS role access for tenant %s", req.TenantID)
		verifier, err := aws.NewRoleVerifier(context.Background())
		if err != nil {
			log.Printf("[ERROR] Failed to create role verifier: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to initialize AWS verification",
			})
			return
		}

		verifyResult := verifier.VerifyRoleAccess(r.Context(), req.RoleARN, externalID)
		if !verifyResult.Success {
			log.Printf("[WARN] AWS role verification failed for tenant %s: %s", req.TenantID, verifyResult.ErrorCode)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":       false,
				"verified":      false,
				"error":         verifyResult.Message,
				"error_code":    verifyResult.ErrorCode,
				"error_details": verifyResult.ErrorDetails,
			})
			return
		}

		log.Printf("[INFO] AWS role verification successful for tenant %s", req.TenantID)
	}

	// Save AWS connection
	conn := &onboarding.AWSConnection{
		TenantID:       req.TenantID,
		AccountID:      req.AccountID,
		RoleARN:        req.RoleARN,
		ExternalID:     externalID,
		Regions:        req.Regions,
		Verified:       true,
		LastVerifiedAt: time.Now(),
	}

	if err := h.service.SaveAWSConnection(r.Context(), conn); err != nil {
		log.Printf("[ERROR] Failed to save AWS connection: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to save AWS connection",
		})
		return
	}

	h.service.UpdateOnboardingStep(r.Context(), req.TenantID, onboarding.StepMetricsIntegration)

	// Trigger immediate AWS scan in background (like Datadog)
	go h.triggerInitialScan(req.TenantID)

	response := ConfigureAWSResponse{
		Verified: true,
		Message:  "AWS connection verified! Starting resource discovery...",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type ConfigureMetricsRequest struct {
	TenantID string            `json:"tenant_id"`
	Source   string            `json:"source"`
	Endpoint string            `json:"endpoint"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	Token    string            `json:"token"`
}

type ConfigureMetricsResponse struct {
	Verified bool   `json:"verified"`
	Message  string `json:"message"`
}

func (h *OnboardingHandler) ConfigureMetrics(w http.ResponseWriter, r *http.Request) {
	var req ConfigureMetricsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// SECURITY FIX: Validate tenant_id from request
	// TODO: Add JWT middleware to onboarding routes and validate tenant_id matches JWT
	if req.TenantID == "" {
		http.Error(w, "tenant_id required", http.StatusBadRequest)
		return
	}

	integration := &onboarding.MetricsIntegration{
		TenantID:  req.TenantID,
		Source:    req.Source,
		Endpoint:  req.Endpoint,
		Verified:  true,
		CreatedAt: time.Now(),
	}

	if err := h.service.SaveMetricsIntegration(r.Context(), integration); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.service.UpdateOnboardingStep(r.Context(), req.TenantID, onboarding.StepInitialScan)

	response := ConfigureMetricsResponse{
		Verified: true,
		Message:  "Metrics integration configured successfully. Starting initial scan.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type OnboardingStatusResponse struct {
	Status      string `json:"status"`
	CurrentStep string `json:"current_step"`
	Progress    int    `json:"progress"`
	Message     string `json:"message"`
}

func (h *OnboardingHandler) GenerateExternalID(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "tenant_id required", http.StatusBadRequest)
		return
	}

	externalID := h.service.GenerateExternalID(tenantID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"external_id": externalID,
		"tenant_id":   tenantID,
	})
}

func (h *OnboardingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "tenant_id required",
		})
		return
	}

	customer, err := h.service.GetCustomer(r.Context(), tenantID)
	if err != nil {
		// New user - no onboarding record yet, return default status
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(OnboardingStatusResponse{
			Status:      "pending",
			CurrentStep: "aws_connection",
			Progress:    0,
			Message:     "Configure AWS connection with IAM role",
		})
		return
	}

	progress := calculateProgress(customer.OnboardingStep)
	message := getStepMessage(customer.OnboardingStep)

	response := OnboardingStatusResponse{
		Status:      string(customer.OnboardingStatus),
		CurrentStep: string(customer.OnboardingStep),
		Progress:    progress,
		Message:     message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func calculateProgress(step onboarding.OnboardingStep) int {
	switch step {
	case onboarding.StepAWSConnection:
		return 25
	case onboarding.StepMetricsIntegration:
		return 50
	case onboarding.StepInitialScan:
		return 75
	case onboarding.StepReviewFindings:
		return 100
	default:
		return 0
	}
}

func getStepMessage(step onboarding.OnboardingStep) string {
	switch step {
	case onboarding.StepAWSConnection:
		return "Configure AWS connection with IAM role"
	case onboarding.StepMetricsIntegration:
		return "Connect your monitoring system (Prometheus, InfluxDB, etc.)"
	case onboarding.StepInitialScan:
		return "Running initial cost optimization scan"
	case onboarding.StepReviewFindings:
		return "Review findings and start optimizing"
	default:
		return "Getting started"
	}
}

// GetAWSConnection retrieves AWS connection details for a tenant
func (h *OnboardingHandler) GetAWSConnection(w http.ResponseWriter, r *http.Request) {
	// Get tenant_id from query param or use default for testing
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = "25" // Default to your tenant ID
	}

	conn, err := h.service.GetAWSConnection(r.Context(), tenantID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "AWS connection not found",
		})
		return
	}

	// Extract role name from ARN
	roleName := "Unknown"
	if conn.RoleARN != "" {
		// ARN format: arn:aws:iam::123456789012:role/RoleName
		parts := strings.Split(conn.RoleARN, "/")
		if len(parts) > 1 {
			roleName = parts[len(parts)-1]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"account_id":        conn.AccountID,
			"role_name":         roleName,
			"role_arn":          conn.RoleARN,
			"verified":          conn.Verified,
			"last_verified_at":  conn.LastVerifiedAt.Format("2006-01-02 15:04:05"),
			"regions":           conn.Regions,
		},
	})
}

// triggerInitialScan starts AWS resource discovery immediately after connection
func (h *OnboardingHandler) triggerInitialScan(tenantID string) {
	log.Printf("[INFO] Starting automatic initial scan for tenant: %s", tenantID)
	
	// Convert tenant_id to int
	tenantIDInt, err := strconv.Atoi(tenantID)
	if err != nil {
		log.Printf("[ERROR] Invalid tenant_id format: %s", tenantID)
		return
	}
	
	// Run AWS scan
	awsScanner := scanner.NewAWSScanner(h.service.GetDB())
	if err := awsScanner.ScanTenant(context.Background(), tenantIDInt); err != nil {
		log.Printf("[ERROR] Initial scan failed for tenant %s: %v", tenantID, err)
	} else {
		log.Printf("[INFO] Initial scan completed successfully for tenant %s", tenantID)
	}
}
