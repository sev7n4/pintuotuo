package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"math/big"
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

const (
	referralRewardRate = 0.05
	referralCodeLength = 8
)

func generateReferralCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, referralCodeLength)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		code[i] = charset[n.Int64()]
	}
	return string(code)
}

func GetMyReferralCode(c *gin.Context) {
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
	cacheKey := cache.ReferralCodeKey(userIDInt)

	if cachedCode, err := cache.Get(ctx, cacheKey); err == nil {
		var referralCode models.ReferralCode
		if err := json.Unmarshal([]byte(cachedCode), &referralCode); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"code": referralCode.Code,
			})
			return
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var referralCode models.ReferralCode
	err := db.QueryRow(
		"SELECT id, user_id, code, created_at, updated_at FROM referral_codes WHERE user_id = $1",
		userIDInt,
	).Scan(&referralCode.ID, &referralCode.UserID, &referralCode.Code, &referralCode.CreatedAt, &referralCode.UpdatedAt)

	if err == sql.ErrNoRows {
		code := generateReferralCode()
		err = db.QueryRow(
			"INSERT INTO referral_codes (user_id, code) VALUES ($1, $2) RETURNING id, user_id, code, created_at, updated_at",
			userIDInt, code,
		).Scan(&referralCode.ID, &referralCode.UserID, &referralCode.Code, &referralCode.CreatedAt, &referralCode.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"REFERRAL_CODE_CREATE_FAILED",
				"Failed to create referral code",
				http.StatusInternalServerError,
				err,
			))
			return
		}

		_, err = db.Exec("UPDATE users SET referral_code = $1 WHERE id = $2", code, userIDInt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"USER_UPDATE_FAILED",
				"Failed to update user referral code",
				http.StatusInternalServerError,
				err,
			))
			return
		}
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if referralCodeJSON, err := json.Marshal(referralCode); err == nil {
		cache.Set(ctx, cacheKey, string(referralCodeJSON), cache.ReferralCodeTTL)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": referralCode.Code,
	})
}

func ValidateReferralCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" || len(code) != referralCodeLength {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	ctx := context.Background()
	cacheKey := "referral:validate:" + code

	if cachedResult, err := cache.Get(ctx, cacheKey); err == nil {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(cachedResult), &result); err == nil {
			c.JSON(http.StatusOK, result)
			return
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var referrerID int
	var referrerName string
	err := db.QueryRow(
		"SELECT rc.user_id, u.name FROM referral_codes rc JOIN users u ON rc.user_id = u.id WHERE rc.code = $1",
		code,
	).Scan(&referrerID, &referrerName)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": "Invalid referral code",
		})
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	result := gin.H{
		"valid":         true,
		"referrer_id":   referrerID,
		"referrer_name": referrerName,
	}

	if resultJSON, marshalErr := json.Marshal(result); marshalErr == nil {
		cache.Set(ctx, cacheKey, string(resultJSON), 5*60)
	}

	c.JSON(http.StatusOK, result)
}

func GetReferralStats(c *gin.Context) {
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
	cacheKey := cache.ReferralStatsKey(userIDInt)

	if cachedStats, err := cache.Get(ctx, cacheKey); err == nil {
		var stats map[string]interface{}
		if err := json.Unmarshal([]byte(cachedStats), &stats); err == nil {
			c.JSON(http.StatusOK, stats)
			return
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var totalReferrals int
	var totalRewards float64
	var pendingRewards float64
	var paidRewards float64

	db.QueryRow("SELECT COUNT(*) FROM referrals WHERE referrer_id = $1", userIDInt).Scan(&totalReferrals)
	db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM referral_rewards WHERE referrer_id = $1", userIDInt).Scan(&totalRewards)
	db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM referral_rewards WHERE referrer_id = $1 AND status = 'pending'", userIDInt).Scan(&pendingRewards)
	db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM referral_rewards WHERE referrer_id = $1 AND status = 'paid'", userIDInt).Scan(&paidRewards)

	stats := gin.H{
		"total_referrals": totalReferrals,
		"total_rewards":   totalRewards,
		"pending_rewards": pendingRewards,
		"paid_rewards":    paidRewards,
	}

	if statsJSON, err := json.Marshal(stats); err == nil {
		cache.Set(ctx, cacheKey, string(statsJSON), cache.ReferralStatsTTL)
	}

	c.JSON(http.StatusOK, stats)
}

func GetReferralList(c *gin.Context) {
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

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 20
	}

	offset := (pageNum - 1) * perPageNum

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rows, err := db.Query(
		`SELECT r.id, r.referrer_id, r.referee_id, r.code_used, r.status, r.created_at,
		 u.name as referee_name
		 FROM referrals r 
		 JOIN users u ON r.referee_id = u.id 
		 WHERE r.referrer_id = $1 
		 ORDER BY r.created_at DESC 
		 LIMIT $2 OFFSET $3`,
		userIDInt, perPageNum, offset,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	type ReferralWithUser struct {
		models.Referral
		RefereeName string `json:"referee_name"`
	}

	var referrals []ReferralWithUser
	for rows.Next() {
		var r ReferralWithUser
		err := rows.Scan(&r.ID, &r.ReferrerID, &r.RefereeID, &r.CodeUsed, &r.Status, &r.CreatedAt, &r.RefereeName)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		referrals = append(referrals, r)
	}

	var total int
	db.QueryRow("SELECT COUNT(*) FROM referrals WHERE referrer_id = $1", userIDInt).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     referrals,
	})
}

func GetReferralRewards(c *gin.Context) {
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

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	status := c.DefaultQuery("status", "")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 20
	}

	offset := (pageNum - 1) * perPageNum

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var rows *sql.Rows
	var err error

	if status != "" {
		rows, err = db.Query(
			`SELECT rr.id, rr.referrer_id, rr.referee_id, rr.order_id, rr.amount, rr.status, rr.created_at, rr.paid_at,
			 u.name as referee_name
			 FROM referral_rewards rr 
			 JOIN users u ON rr.referee_id = u.id 
			 WHERE rr.referrer_id = $1 AND rr.status = $2
			 ORDER BY rr.created_at DESC 
			 LIMIT $3 OFFSET $4`,
			userIDInt, status, perPageNum, offset,
		)
	} else {
		rows, err = db.Query(
			`SELECT rr.id, rr.referrer_id, rr.referee_id, rr.order_id, rr.amount, rr.status, rr.created_at, rr.paid_at,
			 u.name as referee_name
			 FROM referral_rewards rr 
			 JOIN users u ON rr.referee_id = u.id 
			 WHERE rr.referrer_id = $1 
			 ORDER BY rr.created_at DESC 
			 LIMIT $2 OFFSET $3`,
			userIDInt, perPageNum, offset,
		)
	}

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	type RewardWithUser struct {
		models.ReferralReward
		RefereeName string `json:"referee_name"`
	}

	var rewards []RewardWithUser
	for rows.Next() {
		var r RewardWithUser
		var orderID sql.NullInt64
		var paidAt sql.NullTime
		err := rows.Scan(&r.ID, &r.ReferrerID, &r.RefereeID, &orderID, &r.Amount, &r.Status, &r.CreatedAt, &paidAt, &r.RefereeName)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if orderID.Valid {
			r.OrderID = int(orderID.Int64)
		}
		if paidAt.Valid {
			r.PaidAt = paidAt.Time
		}
		rewards = append(rewards, r)
	}

	var total int
	if status != "" {
		db.QueryRow("SELECT COUNT(*) FROM referral_rewards WHERE referrer_id = $1 AND status = $2", userIDInt, status).Scan(&total)
	} else {
		db.QueryRow("SELECT COUNT(*) FROM referral_rewards WHERE referrer_id = $1", userIDInt).Scan(&total)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     rewards,
	})
}

func BindReferralCode(c *gin.Context) {
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
		Code string `json:"code" binding:"required"`
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

	var referredBy sql.NullInt64
	err := db.QueryRow("SELECT referred_by FROM users WHERE id = $1", userIDInt).Scan(&referredBy)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if referredBy.Valid && int(referredBy.Int64) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Already bound to a referrer",
		})
		return
	}

	var referrerID int
	err = db.QueryRow("SELECT user_id FROM referral_codes WHERE code = $1", req.Code).Scan(&referrerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid referral code",
		})
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if referrerID == userIDInt {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot use your own referral code",
		})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE users SET referred_by = $1 WHERE id = $2", referrerID, userIDInt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	_, err = tx.Exec(
		"INSERT INTO referrals (referrer_id, referee_id, code_used) VALUES ($1, $2, $3)",
		referrerID, userIDInt, req.Code,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	_, err = tx.Exec("UPDATE users SET total_referrals = total_referrals + 1 WHERE id = $1", referrerID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if err = tx.Commit(); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.ReferralStatsKey(referrerID))
	cache.Delete(ctx, cache.ReferralStatsKey(userIDInt))

	c.JSON(http.StatusOK, gin.H{
		"message": "Referral code bound successfully",
	})
}

func CalculateReferralReward(orderID int, refereeID int, orderAmount float64) error {
	db := config.GetDB()
	if db == nil {
		return sql.ErrConnDone
	}

	var referrerID int
	err := db.QueryRow("SELECT referred_by FROM users WHERE id = $1", refereeID).Scan(&referrerID)
	if err == sql.ErrNoRows || referrerID == 0 {
		return nil
	}

	rewardAmount := orderAmount * referralRewardRate

	_, err = db.Exec(
		"INSERT INTO referral_rewards (referrer_id, referee_id, order_id, amount) VALUES ($1, $2, $3, $4)",
		referrerID, refereeID, orderID, rewardAmount,
	)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE users SET total_rewards = total_rewards + $1 WHERE id = $2", rewardAmount, referrerID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.ReferralStatsKey(referrerID))

	return nil
}

func PayReferralRewards(c *gin.Context) {
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
		RewardIDs []int `json:"reward_ids" binding:"required"`
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

	tx, err := db.Begin()
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer tx.Rollback()

	for _, rewardID := range req.RewardIDs {
		var referrerID int
		var status string
		queryErr := tx.QueryRow("SELECT referrer_id, status FROM referral_rewards WHERE id = $1", rewardID).Scan(&referrerID, &status)
		if queryErr != nil {
			continue
		}

		if referrerID != userIDInt || status != "pending" {
			continue
		}

		_, execErr := tx.Exec("UPDATE referral_rewards SET status = 'paid', paid_at = $1 WHERE id = $2", time.Now(), rewardID)
		if execErr != nil {
			continue
		}
	}

	if err = tx.Commit(); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.ReferralStatsKey(userIDInt))

	c.JSON(http.StatusOK, gin.H{
		"message": "Rewards paid successfully",
	})
}
