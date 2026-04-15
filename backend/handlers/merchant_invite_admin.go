package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"os"
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
	rows, err := db.Query(`
		SELECT id, code, max_uses, used_count, expires_at, revoked_at, COALESCE(note,''), created_at
		FROM merchant_invites
		ORDER BY id DESC
		LIMIT 100
	`)
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
	}
	var list []inviteRow
	for rows.Next() {
		var r inviteRow
		var exp, rev sql.NullTime
		if err := rows.Scan(&r.ID, &r.Code, &r.MaxUses, &r.UsedCount, &exp, &rev, &r.Note, &r.CreatedAt); err != nil {
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
		list = append(list, r)
	}
	if list == nil {
		list = []inviteRow{}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": list})
}
