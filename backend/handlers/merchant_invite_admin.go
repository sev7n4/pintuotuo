package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

type createMerchantInviteBody struct {
	MaxUses   int    `json:"max_uses"`
	ExpiresIn string `json:"expires_in"` // 如 720h、168h，空表示不过期
	Note      string `json:"note"`
}

type revokeMerchantInviteBody struct {
	Reason string `json:"reason"`
}

// CreateMerchantInvite POST /admin/merchant-invites
func CreateMerchantInvite(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError("FORBIDDEN", "Admin access required", http.StatusForbidden, nil))
		return
	}
	var body createMerchantInviteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	maxUses := body.MaxUses
	if maxUses <= 0 {
		maxUses = 1
	}
	var expiresAt interface{}
	if strings.TrimSpace(body.ExpiresIn) != "" {
		d, err := time.ParseDuration(strings.TrimSpace(body.ExpiresIn))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_EXPIRES_IN", "message": "expires_in 无效，请使用如 720h"})
			return
		}
		t := time.Now().Add(d)
		expiresAt = t
	} else {
		expiresAt = nil
	}

	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	code := hex.EncodeToString(buf)

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	adminID, _ := adminActorID(c)
	note := strings.TrimSpace(body.Note)
	var noteArg interface{} = nil
	if note != "" {
		noteArg = note
	}
	var id int
	err := db.QueryRow(`
		INSERT INTO merchant_invites (code, max_uses, expires_at, note, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, code, maxUses, expiresAt, noteArg, nullIntPtr(adminID)).Scan(&id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	base := strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_APP_BASE_URL")), "/")
	if base == "" {
		base = strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN")), "/")
	}
	registerURL := ""
	if base != "" {
		registerURL = base + "/register?invite=" + code
	}

	suffix := code
	if len(suffix) > 6 {
		suffix = suffix[len(suffix)-6:]
	}
	_ = services.InsertPlatformAuditLog(db, "merchant_invite", id, "create", adminID, c, map[string]interface{}{"code_suffix": suffix})

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":           id,
			"code":         code,
			"max_uses":     maxUses,
			"register_url": registerURL,
		},
	})
}

func nullIntPtr(id int) interface{} {
	if id <= 0 {
		return nil
	}
	return id
}

// ListMerchantInvites GET /admin/merchant-invites
func ListMerchantInvites(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError("FORBIDDEN", "Admin access required", http.StatusForbidden, nil))
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	statusFilter := strings.TrimSpace(c.Query("status"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	limit := 100
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 300 {
			limit = parsed
		}
	}
	where := []string{"1=1"}
	args := []interface{}{}
	argN := 1
	if keyword != "" {
		where = append(where, "(i.code ILIKE $"+strconv.Itoa(argN)+" OR COALESCE(i.note,'') ILIKE $"+strconv.Itoa(argN)+" OR COALESCE(u.email,'') ILIKE $"+strconv.Itoa(argN)+")")
		args = append(args, "%"+keyword+"%")
		argN++
	}
	if statusFilter == "active" {
		where = append(where, "i.revoked_at IS NULL", "(i.expires_at IS NULL OR i.expires_at > NOW())", "i.used_count < i.max_uses")
	}
	if statusFilter == "revoked" {
		where = append(where, "i.revoked_at IS NOT NULL")
	}
	if statusFilter == "expired" {
		where = append(where, "i.revoked_at IS NULL", "i.expires_at IS NOT NULL", "i.expires_at <= NOW()")
	}
	if statusFilter == "used_up" {
		where = append(where, "i.used_count >= i.max_uses")
	}

	args = append(args, limit)
	q := `
		SELECT i.id, i.code, i.max_uses, i.used_count, i.expires_at, i.revoked_at, COALESCE(i.note,''), i.created_at, i.created_by_user_id, COALESCE(u.email,'')
		FROM merchant_invites i
		LEFT JOIN users u ON u.id = i.created_by_user_id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY i.id DESC
		LIMIT $` + strconv.Itoa(argN)
	rows, err := db.Query(q, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()
	type inviteRow struct {
		ID        int        `json:"id"`
		Code      string     `json:"code"`
		MaxUses   int        `json:"max_uses"`
		UsedCount int        `json:"used_count"`
		ExpiresAt *time.Time `json:"expires_at,omitempty"`
		RevokedAt *time.Time `json:"revoked_at,omitempty"`
		Note      string     `json:"note"`
		CreatedAt time.Time  `json:"created_at"`
		Status    string     `json:"status"`
		Creator   string     `json:"creator,omitempty"`
		Register  string     `json:"register_url,omitempty"`
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_APP_BASE_URL")), "/")
	if base == "" {
		base = strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN")), "/")
	}
	var list []inviteRow
	for rows.Next() {
		var r inviteRow
		var exp, rev sql.NullTime
		var createdBy sql.NullInt64
		var creatorEmail string
		if err := rows.Scan(&r.ID, &r.Code, &r.MaxUses, &r.UsedCount, &exp, &rev, &r.Note, &r.CreatedAt, &createdBy, &creatorEmail); err != nil {
			continue
		}
		if exp.Valid {
			t := exp.Time
			r.ExpiresAt = &t
		}
		if rev.Valid {
			t := rev.Time
			r.RevokedAt = &t
		}
		if creatorEmail != "" {
			r.Creator = creatorEmail
		}
		r.Status = inviteStatus(r.RevokedAt, r.ExpiresAt, r.UsedCount, r.MaxUses)
		if base != "" {
			r.Register = base + "/register?invite=" + r.Code
		}
		list = append(list, r)
	}
	if list == nil {
		list = []inviteRow{}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": list})
}

func inviteStatus(revokedAt *time.Time, expiresAt *time.Time, usedCount, maxUses int) string {
	if revokedAt != nil {
		return "revoked"
	}
	if expiresAt != nil && time.Now().After(*expiresAt) {
		return "expired"
	}
	if usedCount >= maxUses {
		return "used_up"
	}
	return "active"
}

// RevokeMerchantInvite POST /admin/merchant-invites/:id/revoke
func RevokeMerchantInvite(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError("FORBIDDEN", "Admin access required", http.StatusForbidden, nil))
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	inviteID, err := strconv.Atoi(id)
	if err != nil || inviteID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	var body revokeMerchantInviteBody
	_ = c.ShouldBindJSON(&body)

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	res, err := db.Exec(`UPDATE merchant_invites SET revoked_at = COALESCE(revoked_at, CURRENT_TIMESTAMP) WHERE id = $1`, inviteID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		middleware.RespondWithError(c, apperrors.NewAppError("NOT_FOUND", "invite not found", http.StatusNotFound, nil))
		return
	}
	adminID, _ := adminActorID(c)
	_ = services.InsertPlatformAuditLog(db, "merchant_invite", inviteID, "revoke", adminID, c, map[string]interface{}{"reason": strings.TrimSpace(body.Reason)})
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}
