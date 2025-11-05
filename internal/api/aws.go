package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/cloudcostoptimizer/yukti/internal/services"
	"github.com/cloudcostoptimizer/yukti/internal/models"
)

func (s *Server) setupAWSRoutes() {
	aws := s.router.Group("/api/v1/aws")
	{
		aws.POST("/sync/resources", s.syncAWSResources)
		aws.POST("/sync/costs", s.syncAWSCosts)
		aws.POST("/sync/pricing", s.syncPricing)
		aws.GET("/pricing/:instanceType", s.getPricing)
	}
}

func (s *Server) syncAWSResources(c *gin.Context) {
	awsService := services.NewSimpleAWSService(s.db)

	if err := awsService.SyncEC2Resources(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "AWS resources synced successfully"})
}

func (s *Server) syncAWSCosts(c *gin.Context) {
	awsService := services.NewSimpleAWSService(s.db)

	if err := awsService.SyncCostData(30); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "AWS cost data synced successfully"})
}

func (s *Server) syncPricing(c *gin.Context) {
	pricingService := services.NewPricingService(s.db)
	
	if err := pricingService.UpdatePricingData(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := pricingService.GenerateOptimizationRecommendations(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pricing data updated and recommendations generated"})
}

func (s *Server) getPricing(c *gin.Context) {
	instanceType := c.Param("instanceType")
	
	var pricing []models.AWSPricing
	if err := s.db.Where("instance_type = ?", instanceType).Find(&pricing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pricing)
}