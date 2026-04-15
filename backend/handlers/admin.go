package handlers

import (
	"database/sql"
	"fmt"
	stdlog "log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
)

const (
	roleAdmin    = "admin"
	roleMerchant = "merchant"
	roleUser     = "user"
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
		TotalUsers          int     `json:"total_users"`
		TotalMerchants      int     `json:"total_merchants"`
		TotalOrders         int     `json:"total_orders"`
		TotalRevenue        float64 `json:"total_revenue"`
		PendingOrders       int     `json:"pending_orders"`
		PaidOrders          int     `json:"paid_orders"`
		CanceledOrders      int     `json:"canceled_orders"`
		MultiItemOrderRatio float64 `json:"multi_item_order_ratio"`
		OrderConversionRate float64 `json:"order_conversion_rate"`
		PaymentSuccessRate  float64 `json:"payment_success_rate"`
		CancellationRate    float64 `json:"cancellation_rate"`
	}

	// Get total users
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)

	// Get total merchants
	db.QueryRow("SELECT COUNT(*) FROM merchants").Scan(&stats.TotalMerchants)

	// Get total orders
	db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&stats.TotalOrders)

	// Get total revenue
	db.QueryRow("SELECT COALESCE(SUM(total_price), 0) FROM orders WHERE status = 'completed'").Scan(&stats.TotalRevenue)

	// Funnel counters for P1 operational monitoring
	db.QueryRow("SELECT COUNT(*) FROM orders WHERE status = 'pending'").Scan(&stats.PendingOrders)
	db.QueryRow("SELECT COUNT(*) FROM orders WHERE status IN ('paid', 'completed')").Scan(&stats.PaidOrders)
	db.QueryRow("SELECT COUNT(*) FROM orders WHERE status = 'cancelled'").Scan(&stats.CanceledOrders)

	if stats.TotalOrders > 0 {
		stats.OrderConversionRate = float64(stats.PaidOrders) / float64(stats.TotalOrders)
		stats.CancellationRate = float64(stats.CanceledOrders) / float64(stats.TotalOrders)
	}

	var totalPayments int
	var successPayments int
	db.QueryRow("SELECT COUNT(*) FROM payments").Scan(&totalPayments)
	db.QueryRow("SELECT COUNT(*) FROM payments WHERE status = 'success'").Scan(&successPayments)
	if totalPayments > 0 {
		stats.PaymentSuccessRate = float64(successPayments) / float64(totalPayments)
	}

	var multiItemOrders int
	db.QueryRow(
		`SELECT COUNT(*) FROM (
		   SELECT order_id
		     FROM order_items
		    GROUP BY order_id
		   HAVING COUNT(*) > 1
		) t`,
	).Scan(&multiItemOrders)
	if stats.TotalOrders > 0 {
		stats.MultiItemOrderRatio = float64(multiItemOrders) / float64(stats.TotalOrders)
	}

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
		`SELECT id, user_id, company_name, business_license, business_license_url, id_card_front_url, id_card_back_url,
		 contact_name, contact_phone, contact_email, address, description, status, review_note AS rejection_reason,
		 business_category, admin_notes, reviewed_at, created_at, updated_at
		 FROM merchants WHERE status IN ('pending', 'reviewing') ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
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
		if err := rows.Scan(&m.ID, &m.UserID, &m.CompanyName, &m.BusinessLicense, &m.BusinessLicenseURL,
			&m.IDCardFrontURL, &m.IDCardBackURL, &m.ContactName,
			&m.ContactPhone, &m.ContactEmail, &m.Address, &m.Description, &m.Status, &m.RejectionReason,
			&m.BusinessCategory, &m.AdminNotes, &m.ReviewedAt, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		merchants = append(merchants, m)
	}

	if merchants == nil {
		merchants = []models.Merchant{}
	}

	var total int
	db.QueryRow("SELECT COUNT(*) FROM merchants WHERE status IN ('pending', 'reviewing')").Scan(&total)

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

	if currentStatus != merchantStatusPending && currentStatus != "reviewing" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_STATUS",
			"Merchant is not in pending or reviewing status",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	var merchant models.Merchant
	err = db.QueryRow(
		`UPDATE merchants SET status = 'active', lifecycle_status = 'active', reviewed_at = CURRENT_TIMESTAMP, review_note = NULL, updated_at = CURRENT_TIMESTAMP 
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

	adminID, _ := adminActorID(c)
	if err := insertMerchantAuditLog(db, merchant.ID, adminID, "approve", merchant.CompanyName, nil); err != nil {
		stdlog.Printf("merchant audit log (approve): %v", err)
	}
	if err := services.InsertPlatformAuditLog(db, "merchant", merchant.ID, "approve", adminID, c, nil); err != nil {
		stdlog.Printf("platform audit log (approve merchant): %v", err)
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

	if currentStatus != merchantStatusPending && currentStatus != "reviewing" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_STATUS",
			"Merchant is not in pending or reviewing status",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	var merchant models.Merchant
	err = db.QueryRow(
		`UPDATE merchants SET status = 'rejected', review_note = $1, updated_at = CURRENT_TIMESTAMP 
		 WHERE id = $2 
		 RETURNING id, user_id, company_name, business_license, contact_name, contact_phone, contact_email, address, description, status, created_at, updated_at`,
		req.Reason, merchantID,
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

	adminID, _ := adminActorID(c)
	var reasonPtr *string
	if req.Reason != "" {
		reasonPtr = &req.Reason
	}
	if err := insertMerchantAuditLog(db, merchant.ID, adminID, "reject", merchant.CompanyName, reasonPtr); err != nil {
		stdlog.Printf("merchant audit log (reject): %v", err)
	}
	if err := services.InsertPlatformAuditLog(db, "merchant", merchant.ID, "reject", adminID, c, map[string]interface{}{"reason": req.Reason}); err != nil {
		stdlog.Printf("platform audit log (reject merchant): %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Merchant rejected successfully",
		"data":    merchant,
		"reason":  req.Reason,
	})
}

func adminActorID(c *gin.Context) (int, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	id, ok := v.(int)
	if !ok {
		return 0, false
	}
	return id, true
}

func insertMerchantAuditLog(db *sql.DB, merchantID, adminID int, action, companyName string, reason *string) error {
	var adminArg interface{}
	if adminID > 0 {
		adminArg = adminID
	} else {
		adminArg = nil
	}
	var snap interface{}
	if companyName != "" {
		snap = companyName
	} else {
		snap = nil
	}
	_, err := db.Exec(
		`INSERT INTO merchant_audit_logs (merchant_id, admin_user_id, action, company_name_snapshot, reason) VALUES ($1, $2, $3, $4, $5)`,
		merchantID, adminArg, action, snap, reason,
	)
	return err
}

// GetMerchantAuditLogs lists admin audit records for merchants.
func GetMerchantAuditLogs(c *gin.Context) {
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

	where := "1=1"
	args := []interface{}{}
	argN := 1
	if mid := strings.TrimSpace(c.Query("merchant_id")); mid != "" {
		if id, err := parseInt(mid); err == nil && id > 0 {
			where += fmt.Sprintf(" AND l.merchant_id = $%d", argN)
			args = append(args, id)
			argN++
		}
	}
	if act := strings.TrimSpace(c.Query("action")); act != "" {
		where += fmt.Sprintf(" AND l.action = $%d", argN)
		args = append(args, act)
		argN++
	}

	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM merchant_audit_logs l WHERE %s`, where)
	var total int
	if err := db.QueryRow(countQ, args...).Scan(&total); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	dataArgs := append(append([]interface{}{}, args...), perPageNum, offset)
	limitPh := fmt.Sprintf("$%d", argN)
	offsetPh := fmt.Sprintf("$%d", argN+1)
	q := fmt.Sprintf(
		`SELECT l.id, l.merchant_id, l.admin_user_id, l.action, l.company_name_snapshot, l.reason, l.created_at, u.email
		 FROM merchant_audit_logs l
		 LEFT JOIN users u ON u.id = l.admin_user_id
		 WHERE %s
		 ORDER BY l.created_at DESC LIMIT %s OFFSET %s`,
		where, limitPh, offsetPh,
	)

	rows, err := db.Query(q, dataArgs...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var logs []models.MerchantAuditLog
	for rows.Next() {
		var logRow models.MerchantAuditLog
		var adminUID sql.NullInt64
		var companySnap, reason sql.NullString
		var adminEmail sql.NullString
		if err := rows.Scan(&logRow.ID, &logRow.MerchantID, &adminUID, &logRow.Action,
			&companySnap, &reason, &logRow.CreatedAt, &adminEmail); err != nil {
			continue
		}
		if adminUID.Valid {
			v := int(adminUID.Int64)
			logRow.AdminUserID = &v
		}
		if companySnap.Valid {
			s := companySnap.String
			logRow.CompanyNameSnapshot = &s
		}
		if reason.Valid {
			s := reason.String
			logRow.Reason = &s
		}
		if adminEmail.Valid {
			s := adminEmail.String
			logRow.AdminEmail = &s
		}
		logs = append(logs, logRow)
	}
	if logs == nil {
		logs = []models.MerchantAuditLog{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     0,
		"message":  "success",
		"data":     logs,
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
	})
}

// GetAdminMerchants lists merchants with optional filters (status, business_category, keyword).
func GetAdminMerchants(c *gin.Context) {
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

	where, args := buildMerchantAdminWhere(c)

	countQ := "SELECT COUNT(*) FROM merchants WHERE " + where
	var total int
	if err := db.QueryRow(countQ, args...).Scan(&total); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	dataArgs := append(append([]interface{}{}, args...), perPageNum, offset)
	limitPh := fmt.Sprintf("$%d", len(args)+1)
	offsetPh := fmt.Sprintf("$%d", len(args)+2)
	q := fmt.Sprintf(
		`SELECT id, user_id, company_name, business_license, business_license_url, id_card_front_url, id_card_back_url,
		 contact_name, contact_phone, contact_email, address, description, status, review_note AS rejection_reason,
		 business_category, admin_notes, reviewed_at, created_at, updated_at
		 FROM merchants WHERE %s ORDER BY created_at DESC LIMIT %s OFFSET %s`,
		where, limitPh, offsetPh,
	)

	rows, err := db.Query(q, dataArgs...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var merchants []models.Merchant
	for rows.Next() {
		var m models.Merchant
		if err := rows.Scan(&m.ID, &m.UserID, &m.CompanyName, &m.BusinessLicense, &m.BusinessLicenseURL,
			&m.IDCardFrontURL, &m.IDCardBackURL, &m.ContactName,
			&m.ContactPhone, &m.ContactEmail, &m.Address, &m.Description, &m.Status, &m.RejectionReason,
			&m.BusinessCategory, &m.AdminNotes, &m.ReviewedAt, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		merchants = append(merchants, m)
	}
	if merchants == nil {
		merchants = []models.Merchant{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     0,
		"message":  "success",
		"data":     merchants,
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
	})
}

func buildMerchantAdminWhere(c *gin.Context) (where string, args []interface{}) {
	where = "1=1"
	args = []interface{}{}
	n := 1
	if s := strings.TrimSpace(c.Query("status")); s != "" {
		where += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, s)
		n++
	}
	if cat := strings.TrimSpace(c.Query("business_category")); cat != "" {
		where += fmt.Sprintf(" AND business_category = $%d", n)
		args = append(args, cat)
		n++
	}
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		where += fmt.Sprintf(" AND (company_name ILIKE $%d OR contact_email ILIKE $%d OR contact_phone ILIKE $%d)", n, n+1, n+2)
		like := "%" + kw + "%"
		args = append(args, like, like, like)
	}
	return where, args
}

// PatchAdminMerchant updates admin-maintained merchant fields (category, internal notes).
func PatchAdminMerchant(c *gin.Context) {
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
		BusinessCategory *string `json:"business_category"`
		AdminNotes       *string `json:"admin_notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	if req.BusinessCategory == nil && req.AdminNotes == nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"At least one of business_category or admin_notes is required",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var companyName string
	err := db.QueryRow("SELECT company_name FROM merchants WHERE id = $1", merchantID).Scan(&companyName)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	setParts := []string{}
	args := []interface{}{}
	n := 1
	if req.BusinessCategory != nil {
		setParts = append(setParts, fmt.Sprintf("business_category = $%d", n))
		args = append(args, *req.BusinessCategory)
		n++
	}
	if req.AdminNotes != nil {
		setParts = append(setParts, fmt.Sprintf("admin_notes = $%d", n))
		args = append(args, *req.AdminNotes)
		n++
	}
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, merchantID)

	q := fmt.Sprintf(
		`UPDATE merchants SET %s WHERE id = $%d RETURNING id, user_id, company_name, business_license, business_license_url, id_card_front_url, id_card_back_url,
		 contact_name, contact_phone, contact_email, address, description, status, review_note AS rejection_reason,
		 business_category, admin_notes, reviewed_at, created_at, updated_at`,
		strings.Join(setParts, ", "), n,
	)

	var m models.Merchant
	err = db.QueryRow(q, args...).Scan(&m.ID, &m.UserID, &m.CompanyName, &m.BusinessLicense, &m.BusinessLicenseURL,
		&m.IDCardFrontURL, &m.IDCardBackURL, &m.ContactName,
		&m.ContactPhone, &m.ContactEmail, &m.Address, &m.Description, &m.Status, &m.RejectionReason,
		&m.BusinessCategory, &m.AdminNotes, &m.ReviewedAt, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"UPDATE_FAILED",
			"Failed to update merchant",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	adminID, _ := adminActorID(c)
	if err := insertMerchantAuditLog(db, m.ID, adminID, "meta_update", m.CompanyName, nil); err != nil {
		stdlog.Printf("merchant audit log (meta_update): %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    m,
	})
}
