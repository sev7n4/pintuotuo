package notification

import (
	"os"
	"strconv"
	"strings"
)

// NotificationServiceFromEnv builds a notification service from environment variables.
// SMTP: SMTP_HOST, SMTP_PORT (default 587), SMTP_USERNAME, SMTP_PASSWORD, SMTP_FROM_NAME, SMTP_FROM_EMAIL, SMTP_USE_TLS (true/1/yes).
// Push: FCM_SERVER_KEY (optional).
// When SMTP_HOST or SMTP_FROM_EMAIL is empty, email sending is disabled (no panic).
func NotificationServiceFromEnv() *NotificationService {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	from := strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL"))
	if host == "" || from == "" {
		return NewNotificationService(nil, pushConfigFromEnv())
	}
	port := 587
	if p := strings.TrimSpace(os.Getenv("SMTP_PORT")); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			port = n
		}
	}
	useTLS := parseBoolEnv(os.Getenv("SMTP_USE_TLS"))
	emailCfg := &EmailConfig{
		SMTPHost:     host,
		SMTPPort:     port,
		SMTPUsername: strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromName:     strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")),
		FromEmail:    from,
		UseTLS:       useTLS,
	}
	if emailCfg.FromName == "" {
		emailCfg.FromName = "拼脱脱"
	}
	return NewNotificationService(emailCfg, pushConfigFromEnv())
}

func pushConfigFromEnv() *PushConfig {
	key := strings.TrimSpace(os.Getenv("FCM_SERVER_KEY"))
	if key == "" {
		return nil
	}
	return &PushConfig{FCMServerKey: key}
}

func parseBoolEnv(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
