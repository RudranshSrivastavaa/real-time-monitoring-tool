// handlers/api_handlers.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"monitoring-tool/models"
	"monitoring-tool/services"
)

type APIHandler struct {
	monitorService *services.MonitorService
	wsHub          *services.WebSocketHub
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(monitorService *services.MonitorService, wsHub *services.WebSocketHub) *APIHandler {
	return &APIHandler{
		monitorService: monitorService,
		wsHub:          wsHub,
	}
}

// GetMonitors handles GET /api/v1/monitors
func (h *APIHandler) GetMonitors(c *gin.Context) {
	monitors, err := h.monitorService.GetMonitors()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve monitors",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    monitors,
		"count":   len(monitors),
	})
}

// CreateMonitor handles POST /api/v1/monitors
func (h *APIHandler) CreateMonitor(c *gin.Context) {
	var req models.CreateMonitorRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Convert request to monitor model
	monitor := req.ToMonitor()

	// Create monitor in database
	if err := h.monitorService.CreateMonitor(monitor); err != nil {
		if err.Error() == "monitor with URL "+req.URL+" already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Monitor already exists",
				"details": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create monitor",
			"details": err.Error(),
		})
		return
	}

	// FIXED: Start monitoring job immediately for the new monitor
	if monitor.IsActive {
		h.monitorService.StartMonitorJob(*monitor, h.wsHub)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Monitor created successfully",
		"data":    monitor,
	})
}

// DeleteMonitor handles DELETE /api/v1/monitors/:id
func (h *APIHandler) DeleteMonitor(c *gin.Context) {
	idParam := c.Param("id")
	
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid monitor ID format",
			"details": err.Error(),
		})
		return
	}

	// Delete monitor (this also stops the monitoring job)
	if err := h.monitorService.DeleteMonitor(objectID); err != nil {
		if err.Error() == "monitor not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Monitor not found",
				"details": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete monitor",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Monitor deleted successfully",
	})
}

// GetMetrics handles GET /api/v1/monitors/:id/metrics
func (h *APIHandler) GetMetrics(c *gin.Context) {
	idParam := c.Param("id")
	
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid monitor ID format",
			"details": err.Error(),
		})
		return
	}

	// Get hours parameter (default to 24 hours)
	hoursParam := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursParam)
	if err != nil || hours < 1 || hours > 168 { // Max 1 week
		hours = 24
	}

	// Retrieve metrics
	metrics, err := h.monitorService.GetMetrics(objectID, hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve metrics",
			"details": err.Error(),
		})
		return
	}

	// Calculate summary statistics
	summary := calculateMetricsSummary(metrics)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
		"summary": summary,
		"count":   len(metrics),
		"hours":   hours,
	})
}

// GetDashboardStats handles GET /api/v1/dashboard/stats
func (h *APIHandler) GetDashboardStats(c *gin.Context) {
    monitors, err := h.monitorService.GetMonitors()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Failed to retrieve dashboard stats",
            "details": err.Error(),
        })
        return
    }

    // Calculate dashboard statistics
    stats := models.DashboardStats{
        TotalMonitors:  len(monitors),
        ActiveMonitors: 0,
        UpMonitors:     0,
        DownMonitors:   0,
        OverallUptime:  0.0,
        AverageResponse: 0.0,
    }

    var totalUptime float64 = 0
    var totalResponseTime int64 = 0
    var activeCount int = 0

    for _, monitor := range monitors {
        if monitor.IsActive {
            stats.ActiveMonitors++
            activeCount++
            
            // 🔥 Use switch instead of if-else
            switch monitor.CurrentStatus {
            case "up":
                stats.UpMonitors++
            case "down":
                stats.DownMonitors++
            // You could add more cases here like "unknown", "warning", etc.
            }

            totalUptime += monitor.UptimePercentage
            totalResponseTime += int64(monitor.CurrentResponse)
        }
    }

    // Calculate averages
    if activeCount > 0 {
        stats.OverallUptime = totalUptime / float64(activeCount)
        stats.AverageResponse = float64(totalResponseTime) / float64(activeCount)
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    stats,
    })
}

// calculateMetricsSummary generates summary statistics for metrics
func calculateMetricsSummary(metrics []models.Metric) map[string]interface{} {
	if len(metrics) == 0 {
		return map[string]interface{}{
			"total_checks":       0,
			"successful_checks":  0,
			"failed_checks":      0,
			"uptime_percentage":  0.0,
			"average_response":   0.0,
			"min_response":       0,
			"max_response":       0,
		}
	}

	var successfulChecks, failedChecks int
	var totalResponseTime, minResponse, maxResponse int64
	
	minResponse = metrics[0].ResponseTime
	maxResponse = metrics[0].ResponseTime

	for _, metric := range metrics {
		if metric.Status == "up" {
			successfulChecks++
		} else {
			failedChecks++
		}

		totalResponseTime += metric.ResponseTime
		
		if metric.ResponseTime < minResponse {
			minResponse = metric.ResponseTime
		}
		if metric.ResponseTime > maxResponse {
			maxResponse = metric.ResponseTime
		}
	}

	uptimePercentage := float64(successfulChecks) / float64(len(metrics)) * 100
	averageResponse := float64(totalResponseTime) / float64(len(metrics))

	return map[string]interface{}{
		"total_checks":       len(metrics),
		"successful_checks":  successfulChecks,
		"failed_checks":      failedChecks,
		"uptime_percentage":  uptimePercentage,
		"average_response":   averageResponse,
		"min_response":       minResponse,
		"max_response":       maxResponse,
	}
}