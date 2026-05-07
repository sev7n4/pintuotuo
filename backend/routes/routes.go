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
		health.GET("/stats", handlers.GetHealthCheckStats)
	}

	providers := router.Group("/health/providers")
	{
		providers.GET("", handlers.GetAllProvidersHealth)
		providers.GET("/:id", handlers.GetProviderHealth)
		providers.POST("/:id/check", middleware.AuthMiddleware(), handlers.TriggerHealthCheck)
		providers.GET("/:id/history", handlers.GetHealthCheckHistory)
	}
}

func RegisterUserRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		users.GET("/auth/capabilities", handlers.GetAuthCapabilities)
		users.POST("/sms/send", handlers.SendSMSCode)
		users.POST("/sms/register", handlers.RegisterWithSMS)
		users.POST("/sms/login", handlers.LoginWithSMS)
		users.POST("/email/magic/send", handlers.SendEmailMagicLink)
		users.GET("/email/magic/verify", handlers.VerifyEmailMagicLink)
		users.GET("/oauth/wechat/start", handlers.WechatOAuthStart)
		users.GET("/oauth/wechat/callback", handlers.WechatOAuthCallback)
		users.GET("/oauth/github/start", handlers.GithubOAuthStart)
		users.GET("/oauth/github/callback", handlers.GithubOAuthCallback)

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
			auth.GET("/me/identities", handlers.GetUserIdentities)
			auth.PUT("/me", handlers.UpdateCurrentUser)
			auth.POST("/me/mfa/totp/setup", handlers.PostAdminTOTPSetup)
			auth.POST("/me/mfa/totp/confirm", handlers.PostAdminTOTPConfirm)
		}

		users.GET("/:id", handlers.GetUserByID)
		users.PUT("/:id", handlers.UpdateUser)
	}
}

func RegisterAPIRoutes(router *gin.RouterGroup) {
	api := router.Group("/proxy")
	api.Use(middleware.AuthMiddleware())
	api.Use(middleware.RateLimitMiddleware())
	{
		api.POST("/chat", handlers.ProxyAPIRequest)
		api.POST("/completions", handlers.ProxyAPIRequest)
		api.GET("/providers", handlers.GetAPIProviders)
		api.GET("/usage", handlers.GetAPIUsageStats)
		api.GET("/trace/:request_id", handlers.GetAPIRequestTrace)
	}
}

// RegisterOpenAICompatRoutes exposes OpenAI-SDK-friendly paths under /openai/v1 (full URL: /api/v1/openai/v1/...).
func RegisterOpenAICompatRoutes(router *gin.RouterGroup) {
	openai := router.Group("/openai/v1")
	openai.Use(middleware.APIKeyOrJWTAuthMiddleware())
	openai.Use(middleware.RateLimitMiddleware())
	{
		openai.GET("/models", handlers.OpenAIListModels)
		openai.POST("/chat/completions", handlers.OpenAIChatCompletions)
		openai.POST("/responses", handlers.OpenAIResponses)
		openai.GET("/responses/:id", handlers.OpenAIResponsesGet)
		openai.DELETE("/responses/:id", handlers.OpenAIResponsesDelete)
		openai.GET("/responses/:id/status", handlers.OpenAIResponsesStatus)
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

// RegisterCatalogRoutes registers marketplace catalog APIs (SKU 粒度，替代 legacy /products)。
func RegisterCatalogRoutes(router *gin.RouterGroup) {
	catalog := router.Group("/catalog")
	{
		catalog.GET("", handlers.ListProducts)
		catalog.GET("/home", handlers.GetHomeData)
		catalog.GET("/hot", handlers.GetHotProducts)
		catalog.GET("/new", handlers.GetNewProducts)
		catalog.GET("/categories", handlers.GetCategories)
		catalog.GET("/search", handlers.SearchProducts)
		catalog.GET("/:id", handlers.GetProductByID)
		catalog.GET("/:id/groups", handlers.GetGroupsBySKU)
		catalog.GET("/scenarios", handlers.GetScenarios)
		catalog.GET("/scenarios/:scenario/spus", handlers.GetSPUsByScenario)
		catalog.GET("/spus/:id/performance", handlers.GetSPUPerformance)
		catalog.GET("/fuel-station-config", handlers.GetFuelStationConfig)

		merchants := catalog.Group("/merchants")
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
		tokens.GET("/lots", handlers.GetTokenLots)
		tokens.GET("/consumption", handlers.GetTokenConsumption)
		tokens.POST("/transfer", handlers.TransferTokens)

		tokens.GET("/recharge/packages", handlers.GetRechargePackages)
		tokens.POST("/recharge", handlers.CreateRechargeOrder)
		tokens.GET("/recharge/orders", handlers.GetRechargeOrders)
		tokens.POST("/recharge/orders/:id/mock-pay", handlers.MockCompleteRechargeOrder)
		tokens.GET("/recharge/orders/:id", handlers.GetRechargeOrder)

		keys := tokens.Group("/keys")
		{
			keys.GET("", handlers.ListAPIKeys)
			keys.POST("", handlers.CreateAPIKey)
			keys.PUT("/:id", handlers.UpdateAPIKey)
			keys.DELETE("/:id", handlers.DeleteAPIKey)
		}

		tokens.GET("/api-usage-guide", handlers.GetAPIUsageGuide)
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
		authMerchants.GET("/model-providers", handlers.GetMerchantModelProviders)

		apiKeys := authMerchants.Group("/api-keys")
		{
			apiKeys.GET("", handlers.ListMerchantAPIKeys)
			apiKeys.POST("", handlers.CreateMerchantAPIKey)
			apiKeys.PUT("/:id", handlers.UpdateMerchantAPIKey)
			apiKeys.DELETE("/:id", handlers.DeleteMerchantAPIKey)
			apiKeys.GET("/usage", handlers.GetMerchantAPIKeyUsage)
			apiKeys.POST("/:id/verify", handlers.VerifyMerchantAPIKey)
			apiKeys.POST("/:id/health-check", handlers.TriggerMerchantAPIKeyHealthCheck)
			apiKeys.GET("/:id/verification", handlers.GetMerchantAPIKeyVerification)
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
	admin.Use(middleware.AdminIPAllowlistMiddleware())
	admin.Use(middleware.AuthMiddleware())
	{
		admin.GET("/users", handlers.GetAdminUsers)
		admin.POST("/users", handlers.CreateAdminUser)
		admin.GET("/stats", handlers.GetAdminStats)
		admin.GET("/merchants/pending", handlers.GetPendingMerchants)
		admin.GET("/merchants/audit-logs", handlers.GetMerchantAuditLogs)
		admin.GET("/merchants/byok-summary", handlers.GetAdminBYOKSummary)
		admin.GET("/byok-routing", handlers.GetBYOKRoutingList)
		admin.PUT("/byok-routing/:id/route-config", handlers.UpdateBYOKRouteConfig)
		admin.POST("/byok-routing/:id/probe", handlers.TriggerBYOKProbe)
		admin.POST("/byok-routing/:id/light-verify", handlers.LightVerifyBYOK)
		admin.POST("/byok-routing/:id/deep-verify", handlers.DeepVerifyBYOK)
		admin.GET("/byok-routing/:id/verification", handlers.GetBYOKVerificationDetails)
		admin.GET("/merchants", handlers.GetAdminMerchants)
		admin.PATCH("/merchants/:id", handlers.PatchAdminMerchant)
		admin.PATCH("/merchants/:id/lifecycle", handlers.PatchAdminMerchantLifecycle)
		admin.POST("/merchants/:id/approve", handlers.ApproveMerchant)
		admin.POST("/merchants/:id/reject", handlers.RejectMerchant)
		admin.GET("/merchant-invites", handlers.ListMerchantInvites)
		admin.POST("/merchant-invites", handlers.CreateMerchantInvite)
		admin.POST("/merchant-invites/:id/revoke", handlers.RevokeMerchantInvite)

		admin.GET("/spus", handlers.ListSPUs)
		admin.GET("/spus/:id", handlers.GetSPUByID)
		admin.POST("/spus", handlers.CreateSPU)
		admin.PUT("/spus/:id", handlers.UpdateSPU)
		admin.DELETE("/spus/:id", handlers.DeleteSPU)
		admin.GET("/spus/:id/scenarios", handlers.GetSPUScenarios)
		admin.PUT("/spus/:id/scenarios", handlers.UpdateSPUScenarios)

		admin.GET("/skus", handlers.ListSKUs)
		admin.GET("/skus/:id", handlers.GetSKUByID)
		admin.POST("/skus", handlers.CreateSKU)
		admin.PUT("/skus/:id", handlers.UpdateSKU)
		admin.DELETE("/skus/:id", handlers.DeleteSKU)

		admin.GET("/model-providers/all", handlers.ListAllModelProviders)
		admin.GET("/model-providers", handlers.GetModelProviders)
		admin.POST("/model-providers", handlers.CreateModelProvider)
		admin.PATCH("/model-providers/:id", handlers.PatchModelProvider)

		admin.GET("/provider-models", handlers.ListProviderModels)
		admin.POST("/model-providers/:code/sync-models", handlers.SyncProviderModels)

		admin.GET("/route-configs/providers", handlers.GetProviderRouteConfigs)
		admin.GET("/route-configs/providers/:code", handlers.GetProviderRouteConfig)
		admin.PUT("/route-configs/providers/:code", handlers.UpdateProviderRouteConfig)
		admin.POST("/route-configs/providers/:code/probe-endpoint", handlers.ProbeEndpoint)
		admin.GET("/route-configs/merchants", handlers.GetMerchantRouteConfigs)
		admin.PUT("/route-configs/merchants/:id", handlers.UpdateMerchantRouteConfig)
		admin.POST("/route-configs/test", handlers.TestRouteDecision)

		admin.GET("/model-catalog-keys", handlers.GetAdminCatalogModelKeys)
		admin.GET("/model-fallback-rules", handlers.ListModelFallbackRules)
		admin.POST("/model-fallback-rules", handlers.CreateModelFallbackRule)
		admin.PATCH("/model-fallback-rules/:id", handlers.PatchModelFallbackRule)
		admin.DELETE("/model-fallback-rules/:id", handlers.DeleteModelFallbackRule)

		admin.GET("/platform-settings", handlers.GetAdminPlatformSettings)
		admin.PUT("/platform-settings", handlers.UpdateAdminPlatformSettings)
		admin.GET("/fuel-station-config", handlers.GetAdminFuelStationConfig)
		admin.PUT("/fuel-station-config", handlers.UpdateAdminFuelStationConfig)
		admin.GET("/fuel-station-templates", handlers.GetAdminFuelStationTemplateLibrary)
		admin.PUT("/fuel-station-templates", handlers.UpdateAdminFuelStationTemplateLibrary)
		admin.GET("/entitlement-packages", handlers.ListAdminEntitlementPackages)
		admin.POST("/entitlement-packages", handlers.CreateAdminEntitlementPackage)
		admin.PUT("/entitlement-packages/:id", handlers.UpdateAdminEntitlementPackage)
		admin.DELETE("/entitlement-packages/:id", handlers.DeleteAdminEntitlementPackage)
	}
}

func RegisterFlashSaleRoutes(router *gin.RouterGroup) {
	flashSales := router.Group("/flash-sales")
	{
		flashSales.GET("/active", handlers.GetActiveFlashSales)
		flashSales.GET("/:id/skus", handlers.GetFlashSaleProducts)
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
		favorites.DELETE("/:sku_id", handlers.RemoveFavorite)
		favorites.GET("/check/:sku_id", handlers.CheckFavorite)
	}
}

func RegisterBrowseHistoryRoutes(router *gin.RouterGroup) {
	history := router.Group("/browse-history")
	history.Use(middleware.AuthMiddleware())
	{
		history.GET("", handlers.GetBrowseHistory)
		history.POST("", handlers.AddBrowseHistory)
		history.DELETE("", handlers.ClearBrowseHistory)
		history.DELETE("/:sku_id", handlers.RemoveBrowseHistoryItem)
	}
}

func RegisterSKURoutes(router *gin.RouterGroup) {
	skus := router.Group("/skus")
	{
		skus.GET("", handlers.ListPublicSKUs)
		skus.GET("/:id", handlers.GetPublicSKUByID)
	}
	router.GET("/entitlement-packages/stats", handlers.BatchGetEntitlementPackageStats)
	router.GET("/entitlement-packages", handlers.ListPublicEntitlementPackages)

	epSocial := router.Group("/entitlement-packages")
	epSocial.Use(middleware.AuthMiddleware())
	{
		epSocial.POST("/:id/favorite", handlers.AddEntitlementPackageFavorite)
		epSocial.DELETE("/:id/favorite", handlers.RemoveEntitlementPackageFavorite)
		epSocial.POST("/:id/like", handlers.ToggleEntitlementPackageLike)
		epSocial.POST("/:id/reviews", handlers.UpsertEntitlementPackageReview)
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
	myEntitlements := router.Group("/users/me/entitlements")
	myEntitlements.Use(middleware.AuthMiddleware())
	{
		myEntitlements.GET("/packages", handlers.GetMyEntitlementPackages)
	}
}

func RegisterSettlementRoutes(router *gin.RouterGroup) {
	merchantSettlements := router.Group("/merchant/settlements")
	merchantSettlements.Use(middleware.AuthMiddleware())
	{
		merchantSettlements.GET("/:id", handlers.GetMerchantSettlementByID)
		merchantSettlements.GET("/:id/items", handlers.GetSettlementItems)
		merchantSettlements.POST("/:id/confirm", handlers.ConfirmSettlement)
		merchantSettlements.POST("/:id/dispute", handlers.SubmitSettlementDispute)
	}

	adminSettlements := router.Group("/admin/settlements")
	adminSettlements.Use(middleware.AuthMiddleware())
	{
		adminSettlements.GET("", handlers.AdminGetSettlements)
		adminSettlements.POST("/generate", handlers.AdminGenerateMonthlySettlements)
		adminSettlements.POST("/generate/merchant", handlers.AdminGenerateSettlementForMerchant)
		adminSettlements.GET("/:id", handlers.AdminGetSettlementByID)
		adminSettlements.GET("/:id/items", handlers.AdminGetSettlementItems)
		adminSettlements.POST("/:id/approve", handlers.AdminApproveSettlement)
		adminSettlements.POST("/:id/mark-paid", handlers.AdminMarkSettlementPaid)
		adminSettlements.POST("/:id/reconcile", handlers.AdminReconcileSettlement)
	}

	adminBillings := router.Group("/admin/billings")
	adminBillings.Use(middleware.AuthMiddleware())
	{
		adminBillings.GET("", handlers.AdminGetBillings)
		adminBillings.GET("/stats", handlers.AdminGetBillingStats)
		adminBillings.GET("/trends", handlers.AdminGetBillingTrends)
		adminBillings.GET("/export", handlers.AdminExportBillings)
		adminBillings.GET("/providers", handlers.AdminGetBillingProviders)
		adminBillings.GET("/models", handlers.AdminGetBillingModels)
	}

	adminUserBillings := router.Group("/admin/user-billings")
	adminUserBillings.Use(middleware.AuthMiddleware())
	{
		adminUserBillings.GET("", handlers.AdminGetUserBillings)
		adminUserBillings.GET("/export", handlers.AdminExportUserBillings)
	}

	adminReconciliation := router.Group("/admin/reconciliation")
	adminReconciliation.Use(middleware.AuthMiddleware())
	{
		adminReconciliation.GET("/ledger", handlers.AdminGetLedgerReconciliation)
		adminReconciliation.GET("/ledger/drift/export", handlers.AdminExportLedgerDriftCSV)
		adminReconciliation.GET("/ledger/drift", handlers.AdminGetLedgerDrift)
		adminReconciliation.POST("/ledger/check", handlers.AdminPostLedgerCheck)
		adminReconciliation.GET("/gmv/trends", handlers.AdminGetGMVTrends)
		adminReconciliation.GET("/gmv", handlers.AdminGetGMVReport)
	}

	merchantBillings := router.Group("/merchant/billings")
	merchantBillings.Use(middleware.AuthMiddleware())
	{
		merchantBillings.GET("", handlers.MerchantGetBillings)
		merchantBillings.GET("/export", handlers.MerchantExportBillings)
	}

	adminDisputes := router.Group("/admin/disputes")
	adminDisputes.Use(middleware.AuthMiddleware())
	{
		adminDisputes.POST("/:id/process", handlers.AdminProcessDispute)
	}

	adminRoutingStrategies := router.Group("/admin/routing-strategies")
	adminRoutingStrategies.Use(middleware.AuthMiddleware())
	{
		adminRoutingStrategies.GET("", handlers.AdminGetRoutingStrategies)
		adminRoutingStrategies.GET("/:id", handlers.AdminGetRoutingStrategy)
		adminRoutingStrategies.POST("", handlers.AdminCreateRoutingStrategy)
		adminRoutingStrategies.POST("/test", handlers.TestRoutingStrategy)
		adminRoutingStrategies.PUT("/:id", handlers.AdminUpdateRoutingStrategy)
		adminRoutingStrategies.DELETE("/:id", handlers.AdminDeleteRoutingStrategy)
	}

	adminAPIKeyStatus := router.Group("/admin/api-key-status")
	adminAPIKeyStatus.Use(middleware.AuthMiddleware())
	{
		adminAPIKeyStatus.GET("", handlers.GetAllAPIKeyStatus)
		adminAPIKeyStatus.GET("/:id", handlers.GetAPIKeyStatus)
		adminAPIKeyStatus.GET("/merchant", handlers.GetAPIKeyStatusByMerchantID)
		adminAPIKeyStatus.POST("/batch", handlers.GetBatchAPIKeyStatus)
		adminAPIKeyStatus.POST("/collect", handlers.TriggerStatusCollect)
	}

	adminRouteDecisionLogs := router.Group("/admin/route-decision-logs")
	adminRouteDecisionLogs.Use(middleware.AuthMiddleware())
	{
		adminRouteDecisionLogs.GET("", handlers.GetRouteDecisionLogs)
		adminRouteDecisionLogs.GET("/:id", handlers.GetRouteDecisionLog)
	}

	adminGateway := router.Group("/admin/gateway")
	adminGateway.Use(middleware.AuthMiddleware())
	{
		adminGateway.GET("/stats", handlers.GetGatewayStats)
	}

	adminRateLimiter := router.Group("/admin/rate-limiter")
	adminRateLimiter.Use(middleware.AuthMiddleware())
	{
		adminRateLimiter.GET("/stats", handlers.GetRateLimiterStats)
		adminRateLimiter.PUT("/config", handlers.UpdateRateLimiterConfig)
		adminRateLimiter.POST("/reset", handlers.ResetRateLimiterStats)
	}

	adminQueue := router.Group("/admin/queue")
	adminQueue.Use(middleware.AuthMiddleware())
	{
		adminQueue.GET("/stats", handlers.GetQueueStats)
		adminQueue.PUT("/config", handlers.UpdateQueueConfig)
		adminQueue.POST("/reset", handlers.ResetQueueStats)
	}
}
