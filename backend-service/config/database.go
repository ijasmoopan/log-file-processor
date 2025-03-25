package config

import (
	"fmt"
	"log"
	"os"

	"github.com/ijasmoopan/intucloud-task/backend-service/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBConfig holds database configuration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDBConfig creates a new database configuration from environment variables
func NewDBConfig() *DBConfig {
	return &DBConfig{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("POSTGRES_SSLMODE"),
	}
}

// GetDSN returns the database connection string
func (c *DBConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// InitDB initializes the database connection and performs migrations
func InitDB() (*gorm.DB, error) {
	dbConfig := NewDBConfig()

	db, err := gorm.Open(postgres.Open(dbConfig.GetDSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Auto-migrate the models
	if err := db.AutoMigrate(&models.FileResult{}); err != nil {
		return nil, fmt.Errorf("failed to perform database migration: %v", err)
	}

	log.Println("Successfully connected to database and performed migrations")
	return db, nil
}
