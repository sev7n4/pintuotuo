package notification

import (
	"testing"
)

func TestNotificationServiceFromEnv_DisabledWithoutSMTP(t *testing.T) {
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_FROM_EMAIL", "")
	t.Setenv("FCM_SERVER_KEY", "")
	s := NotificationServiceFromEnv()
	if s == nil {
		t.Fatal("service should be non-nil")
	}
	if err := s.TrySendSubscriptionExpiringEmail("u@example.com", "U", map[string]interface{}{
		"SPUName": "P", "EndDate": "2026-01-01", "Kind": "7d", "AutoRenewTxt": "否",
	}); err != nil {
		t.Errorf("TrySend with no SMTP should not error: %v", err)
	}
}

func TestNotificationServiceFromEnv_WithSMTP(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_FROM_EMAIL", "from@example.com")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_USERNAME", "u")
	t.Setenv("SMTP_PASSWORD", "p")
	t.Setenv("FCM_SERVER_KEY", "k")
	s := NotificationServiceFromEnv()
	if s == nil || s.email == nil {
		t.Fatal("expected email service when SMTP env set")
	}
}
