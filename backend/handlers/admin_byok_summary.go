package handlers

import (
	"database/sql"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

// BYOKMerchantRollup 管理端按商户聚合的 BYOK 状态。
type BYOKMerchantRollup struct {
	MerchantID          int    `json:"merchant_id"`
	CompanyName         string `json:"company_name"`
	TotalKeyCount       int    `json:"total_key_count"`
	ActiveKeyCount      int    `json:"active_key_count"`
	HasRoutable         bool   `json:"has_routable"`
	NeedAttentionActive int    `json:"need_attention_active"`
	Level               string `json:"level"`
}

// BYOKSummaryResponse 全平台 BYOK 概览 + 按商户聚合。
type BYOKSummaryResponse struct {
	Summary struct {
		ActiveKeysTotal         int `json:"active_keys_total"`
		MerchantsWithActiveKeys int `json:"merchants_with_active_keys"`
		MerchantsHasRoutable    int `json:"merchants_has_routable"`
		MerchantsNeedAttention  int `json:"merchants_need_attention"`
		MerchantsWithNoKeys     int `json:"merchants_with_no_keys"`
	} `json:"summary"`
	ByMerchant []BYOKMerchantRollup `json:"by_merchant"`
}

// GetAdminBYOKSummary returns platform-wide BYOK observability for admin dashboard.
func GetAdminBYOKSummary(c *gin.Context) {
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

	const q = `
SELECT mak.merchant_id, mak.status, mak.health_status, mak.verification_result, mak.verified_at,
       COALESCE(NULLIF(trim(m.company_name), ''), '—') AS company_name
FROM merchant_api_keys mak
INNER JOIN merchants m ON m.id = mak.merchant_id
ORDER BY mak.merchant_id, mak.id`

	rows, err := db.Query(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load api keys"})
		return
	}
	defer rows.Close()

	type row struct {
		MerchantID   int
		Status       string
		Health       string
		Verification string
		VerifiedAt   sql.NullTime
		CompanyName  string
	}

	byMerchant := make(map[int][]services.KeyRowLite)
	companyName := make(map[int]string)

	for rows.Next() {
		var r row
		if err := rows.Scan(&r.MerchantID, &r.Status, &r.Health, &r.Verification, &r.VerifiedAt, &r.CompanyName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan api keys"})
			return
		}
		companyName[r.MerchantID] = r.CompanyName
		byMerchant[r.MerchantID] = append(byMerchant[r.MerchantID], services.KeyRowLite{
			Status:       r.Status,
			Health:       r.Health,
			Verification: r.Verification,
			VerifiedAt:   r.VerifiedAt,
		})
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to iterate api keys"})
		return
	}

	var out BYOKSummaryResponse
	activeKeysTotal := 0
	merchantsWithActive := 0
	merchantsHasRoutable := 0
	merchantsNeedAttention := 0

	for mid, keys := range byMerchant {
		level, hasRoutable, needAtt, activeCount := services.AggregateMerchantBYOK(keys)
		for _, k := range keys {
			if strings.EqualFold(strings.TrimSpace(k.Status), "active") {
				activeKeysTotal++
			}
		}
		if activeCount > 0 {
			merchantsWithActive++
		}
		if hasRoutable {
			merchantsHasRoutable++
		}
		if needAtt > 0 {
			merchantsNeedAttention++
		}

		out.ByMerchant = append(out.ByMerchant, BYOKMerchantRollup{
			MerchantID:          mid,
			CompanyName:         companyName[mid],
			TotalKeyCount:       len(keys),
			ActiveKeyCount:      activeCount,
			HasRoutable:         hasRoutable,
			NeedAttentionActive: needAtt,
			Level:               level,
		})
	}

	// Merchants with no keys at all
	const mq = `SELECT COUNT(*) FROM merchants m WHERE NOT EXISTS (
  SELECT 1 FROM merchant_api_keys mak WHERE mak.merchant_id = m.id
)`
	var noKeyMerchants int
	if err := db.QueryRow(mq).Scan(&noKeyMerchants); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count merchants without keys"})
		return
	}

	out.Summary.ActiveKeysTotal = activeKeysTotal
	out.Summary.MerchantsWithActiveKeys = merchantsWithActive
	out.Summary.MerchantsHasRoutable = merchantsHasRoutable
	out.Summary.MerchantsNeedAttention = merchantsNeedAttention
	out.Summary.MerchantsWithNoKeys = noKeyMerchants

	sort.Slice(out.ByMerchant, func(i, j int) bool {
		return out.ByMerchant[i].MerchantID < out.ByMerchant[j].MerchantID
	})

	c.JSON(http.StatusOK, out)
}
