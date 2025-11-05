package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/cloudcostoptimizer/yukti/internal/services"
	"github.com/cloudcostoptimizer/yukti/internal/models"
)

type Server struct {
	router           *gin.Engine
	finopsService    *services.FinOpsService
	analyticsService *services.AnalyticsService
	batchProcessor   *services.BatchProcessor
	db               *gorm.DB
}

func NewServer(db *gorm.DB) *Server {
	router := gin.Default()
	finopsService := services.NewFinOpsService(db)

	server := &Server{
		router:           router,
		finopsService:    finopsService,
		analyticsService: services.NewAnalyticsService(db),
		batchProcessor:   services.NewBatchProcessor(db),
		db:               db,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")
	{
		api.GET("/health", s.healthCheck)
		api.GET("/cost/summary", s.getCostSummary)
		api.GET("/recommendations", s.getRecommendations)
		api.GET("/resources", s.getResources)
		api.GET("/costs", s.getCosts)
		api.GET("/pricing", s.getAllPricing)
		api.GET("/analytics/cost-by-type", s.getCostByResourceType)
		api.GET("/analytics/cost-trend", s.getCostTrend)
		api.GET("/analytics/utilization", s.getUtilizationMetrics)
		api.POST("/batch/process-costs", s.processCostsBatch)
		api.POST("/batch/process-metrics", s.processMetricsBatch)
		api.POST("/batch/generate-recommendations", s.generateRecommendationsBatch)
	}
	
	// AWS-specific routes
	s.setupAWSRoutes()
	
	// Static files for web interface
	s.router.Static("/static", "./web/static")
	s.router.LoadHTMLGlob("web/templates/*")
	s.router.GET("/", s.dashboard)
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "yukti-finops",
	})
}

func (s *Server) getCostSummary(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter"})
		return
	}

	summary, err := s.finopsService.GetCostSummary(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (s *Server) getRecommendations(c *gin.Context) {
	var recommendations []models.OptimizationRecommendation
	if err := s.db.Select("recommendation_type, current_cost, optimized_cost, potential_savings, confidence").Where("status = ?", "active").Order("potential_savings DESC").Limit(10).Find(&recommendations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, recommendations)
}

func (s *Server) getResources(c *gin.Context) {
	var resources []models.Resource
	if err := s.db.Select("id, resource_id, resource_type, instance_type, region, status, environment").Find(&resources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resources)
}

func (s *Server) getCosts(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	
	var costs []models.ResourceCost
	query := s.db.Select("id, resource_id, date, cost_usd, usage_hours, data_source")
	
	if days > 0 {
		startDate := time.Now().AddDate(0, 0, -days)
		query = query.Where("date >= ?", startDate)
	}
	
	if err := query.Limit(100).Order("date DESC").Find(&costs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, costs)
}

func (s *Server) getAllPricing(c *gin.Context) {
	var pricing []models.AWSPricing
	if err := s.db.Find(&pricing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pricing)
}

func (s *Server) dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Yukti FinOps Dashboard",
	})
}

func (s *Server) getCostByResourceType(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	
	results, err := s.analyticsService.GetCostByResourceType(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (s *Server) getCostTrend(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	
	results, err := s.analyticsService.GetCostTrend(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (s *Server) getUtilizationMetrics(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, _ := strconv.Atoi(daysStr)
	
	results, err := s.analyticsService.GetUtilizationMetrics(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (s *Server) processCostsBatch(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	
	ctx := c.Request.Context()
	result, err := s.batchProcessor.ProcessCostData(ctx, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) processMetricsBatch(c *gin.Context) {
	hoursStr := c.DefaultQuery("hours", "168") // 7 days
	hours, _ := strconv.Atoi(hoursStr)
	
	ctx := c.Request.Context()
	result, err := s.batchProcessor.ProcessMetricsData(ctx, hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) generateRecommendationsBatch(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := s.batchProcessor.GenerateRecommendations(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}