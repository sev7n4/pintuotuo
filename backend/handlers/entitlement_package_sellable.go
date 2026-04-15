package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	apperrors "github.com/pintuotuo/backend/errors"
)

func validateEntitlementPackageSKULines(tx *sql.Tx, items []struct {
	SKUID           int
	DefaultQuantity int
}) error {
	for _, it := range items {
		var skuSt, spuSt string
		var stock int
		err := tx.QueryRow(
			`SELECT s.status, sp.status, s.stock
			 FROM skus s JOIN spus sp ON s.spu_id = sp.id
			 WHERE s.id = $1`,
			it.SKUID,
		).Scan(&skuSt, &spuSt, &stock)
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
		if skuSt != merchantStatusActive {
			return apperrors.NewAppError(
				"ENTITLEMENT_SKU_NOT_SELLABLE",
				fmt.Sprintf("SKU %d 未上架（inactive），仅可选择 SKU 与 SPU 均在售的商品", it.SKUID),
				http.StatusBadRequest,
				nil,
			)
		}
		if spuSt != productStatusActive {
			return apperrors.NewAppError(
				"ENTITLEMENT_SPU_NOT_SELLABLE",
				fmt.Sprintf("SKU %d 所属 SPU 非在售状态", it.SKUID),
				http.StatusBadRequest,
				nil,
			)
		}
		if stock != -1 && stock < it.DefaultQuantity {
			return apperrors.NewAppError(
				"ENTITLEMENT_SKU_INSUFFICIENT_STOCK",
				fmt.Sprintf("SKU %d 可用库存不足（需 %d）", it.SKUID, it.DefaultQuantity),
				http.StatusBadRequest,
				nil,
			)
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
