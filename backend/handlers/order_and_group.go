package handlers

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services/group"
	"github.com/pintuotuo/backend/services/order"
)

// Initialize services
var (
	orderService order.Service
	groupService group.Service
)

func initOrderAndGroupServices() {
	if orderService == nil {
		logger := log.New(os.Stderr, "[OrderHandler] ", log.LstdFlags)
		orderService = order.NewService(config.GetDB(), logger)
	}
	if groupService == nil {
		logger := log.New(os.Stderr, "[GroupHandler] ", log.LstdFlags)
		groupService = group.NewService(config.GetDB(), logger)
	}
}

// CreateOrder creates a new order
func CreateOrder(c *gin.Context) {
	initOrderAndGroupServices()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt := userID.(int)

	var req order.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Call service
	o, err := orderService.CreateOrder(c.Request.Context(), userIDInt, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusCreated, o)
}

// ListOrders lists all orders for current user
func ListOrders(c *gin.Context) {
	initOrderAndGroupServices()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt := userID.(int)

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	status := c.DefaultQuery("status", "all")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	params := &order.ListOrdersParams{
		Page:    pageNum,
		PerPage: perPageNum,
		Status:  status,
	}

	result, err := orderService.ListOrders(c.Request.Context(), userIDInt, params)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetOrderByID retrieves an order by ID
func GetOrderByID(c *gin.Context) {
	initOrderAndGroupServices()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt := userID.(int)
	id := c.Param("id")

	orderID, err := strconv.Atoi(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	o, err := orderService.GetOrderByID(c.Request.Context(), userIDInt, orderID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrOrderNotFound)
		}
		return
	}

	c.JSON(http.StatusOK, o)
}

// CancelOrder cancels an order
func CancelOrder(c *gin.Context) {
	initOrderAndGroupServices()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt := userID.(int)
	id := c.Param("id")

	orderID, err := strconv.Atoi(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	o, err := orderService.CancelOrder(c.Request.Context(), userIDInt, orderID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, o)
}

// CreateGroup creates a new group purchase
func CreateGroup(c *gin.Context) {
	initOrderAndGroupServices()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt := userID.(int)

	var req group.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Call service
	grp, err := groupService.CreateGroup(c.Request.Context(), userIDInt, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusCreated, grp)
}

// ListGroups lists all active groups
func ListGroups(c *gin.Context) {
	initOrderAndGroupServices()

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	status := c.DefaultQuery("status", "active")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	params := &group.ListGroupsParams{
		Page:    pageNum,
		PerPage: perPageNum,
		Status:  status,
	}

	result, err := groupService.ListGroups(c.Request.Context(), params)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetGroupByID retrieves a group by ID
func GetGroupByID(c *gin.Context) {
	initOrderAndGroupServices()

	id := c.Param("id")

	groupID, err := strconv.Atoi(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	grp, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrGroupNotFound)
		}
		return
	}

	c.JSON(http.StatusOK, grp)
}

// JoinGroup adds current user to a group
func JoinGroup(c *gin.Context) {
	initOrderAndGroupServices()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt := userID.(int)
	id := c.Param("id")

	groupID, err := strconv.Atoi(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Call service
	grp, err := groupService.JoinGroup(c.Request.Context(), userIDInt, groupID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, grp)
}

// CancelGroup cancels a group (creator only)
func CancelGroup(c *gin.Context) {
	initOrderAndGroupServices()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt := userID.(int)
	id := c.Param("id")

	groupID, err := strconv.Atoi(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Call service
	err = groupService.CancelGroup(c.Request.Context(), userIDInt, groupID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group cancelled successfully"})
}

// GetGroupProgress retrieves group progress
func GetGroupProgress(c *gin.Context) {
	initOrderAndGroupServices()

	id := c.Param("id")

	groupID, err := strconv.Atoi(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	grp, err := groupService.GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrGroupNotFound)
		}
		return
	}

	// Return progress info
	progress := gin.H{
		"id":           grp.ID,
		"product_id":   grp.ProductID,
		"target_count": grp.TargetCount,
		"current_count": grp.CurrentCount,
		"percentage":   (float64(grp.CurrentCount) / float64(grp.TargetCount)) * 100,
		"status":       grp.Status,
		"deadline":     grp.Deadline,
	}

	c.JSON(http.StatusOK, progress)
}
