package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"yukti/internal/api/middleware"
	"yukti/internal/scanner"
)

type ScanHandler struct {
	db              *sql.DB
	lastScanTime    map[int]time.Time
	scanningTenants map[int]bool
	lastErrors      map[int]string
	mu              sync.RWMutex
}

func NewScanHandler(db *sql.DB) *ScanHandler {
	return &ScanHandler{
		db:              db,
		lastScanTime:    make(map[int]time.Time),
		scanningTenants: make(map[int]bool),
		lastErrors:      make(map[int]string),
	}
}

// TriggerScan manually triggers an AWS scan for the authenticated tenant
func (h *ScanHandler) TriggerScan(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		log.Printf("[ScanAPI] ERROR: Unauthorized scan request - no tenant ID in context")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized - please login again",
		})
		return
	}

	log.Printf("[ScanAPI] ========== SCAN REQUEST RECEIVED ===========")
	log.Printf("[ScanAPI] Tenant ID: %d", tenantID)
	log.Printf("[ScanAPI] User Agent: %s", r.Header.Get("User-Agent"))
	log.Printf("[ScanAPI] Remote IP: %s", r.RemoteAddr)

	// Check if scan already in progress
	h.mu.Lock()
	isScanning := h.scanningTenants[tenantID]
	if !isScanning {
		h.scanningTenants[tenantID] = true
		h.lastScanTime[tenantID] = time.Now()
	}
	h.mu.Unlock()

	if isScanning {
		log.Printf("[ScanAPI] WARNING: Scan already in progress for tenant %d", tenantID)
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Scan already in progress. Please wait.",
			"code":    "SCAN_IN_PROGRESS",
		})
		return
	}

	log.Printf("[ScanAPI] ✅ Scan throttling check passed")

	// Check if AWS connection exists
	var accountID, roleARN, externalID string
	var verified bool
	var lastVerified sql.NullTime
	tenantIDStr := strconv.Itoa(tenantID)

	log.Printf("[ScanAPI] Checking AWS connection for tenant: %s", tenantIDStr)
	err := h.db.QueryRow(`
		SELECT account_id, role_arn, external_id, verified, last_verified_at
		FROM yt_aws_connections
		WHERE tenant_id = $1
		LIMIT 1
	`, tenantIDStr).Scan(&accountID, &roleARN, &externalID, &verified, &lastVerified)

	if err == sql.ErrNoRows {
		log.Printf("[ScanAPI] ERROR: No AWS connection found for tenant %d", tenantID)
		log.Printf("[ScanAPI] CUSTOMER ACTION REQUIRED: Complete onboarding at /onboarding")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No AWS connection configured. Please complete onboarding first.",
			"action":  "Go to Settings > AWS Connection to configure your account",
			"code":    "NO_AWS_CONNECTION",
		})
		return
	}

	if err != nil {
		log.Printf("[ScanAPI] ERROR: Database query failed for tenant %d: %v", tenantID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Database error - please try again",
			"code":    "DATABASE_ERROR",
		})
		return
	}

	log.Printf("[ScanAPI] AWS Connection Found:")
	log.Printf("[ScanAPI]   Account: %s", accountID)
	log.Printf("[ScanAPI]   Role: %s", roleARN)
	log.Printf("[ScanAPI]   Verified: %t", verified)
	if lastVerified.Valid {
		log.Printf("[ScanAPI]   Last Verified: %s", lastVerified.Time.Format("2006-01-02 15:04:05"))
	}

	if !verified {
		log.Printf("[ScanAPI] ERROR: AWS connection not verified for tenant %d", tenantID)
		log.Printf("[ScanAPI] CUSTOMER ACTION REQUIRED: Verify IAM role setup")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "AWS connection not verified. Please check your IAM role setup.",
			"action":  "Verify your IAM role trust policy includes the correct external ID",
			"code":    "AWS_NOT_VERIFIED",
			"details": map[string]string{
				"account_id":  accountID,
				"role_arn":    roleARN,
				"external_id": externalID,
			},
		})
		return
	}

	log.Printf("[ScanAPI] ✓ AWS connection verified - starting background scan")

	// Run AWS scan in background
	go func() {
		defer func() {
			// Always clear scanning flag when done
			h.mu.Lock()
			delete(h.scanningTenants, tenantID)
			h.mu.Unlock()
		}()

		log.Printf("[ScanAPI] Background scan goroutine started for tenant %d", tenantID)
		awsScanner := scanner.NewAWSScanner(h.db)
		if err := awsScanner.ScanTenant(context.Background(), tenantID); err != nil {
			log.Printf("[ScanAPI] ERROR: AWS scan failed for tenant %d: %v", tenantID, err)
			log.Printf("[ScanAPI] TROUBLESHOOTING: Check IAM permissions and AWS service availability")

			// Store error for dashboard display
			h.mu.Lock()
			h.lastErrors[tenantID] = err.Error()
			h.mu.Unlock()
		} else {
			log.Printf("[ScanAPI] ✓ AWS scan completed successfully for tenant %d", tenantID)
			log.Printf("[ScanAPI] Results should now be visible in dashboard")

			// Clear any previous errors
			h.mu.Lock()
			delete(h.lastErrors, tenantID)
			h.mu.Unlock()
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"message": "AWS resource scan initiated successfully",
		"status":  "running",
		"account": accountID,
		"details": map[string]interface{}{
			"scan_id":           fmt.Sprintf("scan-%d-%d", tenantID, time.Now().Unix()),
			"tenant_id":         tenantID,
			"started_at":        time.Now().Format(time.RFC3339),
			"expected_duration": "30-60 seconds",
		},
	}

	log.Printf("[ScanAPI] ✓ Scan request accepted - returning success response")
	json.NewEncoder(w).Encode(response)
}

// GetScanStatus returns the current scan status and recent findings
func (h *ScanHandler) GetScanStatus(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	log.Printf("[ScanAPI] Scan status requested for tenant: %d", tenantID)

	// Get AWS connection status
	var accountID, roleARN string
	var verified bool
	var lastVerified sql.NullTime
	tenantIDStr := strconv.Itoa(tenantID)

	err := h.db.QueryRow(`
		SELECT account_id, role_arn, verified, last_verified_at
		FROM yt_aws_connections
		WHERE tenant_id = $1
		LIMIT 1
	`, tenantIDStr).Scan(&accountID, &roleARN, &verified, &lastVerified)

	connectionStatus := "not_configured"
	if err == nil {
		if verified {
			connectionStatus = "verified"
		} else {
			connectionStatus = "unverified"
		}
	}

	// Get recent findings count
	var findingsCount int
	h.db.QueryRow(`
		SELECT COUNT(*)
		FROM yt_hidden_cost_findings
		WHERE tenant_id = $1
		AND created_at > NOW() - INTERVAL '1 hour'
	`, tenantIDStr).Scan(&findingsCount)

	// Get total findings count
	var totalFindings int
	h.db.QueryRow(`
		SELECT COUNT(*)
		FROM yt_hidden_cost_findings
		WHERE tenant_id = $1
	`, tenantIDStr).Scan(&totalFindings)

	// Get last error if any
	h.mu.RLock()
	lastError := h.lastErrors[tenantID]
	isScanning := h.scanningTenants[tenantID]
	h.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"tenant_id":         tenantID,
			"connection_status": connectionStatus,
			"aws_account":       accountID,
			"role_arn":          roleARN,
			"verified":          verified,
			"last_verified":     lastVerified,
			"recent_findings":   findingsCount,
			"total_findings":    totalFindings,
			"is_scanning":       isScanning,
			"last_error":        lastError,
			"timestamp":         time.Now().Format(time.RFC3339),
			"troubleshooting": map[string]string{
				"logs_command": "docker-compose logs -f backend | grep Scanner",
				"tenant_logs":  fmt.Sprintf("docker-compose logs backend | grep 'tenant.*%d'", tenantID),
				"guide_url":    "/docs/troubleshooting",
			},
		},
	})
}
