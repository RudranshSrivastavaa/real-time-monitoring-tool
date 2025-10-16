package main

import (
	"context"
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
	gin.SetMode(cfg.Environment)

	// Initialize database connection
	db, err := database.InitMongoDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer db.Disconnect(context.Background())

	// Initialize services - Fixed: Remove extra parameters
	monitorService := services.NewMonitorService(db)
	wsHub := services.NewWebSocketHub()

	// Start WebSocket hub
	go wsHub.Run()

	// Start monitoring service
	go monitorService.StartMonitoring(wsHub)

	// Initialize Gin router
	r := gin.Default()

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

	// API Routes
	api := r.Group("/api/v1")
	{
		api.GET("/monitors", apiHandler.GetMonitors)
		api.POST("/monitors", apiHandler.CreateMonitor)
		api.DELETE("/monitors/:id", apiHandler.DeleteMonitor)
		api.GET("/monitors/:id/metrics", apiHandler.GetMetrics)
		api.GET("/dashboard/stats", apiHandler.GetDashboardStats)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":      "healthy",
				"version":     "1.0.0",
				"environment": cfg.Environment,
			})
		})
	}

	// WebSocket endpoint
	r.GET("/ws", wsHandler.HandleWebSocket)

	// Server configuration
	srv := &http.Server{
		Addr:    cfg.GetServerAddress(),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on %s", cfg.GetServerAddress())
		log.Printf("📊 Dashboard: http://localhost:3000")
		log.Printf("🔌 WebSocket: ws://localhost:%s/ws", cfg.Port)
		log.Printf("🏥 Health: http://localhost:%s/api/v1/health", cfg.Port)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited")
}