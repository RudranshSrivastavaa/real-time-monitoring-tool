// config/config.go
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port        string
	Host        string
	Environment string

	// Database configuration
	MongodbURI    string
	DatabaseName  string
	
	// WebSocket configuration
	WSReadTimeout  time.Duration
	WSWriteTimeout time.Duration
	WSPingPeriod   time.Duration
	
	// Monitoring configuration
	DefaultInterval       int    // seconds
	DefaultTimeout        int    // seconds
	MaxConcurrentChecks   int
	MetricsRetentionDays  int
	
	// CORS configuration
	AllowedOrigins []string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		// Server
		Port:        getEnvOrDefault("PORT", "8080"),
		Host:        getEnvOrDefault("HOST", "0.0.0.0"),
		Environment: getEnvOrDefault("GIN_MODE", "debug"),

		// Database
		MongodbURI:   getEnvOrDefault("MONGODB_URI", "mongodb://localhost:27017"),
		DatabaseName: getEnvOrDefault("DATABASE_NAME", "realtime_monitor"),

		// WebSocket
		WSReadTimeout:  time.Duration(getEnvAsInt("WS_READ_TIMEOUT", 60)) * time.Second,
		WSWriteTimeout: time.Duration(getEnvAsInt("WS_WRITE_TIMEOUT", 10)) * time.Second,
		WSPingPeriod:   time.Duration(getEnvAsInt("WS_PING_PERIOD", 54)) * time.Second,

		// Monitoring
		DefaultInterval:      getEnvAsInt("DEFAULT_INTERVAL", 30),
		DefaultTimeout:       getEnvAsInt("DEFAULT_TIMEOUT", 10),
		MaxConcurrentChecks:  getEnvAsInt("MAX_CONCURRENT_CHECKS", 100),
		MetricsRetentionDays: getEnvAsInt("METRICS_RETENTION_DAYS", 30),

		// CORS
		AllowedOrigins: []string{
			getEnvOrDefault("FRONTEND_URL", "http://localhost:3000"),
			"http://localhost:3001", // Alternative frontend port
		},
	}
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return c.Host + ":" + c.Port
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "debug" || c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "release" || c.Environment == "production"
}

// GetMetricsRetentionDuration returns retention duration for metrics
func (c *Config) GetMetricsRetentionDuration() time.Duration {
	return time.Duration(c.MetricsRetentionDays) * 24 * time.Hour
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.MongodbURI == "" {
		return fmt.Errorf("MONGODB_URI is required")
	}
	
	if c.DatabaseName == "" {
		return fmt.Errorf("DATABASE_NAME is required")
	}
	
	if c.DefaultInterval < 5 {
		log.Printf("Warning: DEFAULT_INTERVAL is very low (%ds), this may cause high load", c.DefaultInterval)
	}
	
	if c.DefaultTimeout >= c.DefaultInterval {
		log.Printf("Warning: DEFAULT_TIMEOUT (%ds) should be less than DEFAULT_INTERVAL (%ds)", c.DefaultTimeout, c.DefaultInterval)
	}
	
	return nil
}

// LogConfig logs the current configuration (without sensitive data)
func (c *Config) LogConfig() {
	log.Printf("📋 Configuration loaded:")
	log.Printf("   Server: %s (mode: %s)", c.GetServerAddress(), c.Environment)
	log.Printf("   Database: %s", c.DatabaseName)
	log.Printf("   Default monitoring interval: %ds", c.DefaultInterval)
	log.Printf("   Default timeout: %ds", c.DefaultTimeout)
	log.Printf("   Max concurrent checks: %d", c.MaxConcurrentChecks)
	log.Printf("   Metrics retention: %d days", c.MetricsRetentionDays)
	log.Printf("   Allowed origins: %v", c.AllowedOrigins)
}

// Helper functions

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt returns environment variable as integer or default value
func getEnvAsInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		} else {
			log.Printf("Warning: Invalid integer value for %s: %s, using default %d", key, valueStr, defaultValue)
		}
	}
	return defaultValue
}

// getEnvAsBool returns environment variable as boolean or default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		} else {
			log.Printf("Warning: Invalid boolean value for %s: %s, using default %t", key, valueStr, defaultValue)
		}
	}
	return defaultValue
}

// Development helpers

// GetTestConfig returns configuration for testing
func GetTestConfig() *Config {
	return &Config{
		Port:                "8081",
		Host:                "localhost",
		Environment:         "test",
		MongodbURI:         "mongodb://localhost:27017",
		DatabaseName:       "realtime_monitor_test",
		WSReadTimeout:      30 * time.Second,
		WSWriteTimeout:     5 * time.Second,
		WSPingPeriod:       25 * time.Second,
		DefaultInterval:    10,
		DefaultTimeout:     5,
		MaxConcurrentChecks: 50,
		MetricsRetentionDays: 7,
		AllowedOrigins:     []string{"http://localhost:3000"},
	}
}