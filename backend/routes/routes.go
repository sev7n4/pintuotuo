package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/handlers"
	"github.com/pintuotuo/backend/middleware"
)

func RegisterHealthRoutes(router *gin.RouterGroup) {
	health := router.Group("/health")
	{
		health.GET("", handlers.HealthCheck)
		health.GET("/ready", handlers.ReadyCheck)
		health.GET("/live", handlers.LiveCheck)
		health.GET("/db", handlers.DBStats)
	}
}

func RegisterUserRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		users.POST("/register", handlers.RegisterUser)
		users.POST("/login", handlers.LoginUser)
		users.POST("/logout", handlers.LogoutUser)
		users.POST("/refresh", handlers.RefreshToken)

		users.POST("/password/reset-request", handlers.RequestPasswordReset)
		users.POST("/password/reset", handlers.ResetPassword)

		auth := users.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			auth.GET("/me", handlers.GetCurrentUser)
			auth.PUT("/me", handlers.UpdateCurrentUser)
		}

		users.GET("/:id", handlers.GetUserByID)
		users.PUT("/:id", handlers.UpdateUser)
	}
}

func RegisterAPIRoutes(router *gin.RouterGroup) {
	api := router.Group("/proxy")
	api.Use(middleware.AuthMiddleware())
	{
		api.POST("/chat", handlers.ProxyAPIRequest)
		api.POST("/completions", handlers.ProxyAPIRequest)
		api.GET("/providers", handlers.GetAPIProviders)
		api.GET("/usage", handlers.GetAPIUsageStats)
	}
}

func RegisterConsumptionRoutes(router *gin.RouterGroup) {
	consumption := router.Group("/consumption")
	consumption.Use(middleware.AuthMiddleware())
	{
		consumption.GET("/records", handlers.GetConsumptionRecords)
		consumption.GET("/stats", handlers.GetConsumptionStats)
	}
}

func RegisterProductRoutes(router *gin.RouterGroup) {
	products := router.Group("/products")
	{
		products.GET("", handlers.ListProducts)
		products.GET("/home", handlers.GetHomeData)
		products.GET("/hot", handlers.GetHotProducts)
		products.GET("/new", handlers.GetNewProducts)
		products.GET("/categories", handlers.GetCategories)
		products.GET("/search", handlers.SearchProducts)
		products.GET("/:id", handlers.GetProductByID)

		merchants := products.Group("/merchants")
		merchants.Use(middleware.AuthMiddleware())
		{
			merchants.POST("", handlers.CreateProduct)
			merchants.PUT("/:id", handlers.UpdateProduct)
			merchants.DELETE("/:id", handlers.DeleteProduct)
		}
	}
}

func RegisterOrderRoutes(router *gin.RouterGroup) {
	orders := router.Group("/orders")
	orders.Use(middleware.AuthMiddleware())
	{
		orders.POST("", handlers.CreateOrder)
		orders.GET("", handlers.ListOrders)
		orders.GET("/:id", handlers.GetOrderByID)
		orders.PUT("/:id/cancel", handlers.CancelOrder)
	}
}

func RegisterCartRoutes(router *gin.RouterGroup) {
	cart := router.Group("/cart")
	cart.Use(middleware.AuthMiddleware())
	{
		cart.GET("", handlers.GetCart)
		cart.POST("/items", handlers.AddToCart)
		cart.PUT("/items/:id", handlers.UpdateCartItem)
		cart.DELETE("/items/:id", handlers.RemoveFromCart)
		cart.DELETE("", handlers.ClearCart)
	}
}

func RegisterGroupRoutes(router *gin.RouterGroup) {
	groups := router.Group("/groups")
	groups.Use(middleware.AuthMiddleware())
	{
		groups.POST("", handlers.CreateGroup)
		groups.GET("", handlers.ListGroups)
		groups.GET("/:id", handlers.GetGroupByID)
		groups.POST("/:id/join", handlers.JoinGroup)
		groups.DELETE("/:id", handlers.CancelGroup)
		groups.GET("/:id/progress", handlers.GetGroupProgress)
	}
}

func RegisterTokenRoutes(router *gin.RouterGroup) {
  tokens := router.Group("/tokens")
  {
    tokens.GET("/balance", handlers.GetBalance)
    tokens.GET("/consumption", handlers.GetConsumption)
    tokens.GET("/total-balance", handlers.GetTotalBalance)
    tokens.GET("/transactions", handlers.ListTransactions)
    tokens.POST("/transfer", handlers.TransferTokens)
    tokens.POST("/recharge", handlers.RechargeTokens)
    tokens.POST("/consume", handlers.ConsumeTokens)

    // API Key management
    keys := tokens.Group("/keys")
    {
      keys.GET("", handlers.ListAPIKeys)
      keys.POST("", handlers.CreateAPIKey)
      keys.PUT("/:id", handlers.UpdateAPIKey)
      keys.DELETE("/:id", handlers.DeleteAPIKey)
    }
  }
}

func RegisterPaymentRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payments")
	payments.Use(middleware.AuthMiddleware())
	{
		payments.POST("", handlers.InitiatePayment)
		payments.GET("", handlers.ListPayments)
		payments.GET("/:id", handlers.GetPaymentByID)
		payments.POST("/:id/refund", handlers.RefundPayment)
	}

	// Webhook routes (without authentication)
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/alipay", handlers.HandleAlipayCallback)
		webhooks.POST("/wechat", handlers.HandleWechatCallback)
	}

	// Merchant routes
	merchants := router.Group("/merchants")
	{
		merchants.GET("/:merchant_id/revenue", handlers.GetMerchantRevenue)
	}
}

// RegisterAnalyticsRoutes registers analytics routes
func RegisterAnalyticsRoutes(router *gin.RouterGroup) {
	analytics := router.Group("/analytics")
	{
		analytics.GET("/consumption", handlers.GetUserConsumption)
		analytics.GET("/spending-pattern", handlers.GetUserSpendingPattern)
		analytics.GET("/consumption-history", handlers.GetConsumptionHistory)
		analytics.GET("/revenue", handlers.GetRevenueAnalytics)
		analytics.GET("/top-spenders", handlers.GetTopSpenders)
		analytics.GET("/metrics", handlers.GetPlatformMetrics)
	}
}

func RegisterNotificationRoutes(router *gin.RouterGroup) {
	notifications := router.Group("/notifications")
	notifications.Use(middleware.AuthMiddleware())
	{
		notifications.GET("", handlers.GetNotifications)
		notifications.GET("/unread-count", handlers.GetUnreadCount)
		notifications.PUT("/:id/read", handlers.MarkNotificationRead)
		notifications.PUT("/read-all", handlers.MarkAllNotificationsRead)

		notifications.POST("/device-token", handlers.RegisterDeviceToken)
		notifications.DELETE("/device-token", handlers.UnregisterDeviceToken)
	}
}

func RegisterAdminRoutes(router *gin.RouterGroup) {
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.GET("/users", handlers.GetAdminUsers)
		admin.POST("/users", handlers.CreateAdminUser)
		admin.GET("/stats", handlers.GetAdminStats)
		admin.GET("/merchants/pending", handlers.GetPendingMerchants)
		admin.POST("/merchants/:id/approve", handlers.ApproveMerchant)
		admin.POST("/merchants/:id/reject", handlers.RejectMerchant)
	}
}
