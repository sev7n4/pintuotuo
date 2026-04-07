// AdminGetBillingProviders returns distinct providers from api_usage_logs
func AdminGetBillingProviders(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	billingService := services.NewBillingService(db)
	 providers, err := billingService.GetProviders()
	if err != nil {
        middleware.RespondWithError(c, apperrors.NewAppError(
            "PROVIDERS_QUERY_FAILED",
            "Failed to get providers",
            http.StatusInternalServerError,
            err,
        ))
        return
    }

	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// AdminGetBillingModels returns distinct models for a provider filter
func AdminGetBillingModels(c *gin.Context) {
	if !requireAdminRole(c) {
        return
    }

	provider := c.Query("provider")
	db := config.GetDB()
	if db == nil {
        middleware.RespondWithError(c, apperrors.ErrDatabaseError)
        return
    }

	billingService := services.NewBillingService(db)
    models, err := billingService.GetModels(provider)
	if err != nil {
        middleware.RespondWithError(c, apperrors.NewAppError(
            "MODELS_QUERY_FAILED",
            "Failed to get models",
            http.StatusInternalServerError,
            err,
        ))
        return
    }

	c.JSON(http.StatusOK, gin.H{"models": models})
