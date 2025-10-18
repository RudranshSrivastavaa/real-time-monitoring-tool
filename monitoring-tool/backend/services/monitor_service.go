// services/monitor_service.go
package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	//"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"monitoring-tool/database"
	"monitoring-tool/models"
)

type MonitorService struct {
	db          *database.MongoDB
	httpClient  *http.Client
	activeJobs  map[string]chan bool // for stopping individual monitor jobs
	jobsMutex   sync.RWMutex
	rateLimiter *time.Ticker
    semaphore   chan struct{}
}

// NewMonitorService creates a new monitor service
func NewMonitorService(db *database.MongoDB, maxConcurrent int) *MonitorService {
    return &MonitorService{
        db: db,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
                DisableKeepAlives:   false,
            },
        },
        activeJobs:  make(map[string]chan bool),
        semaphore:   make(chan struct{}, maxConcurrent),
    }
}

// CreateMonitor adds a new monitor to the database
func (ms *MonitorService) CreateMonitor(monitor *models.Monitor) error {
	collection := ms.db.GetCollection(database.MonitorsCollection)
	
	// Check if URL already exists
	filter := bson.M{"url": monitor.URL}
	count, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("monitor with URL %s already exists", monitor.URL)
	}

	// Insert the monitor
	result, err := collection.InsertOne(context.Background(), monitor)
	if err != nil {
		return err
	}

	monitor.ID = result.InsertedID.(primitive.ObjectID)
	log.Printf("✅ Created monitor: %s (%s)", monitor.Name, monitor.URL)
	
	return nil
}

// GetMonitors retrieves all monitors
func (ms *MonitorService) GetMonitors() ([]models.Monitor, error) {
	collection := ms.db.GetCollection(database.MonitorsCollection)
	
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var monitors []models.Monitor
	if err = cursor.All(context.Background(), &monitors); err != nil {
		return nil, err
	}

	return monitors, nil
}

// DeleteMonitor removes a monitor and stops its monitoring job
func (ms *MonitorService) DeleteMonitor(id primitive.ObjectID) error {
	// Stop the monitoring job first
	ms.stopMonitorJob(id.Hex())

	// Delete from database
	collection := ms.db.GetCollection(database.MonitorsCollection)
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("monitor not found")
	}

	log.Printf("🗑️  Deleted monitor: %s", id.Hex())
	return nil
}

// GetMetrics retrieves metrics for a specific monitor
func (ms *MonitorService) GetMetrics(monitorID primitive.ObjectID, hours int) ([]models.Metric, error) {
	collection := ms.db.GetCollection(database.MetricsCollection)
	
	// Query last N hours of data
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	filter := bson.M{
		"monitor_id": monitorID,
		"checked_at": bson.M{"$gte": startTime},
	}

	// Sort by checked_at descending, limit to 1000 records
	opts := options.Find().SetSort(bson.M{"checked_at": -1}).SetLimit(1000)
	
	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var metrics []models.Metric
	if err = cursor.All(context.Background(), &metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

// StartMonitoring starts monitoring all active monitors
func (ms *MonitorService) StartMonitoring(wsHub *WebSocketHub) {
	log.Println("🔄 Starting monitoring service...")

	// Start monitoring existing monitors
	monitors, err := ms.GetMonitors()
	if err != nil {
		log.Printf("Error loading monitors: %v", err)
		return
	}

	for _, monitor := range monitors {
		if monitor.IsActive {
			ms.startMonitorJob(monitor, wsHub)
		}
	}

	log.Printf("✅ Started monitoring %d active endpoints", len(monitors))
}

// startMonitorJob starts a monitoring job for a specific monitor
func (ms *MonitorService) startMonitorJob(monitor models.Monitor, wsHub *WebSocketHub) {
	monitorID := monitor.ID.Hex()
	
	ms.jobsMutex.Lock()
	// Stop existing job if running
	if stopChan, exists := ms.activeJobs[monitorID]; exists {
		close(stopChan)
	}
	
	// Create new stop channel
	stopChan := make(chan bool)
	ms.activeJobs[monitorID] = stopChan
	ms.jobsMutex.Unlock()

	log.Printf("🚀 Started monitoring job: %s (%s)", monitor.Name, monitor.URL)

	// Start monitoring goroutine
	go func() {
		ticker := time.NewTicker(time.Duration(monitor.Interval) * time.Second)
		defer ticker.Stop()

		// Run initial check immediately
		ms.checkEndpoint(monitor, wsHub)

		for {
			select {
			case <-ticker.C:
				ms.checkEndpoint(monitor, wsHub)
			case <-stopChan:
				log.Printf("🛑 Stopped monitoring job: %s", monitor.Name)
				return
			}
		}
	}()
}

// stopMonitorJob stops a monitoring job
func (ms *MonitorService) stopMonitorJob(monitorID string) {
	ms.jobsMutex.Lock()
	defer ms.jobsMutex.Unlock()

	if stopChan, exists := ms.activeJobs[monitorID]; exists {
		close(stopChan)
		delete(ms.activeJobs, monitorID)
	}
}

// checkEndpoint performs a health check on an endpoint
func (ms *MonitorService) checkEndpoint(monitor models.Monitor, wsHub *WebSocketHub) {

	 ms.semaphore <- struct{}{}
    defer func() { <-ms.semaphore }()
	startTime := time.Now()

	// Create HTTP request
	req, err := http.NewRequest(monitor.Method, monitor.URL, nil)
	if err != nil {
		ms.recordMetric(monitor, "down", 0, 0, err.Error(), wsHub)
		return
	}

	// Set timeout for this specific request
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(monitor.Timeout)*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Set user agent
	req.Header.Set("User-Agent", "RealtimeMonitor/1.0")

	// Perform the request
	resp, err := ms.httpClient.Do(req)
	responseTime := time.Since(startTime).Milliseconds()

	if err != nil {
		ms.recordMetric(monitor, "down", 0, responseTime, err.Error(), wsHub)
		return
	}
	defer resp.Body.Close()

	// Determine status based on HTTP status code
	status := "up"
	if resp.StatusCode >= 400 {
		status = "down"
	}

	ms.recordMetric(monitor, status, resp.StatusCode, responseTime, "", wsHub)
}

// recordMetric saves a metric to the database and broadcasts via WebSocket
func (ms *MonitorService) recordMetric(monitor models.Monitor, status string, statusCode int, responseTime int64, errorMsg string, wsHub *WebSocketHub) {
	now := time.Now()

	// Create metric record
	metric := models.Metric{
		MonitorID:    monitor.ID,
		URL:          monitor.URL,
		Status:       status,
		StatusCode:   statusCode,
		ResponseTime: responseTime,
		Error:        errorMsg,
		CheckedAt:    now,
	}

	// Save to database
	collection := ms.db.GetCollection(database.MetricsCollection)
	_, err := collection.InsertOne(context.Background(), metric)
	if err != nil {
		log.Printf("Error saving metric: %v", err)
	}

	// Update monitor's current status
	ms.updateMonitorStatus(monitor.ID, status, statusCode, responseTime, now)

	// Calculate uptime percentage
	uptimePercentage := ms.calculateUptimePercentage(monitor.ID)

	// Broadcast via WebSocket
	update := models.MonitorUpdate{
		MonitorID:        monitor.ID.Hex(),
		Status:           status,
		StatusCode:       statusCode,
		ResponseTime:     responseTime,
		URL:              monitor.URL,
		Error:            errorMsg,
		Timestamp:        now,
		UptimePercentage: uptimePercentage,
	}

	wsHub.Broadcast <- models.WebSocketMessage{
		Type:      "metric_update",
		Data:      update,
		MonitorID: monitor.ID.Hex(),
	}

	// Log status changes
	if status == "down" {
		log.Printf("🔴 DOWN: %s (%s) - %dms - %s", monitor.Name, monitor.URL, responseTime, errorMsg)
	} else {
		log.Printf("🟢 UP: %s (%s) - %dms - HTTP %d", monitor.Name, monitor.URL, responseTime, statusCode)
	}
}

// updateMonitorStatus updates the monitor's current status in the database
func (ms *MonitorService) updateMonitorStatus(monitorID primitive.ObjectID, status string, statusCode int, responseTime int64, lastChecked time.Time) {
	collection := ms.db.GetCollection(database.MonitorsCollection)
	
	update := bson.M{
		"$set": bson.M{
			"current_status":   status,
			"current_status_code": statusCode,
			"current_response": responseTime,
			"last_checked":     lastChecked,
			"updated_at":       time.Now(),
		},
	}

	_, err := collection.UpdateOne(context.Background(), bson.M{"_id": monitorID}, update)
	if err != nil {
		log.Printf("Error updating monitor status: %v", err)
	}
}

// calculateUptimePercentage calculates uptime percentage for the last 24 hours
// calculateUptimePercentage calculates uptime percentage for the last 24 hours
func (ms *MonitorService) calculateUptimePercentage(monitorID primitive.ObjectID) float64 {
	collection := ms.db.GetCollection(database.MetricsCollection)

	startTime := time.Now().Add(-24 * time.Hour)
	filter := bson.M{
		"monitor_id": monitorID,
		"checked_at": bson.M{"$gte": startTime},
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Printf("Error fetching metrics for uptime calculation: %v", err)
		return 100 // fail-open: assume 100% uptime if query fails
	}
	defer cursor.Close(context.Background())

	total := 0
	down := 0

	for cursor.Next(context.Background()) {
		var metric models.Metric
		if err := cursor.Decode(&metric); err == nil {
			total++
			if metric.Status == "down" {
				down++
			}
		}
	}

	if total == 0 {
		return 100
	}

	uptime := (float64(total-down) / float64(total)) * 100
	return uptime
}
// StartMonitorJob starts a monitoring job for a specific monitor (public method)
func (ms *MonitorService) StartMonitorJob(monitor models.Monitor, wsHub *WebSocketHub) {
	ms.startMonitorJob(monitor, wsHub)
}