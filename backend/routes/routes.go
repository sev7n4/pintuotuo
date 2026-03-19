package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/handlers"
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
		// Auth endpoints
		users.POST("/register", handlers.RegisterUser)
		users.POST("/login", handlers.LoginUser)
		users.POST("/logout", handlers.LogoutUser)
		users.POST("/refresh", handlers.RefreshToken)

		// Password reset endpoints (no auth required)
		users.POST("/password/reset-request", handlers.RequestPasswordReset)
		users.POST("/password/reset", handlers.ResetPassword)

		// User management endpoints
		users.GET("/me", handlers.GetCurrentUser)
		users.PUT("/me", handlers.UpdateCurrentUser)
		users.GET("/:id", handlers.GetUserByID)
		users.PUT("/:id", handlers.UpdateUser)
	}
}

// RegisterAPIRoutes registers API proxy routes
func RegisterAPIRoutes(router *gin.RouterGroup) {
	api := router.Group("/proxy")
	{
		api.POST("/chat", handlers.ProxyAPIRequest)
		api.POST("/completions", handlers.ProxyAPIRequest)
		api.GET("/providers", handlers.GetAPIProviders)
		api.GET("/usage", handlers.GetAPIUsageStats)
	}
}

// RegisterConsumptionRoutes registers consumption routes
func RegisterConsumptionRoutes(router *gin.RouterGroup) {
	consumption := router.Group("/consumption")
	{
		consumption.GET("/records", handlers.GetConsumptionRecords)
		consumption.GET("/stats", handlers.GetConsumptionStats)
	}
}

// RegisterProductRoutes registers product-related routes
func RegisterProductRoutes(router *gin.RouterGroup) {
	products := router.Group("/products")
	{
		// Read operations
		products.GET("", handlers.ListProducts)
		products.GET("/home", handlers.GetHomeData)
		products.GET("/hot", handlers.GetHotProducts)
		products.GET("/new", handlers.GetNewProducts)
		products.GET("/categories", handlers.GetCategories)
		products.GET("/search", handlers.SearchProducts)
		products.GET("/:id", handlers.GetProductByID)

		// Merchant operations
		merchants := products.Group("/merchants")
		{
			merchants.POST("", handlers.CreateProduct)
			merchants.PUT("/:id", handlers.UpdateProduct)
			merchants.DELETE("/:id", handlers.DeleteProduct)
		}
	}
}

// RegisterOrderRoutes registers order-related routes
func RegisterOrderRoutes(router *gin.RouterGroup) {
	orders := router.Group("/orders")
	{
		orders.POST("", handlers.CreateOrder)
		orders.GET("", handlers.ListOrders)
		orders.GET("/:id", handlers.GetOrderByID)
		orders.PUT("/:id/cancel", handlers.CancelOrder)
	}
}

// RegisterGroupRoutes registers group purchase routes
func RegisterGroupRoutes(router *gin.RouterGroup) {
	groups := router.Group("/groups")
	{
		groups.POST("", handlers.CreateGroup)
		groups.GET("", handlers.ListGroups)
		groups.GET("/:id", handlers.GetGroupByID)
		groups.POST("/:id/join", handlers.JoinGroup)
		groups.DELETE("/:id", handlers.CancelGroup)
		groups.GET("/:id/progress", handlers.GetGroupProgress)
	}
}

// RegisterTokenRoutes registers token management routes
func RegisterTokenRoutes(router *gin.RouterGroup) {
	tokens := router.Group("/tokens")
	{
		tokens.GET("/balance", handlers.GetTokenBalance)
		tokens.GET("/consumption", handlers.GetTokenConsumption)
		tokens.POST("/transfer", handlers.TransferTokens)

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

// RegisterPaymentRoutes registers payment routes
func RegisterPaymentRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payments")
	{
		payments.POST("", handlers.CreatePayment)
		payments.GET("/:id", handlers.GetPaymentStatus)

		// Webhooks
		webhooks := payments.Group("/webhooks")
		{
			webhooks.POST("/alipay", handlers.AlipayNotify)
			webhooks.POST("/wechat", handlers.WechatNotify)
		}
	}
}

// RegisterReferralRoutes registers referral routes
func RegisterReferralRoutes(router *gin.RouterGroup) {
	referrals := router.Group("/referrals")
	{
		referrals.GET("/code", handlers.GetMyReferralCode)
		referrals.POST("/bind", handlers.BindReferralCode)
		referrals.GET("/validate/:code", handlers.ValidateReferralCode)
		referrals.GET("/stats", handlers.GetReferralStats)
		referrals.GET("/list", handlers.GetReferralList)
		referrals.GET("/rewards", handlers.GetReferralRewards)
		referrals.POST("/rewards/pay", handlers.PayReferralRewards)
	}
}

// RegisterMerchantRoutes registers merchant routes
func RegisterMerchantRoutes(router *gin.RouterGroup) {
	merchants := router.Group("/merchants")
	{
		merchants.POST("/register", handlers.RegisterMerchant)
		merchants.GET("/profile", handlers.GetMerchantProfile)
		merchants.PUT("/profile", handlers.UpdateMerchantProfile)
		merchants.GET("/stats", handlers.GetMerchantStats)
		merchants.GET("/products", handlers.GetMerchantProducts)
		merchants.GET("/orders", handlers.GetMerchantOrders)
		merchants.GET("/settlements", handlers.GetMerchantSettlements)
		merchants.POST("/settlements", handlers.RequestSettlement)
		merchants.GET("/settlements/:id", handlers.GetSettlementDetail)

		// API Key management
		apiKeys := merchants.Group("/api-keys")
		{
			apiKeys.GET("", handlers.ListMerchantAPIKeys)
			apiKeys.POST("", handlers.CreateMerchantAPIKey)
			apiKeys.PUT("/:id", handlers.UpdateMerchantAPIKey)
			apiKeys.DELETE("/:id", handlers.DeleteMerchantAPIKey)
			apiKeys.GET("/usage", handlers.GetMerchantAPIKeyUsage)
		}
	}
}

// RegisterNotificationRoutes registers notification routes
func RegisterNotificationRoutes(router *gin.RouterGroup) {
	notifications := router.Group("/notifications")
	{
		notifications.GET("", handlers.GetNotifications)
		notifications.GET("/unread-count", handlers.GetUnreadCount)
		notifications.PUT("/:id/read", handlers.MarkNotificationRead)
		notifications.PUT("/read-all", handlers.MarkAllNotificationsRead)

		// Device tokens for push notifications
		devices := notifications.Group("/devices")
		{
			devices.POST("", handlers.RegisterDeviceToken)
			devices.DELETE("", handlers.UnregisterDeviceToken)
		}
	}
}
