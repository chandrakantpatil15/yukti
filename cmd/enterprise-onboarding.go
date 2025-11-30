package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type EnterpriseOnboardingService struct {
	activeOnboardings map[string]*OnboardingSession
	mu                sync.RWMutex
	metrics          *OnboardingMetrics
}

type OnboardingSession struct {
	CustomerID    string                 `json:"customer_id"`
	Status        string                 `json:"status"`
	Progress      int                    `json:"progress"`
	StartTime     time.Time              `json:"start_time"`
	LastUpdate    time.Time              `json:"last_update"`
	Phase         string                 `json:"phase"`
	Data          map[string]interface{} `json:"data"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
}

type OnboardingMetrics struct {
	activeOnboardings   prometheus.Gauge
	completedOnboardings prometheus.Counter
	averageOnboardingTime prometheus.Histogram
	errorRate           prometheus.Counter
}

func NewEnterpriseOnboardingService() *EnterpriseOnboardingService {
	metrics := &OnboardingMetrics{
		activeOnboardings: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "yukti_active_onboardings",
			Help: "Number of active customer onboardings",
		}),
		completedOnboardings: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "yukti_completed_onboardings_total",
			Help: "Total number of completed onboardings",
		}),
		averageOnboardingTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "yukti_onboarding_duration_seconds",
			Help: "Time taken for customer onboarding",
			Buckets: []float64{30, 60, 120, 300, 600, 1200, 1800, 3600},
		}),
		errorRate: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "yukti_onboarding_errors_total",
			Help: "Total number of onboarding errors",
		}),
	}

	// Register metrics
	prometheus.MustRegister(metrics.activeOnboardings)
	prometheus.MustRegister(metrics.completedOnboardings)
	prometheus.MustRegister(metrics.averageOnboardingTime)
	prometheus.MustRegister(metrics.errorRate)

	return &EnterpriseOnboardingService{
		activeOnboardings: make(map[string]*OnboardingSession),
		metrics:          metrics,
	}
}

func (s *EnterpriseOnboardingService) StartOnboarding(w http.ResponseWriter, r *http.Request) {
	var request struct {
		CustomerID string `json:"customer_id"`
		AWSConfig  struct {
			AccessKey string `json:"access_key"`
			SecretKey string `json:"secret_key"`
			Region    string `json:"region"`
		} `json:"aws_config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.CustomerID == "" {
		http.Error(w, "customer_id is required", http.StatusBadRequest)
		return
	}

	// Check if onboarding already in progress
	s.mu.RLock()
	if _, exists := s.activeOnboardings[request.CustomerID]; exists {
		s.mu.RUnlock()
		http.Error(w, "Onboarding already in progress", http.StatusConflict)
		return
	}
	s.mu.RUnlock()

	// Create new onboarding session
	session := &OnboardingSession{
		CustomerID: request.CustomerID,
		Status:     "initializing",
		Progress:   0,
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
		Phase:      "starting",
		Data:       make(map[string]interface{}),
	}

	s.mu.Lock()
	s.activeOnboardings[request.CustomerID] = session
	s.mu.Unlock()

	s.metrics.activeOnboardings.Inc()

	// Start progressive onboarding in background
	go s.runProgressiveOnboarding(request.CustomerID, session)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Onboarding started",
		"customer_id": request.CustomerID,
		"status":      "initializing",
		"estimated_ready_time": "2 minutes",
	})
}

func (s *EnterpriseOnboardingService) runProgressiveOnboarding(customerID string, session *OnboardingSession) {
	ctx := context.Background()
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		s.metrics.averageOnboardingTime.Observe(duration.Seconds())
		s.metrics.activeOnboardings.Dec()
		s.metrics.completedOnboardings.Inc()
	}()

	log.Printf("Starting progressive onboarding for customer: %s", customerID)

	// Phase 1: Load recent billing data (Priority 1 - 30 seconds)
	s.updateSession(customerID, "loading_recent_data", 10, "Loading recent billing data...")
	if err := s.loadRecentBillingData(ctx, customerID); err != nil {
		s.handleOnboardingError(customerID, "recent_data_error", err)
		return
	}
	s.updateSession(customerID, "recent_data_loaded", 25, "Recent billing data loaded")

	// Phase 2: Generate quick recommendations (Priority 1 - 60 seconds)
	s.updateSession(customerID, "generating_recommendations", 35, "Analyzing costs and generating recommendations...")
	recommendations := s.generateQuickRecommendations(ctx, customerID)
	s.updateSessionData(customerID, "recommendations", recommendations)
	s.updateSession(customerID, "recommendations_ready", 50, "Initial recommendations ready")

	log.Printf("Customer %s ready for dashboard access (2-minute mark)", customerID)

	// Phase 3: Load historical data (Background - 10 minutes)
	go func() {
		s.updateSession(customerID, "loading_historical", 60, "Loading historical data in background...")
		if err := s.loadHistoricalData(ctx, customerID); err != nil {
			log.Printf("Error loading historical data for %s: %v", customerID, err)
			return
		}
		s.updateSession(customerID, "historical_loaded", 75, "Historical data loaded")
	}()

	// Phase 4: Deep analysis (Background - 30 minutes)
	go func() {
		time.Sleep(5 * time.Minute) // Wait for historical data
		s.updateSession(customerID, "deep_analysis", 85, "Performing comprehensive analysis...")
		if err := s.performDeepAnalysis(ctx, customerID); err != nil {
			log.Printf("Error in deep analysis for %s: %v", customerID, err)
			return
		}
		s.updateSession(customerID, "complete", 100, "Onboarding complete")
	}()
}

func (s *EnterpriseOnboardingService) loadRecentBillingData(ctx context.Context, customerID string) error {
	// Simulate AWS Cost Explorer API call for last 7 days
	log.Printf("Loading recent billing data for customer: %s", customerID)
	
	// Simulate API processing time
	time.Sleep(30 * time.Second)
	
	// Mock recent cost data
	recentCosts := map[string]interface{}{
		"total_cost_7_days": 1250.75,
		"daily_average":     178.68,
		"top_services": []map[string]interface{}{
			{"service": "EC2", "cost": 850.25, "percentage": 68},
			{"service": "RDS", "cost": 200.50, "percentage": 16},
			{"service": "S3", "cost": 100.00, "percentage": 8},
		},
	}
	
	s.updateSessionData(customerID, "recent_costs", recentCosts)
	return nil
}

func (s *EnterpriseOnboardingService) generateQuickRecommendations(ctx context.Context, customerID string) []map[string]interface{} {
	log.Printf("Generating quick recommendations for customer: %s", customerID)
	
	// Simulate analysis time
	time.Sleep(45 * time.Second)
	
	return []map[string]interface{}{
		{
			"type":             "rightsizing",
			"resource_id":      "i-1234567890abcdef0",
			"current_type":     "m5.large",
			"recommended_type": "m5.medium",
			"monthly_savings":  120.50,
			"confidence":       "high",
			"risk":            "low",
		},
		{
			"type":            "termination",
			"resource_id":     "i-0987654321fedcba0",
			"reason":          "unused_for_30_days",
			"monthly_savings": 85.75,
			"confidence":      "high",
			"risk":           "low",
		},
		{
			"type":            "reserved_instances",
			"instance_type":   "m5.large",
			"quantity":        3,
			"monthly_savings": 200.00,
			"confidence":      "medium",
			"risk":           "low",
		},
	}
}

func (s *EnterpriseOnboardingService) loadHistoricalData(ctx context.Context, customerID string) error {
	log.Printf("Loading historical data for customer: %s", customerID)
	time.Sleep(8 * time.Minute) // Simulate longer processing
	return nil
}

func (s *EnterpriseOnboardingService) performDeepAnalysis(ctx context.Context, customerID string) error {
	log.Printf("Performing deep analysis for customer: %s", customerID)
	time.Sleep(15 * time.Minute) // Simulate comprehensive analysis
	return nil
}

func (s *EnterpriseOnboardingService) GetOnboardingStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID := vars["customer_id"]

	s.mu.RLock()
	session, exists := s.activeOnboardings[customerID]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Onboarding session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (s *EnterpriseOnboardingService) updateSession(customerID, status string, progress int, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.activeOnboardings[customerID]; exists {
		session.Status = status
		session.Progress = progress
		session.Phase = message
		session.LastUpdate = time.Now()
	}
}

func (s *EnterpriseOnboardingService) updateSessionData(customerID string, key string, data interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.activeOnboardings[customerID]; exists {
		session.Data[key] = data
		session.LastUpdate = time.Now()
	}
}

func (s *EnterpriseOnboardingService) handleOnboardingError(customerID, status string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.activeOnboardings[customerID]; exists {
		session.Status = status
		session.ErrorMessage = err.Error()
		session.LastUpdate = time.Now()
	}

	s.metrics.errorRate.Inc()
	log.Printf("Onboarding error for customer %s: %v", customerID, err)
}

func main() {
	service := NewEnterpriseOnboardingService()

	r := mux.NewRouter()
	
	// API endpoints
	r.HandleFunc("/api/onboarding/start", service.StartOnboarding).Methods("POST")
	r.HandleFunc("/api/onboarding/status/{customer_id}", service.GetOnboardingStatus).Methods("GET")
	
	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}).Methods("GET")
	
	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

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
		Addr:         ":8086",
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

		log.Println("Shutting down enterprise onboarding service...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Enterprise Onboarding Service starting on port 8086...")
	log.Printf("Endpoints:")
	log.Printf("  POST /api/onboarding/start - Start customer onboarding")
	log.Printf("  GET  /api/onboarding/status/{customer_id} - Get onboarding status")
	log.Printf("  GET  /health - Health check")
	log.Printf("  GET  /metrics - Prometheus metrics")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}