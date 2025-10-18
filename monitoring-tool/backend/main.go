package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"monitoring-tool/config"
	"monitoring-tool/database"
	"monitoring-tool/handlers"
	"monitoring-tool/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatal("Invalid configuration:", err)
	}
	cfg.LogConfig()

	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := database.InitMongoDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer db.Disconnect(context.Background())

	// Initialize MonitorService with max concurrent jobs
	maxConcurrentJobs := 10 // adjust as needed
	monitorService := services.NewMonitorService(db, maxConcurrentJobs)

	// Initialize WebSocket hub
	wsHub := services.NewWebSocketHub()
	go wsHub.Run()
	go monitorService.StartMonitoring(wsHub)

	// Initialize router
	var r *gin.Engine
	if cfg.IsProduction() {
		r = gin.New()
		r.Use(gin.Recovery())

		// Structured logging
		r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("[%s] %s %s %d %s\n",
				param.TimeStamp.Format("2006-01-02 15:04:05"),
				param.Method,
				param.Path,
				param.StatusCode,
				param.Latency,
			)
		}))
	} else {
		r = gin.Default()
	}

	// Set trusted proxies
	if cfg.IsProduction() && len(cfg.TrustedProxies) > 0 {
		r.SetTrustedProxies(cfg.TrustedProxies)
	} else {
		r.SetTrustedProxies(nil)
	}

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize handlers
	apiHandler := handlers.NewAPIHandler(monitorService, wsHub)
	wsHandler := handlers.NewWebSocketHandler(wsHub)

	// API routes
	api := r.Group("/api/v1")
	{
		api.GET("/monitors", apiHandler.GetMonitors)
		api.POST("/monitors", apiHandler.CreateMonitor)
		api.DELETE("/monitors/:id", apiHandler.DeleteMonitor)
		api.GET("/monitors/:id/metrics", apiHandler.GetMetrics)
		api.GET("/dashboard/stats", apiHandler.GetDashboardStats)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":      "healthy",
				"version":     "1.0.0",
				"environment": cfg.Environment,
			})
		})
	}

	// WebSocket endpoint
	r.GET("/ws", wsHandler.HandleWebSocket)

	// HTTP server with timeouts
	srv := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Start server
	go func() {
		log.Printf("🚀 Server starting on %s (mode: %s)", cfg.GetServerAddress(), cfg.Environment)

		var err error
		if cfg.EnableHTTPS {
			if cfg.CertFile == "" || cfg.KeyFile == "" {
				log.Fatal("HTTPS enabled but cert/key files are missing")
			}
			err = srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	shutdownTimeout := 5 * time.Second
	if cfg.IsProduction() {
		shutdownTimeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited")
}
