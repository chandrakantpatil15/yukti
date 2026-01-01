package database

import (
	"database/sql"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"yukti/internal/models"
)

func Connect(databaseURL string) (*sql.DB, error) {
	log.Printf("[INFO] Connecting to database...")
	gormDB, err := connectGorm(databaseURL)
	if err != nil {
		log.Printf("[ERROR] Failed to connect to database: %v", err)
		return nil, err
	}
	log.Printf("[INFO] Database connection established successfully")
	return gormDB.DB()
}

func connectGorm(databaseURL string) (*gorm.DB, error) {
	log.Printf("[DEBUG] Opening GORM connection to PostgreSQL")
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Printf("[ERROR] GORM connection failed: %v", err)
		return nil, err
	}

	// Auto-migrate models
	log.Printf("[INFO] Running database auto-migration...")
	err = db.AutoMigrate(
		&models.Resource{},
		&models.ResourceCost{},
		&models.OptimizationRecommendation{},
		&models.AWSPricing{},
		&models.ResourceMetrics{},
		&models.User{},
	)
	if err != nil {
		log.Printf("[ERROR] Database migration failed: %v", err)
		return nil, err
	}
	log.Printf("[INFO] Database migration completed successfully")

	return db, nil
}