package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Metric represents a single monitoring check result
type Metric struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	MonitorID    primitive.ObjectID `json:"monitor_id" bson:"monitor_id"`
	URL          string             `json:"url" bson:"url"`
	Status       string             `json:"status" bson:"status"`             // up, down
	StatusCode   int                `json:"status_code" bson:"status_code"`   // HTTP status code
	ResponseTime int64              `json:"response_time" bson:"response_time"` // milliseconds
	Error        string             `json:"error,omitempty" bson:"error,omitempty"`
	CheckedAt    time.Time          `json:"checked_at" bson:"checked_at"`
}

// WebSocketMessage represents real-time updates sent via WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`    // "metric_update", "monitor_status", "error"
	Data    interface{} `json:"data"`
	MonitorID string    `json:"monitor_id,omitempty"`
}

// MonitorUpdate represents live status updates
type MonitorUpdate struct {
	MonitorID        string    `json:"monitor_id"`
	Status           string    `json:"status"`           // up, down
	StatusCode       int       `json:"status_code"`
	ResponseTime     int64     `json:"response_time"`
	URL              string    `json:"url"`
	Error            string    `json:"error,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
	UptimePercentage float64   `json:"uptime_percentage"`
}

// DashboardStats represents overall monitoring statistics
type DashboardStats struct {
	TotalMonitors    int     `json:"total_monitors"`
	ActiveMonitors   int     `json:"active_monitors"`
	UpMonitors      int     `json:"up_monitors"`
	DownMonitors    int     `json:"down_monitors"`
	OverallUptime   float64 `json:"overall_uptime"`
	AverageResponse float64 `json:"average_response"`
}

// MetricsQuery represents query parameters for fetching metrics
type MetricsQuery struct {
	MonitorID primitive.ObjectID `json:"monitor_id"`
	StartTime time.Time          `json:"start_time"`
	EndTime   time.Time          `json:"end_time"`
	Limit     int                `json:"limit"`
}