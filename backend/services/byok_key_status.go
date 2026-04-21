package services

import (
	"database/sql"
	"strings"
)

// KeyMeetsStrictAllowlist 与 entitlement 白名单 SQL 一致：
// (verified_at IS NOT NULL OR verification_result = 'verified') AND health IN (healthy, degraded)
func KeyMeetsStrictAllowlist(health, verification string, verifiedAt sql.NullTime) bool {
	vr := strings.ToLower(strings.TrimSpace(verification))
	verifiedLine := verifiedAt.Valid || vr == "verified" || vr == "success"
	h := strings.ToLower(strings.TrimSpace(health))
	if h == "" {
		h = "unknown"
	}
	healthOk := h == "healthy" || h == "degraded"
	return verifiedLine && healthOk
}

// KeyNeedsAttentionActive 与商户端「Strict 权益 / 需立即关注」一致：启用中的 Key 未满足 strict 条件。
func KeyNeedsAttentionActive(status, health, verification string, verifiedAt sql.NullTime) bool {
	if strings.ToLower(strings.TrimSpace(status)) != "active" {
		return false
	}
	return !KeyMeetsStrictAllowlist(health, verification, verifiedAt)
}

// MerchantBYOKLevel 管理端多 Key 聚合：green=存在可路由；yellow=无可路由且存在不健康或验证失败；gray=有启用 Key 但仅未知等；none=无密钥。
const (
	MerchantBYOKLevelNone   = "none"
	MerchantBYOKLevelGray   = "gray"
	MerchantBYOKLevelYellow = "yellow"
	MerchantBYOKLevelGreen  = "green"
)

// KeyRowLite 聚合用最小字段集（与 DB 扫描列对应）。
type KeyRowLite struct {
	Status       string
	Health       string
	Verification string
	VerifiedAt   sql.NullTime
}

// AggregateMerchantBYOK 仅统计 status=active 的密钥参与路由语义；若存在任一把 strict 通过则为 green。
func AggregateMerchantBYOK(keys []KeyRowLite) (level string, hasRoutable bool, needAttentionActive int, activeCount int) {
	var active []KeyRowLite
	for _, k := range keys {
		if strings.ToLower(strings.TrimSpace(k.Status)) == "active" {
			active = append(active, k)
		}
	}
	activeCount = len(active)
	if len(keys) == 0 {
		return MerchantBYOKLevelNone, false, 0, 0
	}
	if activeCount == 0 {
		return MerchantBYOKLevelGray, false, 0, 0
	}

	for _, k := range active {
		if KeyMeetsStrictAllowlist(k.Health, k.Verification, k.VerifiedAt) {
			hasRoutable = true
		}
		if KeyNeedsAttentionActive(k.Status, k.Health, k.Verification, k.VerifiedAt) {
			needAttentionActive++
		}
	}
	if hasRoutable {
		return MerchantBYOKLevelGreen, true, needAttentionActive, activeCount
	}
	for _, k := range active {
		h := strings.ToLower(strings.TrimSpace(k.Health))
		vr := strings.ToLower(strings.TrimSpace(k.Verification))
		if h == "unhealthy" || vr == "failed" {
			return MerchantBYOKLevelYellow, false, needAttentionActive, activeCount
		}
	}
	return MerchantBYOKLevelGray, false, needAttentionActive, activeCount
}
