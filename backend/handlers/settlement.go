package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

func requireAdminRole(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return false
	}
	return true
}

func GetMerchantSettlementByID(c *gin.Context) {
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

	settlementIDStr := c.Param("id")
	settlementID, err := strconv.Atoi(settlementIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settlement ID"})
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchantID int
	err = db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{"error": "Merchant not found"})
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	type SettlementResponse struct {
		ID                  int        `json:"id"`
		MerchantID          int        `json:"merchant_id"`
		PeriodStart         time.Time  `json:"period_start"`
		PeriodEnd           time.Time  `json:"period_end"`
		TotalSales          float64    `json:"total_sales"`
		PlatformFee         float64    `json:"platform_fee"`
		SettlementAmount    float64    `json:"settlement_amount"`
		Status              string     `json:"status"`
		SettledAt           *time.Time `json:"settled_at,omitempty"`
		MerchantConfirmed   bool       `json:"merchant_confirmed"`
		MerchantConfirmedAt *time.Time `json:"merchant_confirmed_at,omitempty"`
		FinanceApproved     bool       `json:"finance_approved"`
		FinanceApprovedAt   *time.Time `json:"finance_approved_at,omitempty"`
		CreatedAt           time.Time  `json:"created_at"`
		UpdatedAt           time.Time  `json:"updated_at"`
	}

	var s SettlementResponse
	err = db.QueryRow(
		`SELECT id, merchant_id, period_start, period_end, total_sales, platform_fee, 
                settlement_amount, status, settled_at, created_at, updated_at,
                merchant_confirmed, merchant_confirmed_at, finance_approved, finance_approved_at
                FROM merchant_settlements WHERE id = $1 AND merchant_id = $2`,
		settlementID, merchantID,
	).Scan(&s.ID, &s.MerchantID, &s.PeriodStart, &s.PeriodEnd, &s.TotalSales,
		&s.PlatformFee, &s.SettlementAmount, &s.Status, &s.SettledAt, &s.CreatedAt, &s.UpdatedAt,
		&s.MerchantConfirmed, &s.MerchantConfirmedAt, &s.FinanceApproved, &s.FinanceApprovedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Settlement not found"})
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"settlement": s})
}

func ConfirmSettlement(c *gin.Context) {
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

	settlementIDStr := c.Param("id")
	settlementID, err := strconv.Atoi(settlementIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settlement ID"})
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchantID int
	err = db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{"error": "Merchant not found"})
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	service := services.GetSettlementService()
	err = service.MerchantConfirm(settlementID, merchantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settlement confirmed successfully"})
}

func SubmitSettlementDispute(c *gin.Context) {
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

	settlementIDStr := c.Param("id")
	settlementID, err := strconv.Atoi(settlementIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settlement ID"})
		return
	}

	var req struct {
		DisputeType    string  `json:"dispute_type" binding:"required"`
		Reason         string  `json:"reason" binding:"required"`
		OriginalAmount float64 `json:"original_amount" binding:"required"`
		DisputedAmount float64 `json:"disputed_amount" binding:"required"`
	}

	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchantID int
	err = db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{"error": "Merchant not found"})
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	service := services.GetSettlementService()
	dispute, err := service.SubmitDispute(settlementID, merchantID, req.DisputeType, req.Reason, req.OriginalAmount, req.DisputedAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Dispute submitted successfully",
		"dispute": dispute,
	})
}

func AdminGetSettlements(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	status := c.Query("status")
	year := c.Query("year")
	month := c.Query("month")

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	query := `SELECT s.id, s.merchant_id, s.period_start, s.period_end, s.total_sales, s.platform_fee, 
              s.settlement_amount, s.status, s.settled_at, s.created_at, s.updated_at,
              s.merchant_confirmed, s.merchant_confirmed_at, s.finance_approved, s.finance_approved_at,
              m.company_name
              FROM merchant_settlements s
              JOIN merchants m ON s.merchant_id = m.id
              WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		query += " AND s.status = $" + strconv.Itoa(argIndex)
		args = append(args, status)
		argIndex++
	}

	if year != "" {
		query += " AND EXTRACT(YEAR FROM s.period_start) = $" + strconv.Itoa(argIndex)
		args = append(args, year)
		argIndex++
	}

	if month != "" {
		query += " AND EXTRACT(MONTH FROM s.period_start) = $" + strconv.Itoa(argIndex)
		args = append(args, month)
	}

	query += " ORDER BY s.period_start DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SETTLEMENTS_QUERY_FAILED",
			"Failed to query settlements",
			http.StatusInternalServerError,
			err,
		))
		return
	}
	defer rows.Close()

	type AdminSettlementResponse struct {
		ID                  int        `json:"id"`
		MerchantID          int        `json:"merchant_id"`
		CompanyName         string     `json:"company_name"`
		PeriodStart         time.Time  `json:"period_start"`
		PeriodEnd           time.Time  `json:"period_end"`
		TotalSales          float64    `json:"total_sales"`
		PlatformFee         float64    `json:"platform_fee"`
		SettlementAmount    float64    `json:"settlement_amount"`
		Status              string     `json:"status"`
		SettledAt           time.Time  `json:"settled_at,omitempty"`
		MerchantConfirmed   bool       `json:"merchant_confirmed"`
		MerchantConfirmedAt *time.Time `json:"merchant_confirmed_at,omitempty"`
		FinanceApproved     bool       `json:"finance_approved"`
		FinanceApprovedAt   *time.Time `json:"finance_approved_at,omitempty"`
		CreatedAt           time.Time  `json:"created_at"`
		UpdatedAt           time.Time  `json:"updated_at"`
	}

	var settlements []AdminSettlementResponse
	for rows.Next() {
		var s AdminSettlementResponse
		err := rows.Scan(&s.ID, &s.MerchantID, &s.PeriodStart, &s.PeriodEnd, &s.TotalSales,
			&s.PlatformFee, &s.SettlementAmount, &s.Status, &s.SettledAt, &s.CreatedAt, &s.UpdatedAt,
			&s.MerchantConfirmed, &s.MerchantConfirmedAt, &s.FinanceApproved, &s.FinanceApprovedAt,
			&s.CompanyName)
		if err != nil {
			continue
		}
		settlements = append(settlements, s)
	}

	if settlements == nil {
		settlements = []AdminSettlementResponse{}
	}

	c.JSON(http.StatusOK, gin.H{"settlements": settlements})
}

func AdminGetSettlementByID(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	settlementIDStr := c.Param("id")
	settlementID, err := strconv.Atoi(settlementIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settlement ID"})
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	type AdminSettlementDetailResponse struct {
		ID                  int        `json:"id"`
		MerchantID          int        `json:"merchant_id"`
		CompanyName         string     `json:"company_name"`
		PeriodStart         time.Time  `json:"period_start"`
		PeriodEnd           time.Time  `json:"period_end"`
		TotalSales          float64    `json:"total_sales"`
		PlatformFee         float64    `json:"platform_fee"`
		SettlementAmount    float64    `json:"settlement_amount"`
		Status              string     `json:"status"`
		MerchantConfirmed   bool       `json:"merchant_confirmed"`
		MerchantConfirmedAt *time.Time `json:"merchant_confirmed_at,omitempty"`
		FinanceApproved     bool       `json:"finance_approved"`
		FinanceApprovedAt   *time.Time `json:"finance_approved_at,omitempty"`
		CreatedAt           time.Time  `json:"created_at"`
		UpdatedAt           time.Time  `json:"updated_at"`
	}

	var s AdminSettlementDetailResponse
	err = db.QueryRow(
		`SELECT s.id, s.merchant_id, s.period_start, s.period_end, s.total_sales, s.platform_fee, 
                s.settlement_amount, s.status, s.merchant_confirmed, s.merchant_confirmed_at, 
                s.finance_approved, s.finance_approved_at, s.created_at, s.updated_at,
                m.company_name
                FROM merchant_settlements s
                JOIN merchants m ON s.merchant_id = m.id
                WHERE s.id = $1`,
		settlementID,
	).Scan(&s.ID, &s.MerchantID, &s.PeriodStart, &s.PeriodEnd, &s.TotalSales,
		&s.PlatformFee, &s.SettlementAmount, &s.Status, &s.MerchantConfirmed, &s.MerchantConfirmedAt,
		&s.FinanceApproved, &s.FinanceApprovedAt, &s.CreatedAt, &s.UpdatedAt, &s.CompanyName)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Settlement not found"})
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"settlement": s})
}

func AdminApproveSettlement(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	adminUserID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	adminUserIDInt, ok := adminUserID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	settlementIDStr := c.Param("id")
	settlementID, err := strconv.Atoi(settlementIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settlement ID"})
		return
	}

	service := services.GetSettlementService()
	err = service.FinanceApprove(settlementID, adminUserIDInt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settlement approved successfully"})
}

func AdminMarkSettlementPaid(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	adminUserID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	adminUserIDInt, ok := adminUserID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	settlementIDStr := c.Param("id")
	settlementID, err := strconv.Atoi(settlementIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settlement ID"})
		return
	}

	service := services.GetSettlementService()
	err = service.MarkAsPaid(settlementID, adminUserIDInt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settlement marked as paid successfully"})
}

func AdminProcessDispute(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	adminUserID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	adminUserIDInt, ok := adminUserID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	disputeIDStr := c.Param("id")
	disputeID, err := strconv.Atoi(disputeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dispute ID"})
		return
	}

	var req struct {
		Resolution     string  `json:"resolution" binding:"required"`
		AdjustedAmount float64 `json:"adjusted_amount" binding:"required"`
	}

	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	service := services.GetSettlementService()
	err = service.ProcessDispute(disputeID, adminUserIDInt, req.Resolution, req.AdjustedAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dispute processed successfully"})
}

func AdminReconcileSettlement(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	settlementIDStr := c.Param("id")
	settlementID, err := strconv.Atoi(settlementIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settlement ID"})
		return
	}

	service := services.GetSettlementService()
	reconciliation, err := service.ReconcileOrders(settlementID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Reconciliation completed",
		"reconciliation": reconciliation,
	})
}

func AdminGenerateMonthlySettlements(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	var req struct {
		Year  int `json:"year" binding:"required"`
		Month int `json:"month" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	periodStart := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	service := services.GetSettlementService()
	settlements, err := service.GenerateMonthlySettlements(periodStart, periodEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Monthly settlements generated successfully",
		"count":       len(settlements),
		"settlements": settlements,
	})
}

func GetSettlementItems(c *gin.Context) {
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

	settlementIDStr := c.Param("id")
	settlementID, err := strconv.Atoi(settlementIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settlement ID"})
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchantID int
	err = db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{"error": "Merchant not found"})
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var settlementMerchantID int
	err = db.QueryRow("SELECT merchant_id FROM merchant_settlements WHERE id = $1", settlementID).Scan(&settlementMerchantID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Settlement not found"})
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if settlementMerchantID != merchantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view this settlement"})
		return
	}

	service := services.GetSettlementService()
	items, err := service.GetSettlementItems(settlementID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

func AdminGenerateSettlementForMerchant(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	var req struct {
		MerchantID int `json:"merchant_id" binding:"required"`
		Year       int `json:"year" binding:"required"`
		Month      int `json:"month" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	periodStart := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	service := services.GetSettlementService()
	settlement, err := service.GenerateSettlementForMerchant(req.MerchantID, periodStart, periodEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Settlement generated successfully",
		"settlement": settlement,
	})
}

func AdminGetBillingRecords(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}

	merchantIDStr := c.Query("merchant_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	merchantID := 0
	if merchantIDStr != "" {
		var err error
		merchantID, err = strconv.Atoi(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant_id"})
			return
		}
	}

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, use YYYY-MM-DD"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, use YYYY-MM-DD"})
		return
	}

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	service := services.GetSettlementService()
	records, err := service.GetBillingRecords(merchantID, startDate, endDate, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records":   records,
		"page":      page,
		"page_size": pageSize,
	})
}
