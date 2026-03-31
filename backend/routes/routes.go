package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/handlers"
	"github.com/pintuotuo/backend/middleware"
)

func RegisterUploadRoutes(router *gin.RouterGroup) {
	upload := router.Group("/upload")
	upload.Use(middleware.AuthMiddleware())
	{
		upload.POST("", handlers.UploadFile)
	}
}

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
	authMerchants := router.Group("/merchants")
	authMerchants.Use(middleware.AuthMiddleware())
	{
		authMerchants.POST("/register", handlers.RegisterMerchant)
		authMerchants.GET("/profile", handlers.GetMerchantProfile)
		authMerchants.PUT("/profile", handlers.UpdateMerchantProfile)
		authMerchants.GET("/stats", handlers.GetMerchantStats)
		authMerchants.GET("/products", handlers.GetMerchantProducts)
		authMerchants.GET("/orders", handlers.GetMerchantOrders)
		authMerchants.GET("/settlements", handlers.GetMerchantSettlements)
		authMerchants.POST("/settlements", handlers.RequestSettlement)
		authMerchants.GET("/settlements/:id", handlers.GetSettlementDetail)
		authMerchants.POST("/documents", handlers.SubmitMerchantDocuments)
		authMerchants.GET("/status", handlers.GetMerchantStatus)

		apiKeys := authMerchants.Group("/api-keys")
		{
			apiKeys.GET("", handlers.ListMerchantAPIKeys)
			apiKeys.POST("", handlers.CreateMerchantAPIKey)
			apiKeys.PUT("/:id", handlers.UpdateMerchantAPIKey)
			apiKeys.DELETE("/:id", handlers.DeleteMerchantAPIKey)
			apiKeys.GET("/usage", handlers.GetMerchantAPIKeyUsage)
		}

		skus := authMerchants.Group("/skus")
		{
			skus.GET("", handlers.ListMerchantSKUs)
			skus.GET("/available", handlers.GetAvailableSKUs)
			skus.POST("", handlers.CreateMerchantSKU)
			skus.PUT("/:id", handlers.UpdateMerchantSKU)
			skus.DELETE("/:id", handlers.DeleteMerchantSKU)
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
		admin.GET("/merchants/audit-logs", handlers.GetMerchantAuditLogs)
		admin.GET("/merchants", handlers.GetAdminMerchants)
		admin.PATCH("/merchants/:id", handlers.PatchAdminMerchant)
		admin.POST("/merchants/:id/approve", handlers.ApproveMerchant)
		admin.POST("/merchants/:id/reject", handlers.RejectMerchant)

		admin.GET("/spus", handlers.ListSPUs)
		admin.GET("/spus/:id", handlers.GetSPUByID)
		admin.POST("/spus", handlers.CreateSPU)
		admin.PUT("/spus/:id", handlers.UpdateSPU)
		admin.DELETE("/spus/:id", handlers.DeleteSPU)

		admin.GET("/skus", handlers.ListSKUs)
		admin.GET("/skus/:id", handlers.GetSKUByID)
		admin.POST("/skus", handlers.CreateSKU)
		admin.PUT("/skus/:id", handlers.UpdateSKU)
		admin.DELETE("/skus/:id", handlers.DeleteSKU)

		admin.GET("/model-providers", handlers.GetModelProviders)
	}
}

func RegisterFlashSaleRoutes(router *gin.RouterGroup) {
	flashSales := router.Group("/flash-sales")
	{
		flashSales.GET("/active", handlers.GetActiveFlashSales)
		flashSales.GET("/:id/products", handlers.GetFlashSaleProducts)
	}

	authFlashSales := router.Group("/flash-sales")
	authFlashSales.Use(middleware.AuthMiddleware())
	{
		authFlashSales.POST("", handlers.CreateFlashSale)
		authFlashSales.PUT("/:id/status", handlers.UpdateFlashSaleStatus)
	}
}

func RegisterFavoriteRoutes(router *gin.RouterGroup) {
	favorites := router.Group("/favorites")
	favorites.Use(middleware.AuthMiddleware())
	{
		favorites.GET("", handlers.GetFavorites)
		favorites.POST("", handlers.AddFavorite)
		favorites.DELETE("/:product_id", handlers.RemoveFavorite)
		favorites.GET("/check/:product_id", handlers.CheckFavorite)
	}
}

func RegisterBrowseHistoryRoutes(router *gin.RouterGroup) {
	history := router.Group("/browse-history")
	history.Use(middleware.AuthMiddleware())
	{
		history.GET("", handlers.GetBrowseHistory)
		history.POST("", handlers.AddBrowseHistory)
		history.DELETE("", handlers.ClearBrowseHistory)
		history.DELETE("/:product_id", handlers.RemoveBrowseHistoryItem)
	}
}

func RegisterSKURoutes(router *gin.RouterGroup) {
	skus := router.Group("/skus")
	{
		skus.GET("", handlers.ListPublicSKUs)
		skus.GET("/:id", handlers.GetPublicSKUByID)
	}

	computePoints := router.Group("/compute-points")
	computePoints.Use(middleware.AuthMiddleware())
	{
		computePoints.GET("/balance", handlers.GetComputePointBalance)
		computePoints.GET("/transactions", handlers.GetComputePointTransactions)
	}

	subscriptions := router.Group("/subscriptions")
	subscriptions.Use(middleware.AuthMiddleware())
	{
		subscriptions.GET("", handlers.GetUserSubscriptions)
	}
}
