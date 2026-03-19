package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
)

func GetNotifications(c *gin.Context) {
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

	unreadOnly := c.Query("unread") == "true"
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit > 100 {
		limit = 100
	}

	query := "SELECT id, type, title, content, data, is_read, read_at, created_at FROM notifications WHERE user_id = $1"
	args := []interface{}{userIDInt}

	if unreadOnly {
		query += " AND is_read = FALSE"
	}

	query += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)+1)
	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	notifications := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var notifType, title, content string
		var data []byte
		var isRead bool
		var readAt sql.NullTime
		var createdAt time.Time

		err := rows.Scan(&id, &notifType, &title, &content, &data, &isRead, &readAt, &createdAt)
		if err != nil {
			continue
		}

		var dataMap map[string]interface{}
		if data != nil {
			json.Unmarshal(data, &dataMap)
		}

		notifications = append(notifications, map[string]interface{}{
			"id":         id,
			"type":       notifType,
			"title":      title,
			"content":    content,
			"data":       dataMap,
			"is_read":    isRead,
			"read_at":    readAt.Time,
			"created_at": createdAt,
		})
	}

	c.JSON(http.StatusOK, notifications)
}

func GetUnreadCount(c *gin.Context) {
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

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE", userIDInt).Scan(&count)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func MarkNotificationRead(c *gin.Context) {
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

	notificationID := c.Param("id")

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	result, err := db.Exec(
		"UPDATE notifications SET is_read = TRUE, read_at = $1 WHERE id = $2 AND user_id = $3",
		time.Now(), notificationID, userIDInt,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"NOTIFICATION_NOT_FOUND",
			"Notification not found",
			http.StatusNotFound,
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func MarkAllNotificationsRead(c *gin.Context) {
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

	_, err := db.Exec(
		"UPDATE notifications SET is_read = TRUE, read_at = $1 WHERE user_id = $2 AND is_read = FALSE",
		time.Now(), userIDInt,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

func RegisterDeviceToken(c *gin.Context) {
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

	var req struct {
		DeviceType string `json:"device_type" binding:"required"`
		Token      string `json:"token" binding:"required"`
		DeviceName string `json:"device_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	_, err := db.Exec(
		`INSERT INTO user_device_tokens (user_id, device_type, token, device_name, last_used_at) 
		 VALUES ($1, $2, $3, $4, $5) 
		 ON CONFLICT (user_id, token) DO UPDATE SET is_active = TRUE, last_used_at = $5`,
		userIDInt, req.DeviceType, req.Token, req.DeviceName, time.Now(),
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"TOKEN_REGISTER_FAILED",
			"Failed to register device token",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device token registered"})
}

func UnregisterDeviceToken(c *gin.Context) {
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

	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	_, err := db.Exec(
		"UPDATE user_device_tokens SET is_active = FALSE WHERE user_id = $1 AND token = $2",
		userIDInt, req.Token,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device token unregistered"})
}

func CreateNotification(userID int, notifType, title, content string, data map[string]interface{}) error {
	db := config.GetDB()
	if db == nil {
		return nil
	}

	var dataJSON []byte
	if data != nil {
		dataJSON, _ = json.Marshal(data)
	}

	_, err := db.Exec(
		"INSERT INTO notifications (user_id, type, title, content, data) VALUES ($1, $2, $3, $4, $5)",
		userID, notifType, title, content, dataJSON,
	)
	return err
}
