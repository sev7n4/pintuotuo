package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

// GetModels returns all model categories (level=1)
func GetModels(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.NewAppError("DB_ERROR", "Database not available", http.StatusInternalServerError, nil))
		return
	}

	rows, err := db.Query(`
		SELECT id, name, level, description, icon, sort_order, is_active, created_at, updated_at
		FROM categories
		WHERE level = 1 AND is_active = true
		ORDER BY sort_order ASC
	`)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError("QUERY_ERROR", "Failed to query models", http.StatusInternalServerError, err))
		return
	}
	defer rows.Close()

	var modelList []models.Category
	for rows.Next() {
		var m models.Category
		err := rows.Scan(&m.ID, &m.Name, &m.Level, &m.Description, &m.Icon, &m.SortOrder, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			continue
		}
		modelList = append(modelList, m)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    modelList,
	})
}

// GetPackages returns all package categories (level=2)
func GetPackages(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.NewAppError("DB_ERROR", "Database not available", http.StatusInternalServerError, nil))
		return
	}

	rows, err := db.Query(`
		SELECT id, name, level, description, icon, sort_order, is_active, created_at, updated_at
		FROM categories
		WHERE level = 2 AND is_active = true
		ORDER BY sort_order ASC
	`)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError("QUERY_ERROR", "Failed to query packages", http.StatusInternalServerError, err))
		return
	}
	defer rows.Close()

	var packageList []models.Category
	for rows.Next() {
		var p models.Category
		err := rows.Scan(&p.ID, &p.Name, &p.Level, &p.Description, &p.Icon, &p.SortOrder, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		packageList = append(packageList, p)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    packageList,
	})
}

// GetAllCategories returns all categories (admin only)
func GetAllCategories(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.NewAppError("DB_ERROR", "Database not available", http.StatusInternalServerError, nil))
		return
	}

	levelStr := c.Query("level")
	query := `
		SELECT id, name, level, description, icon, sort_order, is_active, created_at, updated_at
		FROM categories
	`
	args := []interface{}{}

	if levelStr != "" {
		level, err := strconv.Atoi(levelStr)
		if err == nil {
			query += " WHERE level = $1"
			args = append(args, level)
		}
	}

	query += " ORDER BY level ASC, sort_order ASC"

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError("QUERY_ERROR", "Failed to query categories", http.StatusInternalServerError, err))
		return
	}
	defer rows.Close()

	var categoryList []models.Category
	for rows.Next() {
		var cat models.Category
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Level, &cat.Description, &cat.Icon, &cat.SortOrder, &cat.IsActive, &cat.CreatedAt, &cat.UpdatedAt)
		if err != nil {
			continue
		}
		categoryList = append(categoryList, cat)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categoryList,
	})
}

// CreateCategoryRequest represents the request body for creating a category
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Level       int    `json:"level" binding:"required,oneof=1 2"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sort_order"`
}

// CreateCategory creates a new category (admin only)
func CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError("INVALID_REQUEST", "Invalid request body", http.StatusBadRequest, err))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.NewAppError("DB_ERROR", "Database not available", http.StatusInternalServerError, nil))
		return
	}

	var id int
	err := db.QueryRow(`
		INSERT INTO categories (name, level, description, icon, sort_order, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING id
	`, req.Name, req.Level, req.Description, req.Icon, req.SortOrder).Scan(&id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError("CREATE_ERROR", "Failed to create category", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":          id,
			"name":        req.Name,
			"level":       req.Level,
			"description": req.Description,
			"icon":        req.Icon,
			"sort_order":  req.SortOrder,
			"is_active":   true,
		},
		"message": "Category created successfully",
	})
}

// UpdateCategoryRequest represents the request body for updating a category
type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   *int   `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

// UpdateCategory updates a category (admin only)
func UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError("INVALID_ID", "Invalid category ID", http.StatusBadRequest, err))
		return
	}

	var req UpdateCategoryRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError("INVALID_REQUEST", "Invalid request body", http.StatusBadRequest, err))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.NewAppError("DB_ERROR", "Database not available", http.StatusInternalServerError, nil))
		return
	}

	query := "UPDATE categories SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}
	argIndex := 1

	if req.Name != "" {
		query += ", name = $" + strconv.Itoa(argIndex)
		args = append(args, req.Name)
		argIndex++
	}
	if req.Description != "" {
		query += ", description = $" + strconv.Itoa(argIndex)
		args = append(args, req.Description)
		argIndex++
	}
	if req.Icon != "" {
		query += ", icon = $" + strconv.Itoa(argIndex)
		args = append(args, req.Icon)
		argIndex++
	}
	if req.SortOrder != nil {
		query += ", sort_order = $" + strconv.Itoa(argIndex)
		args = append(args, *req.SortOrder)
		argIndex++
	}
	if req.IsActive != nil {
		query += ", is_active = $" + strconv.Itoa(argIndex)
		args = append(args, *req.IsActive)
		argIndex++
	}

	query += " WHERE id = $" + strconv.Itoa(argIndex)
	args = append(args, id)

	result, err := db.Exec(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError("UPDATE_ERROR", "Failed to update category", http.StatusInternalServerError, err))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.RespondWithError(c, apperrors.NewAppError("NOT_FOUND", "Category not found", http.StatusNotFound, nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category updated successfully",
	})
}

// DeleteCategory deletes a category (admin only)
func DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError("INVALID_ID", "Invalid category ID", http.StatusBadRequest, err))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.NewAppError("DB_ERROR", "Database not available", http.StatusInternalServerError, nil))
		return
	}

	result, err := db.Exec("DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError("DELETE_ERROR", "Failed to delete category", http.StatusInternalServerError, err))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.RespondWithError(c, apperrors.NewAppError("NOT_FOUND", "Category not found", http.StatusNotFound, nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category deleted successfully",
	})
}
