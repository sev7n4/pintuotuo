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
		products.GET("/:id/groups", handlers.GetGroupsByProduct)

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
	tokens.Use(middleware.AuthMiddleware())
	{
		tokens.GET("/balance", handlers.GetTokenBalance)
		tokens.GET("/consumption", handlers.GetTokenConsumption)
		tokens.POST("/transfer", handlers.TransferTokens)

		tokens.GET("/recharge/packages", handlers.GetRechargePackages)
		tokens.POST("/recharge", handlers.CreateRechargeOrder)
		tokens.GET("/recharge/orders", handlers.GetRechargeOrders)
		tokens.GET("/recharge/orders/:id", handlers.GetRechargeOrder)

		keys := tokens.Group("/keys")
		{
			keys.GET("", handlers.ListAPIKeys)
			keys.POST("", handlers.CreateAPIKey)
			keys.PUT("/:id", handlers.UpdateAPIKey)
			keys.DELETE("/:id", handlers.DeleteAPIKey)
		}
	}

	router.POST("/tokens/recharge/callback", handlers.HandleRechargeCallback)
}

func RegisterPaymentRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payments")
	payments.Use(middleware.AuthMiddleware())
	{
		payments.POST("", handlers.CreatePayment)
		payments.GET("/:id", handlers.GetPaymentStatus)
	}

	webhooks := router.Group("/payments/webhooks")
	{
		webhooks.POST("/alipay", handlers.AlipayNotify)
		webhooks.POST("/wechat", handlers.WechatNotify)
	}
}

func RegisterReferralRoutes(router *gin.RouterGroup) {
	referrals := router.Group("/referrals")
	referrals.Use(middleware.AuthMiddleware())
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

func RegisterMerchantRoutes(router *gin.RouterGroup) {
	merchants := router.Group("/merchants")
	{
		merchants.POST("/register", handlers.RegisterMerchant)
	}

	authMerchants := router.Group("/merchants")
	authMerchants.Use(middleware.AuthMiddleware())
	{
		authMerchants.GET("/profile", handlers.GetMerchantProfile)
		authMerchants.PUT("/profile", handlers.UpdateMerchantProfile)
		authMerchants.GET("/stats", handlers.GetMerchantStats)
		authMerchants.GET("/products", handlers.GetMerchantProducts)
		authMerchants.GET("/orders", handlers.GetMerchantOrders)
		authMerchants.GET("/settlements", handlers.GetMerchantSettlements)
		authMerchants.POST("/settlements", handlers.RequestSettlement)
		authMerchants.GET("/settlements/:id", handlers.GetSettlementDetail)

		apiKeys := authMerchants.Group("/api-keys")
		{
			apiKeys.GET("", handlers.ListMerchantAPIKeys)
			apiKeys.POST("", handlers.CreateMerchantAPIKey)
			apiKeys.PUT("/:id", handlers.UpdateMerchantAPIKey)
			apiKeys.DELETE("/:id", handlers.DeleteMerchantAPIKey)
			apiKeys.GET("/usage", handlers.GetMerchantAPIKeyUsage)
		}
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
