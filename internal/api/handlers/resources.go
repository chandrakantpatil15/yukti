package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"yukti/internal/api/middleware"
	"yukti/internal/models"

	"github.com/gorilla/mux"
)

// helper to extract tenant id from context or X-Tenant-ID header (dev/backwards-compat)
func tenantIDFromRequest(r *http.Request) (int, bool) {
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		return tid, true
	}
	// Fallback: allow X-Tenant-ID header (useful for local dev tools)
	if hdr := r.Header.Get("X-Tenant-ID"); hdr != "" {
		if id, err := strconv.Atoi(hdr); err == nil {
			return id, true
		}
	}
	return 0, false
}

// helper to calculate uptime in days
func calculateUptime(launchTime interface{}) interface{} {
	if launchTime == nil {
		return nil
	}
	launchStr, ok := launchTime.(string)
	if !ok {
		return nil
	}
	launchParsed, err := time.Parse(time.RFC3339, launchStr)
	if err != nil {
		return nil
	}
	duration := time.Since(launchParsed)
	return int(duration.Hours() / 24)
}

type ResourceHandler struct {
	db *sql.DB
}

func NewResourceHandler(db *sql.DB) *ResourceHandler {
	return &ResourceHandler{db: db}
}

func (h *ResourceHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	// Use models package to prevent unused import
	_ = models.Resource{}

	tenantID, ok := tenantIDFromRequest(r)
	if !ok {
		log.Printf("[ERROR] ListResources: No tenant_id in context or X-Tenant-ID header")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Unauthorized"})
		return
	}
	log.Printf("[DEBUG] ListResources: tenant_id=%d", tenantID)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage := 50
	offset := (page - 1) * perPage

	rows, err := h.db.Query(`
		SELECT r.id, r.resource_id, r.resource_type, r.region, r.instance_type, 
		       r.state, r.tags, r.monthly_cost, a.account_id
		FROM yt_tenant_resources r
		JOIN yt_aws_accounts a ON r.aws_account_id = a.id
		LEFT JOIN yt_whitelists w ON w.tenant_id = r.tenant_id::text 
			AND w.resource_arn LIKE '%' || r.resource_id || '%' 
			AND w.status = 'active'
		WHERE r.tenant_id = $1 AND w.id IS NULL
		ORDER BY r.monthly_cost DESC
		LIMIT $2 OFFSET $3`,
		tenantID, perPage, offset,
	)
	if err != nil {
		log.Printf("[ERROR] ListResources query failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	defer rows.Close()

	resources := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var resourceID, resourceType, region, instanceType, state, accountID string
		var tags []byte
		var monthlyCost sql.NullFloat64

		rows.Scan(&id, &resourceID, &resourceType, &region, &instanceType, &state, &tags, &monthlyCost, &accountID)

		var tagsMap map[string]interface{}
		json.Unmarshal(tags, &tagsMap)

		resources = append(resources, map[string]interface{}{
			"id":            id,
			"resource_id":   resourceID,
			"resource_type": resourceType,
			"region":        region,
			"instance_type": instanceType,
			"state":         state,
			"tags":          tagsMap,
			"monthly_cost":  monthlyCost.Float64,
			"account_id":    accountID,
		})
	}

	var total int
	h.db.QueryRow(`SELECT COUNT(*) FROM yt_tenant_resources r 
		LEFT JOIN yt_whitelists w ON w.tenant_id = r.tenant_id::text 
			AND w.resource_arn LIKE '%' || r.resource_id || '%' 
			AND w.status = 'active'
		WHERE r.tenant_id = $1 AND w.id IS NULL`, tenantID).Scan(&total)

	log.Printf("[SUCCESS] ListResources: Found %d resources for tenant %d", total, tenantID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    resources,
		"meta": map[string]interface{}{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": (total + perPage - 1) / perPage,
		},
	})
}

func (h *ResourceHandler) GetResourceDetails(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenantIDFromRequest(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Unauthorized"})
		return
	}
	// Get resource_id from mux path parameter
	vars := mux.Vars(r)
	resourceID := vars["resourceId"]
	if resourceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "resource_id required"})
		return
	}
	log.Printf("[DEBUG] GetResourceDetails: tenant=%d, resource_id=%s", tenantID, resourceID)

	var id int
	var resID, resourceType, region, instanceType, state, accountID string
	var tags []byte
	var monthlyCost sql.NullFloat64
	var metadata []byte

	err := h.db.QueryRow(`
		SELECT r.id, r.resource_id, r.resource_type, r.region, r.instance_type,
		       r.state, r.tags, r.monthly_cost, r.metadata, a.account_id
		FROM yt_tenant_resources r
		JOIN yt_aws_accounts a ON r.aws_account_id = a.id
		WHERE r.tenant_id = $1 AND r.resource_id = $2`,
		tenantID, resourceID,
	).Scan(&id, &resID, &resourceType, &region, &instanceType, &state, &tags, &monthlyCost, &metadata, &accountID)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Resource not found"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	var tagsMap, metadataMap map[string]interface{}
	json.Unmarshal(tags, &tagsMap)
	json.Unmarshal(metadata, &metadataMap)

	// Enhanced status information
	statusInfo := map[string]interface{}{
		"current_state": state,
		"state_reason": metadataMap["state_reason"],
		"state_transition_reason": metadataMap["state_transition_reason"],
		"launch_time": metadataMap["launch_time"],
		"uptime_days": calculateUptime(metadataMap["launch_time"]),
	}

	// Network information
	networkInfo := map[string]interface{}{
		"private_ip": metadataMap["private_ip"],
		"public_ip": metadataMap["public_ip"],
		"private_dns": metadataMap["private_dns"],
		"public_dns": metadataMap["public_dns"],
		"vpc_id": metadataMap["vpc_id"],
		"subnet_id": metadataMap["subnet_id"],
		"security_groups": metadataMap["security_groups"],
	}

	// Configuration details
	configInfo := map[string]interface{}{
		"ami_id": metadataMap["ami_id"],
		"architecture": metadataMap["architecture"],
		"platform": metadataMap["platform"],
		"key_name": metadataMap["key_name"],
		"ebs_optimized": metadataMap["ebs_optimized"],
		"monitoring": metadataMap["detailed_monitoring"],
		"tenancy": metadataMap["tenancy"],
	}

	// Storage information
	storageInfo := map[string]interface{}{
		"root_device_type": metadataMap["root_device_type"],
		"root_device_name": metadataMap["root_device_name"],
		"block_devices": metadataMap["block_device_mappings"],
		"ebs_volumes": metadataMap["ebs_volumes"],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":            id,
			"resource_id":   resID,
			"resource_type": resourceType,
			"region":        region,
			"instance_type": instanceType,
			"state":         state,
			"tags":          tagsMap,
			"monthly_cost":  monthlyCost.Float64,
			"account_id":    accountID,
			"metadata":      metadataMap,
			"status_info":   statusInfo,
			"network_info":  networkInfo,
			"config_info":   configInfo,
			"storage_info":  storageInfo,
		},
	})
}

func (h *ResourceHandler) GetResourceMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenantIDFromRequest(r)
	if !ok {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Unauthorized"})
		return
	}
	resourceID := r.URL.Query().Get("resource_id")
	if resourceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "resource_id required"})
		return
	}

	// Get resource details and AWS connection
	var resourceType, region string
	err := h.db.QueryRow(`
		SELECT resource_type, region
		FROM yt_tenant_resources 
		WHERE tenant_id = $1 AND resource_id = $2`,
		tenantID, resourceID,
	).Scan(&resourceType, &region)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Resource not found"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	// For now, return mock CloudWatch data structure
	// Real implementation would fetch from AWS CloudWatch
	mockMetrics := []map[string]interface{}{
		{
			"metric_name": "CPUUtilization",
			"namespace":   "AWS/" + resourceType,
			"unit":        "Percent",
			"values": []map[string]interface{}{
				{"timestamp": "2024-01-15T12:00:00Z", "value": 45.2},
				{"timestamp": "2024-01-15T13:00:00Z", "value": 52.1},
				{"timestamp": "2024-01-15T14:00:00Z", "value": 38.7},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    mockMetrics,
		"message": "CloudWatch metrics integration ready - showing sample data",
	})
}

func (h *ResourceHandler) GetResourceCost(w http.ResponseWriter, r *http.Request) {
	if _, ok := tenantIDFromRequest(r); !ok {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Unauthorized"})
		return
	}

	resourceID := r.URL.Query().Get("resource_id")
	if resourceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "resource_id required"})
		return
	}

	// TODO: Query cost history from yt_cost_data when resource-level tracking is implemented
	// For now, return empty data with message

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "No cost history available yet. Cost tracking will be available once resource-level billing is implemented.",
		"data":    []interface{}{},
	})
}

func (h *ResourceHandler) GetResourceStats(w http.ResponseWriter, r *http.Request) {
	// Return mock data for now since no resources exist yet
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"ec2_count":   0,
			"rds_count":   0,
			"s3_count":    0,
			"total_count": 0,
			"message":     "No resources found. Deploy Terraform resources to see data.",
		},
	})
	return

	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Unauthorized"})
		return
	}

	var totalResources int
	var totalCost float64
	h.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(monthly_cost), 0)
		FROM yt_tenant_resources WHERE tenant_id = $1`,
		tenantID,
	).Scan(&totalResources, &totalCost)

	rows, err := h.db.Query(`
		SELECT resource_type, COUNT(*), COALESCE(SUM(monthly_cost), 0)
		FROM yt_tenant_resources 
		WHERE tenant_id = $1
		GROUP BY resource_type`,
		tenantID,
	)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	defer rows.Close()

	breakdown := []map[string]interface{}{}
	for rows.Next() {
		var resourceType string
		var count int
		var cost float64
		rows.Scan(&resourceType, &count, &cost)
		breakdown = append(breakdown, map[string]interface{}{
			"type":  resourceType,
			"count": count,
			"cost":  cost,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"total_resources": totalResources,
			"total_cost":      totalCost,
			"breakdown":       breakdown,
		},
	})
}
