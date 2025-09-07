package database

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

// Config holds database configuration
type Config struct {
	Host      string
	Port      int
	User      string
	Password  string
	DBName    string
	SSLMode   string
	UseSQLite bool
}

// NewConfig creates a new database configuration from environment variables
func NewConfig() *Config {
	port, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	useSQLite := getEnv("USE_SQLITE", "false") == "true" // Default to false to avoid CGO issues

	return &Config{
		Host:      getEnv("DB_HOST", "localhost"),
		Port:      port,
		User:      getEnv("DB_USER", "postgres"),
		Password:  getEnv("DB_PASSWORD", "password"),
		DBName:    getEnv("DB_NAME", "agent_payments"),
		SSLMode:   getEnv("DB_SSLMODE", "disable"),
		UseSQLite: useSQLite,
	}
}

// DSN returns the database connection string
func (c *Config) DSN() string {
	if c.UseSQLite {
		return fmt.Sprintf("%s.db", c.DBName)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Connect establishes a database connection
func Connect(config *Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	if config.UseSQLite {
		db, err = gorm.Open(sqlite.Open(config.DSN()), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		db, err = gorm.Open(postgres.Open(config.DSN()), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool (skip for SQLite)
	if !config.UseSQLite {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get sql.DB: %w", err)
		}

		// Set connection pool settings
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	log.Println("Database connection established successfully")
	return db, nil
}

// HealthCheck performs a database health check
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
