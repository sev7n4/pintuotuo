package handlers

import (
  "context"
  "log"
  "net/http"
  "os"
  "strconv"

  "github.com/gin-gonic/gin"
  "github.com/pintuotuo/backend/config"
  apperrors "github.com/pintuotuo/backend/errors"
  "github.com/pintuotuo/backend/middleware"
  "github.com/pintuotuo/backend/services/token"
)

// Initialize token service
var tokenService token.Service

func initTokenService() {
  if tokenService == nil {
    logger := log.New(os.Stderr, "[TokenHandler] ", log.LstdFlags)
    tokenService = token.NewService(config.GetDB(), logger)
  }
}

// GetBalance retrieves user token balance
// GET /v1/tokens/balance
func GetBalance(c *gin.Context) {
  initTokenService()

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
  balance, err := tokenService.GetBalance(ctx, userIDInt)
  if err != nil {
    if appErr, ok := err.(*apperrors.AppError); ok {
      middleware.RespondWithError(c, appErr)
    } else {
      middleware.RespondWithError(c, apperrors.ErrDatabaseError)
    }
    return
  }

  c.JSON(http.StatusOK, balance)
}

// GetTotalBalance retrieves user total token balance
// GET /v1/tokens/total-balance
func GetTotalBalance(c *gin.Context) {
  initTokenService()

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
  totalBalance, err := tokenService.GetTotalBalance(ctx, userIDInt)
  if err != nil {
    if appErr, ok := err.(*apperrors.AppError); ok {
      middleware.RespondWithError(c, appErr)
    } else {
      middleware.RespondWithError(c, apperrors.ErrDatabaseError)
    }
    return
  }

  c.JSON(http.StatusOK, gin.H{"total_balance": totalBalance})
}

// GetConsumption retrieves user token consumption history
// GET /v1/tokens/consumption
func GetConsumption(c *gin.Context) {
  initTokenService()

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

  // Parse pagination parameters
  page := 1
  pageSize := 20

  if pageStr := c.Query("page"); pageStr != "" {
    if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
      page = p
    }
  }

  if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
    if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
      pageSize = ps
    }
  }

  ctx := context.Background()
  result, err := tokenService.GetConsumption(ctx, userIDInt, &token.GetConsumptionParams{
    UserID:   userIDInt,
    PageSize: pageSize,
    Page:     page,
  })

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

// ListTransactions retrieves user token transactions with optional filtering
// GET /v1/tokens/transactions
func ListTransactions(c *gin.Context) {
  initTokenService()

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

  // Parse optional type filter
  txType := c.Query("type")

  limit := 100
  if limitStr := c.Query("limit"); limitStr != "" {
    if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
      limit = l
    }
  }

  ctx := context.Background()
  transactions, err := tokenService.GetTransactions(ctx, userIDInt, &token.TransactionParams{
    UserID: userIDInt,
    Type:   txType,
    Limit:  limit,
  })

  if err != nil {
    if appErr, ok := err.(*apperrors.AppError); ok {
      middleware.RespondWithError(c, appErr)
    } else {
      middleware.RespondWithError(c, apperrors.ErrDatabaseError)
    }
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "transactions": transactions,
    "count":        len(transactions),
  })
}

// RechargeTokens recharges tokens for a user (admin only)
// POST /v1/tokens/recharge
func RechargeTokens(c *gin.Context) {
  initTokenService()

  // This should be admin-only in production
  userRole, exists := c.Get("user_role")
  if !exists || userRole != "admin" {
    middleware.RespondWithError(c, apperrors.ErrInvalidToken)
    return
  }

  var req token.RechargeTokensRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
    return
  }

  ctx := context.Background()
  balance, err := tokenService.RechargeTokens(ctx, &req)
  if err != nil {
    if appErr, ok := err.(*apperrors.AppError); ok {
      middleware.RespondWithError(c, appErr)
    } else {
      middleware.RespondWithError(c, apperrors.ErrDatabaseError)
    }
    return
  }

  c.JSON(http.StatusOK, balance)
}

// ConsumeTokens consumes tokens for a user (internal use)
// POST /v1/tokens/consume
func ConsumeTokens(c *gin.Context) {
  initTokenService()

  var req token.ConsumeTokensRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
    return
  }

  ctx := context.Background()
  balance, err := tokenService.ConsumeTokens(ctx, &req)
  if err != nil {
    if appErr, ok := err.(*apperrors.AppError); ok {
      middleware.RespondWithError(c, appErr)
    } else {
      middleware.RespondWithError(c, apperrors.ErrDatabaseError)
    }
    return
  }

  c.JSON(http.StatusOK, balance)
}

// TransferTokens transfers tokens between users
// POST /v1/tokens/transfer
func TransferTokens(c *gin.Context) {
  initTokenService()

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
    RecipientID int     `json:"recipient_id" binding:"required,gt=0"`
    Amount      float64 `json:"amount" binding:"required,gt=0"`
  }

  if err := c.ShouldBindJSON(&req); err != nil {
    middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
    return
  }

  ctx := context.Background()
  transferReq := &token.TransferTokensRequest{
    SenderID:    userIDInt,
    RecipientID: req.RecipientID,
    Amount:      req.Amount,
  }

  err := tokenService.TransferTokens(ctx, transferReq)
  if err != nil {
    if appErr, ok := err.(*apperrors.AppError); ok {
      middleware.RespondWithError(c, appErr)
    } else {
      middleware.RespondWithError(c, apperrors.ErrDatabaseError)
    }
    return
  }

  c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}
