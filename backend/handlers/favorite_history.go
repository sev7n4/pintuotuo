package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
)

func GetFavorites(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int
	err := db.QueryRow(
		`SELECT (SELECT COUNT(*)::int FROM favorites WHERE user_id = $1)
		       + (SELECT COUNT(*)::int FROM entitlement_package_favorites WHERE user_id = $1)`,
		userID,
	).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to count favorites",
		})
		return
	}

	mergedRows, err := db.Query(
		`SELECT id, item_type, created_at FROM (
			SELECT f.id, 'sku'::text AS item_type, f.created_at
			  FROM favorites f
			 WHERE f.user_id = $1
			UNION ALL
			SELECT epf.id, 'entitlement_package'::text, epf.created_at
			  FROM entitlement_package_favorites epf
			 WHERE epf.user_id = $1
		) u
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, pageSize, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch favorites",
		})
		return
	}
	defer mergedRows.Close()

	type merged struct {
		id        int
		itemType  string
		createdAt time.Time
	}
	order := make([]merged, 0)
	var skuFavIDs []int
	var epFavIDs []int
	for mergedRows.Next() {
		var m merged
		if scanErr := mergedRows.Scan(&m.id, &m.itemType, &m.createdAt); scanErr != nil {
			continue
		}
		order = append(order, m)
		switch m.itemType {
		case "sku":
			skuFavIDs = append(skuFavIDs, m.id)
		case "entitlement_package":
			epFavIDs = append(epFavIDs, m.id)
		}
	}
	if err := mergedRows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch favorites",
		})
		return
	}

	skuByFavID := map[int]models.UnifiedFavoriteResponse{}
	if len(skuFavIDs) > 0 {
		q := `
			SELECT f.id, f.sku_id, f.created_at,
				   s.id, COALESCE(s.merchant_id, 0), s.spu_id, sp.name || ' · ' || s.sku_code, COALESCE(sp.description, ''),
				   s.retail_price, COALESCE(s.original_price, s.retail_price),
				   CASE WHEN s.stock = -1 THEN 999999 ELSE s.stock END, COALESCE(s.sales_count, 0), COALESCE(sp.model_tier, ''),
				   s.status, s.created_at, s.updated_at
			  FROM favorites f
			  JOIN skus s ON f.sku_id = s.id
			  JOIN spus sp ON s.spu_id = sp.id
			 WHERE f.user_id = $1 AND f.id = ANY($2)
		`
		srows, qerr := db.Query(q, userID, pq.Array(skuFavIDs))
		if qerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch favorites",
			})
			return
		}
		for srows.Next() {
			var favID, skuID int
			var product models.Product
			var createdAt sql.NullTime
			if scanErr := srows.Scan(
				&favID, &skuID, &createdAt,
				&product.ID, &product.MerchantID, &product.SpuID, &product.Name, &product.Description,
				&product.Price, &product.OriginalPrice, &product.Stock, &product.SoldCount,
				&product.Category, &product.Status, &product.CreatedAt, &product.UpdatedAt,
			); scanErr != nil {
				continue
			}
			u := models.UnifiedFavoriteResponse{
				ItemType: "sku",
				ID:       favID,
				SKUID:    &skuID,
				Product:  &product,
			}
			if createdAt.Valid {
				u.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z07:00")
			}
			skuByFavID[favID] = u
		}
		_ = srows.Close()
	}

	epByFavID := map[int]models.UnifiedFavoriteResponse{}
	if len(epFavIDs) > 0 {
		q := `
			SELECT epf.id, epf.created_at, ep.id, ep.package_code, ep.name,
			       COALESCE(ep.marketing_line, ''), ep.status
			  FROM entitlement_package_favorites epf
			  JOIN entitlement_packages ep ON ep.id = epf.package_id
			 WHERE epf.user_id = $1 AND epf.id = ANY($2)
		`
		erows, qerr := db.Query(q, userID, pq.Array(epFavIDs))
		if qerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch favorites",
			})
			return
		}
		for erows.Next() {
			var favID int
			var createdAt sql.NullTime
			var epBrief models.EntitlementPackageFavoriteBrief
			if scanErr := erows.Scan(
				&favID, &createdAt,
				&epBrief.ID, &epBrief.PackageCode, &epBrief.Name, &epBrief.MarketingLine, &epBrief.Status,
			); scanErr != nil {
				continue
			}
			epid := epBrief.ID
			u := models.UnifiedFavoriteResponse{
				ItemType:             "entitlement_package",
				ID:                   favID,
				EntitlementPackageID: &epid,
				EntitlementPackage:   &epBrief,
			}
			if createdAt.Valid {
				u.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z07:00")
			}
			epByFavID[favID] = u
		}
		_ = erows.Close()
	}

	items := make([]models.UnifiedFavoriteResponse, 0, len(order))
	for _, row := range order {
		switch row.itemType {
		case "sku":
			if u, ok := skuByFavID[row.id]; ok {
				items = append(items, u)
			}
		case "entitlement_package":
			if u, ok := epByFavID[row.id]; ok {
				items = append(items, u)
			}
		default:
			// ignore unknown
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":      items,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + pageSize - 1) / pageSize,
		},
	})
}

func AddFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		SKUID int `json:"sku_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	if req.SKUID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "sku_id is required",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	var productExists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE s.id = $1 AND s.status = 'active' AND sp.status = 'active')`, req.SKUID).Scan(&productExists)
	if err != nil || !productExists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Product not found",
		})
		return
	}

	var existingID int
	err = db.QueryRow(`SELECT id FROM favorites WHERE user_id = $1 AND sku_id = $2`, userID, req.SKUID).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Already in favorites",
			"data": gin.H{
				"id":     existingID,
				"sku_id": req.SKUID,
			},
		})
		return
	}

	insertQuery := `
		INSERT INTO favorites (user_id, sku_id, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id
	`
	var newID int
	err = db.QueryRow(insertQuery, userID, req.SKUID).Scan(&newID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to add favorite",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "Added to favorites",
		"data": gin.H{
			"id":     newID,
			"sku_id": req.SKUID,
		},
	})
}

func RemoveFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	skuIDStr := c.Param("sku_id")
	skuID, err := strconv.Atoi(skuIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid sku_id",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	query := `DELETE FROM favorites WHERE user_id = $1 AND sku_id = $2`
	result, err := db.Exec(query, userID, skuID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to remove favorite",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Favorite not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Removed from favorites",
	})
}

func CheckFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	skuIDStr := c.Param("sku_id")
	skuID, err := strconv.Atoi(skuIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid sku_id",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	var isFavorite bool
	err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND sku_id = $2)`, userID, skuID).Scan(&isFavorite)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to check favorite",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"is_favorite": isFavorite,
		},
	})
}

func GetBrowseHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM browse_history WHERE user_id = $1`
	var total int
	err := db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to count browse history",
		})
		return
	}

	query := `
		SELECT bh.id, bh.sku_id, bh.view_count, bh.updated_at,
			   s.id, COALESCE(s.merchant_id, 0), s.spu_id, sp.name || ' · ' || s.sku_code, COALESCE(sp.description, ''),
			   s.retail_price, COALESCE(s.original_price, s.retail_price),
			   CASE WHEN s.stock = -1 THEN 999999 ELSE s.stock END, COALESCE(s.sales_count, 0), COALESCE(sp.model_tier, ''),
			   s.status, s.created_at, s.updated_at
		FROM browse_history bh
		JOIN skus s ON bh.sku_id = s.id
		JOIN spus sp ON s.spu_id = sp.id
		WHERE bh.user_id = $1
		ORDER BY bh.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, userID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch browse history",
		})
		return
	}
	defer rows.Close()

	var items []models.BrowseHistoryResponse
	for rows.Next() {
		var item models.BrowseHistoryResponse
		var product models.Product
		var updatedAt sql.NullTime

		err := rows.Scan(
			&item.ID, &item.SKUID, &item.ViewCount, &updatedAt,
			&product.ID, &product.MerchantID, &product.SpuID, &product.Name, &product.Description,
			&product.Price, &product.OriginalPrice, &product.Stock, &product.SoldCount,
			&product.Category, &product.Status, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if updatedAt.Valid {
			item.ViewedAt = updatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		item.Product = product
		items = append(items, item)
	}

	if items == nil {
		items = []models.BrowseHistoryResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":      items,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + pageSize - 1) / pageSize,
		},
	})
}

func AddBrowseHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		SKUID int `json:"sku_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	if req.SKUID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "sku_id is required",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	var productExists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE s.id = $1 AND s.status = 'active' AND sp.status = 'active')`, req.SKUID).Scan(&productExists)
	if err != nil || !productExists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Product not found",
		})
		return
	}

	upsertQuery := `
		INSERT INTO browse_history (user_id, sku_id, view_count, created_at, updated_at)
		VALUES ($1, $2, 1, NOW(), NOW())
		ON CONFLICT (user_id, sku_id) 
		DO UPDATE SET view_count = browse_history.view_count + 1, updated_at = NOW()
	`
	_, err = db.Exec(upsertQuery, userID, req.SKUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to add browse history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Browse history recorded",
	})
}

func ClearBrowseHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	query := `DELETE FROM browse_history WHERE user_id = $1`
	_, err := db.Exec(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to clear browse history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Browse history cleared",
	})
}

func RemoveBrowseHistoryItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	skuIDStr := c.Param("sku_id")
	skuID, err := strconv.Atoi(skuIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid sku_id",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	query := `DELETE FROM browse_history WHERE user_id = $1 AND sku_id = $2`
	result, err := db.Exec(query, userID, skuID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to remove browse history item",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Browse history item not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Browse history item removed",
	})
}
