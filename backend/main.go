package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/db"
	_ "github.com/pintuotuo/backend/docs"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/routes"
	"github.com/pintuotuo/backend/scheduler"
	"github.com/pintuotuo/backend/utils"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var orderScheduler *scheduler.OrderScheduler
var settlementScheduler *scheduler.SettlementScheduler

func init() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if err := utils.InitEncryption(); err != nil {
		log.Fatalf("Failed to initialize encryption: %v", err)
	}
}

func main() {
	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	if err := cache.Init(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer cache.Close()

	orderScheduler = scheduler.NewOrderScheduler(
		5*time.Minute,
		30*time.Minute,
	)
	orderScheduler.Start()
	defer orderScheduler.Stop()

	settlementScheduler = scheduler.NewSettlementScheduler(1 * time.Hour)
	settlementScheduler.Start()
	defer settlementScheduler.Stop()

	router := gin.Default()

	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.ErrorHandlingMiddleware())
	router.Use(middleware.LoggingMiddleware())

	v1 := router.Group("/api/v1")
	{
		routes.RegisterHealthRoutes(v1)
		routes.RegisterUserRoutes(v1)
		routes.RegisterProductRoutes(v1)
		routes.RegisterOrderRoutes(v1)
		routes.RegisterGroupRoutes(v1)
		routes.RegisterTokenRoutes(v1)
		routes.RegisterPaymentRoutes(v1)
		routes.RegisterReferralRoutes(v1)
		routes.RegisterMerchantRoutes(v1)
		routes.RegisterAPIRoutes(v1)
		routes.RegisterConsumptionRoutes(v1)
		routes.RegisterNotificationRoutes(v1)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Pintuotuo Backend Server starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
