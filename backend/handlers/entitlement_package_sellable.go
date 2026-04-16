package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/services"
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
	if !allOK {
		return false, firstReason
	}
	policyLines := make([]services.OrderLinePolicyInput, 0, len(items))
	for _, it := range items {
		policyLines = append(policyLines, services.OrderLinePolicyInput{
			SKUType:         it.SKUType,
			ModelProvider:   it.ModelProvider,
			ModelName:       it.ModelName,
			ProviderModelID: it.ProviderModelID,
		})
	}
	if err := services.ValidateFuelPackBundle(policyLines); err != nil {
		return false, err.Error()
	}
	return true, ""
}

func validateEntitlementPackageBundlePolicy(tx *sql.Tx, items []struct {
	SKUID           int
	DefaultQuantity int
}) error {
	lines := make([]services.OrderLinePolicyInput, 0, len(items))
	for _, it := range items {
		var skuType, modelProvider, modelName, providerModelID string
		err := tx.QueryRow(
			`SELECT s.sku_type, COALESCE(sp.model_provider, ''), COALESCE(sp.model_name, ''), COALESCE(sp.provider_model_id, '')
			 FROM skus s
			 JOIN spus sp ON s.spu_id = sp.id
			 WHERE s.id = $1`,
			it.SKUID,
		).Scan(&skuType, &modelProvider, &modelName, &providerModelID)
		if err == sql.ErrNoRows {
			return apperrors.NewAppError(
				"ENTITLEMENT_SKU_NOT_FOUND",
				fmt.Sprintf("SKU %d 不存在", it.SKUID),
				http.StatusBadRequest,
				nil,
			)
		}
		if err != nil {
			return err
		}
		lines = append(lines, services.OrderLinePolicyInput{
			SKUType:         skuType,
			ModelProvider:   modelProvider,
			ModelName:       modelName,
			ProviderModelID: providerModelID,
		})
	}
	if err := services.ValidateFuelPackBundle(lines); err != nil {
		return apperrors.NewAppError(
			"FUEL_PACK_PURCHASE_RESTRICTED",
			err.Error(),
			http.StatusBadRequest,
			nil,
		)
	}
	return nil
}
