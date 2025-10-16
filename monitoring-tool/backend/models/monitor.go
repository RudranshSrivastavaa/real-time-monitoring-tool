package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Monitor represents an endpoint to monitor
type Monitor struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	URL         string             `json:"url" bson:"url"`
	Method      string             `json:"method" bson:"method"`           // GET, POST, etc.
	Interval    int                `json:"interval" bson:"interval"`       // seconds
	Timeout     int                `json:"timeout" bson:"timeout"`         // seconds
	Status      string             `json:"status" bson:"status"`           // active, paused, error
	IsActive    bool               `json:"is_active" bson:"is_active"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
	LastChecked *time.Time         `json:"last_checked,omitempty" bson:"last_checked,omitempty"`
	
	// Current status info (for quick dashboard display)
	CurrentStatus     string  `json:"current_status" bson:"current_status"`         // up, down, unknown
	CurrentResponse   int     `json:"current_response" bson:"current_response"`     // response time in ms
	UptimePercentage  float64 `json:"uptime_percentage" bson:"uptime_percentage"`
}

// CreateMonitorRequest represents the request to create a new monitor
type CreateMonitorRequest struct {
	Name     string `json:"name" binding:"required"`
	URL      string `json:"url" binding:"required"`
	Method   string `json:"method"`
	Interval int    `json:"interval"`
	Timeout  int    `json:"timeout"`
}

// Validate sets default values and validates the monitor request
func (req *CreateMonitorRequest) Validate() {
	if req.Method == "" {
		req.Method = "GET"
	}
	if req.Interval == 0 {
		req.Interval = 30 // 30 seconds default
	}
	if req.Timeout == 0 {
		req.Timeout = 10 // 10 seconds default
	}
}

// ToMonitor converts a request to a Monitor model
func (req *CreateMonitorRequest) ToMonitor() *Monitor {
	req.Validate()
	now := time.Now()
	
	return &Monitor{
		Name:              req.Name,
		URL:               req.URL,
		Method:            req.Method,
		Interval:          req.Interval,
		Timeout:           req.Timeout,
		Status:            "active",
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
		CurrentStatus:     "unknown",
		CurrentResponse:   0,
		UptimePercentage:  100.0,
	}
}
