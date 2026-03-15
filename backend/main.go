package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/routes"
)

func init() {
	// Load environment variables
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}

func main() {
	// Initialize dependencies
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer config.CloseDB()

	// Setup Gin router
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.ErrorHandlingMiddleware())
	router.Use(middleware.LoggingMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"message": "Pintuotuo Backend Server is running",
			"timestamp": getCurrentTime(),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// User routes
		routes.RegisterUserRoutes(v1)

		// Product routes
		routes.RegisterProductRoutes(v1)

		// Order routes
		routes.RegisterOrderRoutes(v1)

		// Group routes
		routes.RegisterGroupRoutes(v1)

		// Token routes
		routes.RegisterTokenRoutes(v1)

		// Payment routes (includes payments, webhooks, and merchants)
		routes.RegisterPaymentRoutes(v1)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Pintuotuo Backend Server starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getCurrentTime() string {
	// Simple timestamp function
	return "2026-03-14T23:00:00Z"
}
