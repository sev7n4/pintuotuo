package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	apperrors "github.com/pintuotuo/backend/errors"
)

// validateEntitlementPackageSKUReferences only checks that each sku_id exists (admin config save; sellability enforced at checkout).
func validateEntitlementPackageSKUReferences(tx *sql.Tx, items []struct {
	SKUID           int
	DefaultQuantity int
}) error {
	for _, it := range items {
		var one int
		err := tx.QueryRow(`SELECT 1 FROM skus WHERE id = $1`, it.SKUID).Scan(&one)
		if err != nil {
			if err == sql.ErrNoRows {
				return apperrors.NewAppError(
					"ENTITLEMENT_SKU_NOT_FOUND",
					fmt.Sprintf("SKU %d 不存在", it.SKUID),
					http.StatusBadRequest,
					nil,
				)
			}
			return err
		}
	}
	return nil
}

func applyLinePurchasability(it *EntitlementPackageItem) {
	ok, reason := linePurchasable(it.SKUStatus, it.SPUStatus, it.Stock, it.DefaultQuantity)
	it.LinePurchasable = ok
	if !ok {
		it.LineIssue = reason
	}
}

func linePurchasable(skuStatus, spuStatus string, stock, qty int) (bool, string) {
	if skuStatus != merchantStatusActive {
		return false, "SKU 未上架"
	}
	if spuStatus != productStatusActive {
		return false, "所属 SPU 非在售"
	}
	if stock != -1 && stock < qty {
		return false, "库存不足"
	}
	return true, ""
}

// enrichEntitlementPackageItems sets LinePurchasable / LineIssue on each item and returns overall purchasability.
func enrichEntitlementPackageItems(items []EntitlementPackageItem) (purchasable bool, unavailableReason string) {
	if len(items) == 0 {
		return false, "权益包未配置 SKU"
	}
	allOK := true
	var firstReason string
	for i := range items {
		applyLinePurchasability(&items[i])
		if !items[i].LinePurchasable {
			allOK = false
			if firstReason == "" {
				firstReason = items[i].LineIssue
			}
		}
	}
	return allOK, firstReason
}
