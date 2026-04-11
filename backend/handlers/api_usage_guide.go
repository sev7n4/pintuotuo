package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
)

// APIUsageGuideItem is one entitlement-derived row for OpenAI-compatible calling hints.
type APIUsageGuideItem struct {
	Source               string `json:"source"`
	SPUName              string `json:"spu_name,omitempty"`
	SKUCode              string `json:"sku_code,omitempty"`
	ProviderCode         string `json:"provider_code"`
	ModelExample         string `json:"model_example"`
	ProviderSlashExample string `json:"provider_slash_example,omitempty"`
}

// APIUsageGuideResponse is returned by GET /tokens/api-usage-guide.
type APIUsageGuideResponse struct {
	Items               []APIUsageGuideItem `json:"items"`
	DefaultModelExample string              `json:"default_model_example,omitempty"`
	Disclaimer          string              `json:"disclaimer"`
}

// GetAPIUsageGuide aggregates model hints from active subscriptions and paid/completed orders.
func GetAPIUsageGuide(c *gin.Context) {
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

	type rowKey struct {
		Prov  string
		Model string
	}
	seen := make(map[rowKey]struct{})
	items := make([]APIUsageGuideItem, 0)

	appendRows := func(rows *sql.Rows, source string) error {
		defer rows.Close()
		for rows.Next() {
			var spuName, skuCode, provider, providerModelID, modelName sql.NullString
			if err := rows.Scan(&spuName, &skuCode, &provider, &providerModelID, &modelName); err != nil {
				return err
			}
			pc := strings.ToLower(strings.TrimSpace(provider.String))
			if pc == "" {
				continue
			}
			me := strings.TrimSpace(providerModelID.String)
			if me == "" {
				me = strings.TrimSpace(modelName.String)
			}
			if me == "" {
				continue
			}
			k := rowKey{Prov: pc, Model: strings.ToLower(me)}
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}

			slash := pc + "/" + me
			item := APIUsageGuideItem{
				Source:               source,
				ProviderCode:         pc,
				ModelExample:         me,
				ProviderSlashExample: slash,
			}
			if spuName.Valid {
				item.SPUName = spuName.String
			}
			if skuCode.Valid {
				item.SKUCode = skuCode.String
			}
			items = append(items, item)
		}
		return nil
	}

	// Active subscriptions
	subRows, err := db.Query(
		`SELECT sp.name, s.sku_code, sp.model_provider, sp.provider_model_id, sp.model_name
		 FROM user_subscriptions us
		 JOIN skus s ON us.sku_id = s.id
		 JOIN spus sp ON s.spu_id = sp.id
		 WHERE us.user_id = $1 AND us.status = 'active'`,
		userIDInt,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if appendErr := appendRows(subRows, "subscription"); appendErr != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	orderRows, err := db.Query(
		`SELECT sp.name, s.sku_code, sp.model_provider, sp.provider_model_id, sp.model_name
		 FROM orders o
		 JOIN skus s ON o.sku_id = s.id
		 JOIN spus sp ON s.spu_id = sp.id
		 WHERE o.user_id = $1 AND o.sku_id IS NOT NULL
		   AND o.status IN ('paid', 'completed')`,
		userIDInt,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if appendErr := appendRows(orderRows, "order"); appendErr != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	out := APIUsageGuideResponse{
		Items:      items,
		Disclaimer: "以下为根据您当前有效订阅与已支付订单汇总的模型调用示例；实际以路由与商户可用密钥为准。也可使用「厂商代码/模型名」形式显式指定上游。",
	}
	if len(items) > 0 {
		out.DefaultModelExample = items[0].ProviderSlashExample
		if out.DefaultModelExample == "" {
			out.DefaultModelExample = items[0].ModelExample
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}
