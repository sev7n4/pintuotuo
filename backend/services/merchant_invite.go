package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInviteInvalid 邀请码不可用（不存在、已撤销、过期或次数用尽）。
var ErrInviteInvalid = errors.New("merchant invite invalid")

// ConsumeMerchantInviteTx 在事务内校验并占用一条邀请（used_count+1）。code 须非空且已 trim。
func ConsumeMerchantInviteTx(tx *sql.Tx, code string) (inviteID int, err error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0, ErrInviteInvalid
	}
	var maxUses, used int
	var expiresAt sql.NullTime
	var revokedAt sql.NullTime
	err = tx.QueryRow(`
		SELECT id, max_uses, used_count, expires_at, revoked_at
		FROM merchant_invites
		WHERE code = $1
		FOR UPDATE
	`, code).Scan(&inviteID, &maxUses, &used, &expiresAt, &revokedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInviteInvalid
		}
		return 0, err
	}
	if revokedAt.Valid {
		return 0, ErrInviteInvalid
	}
	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return 0, ErrInviteInvalid
	}
	if used >= maxUses {
		return 0, ErrInviteInvalid
	}
	_, err = tx.Exec(`UPDATE merchant_invites SET used_count = used_count + 1 WHERE id = $1`, inviteID)
	if err != nil {
		return 0, fmt.Errorf("update invite: %w", err)
	}
	return inviteID, nil
}
