package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/handlers"
)

// RegisterUserRoutes registers user-related routes
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

// RegisterProductRoutes registers product-related routes
func RegisterProductRoutes(router *gin.RouterGroup) {
	products := router.Group("/products")
	{
		// Read operations
		products.GET("", handlers.ListProducts)
		products.GET("/:id", handlers.GetProductByID)
		products.GET("/search", handlers.SearchProducts)

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
		payments.POST("", handlers.InitiatePayment)
		payments.GET("/:id", handlers.GetPaymentByID)
		payments.POST("/:id/refund", handlers.RefundPayment)

		// Webhooks
		webhooks := payments.Group("/webhooks")
		{
			webhooks.POST("/alipay", handlers.HandleAlipayCallback)
			webhooks.POST("/wechat", handlers.HandleWechatCallback)
		}
	}
}
