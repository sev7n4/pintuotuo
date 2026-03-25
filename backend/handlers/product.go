package handlers

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services/product"
)

// Initialize product service
var productService product.Service

func initProductService() {
	if productService == nil {
		logger := log.New(os.Stderr, "[ProductHandler] ", log.LstdFlags)
		productService = product.NewService(config.GetDB(), logger)
	}
}

// ListProducts retrieves product list with pagination
func ListProducts(c *gin.Context) {
	initProductService()

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	status := c.DefaultQuery("status", "active")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	params := &product.ListProductsParams{
		Page:    pageNum,
		PerPage: perPageNum,
		Status:  status,
	}

	result, err := productService.ListProducts(c.Request.Context(), params)
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

// GetProductByID retrieves a single product by ID with caching
func GetProductByID(c *gin.Context) {
	initProductService()

	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	prod, err := productService.GetProductByID(c.Request.Context(), idInt)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		}
		return
	}

	// Map service product to model
	c.JSON(http.StatusOK, mapServiceProductToModel(prod))
}

// SearchProducts searches for products by query
func SearchProducts(c *gin.Context) {
	initProductService()

	query := c.Query("q")
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	params := &product.SearchProductsParams{
		Query:   query,
		Page:    pageNum,
		PerPage: perPageNum,
	}

	result, err := productService.SearchProducts(c.Request.Context(), params)
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

// CreateProduct creates a new product (merchant only)
func CreateProduct(c *gin.Context) {
	initProductService()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req product.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	prod, err := productService.CreateProduct(c.Request.Context(), userIDInt, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"PRODUCT_CREATION_FAILED",
				"Failed to create product",
				http.StatusInternalServerError,
				err,
			))
		}
		return
	}

	c.JSON(http.StatusCreated, mapServiceProductToModel(prod))
}

// UpdateProduct updates a product (merchant only)
func UpdateProduct(c *gin.Context) {
	initProductService()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	var req product.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Get user's merchant ID
	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	prod, err := productService.UpdateProduct(c.Request.Context(), userIDInt, idInt, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"PRODUCT_UPDATE_FAILED",
				"Failed to update product",
				http.StatusInternalServerError,
				err,
			))
		}
		return
	}

	c.JSON(http.StatusOK, mapServiceProductToModel(prod))
}

// DeleteProduct deletes a product (merchant only)
func DeleteProduct(c *gin.Context) {
	initProductService()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	err = productService.DeleteProduct(c.Request.Context(), userIDInt, idInt)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"PRODUCT_DELETE_FAILED",
				"Failed to delete product",
				http.StatusInternalServerError,
				err,
			))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// Helper function to convert string ID to int
func idToInt(id string) int {
	idInt, _ := strconv.Atoi(id)
	return idInt
}

// mapServiceProductToModel maps service product to model product
func mapServiceProductToModel(p *product.Product) gin.H {
	return gin.H{
		"id":              p.ID,
		"merchant_id":     p.MerchantID,
		"name":            p.Name,
		"description":     p.Description,
		"price":           p.Price,
		"original_price":  p.OriginalPrice,
		"stock":           p.Stock,
		"status":          p.Status,
		"created_at":      p.CreatedAt,
		"updated_at":      p.UpdatedAt,
	}
}
