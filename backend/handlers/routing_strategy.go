package handlers

import (
	"net/http"
	"strconv"

	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"

	"github.com/gin-gonic/gin"
)

func AdminGetRoutingStrategies(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 20
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	service := services.NewRoutingStrategyService(db)
	strategies, total, err := service.GetStrategies(page, pageSize)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"STRATEGIES_QUERY_FAILED",
			"Failed to query strategies",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"strategies": strategies,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

func AdminGetRoutingStrategy(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_STRATEGY_ID",
			"Invalid strategy ID",
			http.StatusBadRequest,
			err,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	service := services.NewRoutingStrategyService(db)
	strategy, err := service.GetStrategyByID(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"STRATEGY_NOT_FOUND",
			"Strategy not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{"strategy": strategy})
}

func AdminCreateRoutingStrategy(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	var req services.RoutingStrategyConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"Invalid request body",
			http.StatusBadRequest,
			err,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	service := services.NewRoutingStrategyService(db)
	if err := service.CreateStrategy(&req); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"STRATEGY_CREATION_FAILED",
			"Failed to create strategy",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"strategy": req})
}

func AdminUpdateRoutingStrategy(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_STRATEGY_ID",
			"Invalid strategy ID",
			http.StatusBadRequest,
			err,
		))
		return
	}

	var req services.RoutingStrategyConfig
	if err = c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"Invalid request body",
			http.StatusBadRequest,
			err,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	service := services.NewRoutingStrategyService(db)
	err = service.UpdateStrategy(id, &req)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"STRATEGY_UPDATE_FAILED",
			"Failed to update strategy",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	updated, err := service.GetStrategyByID(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"STRATEGY_QUERY_FAILED",
			"Failed to load updated strategy",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{"strategy": updated})
}

func AdminDeleteRoutingStrategy(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_STRATEGY_ID",
			"Invalid strategy ID",
			http.StatusBadRequest,
			err,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	service := services.NewRoutingStrategyService(db)
	err = service.DeleteStrategy(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"STRATEGY_DELETION_FAILED",
			"Failed to delete strategy",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Strategy deleted successfully"})
}
