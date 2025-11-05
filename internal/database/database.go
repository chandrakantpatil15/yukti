package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/cloudcostoptimizer/yukti/internal/models"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.Resource{},
		&models.ResourceCost{},
		&models.OptimizationRecommendation{},
		&models.AWSPricing{},
		&models.ResourceMetrics{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}