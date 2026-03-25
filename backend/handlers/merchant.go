package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

func RegisterMerchant(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		CompanyName     string `json:"company_name" binding:"required"`
		BusinessLicense string `json:"business_license"`
		ContactName     string `json:"contact_name"`
		ContactPhone    string `json:"contact_phone"`
		ContactEmail    string `json:"contact_email"`
		Address         string `json:"address"`
		Description     string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var existingID int
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already registered as merchant"})
		return
	}

	var merchant models.Merchant
	err = db.QueryRow(
		`INSERT INTO merchants (user_id, company_name, business_license, contact_name, contact_phone, contact_email, address, description, status) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending') 
		 RETURNING id, user_id, company_name, business_license, contact_name, contact_phone, contact_email, address, description, status, created_at, updated_at`,
		userIDInt, req.CompanyName, req.BusinessLicense, req.ContactName, req.ContactPhone, req.ContactEmail, req.Address, req.Description,
	).Scan(&merchant.ID, &merchant.UserID, &merchant.CompanyName, &merchant.BusinessLicense, &merchant.ContactName,
		&merchant.ContactPhone, &merchant.ContactEmail, &merchant.Address, &merchant.Description, &merchant.Status,
		&merchant.CreatedAt, &merchant.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_CREATE_FAILED",
			"Failed to create merchant",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	_, err = db.Exec("UPDATE users SET is_merchant = true, merchant_id = $1, role = 'merchant' WHERE id = $2", merchant.ID, userIDInt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"USER_UPDATE_FAILED",
			"Failed to update user",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusCreated, merchant)
}

func GetMerchantProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	ctx := context.Background()
	cacheKey := cache.MerchantKey(userIDInt)

	if cachedMerchant, err := cache.Get(ctx, cacheKey); err == nil {
		var merchant models.Merchant
		if err := json.Unmarshal([]byte(cachedMerchant), &merchant); err == nil {
			c.JSON(http.StatusOK, merchant)
			return
		}
	}

	db := config.GetDB()
	if db == nil {
		log.Printf("GetMerchantProfile: Database connection is nil for user %d", userIDInt)
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchant models.Merchant
	var verifiedAt sql.NullTime
	err := db.QueryRow(
		`SELECT id, user_id, company_name, business_license, contact_name, contact_phone, contact_email, address, description, logo_url, status, verified_at, created_at, updated_at 
		 FROM merchants WHERE user_id = $1`,
		userIDInt,
	).Scan(&merchant.ID, &merchant.UserID, &merchant.CompanyName, &merchant.BusinessLicense, &merchant.ContactName,
		&merchant.ContactPhone, &merchant.ContactEmail, &merchant.Address, &merchant.Description, &merchant.LogoURL,
		&merchant.Status, &verifiedAt, &merchant.CreatedAt, &merchant.UpdatedAt)

	if err == sql.ErrNoRows {
		log.Printf("GetMerchantProfile: Merchant not found for user_id %d", userIDInt)
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant profile not found",
			http.StatusNotFound,
			err,
		))
		return
	} else if err != nil {
		log.Printf("GetMerchantProfile: Database query error for user_id %d: %v", userIDInt, err)
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if verifiedAt.Valid {
		merchant.VerifiedAt = verifiedAt.Time
	}

	if merchantJSON, err := json.Marshal(merchant); err == nil {
		cache.Set(ctx, cacheKey, string(merchantJSON), cache.MerchantCacheTTL)
	}

	c.JSON(http.StatusOK, merchant)
}

func UpdateMerchantProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		CompanyName  string `json:"company_name"`
		ContactName  string `json:"contact_name"`
		ContactPhone string `json:"contact_phone"`
		ContactEmail string `json:"contact_email"`
		Address      string `json:"address"`
		Description  string `json:"description"`
		LogoURL      string `json:"logo_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchant models.Merchant
	var verifiedAt sql.NullTime
	err := db.QueryRow(
		`UPDATE merchants SET 
		 company_name = COALESCE(NULLIF($1, ''), company_name),
		 contact_name = COALESCE(NULLIF($2, ''), contact_name),
		 contact_phone = COALESCE(NULLIF($3, ''), contact_phone),
		 contact_email = COALESCE(NULLIF($4, ''), contact_email),
		 address = COALESCE(NULLIF($5, ''), address),
		 description = COALESCE(NULLIF($6, ''), description),
		 logo_url = COALESCE(NULLIF($7, ''), logo_url),
		 updated_at = CURRENT_TIMESTAMP
		 WHERE user_id = $8
		 RETURNING id, user_id, company_name, business_license, contact_name, contact_phone, contact_email, address, description, logo_url, status, verified_at, created_at, updated_at`,
		req.CompanyName, req.ContactName, req.ContactPhone, req.ContactEmail, req.Address, req.Description, req.LogoURL, userIDInt,
	).Scan(&merchant.ID, &merchant.UserID, &merchant.CompanyName, &merchant.BusinessLicense, &merchant.ContactName,
		&merchant.ContactPhone, &merchant.ContactEmail, &merchant.Address, &merchant.Description, &merchant.LogoURL,
		&merchant.Status, &verifiedAt, &merchant.CreatedAt, &merchant.UpdatedAt)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant profile not found",
			http.StatusNotFound,
			err,
		))
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if verifiedAt.Valid {
		merchant.VerifiedAt = verifiedAt.Time
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.MerchantKey(userIDInt))

	c.JSON(http.StatusOK, merchant)
}

func GetMerchantStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchantID int
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant profile not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	var totalProducts, activeProducts int
	var totalSales, monthSales float64
	var totalOrders, monthOrders int

	db.QueryRow("SELECT COUNT(*) FROM products WHERE merchant_id = $1", merchantID).Scan(&totalProducts)
	db.QueryRow("SELECT COUNT(*) FROM products WHERE merchant_id = $1 AND status = 'active'", merchantID).Scan(&activeProducts)

	db.QueryRow(
		`SELECT COALESCE(SUM(total_price), 0) FROM orders o 
		 JOIN products p ON o.product_id = p.id 
		 WHERE p.merchant_id = $1 AND o.status IN ('paid', 'completed')`,
		merchantID,
	).Scan(&totalSales)

	db.QueryRow(
		`SELECT COALESCE(SUM(total_price), 0) FROM orders o 
		 JOIN products p ON o.product_id = p.id 
		 WHERE p.merchant_id = $1 AND o.status IN ('paid', 'completed') AND o.created_at >= $2`,
		merchantID, time.Now().AddDate(0, -1, 0),
	).Scan(&monthSales)

	db.QueryRow(
		`SELECT COUNT(*) FROM orders o 
		 JOIN products p ON o.product_id = p.id 
		 WHERE p.merchant_id = $1 AND o.status IN ('paid', 'completed')`,
		merchantID,
	).Scan(&totalOrders)

	db.QueryRow(
		`SELECT COUNT(*) FROM orders o 
		 JOIN products p ON o.product_id = p.id 
		 WHERE p.merchant_id = $1 AND o.status IN ('paid', 'completed') AND o.created_at >= $2`,
		merchantID, time.Now().AddDate(0, -1, 0),
	).Scan(&monthOrders)

	c.JSON(http.StatusOK, gin.H{
		"total_products":  totalProducts,
		"active_products": activeProducts,
		"total_sales":     totalSales,
		"month_sales":     monthSales,
		"total_orders":    totalOrders,
		"month_orders":    monthOrders,
	})
}

func GetMerchantProducts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchantID int
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant profile not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	status := c.DefaultQuery("status", "")
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")

	pageNum := 1
	perPageNum := 20
	p, parseErr := parseInt(page)
	if parseErr == nil && p > 0 {
		pageNum = p
	}
	pp, parseErr2 := parseInt(perPage)
	if parseErr2 == nil && pp > 0 && pp <= 100 {
		perPageNum = pp
	}

	offset := (pageNum - 1) * perPageNum

	var rows *sql.Rows
	if status != "" && status != "all" {
		rows, err = db.Query(
			`SELECT id, merchant_id, name, description, price, COALESCE(original_price, price), stock, COALESCE(sold_count, 0), COALESCE(category, ''), status, created_at, updated_at 
			 FROM products WHERE merchant_id = $1 AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
			merchantID, status, perPageNum, offset,
		)
	} else {
		rows, err = db.Query(
			`SELECT id, merchant_id, name, description, price, COALESCE(original_price, price), stock, COALESCE(sold_count, 0), COALESCE(category, ''), status, created_at, updated_at 
			 FROM products WHERE merchant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			merchantID, perPageNum, offset,
		)
	}

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Price, &p.OriginalPrice,
			&p.Stock, &p.SoldCount, &p.Category, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		products = append(products, p)
	}

	var total int
	if status != "" && status != "all" {
		db.QueryRow("SELECT COUNT(*) FROM products WHERE merchant_id = $1 AND status = $2", merchantID, status).Scan(&total)
	} else {
		db.QueryRow("SELECT COUNT(*) FROM products WHERE merchant_id = $1", merchantID).Scan(&total)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     products,
	})
}

func GetMerchantOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchantID int
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant profile not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	status := c.DefaultQuery("status", "")
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")

	pageNum := 1
	perPageNum := 20
	pVal, parseErr3 := parseInt(page)
	if parseErr3 == nil && pVal > 0 {
		pageNum = pVal
	}
	ppVal, parseErr4 := parseInt(perPage)
	if parseErr4 == nil && ppVal > 0 && ppVal <= 100 {
		perPageNum = ppVal
	}

	offset := (pageNum - 1) * perPageNum

	var rows *sql.Rows
	if status != "" && status != "all" {
		rows, err = db.Query(
			`SELECT o.id, o.user_id, o.product_id, o.group_id, o.quantity, o.total_price, o.status, o.created_at, o.updated_at, p.name as product_name
			 FROM orders o 
			 JOIN products p ON o.product_id = p.id 
			 WHERE p.merchant_id = $1 AND o.status = $2 
			 ORDER BY o.created_at DESC LIMIT $3 OFFSET $4`,
			merchantID, status, perPageNum, offset,
		)
	} else {
		rows, err = db.Query(
			`SELECT o.id, o.user_id, o.product_id, o.group_id, o.quantity, o.total_price, o.status, o.created_at, o.updated_at, p.name as product_name
			 FROM orders o 
			 JOIN products p ON o.product_id = p.id 
			 WHERE p.merchant_id = $1 
			 ORDER BY o.created_at DESC LIMIT $2 OFFSET $3`,
			merchantID, perPageNum, offset,
		)
	}

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	type OrderWithProduct struct {
		models.Order
		ProductName string `json:"product_name"`
	}

	var orders []OrderWithProduct
	for rows.Next() {
		var o OrderWithProduct
		err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.GroupID, &o.Quantity, &o.TotalPrice,
			&o.Status, &o.CreatedAt, &o.UpdatedAt, &o.ProductName)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		orders = append(orders, o)
	}

	var total int
	if status != "" && status != "all" {
		db.QueryRow(
			`SELECT COUNT(*) FROM orders o JOIN products p ON o.product_id = p.id WHERE p.merchant_id = $1 AND o.status = $2`,
			merchantID, status,
		).Scan(&total)
	} else {
		db.QueryRow(
			`SELECT COUNT(*) FROM orders o JOIN products p ON o.product_id = p.id WHERE p.merchant_id = $1`,
			merchantID,
		).Scan(&total)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     orders,
	})
}

func GetMerchantSettlements(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchantID int
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant profile not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	rows, err := db.Query(
		`SELECT id, merchant_id, period_start, period_end, total_sales, platform_fee, settlement_amount, status, settled_at, created_at, updated_at 
		 FROM merchant_settlements WHERE merchant_id = $1 ORDER BY period_end DESC LIMIT 12`,
		merchantID,
	)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var settlements []models.MerchantSettlement
	for rows.Next() {
		var s models.MerchantSettlement
		var settledAt sql.NullTime
		err := rows.Scan(&s.ID, &s.MerchantID, &s.PeriodStart, &s.PeriodEnd, &s.TotalSales, &s.PlatformFee,
			&s.SettlementAmount, &s.Status, &settledAt, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if settledAt.Valid {
			s.SettledAt = settledAt.Time
		}
		settlements = append(settlements, s)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": settlements,
	})
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
