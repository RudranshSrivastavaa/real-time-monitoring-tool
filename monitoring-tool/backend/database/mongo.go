// database/mongo.go
package database

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"monitoring-tool/config"
)

// MongoDB holds the database connection
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// InitMongoDB initializes MongoDB connection with configuration
// Update InitMongoDB for production
func InitMongoDB(cfg *config.Config) (*MongoDB, error) {
	clientOptions := options.Client().
		ApplyURI(cfg.MongodbURI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second).
		SetServerSelectionTimeout(30 * time.Second). // Increased for Atlas
		SetConnectTimeout(30 * time.Second).         // Increased for Atlas
		SetSocketTimeout(30 * time.Second)           // Added for Atlas

	// Add retry writes and reads for production and Atlas
	if cfg.IsProduction() || isAtlasURI(cfg.MongodbURI) {
		clientOptions.SetRetryWrites(true).SetRetryReads(true)
		// Atlas-specific optimizations
		clientOptions.SetHeartbeatInterval(10 * time.Second)
		clientOptions.SetLocalThreshold(15 * time.Second)
	}

	// Increase timeout for Atlas connections
	timeout := 30 * time.Second
	if isAtlasURI(cfg.MongodbURI) {
		timeout = 60 * time.Second // Longer timeout for Atlas
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Test the connection with retries for Atlas
	maxRetries := 3
	if isAtlasURI(cfg.MongodbURI) {
		maxRetries = 5 // More retries for Atlas
	}

	for i := 0; i < maxRetries; i++ {
		err = client.Ping(ctx, nil)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			log.Printf("MongoDB ping attempt %d failed, retrying in 5 seconds...", i+1)
			time.Sleep(5 * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB after %d attempts: %v", maxRetries, err)
	}

	database := client.Database(cfg.DatabaseName)

	// Create indexes
	if err := createIndexes(database); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	// Mask URI for production logging
	maskedURI := cfg.MongodbURI
	if cfg.IsProduction() {
		maskedURI = "mongodb://***:***@[host]/[database]"
	}
	log.Printf("✅ Connected to MongoDB: %s (database: %s)", maskedURI, cfg.DatabaseName)

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

// createIndexes creates database indexes for optimal query performance
func createIndexes(db *mongo.Database) error {
	ctx := context.Background()

	// Index for monitors collection
	monitorsCollection := db.Collection("monitors")
	_, err := monitorsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    map[string]int{"url": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]int{"is_active": 1},
		},
		{
			Keys: map[string]int{"status": 1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create monitors indexes: %v", err)
	}

	// Index for metrics collection (time-series optimization)
	metricsCollection := db.Collection("metrics")
	_, err = metricsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: map[string]int{"monitor_id": 1, "checked_at": -1},
		},
		{
			Keys: map[string]int{"checked_at": -1},
		},
		{
			Keys: map[string]int{"monitor_id": 1},
		},
		{
			Keys:    map[string]int{"checked_at": 1},
			Options: options.Index().SetExpireAfterSeconds(30 * 24 * 60 * 60), // 30 days retention
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create metrics indexes: %v", err)
	}

	return nil
}

// Disconnect closes the MongoDB connection
func (m *MongoDB) Disconnect(ctx context.Context) error {
	if m.Client != nil {
		return m.Client.Disconnect(ctx)
	}
	return nil
}

// GetCollection returns a MongoDB collection
func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}

// Collections constants
const (
	MonitorsCollection = "monitors"
	MetricsCollection  = "metrics"
)

// Health checks database connection
func (m *MongoDB) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.Client.Ping(ctx, nil)
}

// isAtlasURI checks if the URI is for MongoDB Atlas
func isAtlasURI(uri string) bool {
	return strings.Contains(uri, "mongodb.net") || strings.Contains(uri, "mongodb+srv://")
}
