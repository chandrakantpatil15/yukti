package api

import (
	"database/sql"
	"log"
	"net/http"

	"yukti/internal/api/routes"
	"yukti/internal/config"
	"yukti/internal/models"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	db     *sql.DB
	router *mux.Router
	config *config.Config
}

func NewServer(db *sql.DB) *Server {
	// Use models package to prevent unused import
	_ = models.Resource{}
	
	cfg := config.Load()
	log.Printf("[INFO] Initializing API server...")
	s := &Server{
		db:     db,
		router: mux.NewRouter(),
		config: cfg,
	}
	log.Printf("[INFO] Setting up routes...")
	routes.SetupRoutes(s.router, db)
	log.Printf("[INFO] API server initialized successfully")
	return s
}

func (s *Server) Run(addr string) error {
	log.Printf("[INFO] Configuring CORS for origins: %v", s.config.CORSAllowedOrigins)
	
	allowedHeaders := []string{
		"Authorization",
		"Content-Type",
		"X-Admin-Key",
		"X-Admin-User",
		"X-Tenant-ID",
		"X-Include-Filters",
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   s.config.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   allowedHeaders,
		AllowCredentials: true,
		MaxAge:           300,
	})
	handler := c.Handler(s.router)
	log.Printf("[INFO] Starting HTTP server on %s", addr)
	err := http.ListenAndServe(addr, handler)
	if err != nil {
		log.Printf("[ERROR] Server failed to start: %v", err)
	}
	return err
}
