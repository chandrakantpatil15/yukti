package routes

import (
	"database/sql"
	"log"
	"net/http"

	"yukti/internal/api/handlers"
	"yukti/internal/api/middleware"
	"yukti/internal/onboarding"
	"yukti/internal/whitelist"

	"github.com/gorilla/mux"
)

func SetupRoutes(router *mux.Router, db *sql.DB) {
	log.Printf("[INFO] Setting up API routes...")
	// Middleware
	log.Printf("[DEBUG] Initializing middleware...")
	authMw := middleware.NewAuthMiddleware(db)
	jwtAuthMw := middleware.NewJWTAuthMiddleware(db)
	adminAuthMw := middleware.NewAdminAuthMiddleware(db)
	tenantIsolation := middleware.NewTenantIsolationMiddleware(db)
	rateLimiter := middleware.NewRateLimiter(100)

	// Handlers
	log.Printf("[DEBUG] Initializing handlers...")
	resourceHandler := handlers.NewResourceHandler(db)
	recommendationHandler := handlers.NewRecommendationHandler(db)
	adminHandler := handlers.NewAdminHandler(db)
	customerHandler := handlers.NewCustomerHandler(db)
	auditHandler := handlers.NewAuditHandler(db)
	authHandler := handlers.NewAuthHandler(db)
	filterHandler := handlers.NewFilterHandler(db)
	billingHandler, err := handlers.NewBillingHandler(db)
	if err != nil {
		log.Printf("[WARN] Failed to initialize billing handler: %v", err)
	}
	// Services + handlers requiring services
	onboardingService := onboarding.NewService(db)
	onboardingHandler := handlers.NewOnboardingHandler(onboardingService)
	whitelistService := whitelist.NewService(db)
	whitelistHandler := handlers.NewWhitelistHandler(whitelistService)
	mlProxyHandler := handlers.NewMLProxyHandler()
	scanHandler := handlers.NewScanHandler()

	// Public routes
	log.Printf("[DEBUG] Registering public routes...")
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Health check from IP: %s", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")

	// Auth routes (public)
	log.Printf("[DEBUG] Registering auth routes...")
	router.HandleFunc("/api/v1/auth/signup", authHandler.Signup).Methods("POST")
	router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/v1/auth/logout", authHandler.Logout).Methods("POST")
	
	// Auth routes (protected - JWT required)
	router.Handle("/api/v1/auth/api-keys",
		jwtAuthMw.RequireAuth(middleware.RequireRole("admin")(http.HandlerFunc(authHandler.CreateAPIKey)))).Methods("POST")

	// Admin routes (protected)
	log.Printf("[DEBUG] Registering admin routes...")
	router.Handle("/api/admin/customers", adminAuthMw.RequireAdmin(http.HandlerFunc(adminHandler.GetCustomers))).Methods("GET")
	router.Handle("/api/admin/metrics", adminAuthMw.RequireAdmin(http.HandlerFunc(adminHandler.GetMetrics))).Methods("GET")
	router.Handle("/api/admin/impersonate", adminAuthMw.RequireAdmin(http.HandlerFunc(adminHandler.ImpersonateTenant))).Methods("POST")
	router.Handle("/api/admin/audit-logs", adminAuthMw.RequireAdmin(http.HandlerFunc(auditHandler.GetAuditLogs))).Methods("GET")

	// Customer routes
	log.Printf("[DEBUG] Registering customer routes...")
	router.HandleFunc("/api/customers", customerHandler.CreateCustomer).Methods("POST")
	router.Handle("/api/customers/dashboard", tenantIsolation.ValidateTenant(http.HandlerFunc(customerHandler.GetDashboard))).Methods("GET")
	router.Handle("/api/customers/findings", tenantIsolation.ValidateTenant(http.HandlerFunc(customerHandler.GetFindings))).Methods("GET")

	// Onboarding routes
	log.Printf("[DEBUG] Registering onboarding routes...")
	router.HandleFunc("/api/onboarding/customer", onboardingHandler.CreateCustomer).Methods("POST")
	router.HandleFunc("/api/onboarding/aws", onboardingHandler.ConfigureAWS).Methods("POST")
	router.HandleFunc("/api/onboarding/metrics", onboardingHandler.ConfigureMetrics).Methods("POST")
	router.HandleFunc("/api/onboarding/status", onboardingHandler.GetStatus).Methods("GET")

	// Protected routes
	log.Printf("[DEBUG] Registering protected routes...")
	router.Handle("/api/v1/resources",
		rateLimiter.Limit(authMw.TenantAuth(http.HandlerFunc(resourceHandler.ListResources)))).Methods("GET")

	router.Handle("/api/v1/resources/stats",
		rateLimiter.Limit(authMw.TenantAuth(http.HandlerFunc(resourceHandler.GetResourceStats)))).Methods("GET")

	router.Handle("/api/v1/recommendations",
		rateLimiter.Limit(authMw.TenantAuth(http.HandlerFunc(recommendationHandler.ListRecommendations)))).Methods("GET")

	// ML proxy routes
	log.Printf("[DEBUG] Registering ML proxy routes...")
	router.Handle("/api/v1/ml/anomaly-detect", rateLimiter.Limit(http.HandlerFunc(mlProxyHandler.AnomalyDetect))).Methods("POST")
	router.Handle("/api/v1/ml/forecast", rateLimiter.Limit(http.HandlerFunc(mlProxyHandler.Forecast))).Methods("POST")

	// Scan orchestration (stub)
	log.Printf("[DEBUG] Registering scan routes...")
	router.HandleFunc("/api/scan", scanHandler.TriggerScan).Methods("POST")

	// Whitelist routes (tenant-scoped)
	log.Printf("[DEBUG] Registering whitelist routes...")
	router.Handle("/api/whitelists", tenantIsolation.ValidateTenant(http.HandlerFunc(whitelistHandler.ListWhitelists))).Methods("GET")
	router.Handle("/api/whitelists", tenantIsolation.ValidateTenant(http.HandlerFunc(whitelistHandler.CreateWhitelist))).Methods("POST")
	router.Handle("/api/whitelists", tenantIsolation.ValidateTenant(http.HandlerFunc(whitelistHandler.RevokeWhitelist))).Methods("DELETE")
	router.Handle("/api/whitelists/{id}", tenantIsolation.ValidateTenant(http.HandlerFunc(whitelistHandler.RevokeWhitelist))).Methods("DELETE")

	// Filter routes (JWT protected, data-driven for UI)
	log.Printf("[DEBUG] Registering filter routes...")
	router.Handle("/api/v1/filters/resource-types",
		jwtAuthMw.RequireAuth(http.HandlerFunc(filterHandler.GetResourceTypes))).Methods("GET")
	router.Handle("/api/v1/filters/tags",
		jwtAuthMw.RequireAuth(http.HandlerFunc(filterHandler.GetTags))).Methods("GET")
	router.Handle("/api/v1/filters/services",
		jwtAuthMw.RequireAuth(http.HandlerFunc(filterHandler.GetServices))).Methods("GET")
	router.Handle("/api/v1/filters/accounts",
		jwtAuthMw.RequireAuth(http.HandlerFunc(filterHandler.GetAccounts))).Methods("GET")
	router.Handle("/api/v1/filters/regions",
		jwtAuthMw.RequireAuth(http.HandlerFunc(filterHandler.GetRegions))).Methods("GET")

	// Billing routes (protected - JWT required)
	if billingHandler != nil {
		log.Printf("[DEBUG] Registering billing routes...")
		router.Handle("/api/v1/billing/checkout-session",
			jwtAuthMw.RequireAuth(http.HandlerFunc(billingHandler.CreateCheckoutSession))).Methods("POST")
		router.Handle("/api/v1/billing/info",
			jwtAuthMw.RequireAuth(http.HandlerFunc(billingHandler.GetBillingInfo))).Methods("GET")
		// Webhook route (no auth - Stripe signature verification)
		router.HandleFunc("/api/v1/webhooks/stripe", billingHandler.HandleStripeWebhook).Methods("POST")
	}

	log.Printf("[INFO] All routes registered successfully")
}
