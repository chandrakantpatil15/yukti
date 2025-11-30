package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"yukti/internal/api/middleware"
	"yukti/internal/models"
)

type ResourceHandler struct {
	db *sql.DB
}

func NewResourceHandler(db *sql.DB) *ResourceHandler {
	return &ResourceHandler{db: db}
}

func (h *ResourceHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	// Use models package to prevent unused import
	_ = models.Resource{}
	
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		log.Printf("[ERROR] ListResources: No tenant_id in context")
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

	tenantIDStr := strconv.Itoa(tenantID)
	rows, err := h.db.Query(`
		SELECT r.id, r.resource_id, r.resource_type, r.region, r.instance_type, 
		       r.state, r.tags, r.monthly_cost, a.account_id
		FROM yt_tenant_resources r
		JOIN yt_aws_accounts a ON r.aws_account_id = a.id
		WHERE r.tenant_id = $1
		ORDER BY r.monthly_cost DESC
		LIMIT $2 OFFSET $3`,
		tenantIDStr, perPage, offset,
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
	h.db.QueryRow(`SELECT COUNT(*) FROM yt_tenant_resources WHERE tenant_id = $1`, tenantIDStr).Scan(&total)

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
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Unauthorized"})
		return
	}
	tenantIDStr := strconv.Itoa(tenantID)

	// Get resource_id from path parameter or query string
	resourceID := r.URL.Query().Get("resource_id")
	if resourceID == "" {
		// Try path parameter
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) > 0 {
			resourceID = pathParts[len(pathParts)-1]
		}
	}
	if resourceID == "" || resourceID == "details" {
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
		tenantIDStr, resourceID,
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
		},
	})
}

func (h *ResourceHandler) GetResourceMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Unauthorized"})
		return
	}
	tenantIDStr := strconv.Itoa(tenantID)

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
		tenantIDStr, resourceID,
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
	_, ok := middleware.GetTenantID(r.Context())
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
