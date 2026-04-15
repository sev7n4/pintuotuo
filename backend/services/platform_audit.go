package services

import (
	"database/sql"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

// InsertPlatformAuditLog 写入 audit_logs（平台级敏感操作）。
func InsertPlatformAuditLog(db *sql.DB, entityType string, entityID int, action string, operatorID int, c *gin.Context, metadata map[string]interface{}) error {
	if db == nil {
		return nil
	}
	var meta interface{}
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		meta = b
	}
	ip := ""
	if c != nil {
		ip = c.ClientIP()
	}
	var ua interface{}
	if c != nil && c.Request != nil {
		ua = c.Request.UserAgent()
	}
	_, err := db.Exec(`
		INSERT INTO audit_logs (entity_type, entity_id, action, operator_id, operator_type, ip_address, user_agent, metadata)
		VALUES ($1, $2, $3, $4, 'admin', $5, $6, $7)
	`, entityType, entityID, action, nullIfZero(operatorID), ip, ua, meta)
	return err
}

func nullIfZero(id int) interface{} {
	if id <= 0 {
		return nil
	}
	return id
}
