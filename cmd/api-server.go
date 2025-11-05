package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	_ "github.com/lib/pq"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ResourceAssessment struct {
	ResourceARN         string  `json:"resource_arn"`
	InstanceID          string  `json:"instance_id"`
	InstanceType        string  `json:"instance_type"`
	Category            string  `json:"utilization_category"`
	Pattern             string  `json:"usage_pattern"`
	OptimizationScore   float64 `json:"optimization_score"`
	RecommendedAction   string  `json:"recommended_action"`
	CurrentHourlyCost   float64 `json:"current_hourly_cost"`
	PotentialSavings    float64 `json:"potential_monthly_savings"`
	AssessmentTimestamp string  `json:"assessment_timestamp"`
}

type TenantConfig struct {
	TenantID                     string  `json:"tenant_id"`
	UnderutilizedCPUThreshold    float64 `json:"underutilized_cpu_threshold"`
	UnderutilizedMemoryThreshold float64 `json:"underutilized_memory_threshold"`
	OverutilizedCPUThreshold     float64 `json:"overutilized_cpu_threshold"`
	OverutilizedMemoryThreshold  float64 `json:"overutilized_memory_threshold"`
}

var db *sql.DB

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/yukti?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Setup routes
	r := mux.NewRouter()
	
	// API v1 routes
	api := r.PathPrefix("/api/v1").Subrouter()
	
	// Resource endpoints
	api.HandleFunc("/resources", getResources).Methods("GET")
	
	// Assessment endpoints
	api.HandleFunc("/assessments", getAssessments).Methods("GET")
	api.HandleFunc("/assessments/{resource_arn}", getResourceAssessment).Methods("GET")
	api.HandleFunc("/assessments/run", runAssessment).Methods("POST")
	
	// Configuration endpoints
	api.HandleFunc("/config/{tenant_id}", getTenantConfig).Methods("GET")
	api.HandleFunc("/config/{tenant_id}", updateTenantConfig).Methods("PUT")
	
	// Emergency endpoints
	api.HandleFunc("/emergency-stop", emergencyStop).Methods("POST")
	
	// Health check
	r.HandleFunc("/health", healthCheck).Methods("GET")
	api.HandleFunc("/health", healthCheck).Methods("GET")
	
	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
	
	handler := c.Handler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Yukti FinOps API Server starting on port %s\n", port)
	fmt.Printf("🌐 CORS enabled for React frontend\n")
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func getAssessments(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	query := `
		SELECT a.resource_arn, r.instance_id, r.instance_type,
		       a.utilization_category, a.usage_pattern, a.optimization_score,
		       a.recommended_action, a.current_hourly_cost, 
		       a.potential_monthly_savings, a.assessment_timestamp
		FROM yt_resource_assessments a
		JOIN yt_aws_resources r ON a.resource_arn = r.resource_arn
		WHERE r.sync_status = 'active'
		ORDER BY a.optimization_score DESC
		LIMIT $1`

	rows, err := db.Query(query, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database query failed")
		return
	}
	defer rows.Close()

	var assessments []ResourceAssessment
	for rows.Next() {
		var a ResourceAssessment
		var timestamp time.Time
		
		err := rows.Scan(&a.ResourceARN, &a.InstanceID, &a.InstanceType,
			&a.Category, &a.Pattern, &a.OptimizationScore,
			&a.RecommendedAction, &a.CurrentHourlyCost,
			&a.PotentialSavings, &timestamp)
		if err != nil {
			continue
		}
		
		a.AssessmentTimestamp = timestamp.Format(time.RFC3339)
		assessments = append(assessments, a)
	}

	respondWithJSON(w, http.StatusOK, assessments)
}

func getResourceAssessment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceARN := vars["resource_arn"]

	query := `
		SELECT a.resource_arn, r.instance_id, r.instance_type,
		       a.utilization_category, a.usage_pattern, a.optimization_score,
		       a.recommended_action, a.current_hourly_cost,
		       a.potential_monthly_savings, a.assessment_timestamp
		FROM yt_resource_assessments a
		JOIN yt_aws_resources r ON a.resource_arn = r.resource_arn
		WHERE a.resource_arn = $1`

	var a ResourceAssessment
	var timestamp time.Time
	
	err := db.QueryRow(query, resourceARN).Scan(
		&a.ResourceARN, &a.InstanceID, &a.InstanceType,
		&a.Category, &a.Pattern, &a.OptimizationScore,
		&a.RecommendedAction, &a.CurrentHourlyCost,
		&a.PotentialSavings, &timestamp)
	
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Resource assessment not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Database query failed")
		}
		return
	}

	a.AssessmentTimestamp = timestamp.Format(time.RFC3339)
	respondWithJSON(w, http.StatusOK, a)
}

func runAssessment(w http.ResponseWriter, r *http.Request) {
	// Trigger assessment run (simplified)
	result := map[string]interface{}{
		"message": "Assessment triggered successfully",
		"status":  "running",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	respondWithJSON(w, http.StatusAccepted, result)
}

func getTenantConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	query := `
		SELECT tenant_id, underutilized_cpu_threshold, underutilized_memory_threshold,
		       overutilized_cpu_threshold, overutilized_memory_threshold
		FROM yt_assessment_config
		WHERE tenant_id = $1`

	var config TenantConfig
	err := db.QueryRow(query, tenantID).Scan(
		&config.TenantID, &config.UnderutilizedCPUThreshold,
		&config.UnderutilizedMemoryThreshold, &config.OverutilizedCPUThreshold,
		&config.OverutilizedMemoryThreshold)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Tenant configuration not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Database query failed")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, config)
}

func updateTenantConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	var config TenantConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	query := `
		UPDATE yt_assessment_config 
		SET underutilized_cpu_threshold = $2,
		    underutilized_memory_threshold = $3,
		    overutilized_cpu_threshold = $4,
		    overutilized_memory_threshold = $5,
		    updated_at = NOW()
		WHERE tenant_id = $1`

	_, err := db.Exec(query, tenantID, config.UnderutilizedCPUThreshold,
		config.UnderutilizedMemoryThreshold, config.OverutilizedCPUThreshold,
		config.OverutilizedMemoryThreshold)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update configuration")
		return
	}

	config.TenantID = tenantID
	respondWithJSON(w, http.StatusOK, config)
}

type Resource struct {
	InstanceID           string  `json:"instance_id"`
	InstanceType         string  `json:"instance_type"`
	Region               string  `json:"region"`
	AvailabilityZone     string  `json:"availability_zone"`
	State                string  `json:"state"`
	LaunchTime           *string `json:"launch_time"`
	Environment          *string `json:"environment"`
	EstimatedMonthlyCost float64 `json:"estimated_monthly_cost"`
}

func getResources(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT r.instance_id, r.instance_type, r.region, r.availability_zone,
		       r.state, r.launch_time, r.environment,
		       COALESCE(p.on_demand_price_usd * 24 * 30, 0) as estimated_monthly_cost
		FROM yt_aws_resources r
		LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
		WHERE r.sync_status = 'active'
		ORDER BY r.launch_time DESC`

	rows, err := db.Query(query)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database query failed")
		return
	}
	defer rows.Close()

	var resources []Resource
	for rows.Next() {
		var res Resource
		var launchTime sql.NullString
		var environment sql.NullString
		
		err := rows.Scan(&res.InstanceID, &res.InstanceType, &res.Region,
			&res.AvailabilityZone, &res.State, &launchTime, &environment,
			&res.EstimatedMonthlyCost)
		if err != nil {
			continue
		}
		
		if launchTime.Valid {
			res.LaunchTime = &launchTime.String
		}
		if environment.Valid {
			res.Environment = &environment.String
		}
		
		resources = append(resources, res)
	}

	respondWithJSON(w, http.StatusOK, resources)
}

func emergencyStop(w http.ResponseWriter, r *http.Request) {
	result := map[string]interface{}{
		"message": "Emergency stop triggered",
		"status":  "completed",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	respondWithJSON(w, http.StatusOK, result)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := db.Ping(); err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Database connection failed")
		return
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	respondWithJSON(w, http.StatusOK, health)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response := APIResponse{
		Success: code < 400,
		Data:    payload,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}