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

	// Enable strict slash matching
	router.StrictSlash(true)
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
	authHandler := handlers.NewAuthHandler(db) // Use main auth handler
	filterHandler := handlers.NewFilterHandler(db)
	teamHandler := handlers.NewTeamHandler(db)
	// Sync status handler
	syncHandler := handlers.NewSyncHandler(db)
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
	scanHandler := handlers.NewScanHandler(db)

	// Public routes
	log.Printf("[DEBUG] Registering public routes...")
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] Health check from IP: %s", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")

	// Auth routes (public)
	log.Printf("[DEBUG] Registering auth routes...")
	router.HandleFunc("/api/v1/auth/signup", authHandler.Signup).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/logout", authHandler.Logout).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/verify-email", authHandler.VerifyEmail).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/resend-code", authHandler.ResendOTP).Methods("POST", "OPTIONS")
	log.Printf("[DEBUG] Auth routes registered: /api/v1/auth/signup, /api/v1/auth/login, /api/v1/auth/logout, /api/v1/auth/verify-email, /api/v1/auth/resend-code")

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
	// Require JWT auth so tenant info is available in context for customer handlers
	router.Handle("/api/customers/dashboard", jwtAuthMw.RequireAuth(tenantIsolation.ValidateTenant(http.HandlerFunc(customerHandler.GetDashboard)))).Methods("GET")
	router.Handle("/api/customers/findings", jwtAuthMw.RequireAuth(tenantIsolation.ValidateTenant(http.HandlerFunc(customerHandler.GetFindings)))).Methods("GET")

	// Onboarding routes
	log.Printf("[DEBUG] Registering onboarding routes...")
	router.HandleFunc("/api/onboarding/customer", onboardingHandler.CreateCustomer).Methods("POST")
	// Allow GET to retrieve AWS connection (handler checks method) and POST to configure
	router.HandleFunc("/api/onboarding/aws", onboardingHandler.ConfigureAWS).Methods("GET", "POST", "OPTIONS")
	router.HandleFunc("/api/onboarding/metrics", onboardingHandler.ConfigureMetrics).Methods("POST")
	router.HandleFunc("/api/onboarding/status", onboardingHandler.GetStatus).Methods("GET")

	// Protected routes
	log.Printf("[DEBUG] Registering protected routes...")
	// Backwards-compatible public resources endpoint (accepts X-Tenant-ID header)
	router.HandleFunc("/api/resources", resourceHandler.ListResources).Methods("GET")
	router.Handle("/api/v1/resources",
		jwtAuthMw.RequireAuth(http.HandlerFunc(resourceHandler.ListResources))).Methods("GET")

	// Resource details endpoint
	router.Handle("/api/v1/resources/{resourceId}",
		jwtAuthMw.RequireAuth(http.HandlerFunc(resourceHandler.GetResourceDetails))).Methods("GET")

	router.Handle("/api/v1/resources/stats",
		jwtAuthMw.RequireAuth(http.HandlerFunc(resourceHandler.GetResourceStats))).Methods("GET")

	router.Handle("/api/v1/recommendations",
		rateLimiter.Limit(authMw.TenantAuth(http.HandlerFunc(recommendationHandler.ListRecommendations)))).Methods("GET")

	// ML proxy routes
	log.Printf("[DEBUG] Registering ML proxy routes...")
	router.Handle("/api/v1/ml/anomaly-detect", rateLimiter.Limit(http.HandlerFunc(mlProxyHandler.AnomalyDetect))).Methods("POST")
	router.Handle("/api/v1/ml/forecast", rateLimiter.Limit(http.HandlerFunc(mlProxyHandler.Forecast))).Methods("POST")

	// Sync status (available to authenticated users)
	router.Handle("/api/internal/sync/status", jwtAuthMw.RequireAuth(http.HandlerFunc(syncHandler.GetStatus))).Methods("GET")

	// Scan orchestration (stub)
	log.Printf("[DEBUG] Registering scan routes...")
	// Protect scan routes with JWT auth so tenant ID is available in context
	router.Handle("/api/scan", jwtAuthMw.RequireAuth(http.HandlerFunc(scanHandler.TriggerScan))).Methods("POST")
	router.Handle("/api/scan/status", jwtAuthMw.RequireAuth(http.HandlerFunc(scanHandler.GetScanStatus))).Methods("GET")

	// Whitelist routes (JWT protected)
	log.Printf("[DEBUG] Registering whitelist routes...")
	router.Handle("/api/v1/whitelists", jwtAuthMw.RequireAuth(http.HandlerFunc(whitelistHandler.ListWhitelists))).Methods("GET")
	router.Handle("/api/v1/whitelists", jwtAuthMw.RequireAuth(http.HandlerFunc(whitelistHandler.CreateWhitelist))).Methods("POST")
	router.Handle("/api/v1/whitelists/{id}", jwtAuthMw.RequireAuth(http.HandlerFunc(whitelistHandler.RevokeWhitelist))).Methods("DELETE")

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

	// Team management routes (JWT protected)
	log.Printf("[DEBUG] Registering team routes...")
	router.Handle("/api/v1/team/members",
		jwtAuthMw.RequireAuth(http.HandlerFunc(teamHandler.ListMembers))).Methods("GET")
	router.Handle("/api/v1/team/invitations",
		jwtAuthMw.RequireAuth(http.HandlerFunc(teamHandler.ListInvitations))).Methods("GET")
	router.Handle("/api/v1/team/invite",
		jwtAuthMw.RequireAuth(middleware.RequireRole("admin", "owner")(http.HandlerFunc(teamHandler.InviteUser)))).Methods("POST")
	router.Handle("/api/v1/team/accept-invite",
		jwtAuthMw.RequireAuth(http.HandlerFunc(teamHandler.AcceptInvite))).Methods("POST")
	router.HandleFunc("/api/v1/team/invite-details", teamHandler.GetInviteDetails).Methods("GET") // Public endpoint
	router.Handle("/api/v1/team/members/{id}/role",
		jwtAuthMw.RequireAuth(middleware.RequireRole("admin", "owner")(http.HandlerFunc(teamHandler.UpdateRole)))).Methods("PUT")
	router.Handle("/api/v1/team/members/{id}",
		jwtAuthMw.RequireAuth(middleware.RequireRole("admin", "owner")(http.HandlerFunc(teamHandler.RemoveUser)))).Methods("DELETE")
	router.Handle("/api/v1/team/invitations/{id}/resend",
		jwtAuthMw.RequireAuth(middleware.RequireRole("admin", "owner")(http.HandlerFunc(teamHandler.ResendInvite)))).Methods("POST")
	router.Handle("/api/v1/team/invitations/{id}",
		jwtAuthMw.RequireAuth(middleware.RequireRole("admin", "owner")(http.HandlerFunc(teamHandler.RevokeInvitation)))).Methods("DELETE")

	// IaC generation routes
	log.Printf("[DEBUG] Registering IaC routes...")
	iacHandler := handlers.NewIaCHandler("us-east-1")
	router.Handle("/api/v1/iac/generate",
		jwtAuthMw.RequireAuth(http.HandlerFunc(iacHandler.GenerateIaC))).Methods("POST")
	router.Handle("/api/v1/iac/bulk-generate",
		jwtAuthMw.RequireAuth(http.HandlerFunc(iacHandler.BulkGenerate))).Methods("POST")

	// Billing routes (protected - JWT required)
	if billingHandler != nil {
		log.Printf("[DEBUG] Registering billing routes...")
		// Admin billing management
		router.Handle("/api/admin/billing",
			adminAuthMw.RequireAdmin(http.HandlerFunc(billingHandler.ListBillings))).Methods("GET")
		router.Handle("/api/admin/billing",
			adminAuthMw.RequireAdmin(http.HandlerFunc(billingHandler.CreateBilling))).Methods("POST")
		router.Handle("/api/admin/billing/stats",
			adminAuthMw.RequireAdmin(http.HandlerFunc(billingHandler.GetBillingStats))).Methods("GET")
		router.Handle("/api/admin/billing/export",
			adminAuthMw.RequireAdmin(http.HandlerFunc(billingHandler.ExportBillings))).Methods("GET")
		router.Handle("/api/admin/billing/{id}",
			adminAuthMw.RequireAdmin(http.HandlerFunc(billingHandler.GetBilling))).Methods("GET")
		router.Handle("/api/admin/billing/{id}",
			adminAuthMw.RequireAdmin(http.HandlerFunc(billingHandler.UpdateBilling))).Methods("PUT")
		router.Handle("/api/admin/billing/{id}",
			adminAuthMw.RequireAdmin(http.HandlerFunc(billingHandler.DeleteBilling))).Methods("DELETE")
		router.Handle("/api/admin/billing/{id}/mark-paid",
			adminAuthMw.RequireAdmin(http.HandlerFunc(billingHandler.MarkAsPaid))).Methods("POST")

		// Tenant-scoped billing info for frontend
		router.Handle("/api/v1/billing/info",
			tenantIsolation.ValidateTenant(http.HandlerFunc(billingHandler.GetBillingInfo))).Methods("GET")
	}

	log.Printf("[INFO] All routes registered successfully")
}
