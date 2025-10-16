// database/mongo.go
package database

import (
	"context"
	"fmt"
	"log"
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
func InitMongoDB(cfg *config.Config) (*MongoDB, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(cfg.MongodbURI)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	database := client.Database(cfg.DatabaseName)

	// Create indexes for better performance
	if err := createIndexes(database); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	log.Printf("✅ Connected to MongoDB: %s (database: %s)", cfg.MongodbURI, cfg.DatabaseName)

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