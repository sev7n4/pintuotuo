package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/db"
	_ "github.com/pintuotuo/backend/docs"
	"github.com/pintuotuo/backend/handlers"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/notification"
	"github.com/pintuotuo/backend/routes"
	"github.com/pintuotuo/backend/scheduler"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/tracing"
	"github.com/pintuotuo/backend/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var orderScheduler *scheduler.OrderScheduler
var settlementScheduler *scheduler.SettlementScheduler
var subscriptionScheduler *scheduler.SubscriptionScheduler
var responseCleanupScheduler *scheduler.ResponseCleanupScheduler

func init() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if err := middleware.ValidateSecurityConfig(); err != nil {
		log.Fatalf("Failed security config validation: %v", err)
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

	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to initialize config database: %v", err)
	}
	defer config.CloseDB()

	services.GetSmartRouter().ReloadRoutingStrategies()
	services.StartRoutingStrategiesListener(config.DatabaseURL())
	services.StartPlatformSettingsListener(config.DatabaseURL())
	services.InitLitellmCache()

	if err := cache.Init(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer cache.Close()

	handlers.InitPaymentService()

	orderScheduler = scheduler.NewOrderScheduler(
		5*time.Minute,
		30*time.Minute,
	)
	orderScheduler.Start()
	defer orderScheduler.Stop()

	settlementScheduler = scheduler.NewSettlementScheduler(1 * time.Hour)
	settlementScheduler.Start()
	defer settlementScheduler.Stop()

	notifySvc := notification.NotificationServiceFromEnv()
	subscriptionScheduler = scheduler.NewSubscriptionScheduler(1*time.Hour, notifySvc)
	subscriptionScheduler.Start()
	defer subscriptionScheduler.Stop()

	responseCleanupScheduler = scheduler.NewResponseCleanupScheduler(24 * time.Hour)
	responseCleanupScheduler.Start()
	defer responseCleanupScheduler.Stop()

	services.GetHealthScheduler().Start()
	defer services.GetHealthScheduler().Stop()

	shutdownTracing, err := tracing.Init(context.Background())
	if err != nil {
		log.Fatalf("Failed to init tracing: %v", err)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			log.Printf("tracing shutdown: %v", err)
		}
	}()

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "pintuotuo-backend"
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(otelgin.Middleware(serviceName))
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.ErrorHandlingMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.TracingResponseHeaders())

	router.Static("/uploads", "./uploads")

	v1 := router.Group("/api/v1")
	{
		routes.RegisterHealthRoutes(v1)
		routes.RegisterUploadRoutes(v1)
		routes.RegisterUserRoutes(v1)
		routes.RegisterCatalogRoutes(v1)
		routes.RegisterCartRoutes(v1)
		routes.RegisterOrderRoutes(v1)
		routes.RegisterGroupRoutes(v1)
		routes.RegisterTokenRoutes(v1)
		routes.RegisterPaymentRoutes(v1)
		routes.RegisterReferralRoutes(v1)
		routes.RegisterMerchantRoutes(v1)
		routes.RegisterAPIRoutes(v1)
		routes.RegisterOpenAICompatRoutes(v1)
		routes.RegisterAnthropicCompatRoutes(v1)
		routes.RegisterConsumptionRoutes(v1)
		routes.RegisterNotificationRoutes(v1)
		routes.RegisterAdminRoutes(v1)
		routes.RegisterFlashSaleRoutes(v1)
		routes.RegisterFavoriteRoutes(v1)
		routes.RegisterBrowseHistoryRoutes(v1)
		routes.RegisterSKURoutes(v1)
		routes.RegisterSettlementRoutes(v1)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Pintuotuo Backend Server starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
