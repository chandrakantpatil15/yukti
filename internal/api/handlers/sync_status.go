package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type SyncHandler struct {
	db *sql.DB
}

func NewSyncHandler(db *sql.DB) *SyncHandler {
	return &SyncHandler{db: db}
}

// SyncStatusResponse is the payload returned to the frontend
type SyncStatusResponse struct {
	LastSyncStart    string           `json:"last_sync_start,omitempty"`
	LastSyncEnd      string           `json:"last_sync_end,omitempty"`
	Status           string           `json:"status,omitempty"`
	ProcessedRecords int              `json:"processed_records"`
	ErrorCount       int              `json:"error_count"`
	LastService      string           `json:"last_service,omitempty"`
	LastRegion       string           `json:"last_region,omitempty"`
	ErrorMessage     string           `json:"error_message,omitempty"`
	ServiceCounts    map[string]int64 `json:"service_counts,omitempty"`
}

// GetStatus returns the latest global AWS pricing sync status and counts
func (h *SyncHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	// Query latest sync status
	row := h.db.QueryRow(`SELECT sync_start_time, sync_end_time, status, processed_records, error_count, last_service, last_region, error_message FROM yt_aws_pricing_sync_status ORDER BY sync_start_time DESC LIMIT 1`)

	var start sql.NullString
	var end sql.NullString
	var status sql.NullString
	var processed int
	var errors int
	var lastService sql.NullString
	var lastRegion sql.NullString
	var errorMessage sql.NullString

	err := row.Scan(&start, &end, &status, &processed, &errors, &lastService, &lastRegion, &errorMessage)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[ERROR] fetching sync status: %v", err)
		http.Error(w, "failed to fetch sync status", http.StatusInternalServerError)
		return
	}

	// Query active pricing counts by service
	counts := make(map[string]int64)
	rows, err := h.db.Query(`SELECT service_code, COUNT(*) as cnt FROM yt_aws_pricing WHERE is_active = true GROUP BY service_code`)
	if err != nil {
		log.Printf("[ERROR] fetching service counts: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var svc string
			var cnt int64
			if err := rows.Scan(&svc, &cnt); err != nil {
				log.Printf("[WARN] scan service count: %v", err)
				continue
			}
			counts[svc] = cnt
		}
	}

	resp := SyncStatusResponse{
		ServiceCounts:    counts,
		ProcessedRecords: processed,
		ErrorCount:       errors,
	}
	if start.Valid {
		resp.LastSyncStart = start.String
	}
	if end.Valid {
		resp.LastSyncEnd = end.String
	}
	if status.Valid {
		resp.Status = status.String
	}
	if lastService.Valid {
		resp.LastService = lastService.String
	}
	if lastRegion.Valid {
		resp.LastRegion = lastRegion.String
	}
	if errorMessage.Valid {
		resp.ErrorMessage = errorMessage.String
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
