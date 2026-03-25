package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

const (
	roleAdmin             = "admin"
	roleMerchant          = "merchant"
	roleUser              = "user"
	merchantStatusPending = "pending"
)

// GetAdminUsers retrieves all users (admin only)
func GetAdminUsers(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rows, err := db.Query(
		"SELECT id, email, name, role, created_at, updated_at FROM users ORDER BY created_at DESC",
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			continue
		}
		users = append(users, user)
	}

	if users == nil {
		users = []models.User{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    users,
	})
}

// CreateAdminUser creates a new admin user (admin only)
func CreateAdminUser(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required,min=2"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role"`
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

	// Check if user exists
	var existingUser models.User
	err := db.QueryRow(
		"SELECT id FROM users WHERE email = $1",
		req.Email,
	).Scan(&existingUser.ID)
	if err == nil {
		middleware.RespondWithError(c, apperrors.ErrUserAlreadyExists)
		return
	}

	// Hash password
	passwordHash := hashPassword(req.Password)

	// Determine role (default to admin)
	role := roleAdmin
	if req.Role == roleMerchant || req.Role == roleUser {
		role = req.Role
	}

	// Create user
	var user models.User
	err = db.QueryRow(
		"INSERT INTO users (email, name, password_hash, role, status) VALUES ($1, $2, $3, $4, $5) RETURNING id, email, name, role, created_at, updated_at",
		req.Email, req.Name, passwordHash, role, "active",
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"USER_CREATION_FAILED",
			"Failed to create user",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	// Create token balance
	_, err = db.Exec(
		"INSERT INTO tokens (user_id, balance) VALUES ($1, $2)",
		user.ID, 0,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"TOKEN_INIT_FAILED",
			"Failed to initialize token balance",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    user,
	})
}

// GetAdminStats retrieves platform statistics (admin only)
func GetAdminStats(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var stats struct {
		TotalUsers     int     `json:"total_users"`
		TotalMerchants int     `json:"total_merchants"`
		TotalOrders    int     `json:"total_orders"`
		TotalRevenue   float64 `json:"total_revenue"`
	}

	// Get total users
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)

	// Get total merchants
	db.QueryRow("SELECT COUNT(*) FROM merchants").Scan(&stats.TotalMerchants)

	// Get total orders
	db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&stats.TotalOrders)

	// Get total revenue
	db.QueryRow("SELECT COALESCE(SUM(total_price), 0) FROM orders WHERE status = 'completed'").Scan(&stats.TotalRevenue)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

func GetPendingMerchants(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")

	pageNum := 1
	perPageNum := 20
	if p, err := parseInt(page); err == nil && p > 0 {
		pageNum = p
	}
	if pp, err := parseInt(perPage); err == nil && pp > 0 && pp <= 100 {
		perPageNum = pp
	}

	offset := (pageNum - 1) * perPageNum

	rows, err := db.Query(
		`SELECT id, user_id, company_name, business_license, contact_name, contact_phone, contact_email, address, description, status, created_at, updated_at 
		 FROM merchants WHERE status = 'pending' ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		perPageNum, offset,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var merchants []models.Merchant
	for rows.Next() {
		var m models.Merchant
		if err := rows.Scan(&m.ID, &m.UserID, &m.CompanyName, &m.BusinessLicense, &m.ContactName,
			&m.ContactPhone, &m.ContactEmail, &m.Address, &m.Description, &m.Status, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		merchants = append(merchants, m)
	}

	if merchants == nil {
		merchants = []models.Merchant{}
	}

	var total int
	db.QueryRow("SELECT COUNT(*) FROM merchants WHERE status = 'pending'").Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"code":     0,
		"message":  "success",
		"data":     merchants,
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
	})
}

func ApproveMerchant(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	merchantID := c.Param("id")
	if merchantID == "" {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var currentStatus string
	err := db.QueryRow("SELECT status FROM merchants WHERE id = $1", merchantID).Scan(&currentStatus)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	if currentStatus != merchantStatusPending {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_STATUS",
			"Merchant is not in pending status",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	var merchant models.Merchant
	err = db.QueryRow(
		`UPDATE merchants SET status = 'active', verified_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		 WHERE id = $1 
		 RETURNING id, user_id, company_name, business_license, contact_name, contact_phone, contact_email, address, description, status, created_at, updated_at`,
		merchantID,
	).Scan(&merchant.ID, &merchant.UserID, &merchant.CompanyName, &merchant.BusinessLicense, &merchant.ContactName,
		&merchant.ContactPhone, &merchant.ContactEmail, &merchant.Address, &merchant.Description, &merchant.Status,
		&merchant.CreatedAt, &merchant.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"APPROVE_FAILED",
			"Failed to approve merchant",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Merchant approved successfully",
		"data":    merchant,
	})
}

func RejectMerchant(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	merchantID := c.Param("id")
	if merchantID == "" {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = ""
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var currentStatus string
	err := db.QueryRow("SELECT status FROM merchants WHERE id = $1", merchantID).Scan(&currentStatus)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	if currentStatus != merchantStatusPending {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_STATUS",
			"Merchant is not in pending status",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	var merchant models.Merchant
	err = db.QueryRow(
		`UPDATE merchants SET status = 'rejected', updated_at = CURRENT_TIMESTAMP 
		 WHERE id = $1 
		 RETURNING id, user_id, company_name, business_license, contact_name, contact_phone, contact_email, address, description, status, created_at, updated_at`,
		merchantID,
	).Scan(&merchant.ID, &merchant.UserID, &merchant.CompanyName, &merchant.BusinessLicense, &merchant.ContactName,
		&merchant.ContactPhone, &merchant.ContactEmail, &merchant.Address, &merchant.Description, &merchant.Status,
		&merchant.CreatedAt, &merchant.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"REJECT_FAILED",
			"Failed to reject merchant",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Merchant rejected successfully",
		"data":    merchant,
		"reason":  req.Reason,
	})
}
