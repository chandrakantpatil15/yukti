package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"yukti/internal/api/middleware"
)

type FilterHandler struct {
	db *sql.DB
}

func NewFilterHandler(db *sql.DB) *FilterHandler {
	return &FilterHandler{db: db}
}

// GetResourceTypes returns distinct resource types for a tenant
// GET /api/v1/filters/resource-types
func (h *FilterHandler) GetResourceTypes(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] GetResourceTypes request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes cache

	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "tenant_id required",
		})
		return
	}

	rows, err := h.db.Query(`
		SELECT DISTINCT resource_type, COUNT(*) as count
		FROM yt_tenant_resources
		WHERE tenant_id = $1
		GROUP BY resource_type
		ORDER BY resource_type
	`, tenantID)
	if err != nil {
		log.Printf("[ERROR] Failed to query resource types: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to fetch resource types",
		})
		return
	}
	defer rows.Close()

	types := []map[string]interface{}{}
	for rows.Next() {
		var resourceType string
		var count int
		if err := rows.Scan(&resourceType, &count); err != nil {
			continue
		}
		types = append(types, map[string]interface{}{
			"key":   resourceType,
			"label": strings.ToUpper(resourceType),
			"count": count,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    types,
	})
}

// GetTags returns tag keys and values for a tenant
// GET /api/v1/filters/tags?limit=10
func (h *FilterHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] GetTags request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")

	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "tenant_id required",
		})
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get tag keys with counts
	rows, err := h.db.Query(`
		SELECT DISTINCT jsonb_object_keys(tags) as tag_key
		FROM yt_tenant_resources
		WHERE tenant_id = $1 AND tags != '{}'::jsonb
		ORDER BY tag_key
		LIMIT $2
	`, tenantID, limit*2) // Get more keys, then limit values per key
	if err != nil {
		log.Printf("[ERROR] Failed to query tag keys: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to fetch tags",
		})
		return
	}
	defer rows.Close()

	tagKeys := []map[string]interface{}{}
	tagValues := make(map[string][]string)

	for rows.Next() {
		var tagKey string
		if err := rows.Scan(&tagKey); err != nil {
			continue
		}

		// Count resources with this tag key
		var count int
		h.db.QueryRow(`
			SELECT COUNT(*) FROM yt_tenant_resources
			WHERE tenant_id = $1 AND tags ? $2
		`, tenantID, tagKey).Scan(&count)

		tagKeys = append(tagKeys, map[string]interface{}{
			"key":   tagKey,
			"count": count,
		})

		// Get distinct values for this tag key
		valueRows, err := h.db.Query(`
			SELECT DISTINCT tags->>$1 as value
			FROM yt_tenant_resources
			WHERE tenant_id = $2 AND tags ? $1 AND tags->>$1 IS NOT NULL
			ORDER BY value
			LIMIT $3
		`, tagKey, tenantID, limit)
		if err == nil {
			values := []string{}
			for valueRows.Next() {
				var value string
				if err := valueRows.Scan(&value); err == nil && value != "" {
					values = append(values, value)
				}
			}
			valueRows.Close()
			if len(values) > 0 {
				tagValues[tagKey] = values
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"tag_keys":   tagKeys,
			"tag_values": tagValues,
		},
	})
}

// GetServices returns distinct services from cost data
// GET /api/v1/filters/services
func (h *FilterHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] GetServices request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")

	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "tenant_id required",
		})
		return
	}

	rows, err := h.db.Query(`
		SELECT DISTINCT service, SUM(cost) as total_cost
		FROM yt_cost_data
		WHERE tenant_id = $1
		GROUP BY service
		ORDER BY total_cost DESC
	`, tenantID)
	if err != nil {
		log.Printf("[ERROR] Failed to query services: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to fetch services",
		})
		return
	}
	defer rows.Close()

	services := []map[string]interface{}{}
	for rows.Next() {
		var service string
		var totalCost float64
		if err := rows.Scan(&service, &totalCost); err != nil {
			continue
		}
		services = append(services, map[string]interface{}{
			"key":        service,
			"label":      service,
			"total_cost": totalCost,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    services,
	})
}

// GetAccounts returns AWS accounts for a tenant
// GET /api/v1/filters/accounts
func (h *FilterHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] GetAccounts request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")

	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "tenant_id required",
		})
		return
	}

	rows, err := h.db.Query(`
		SELECT id, account_id, account_name, status
		FROM yt_aws_accounts
		WHERE tenant_id = $1
		ORDER BY account_name, account_id
	`, tenantID)
	if err != nil {
		log.Printf("[ERROR] Failed to query accounts: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to fetch accounts",
		})
		return
	}
	defer rows.Close()

	accounts := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var accountID, accountName, status string
		if err := rows.Scan(&id, &accountID, &accountName, &status); err != nil {
			continue
		}
		accounts = append(accounts, map[string]interface{}{
			"id":          id,
			"account_id":  accountID,
			"account_name": accountName,
			"status":      status,
			"label":       accountName + " (" + accountID + ")",
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    accounts,
	})
}

// GetRegions returns distinct regions for a tenant
// GET /api/v1/filters/regions
func (h *FilterHandler) GetRegions(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] GetRegions request from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")

	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "tenant_id required",
		})
		return
	}

	rows, err := h.db.Query(`
		SELECT DISTINCT region, COUNT(*) as count
		FROM yt_tenant_resources
		WHERE tenant_id = $1
		GROUP BY region
		ORDER BY region
	`, tenantID)
	if err != nil {
		log.Printf("[ERROR] Failed to query regions: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to fetch regions",
		})
		return
	}
	defer rows.Close()

	regions := []map[string]interface{}{}
	for rows.Next() {
		var region string
		var count int
		if err := rows.Scan(&region, &count); err != nil {
			continue
		}
		regions = append(regions, map[string]interface{}{
			"key":   region,
			"label": region,
			"count": count,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    regions,
	})
}


