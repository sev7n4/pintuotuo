package notification

import (
	"strings"
	"testing"
)

func TestInitEmailTemplates(t *testing.T) {
	InitEmailTemplates()

	if emailTemplates == nil {
		t.Fatal("emailTemplates should not be nil after InitEmailTemplates")
	}

	if emailTemplates.Welcome == nil {
		t.Error("Welcome template should not be nil")
	}

	if emailTemplates.ResetPassword == nil {
		t.Error("ResetPassword template should not be nil")
	}

	if emailTemplates.OrderConfirm == nil {
		t.Error("OrderConfirm template should not be nil")
	}

	if emailTemplates.PaymentSuccess == nil {
		t.Error("PaymentSuccess template should not be nil")
	}

	if emailTemplates.LowBalance == nil {
		t.Error("LowBalance template should not be nil")
	}
}

func TestEmailTemplates_Render(t *testing.T) {
	InitEmailTemplates()

	t.Run("Welcome template", func(t *testing.T) {
		data := map[string]string{"Name": "测试用户"}
		var buf strings.Builder
		err := emailTemplates.Welcome.Execute(&buf, data)
		if err != nil {
			t.Errorf("Failed to execute Welcome template: %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "测试用户") {
			t.Error("Welcome template should contain the name")
		}
		if !strings.Contains(result, "欢迎加入") {
			t.Error("Welcome template should contain welcome message")
		}
	})

	t.Run("ResetPassword template", func(t *testing.T) {
		data := map[string]string{"Code": "123456"}
		var buf strings.Builder
		err := emailTemplates.ResetPassword.Execute(&buf, data)
		if err != nil {
			t.Errorf("Failed to execute ResetPassword template: %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "123456") {
			t.Error("ResetPassword template should contain the code")
		}
	})

	t.Run("OrderConfirm template", func(t *testing.T) {
		data := map[string]interface{}{
			"Name":        "测试用户",
			"OrderID":     "ORDER123",
			"ProductName": "测试商品",
			"Quantity":    2,
			"Amount":      99.99,
		}
		var buf strings.Builder
		err := emailTemplates.OrderConfirm.Execute(&buf, data)
		if err != nil {
			t.Errorf("Failed to execute OrderConfirm template: %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "ORDER123") {
			t.Error("OrderConfirm template should contain the order ID")
		}
		if !strings.Contains(result, "测试商品") {
			t.Error("OrderConfirm template should contain the product name")
		}
	})

	t.Run("PaymentSuccess template", func(t *testing.T) {
		data := map[string]interface{}{
			"Name":      "测试用户",
			"OrderID":   "ORDER123",
			"Amount":    99.99,
			"PayMethod": "支付宝",
			"PaidAt":    "2026-03-19 18:00:00",
		}
		var buf strings.Builder
		err := emailTemplates.PaymentSuccess.Execute(&buf, data)
		if err != nil {
			t.Errorf("Failed to execute PaymentSuccess template: %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "支付成功") {
			t.Error("PaymentSuccess template should contain success message")
		}
	})

	t.Run("LowBalance template", func(t *testing.T) {
		data := map[string]interface{}{
			"Name":      "测试用户",
			"Balance":   5.50,
			"Threshold": 10.00,
		}
		var buf strings.Builder
		err := emailTemplates.LowBalance.Execute(&buf, data)
		if err != nil {
			t.Errorf("Failed to execute LowBalance template: %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "余额不足") {
			t.Error("LowBalance template should contain low balance message")
		}
	})
}

func TestNotificationService_GetPushTitle(t *testing.T) {
	service := &NotificationService{}

	tests := []struct {
		notifType NotificationType
		want      string
	}{
		{NotificationWelcome, "欢迎加入拼脱脱！"},
		{NotificationResetPassword, "密码重置验证码"},
		{NotificationOrderConfirm, "订单创建成功"},
		{NotificationPaymentSuccess, "支付成功"},
		{NotificationLowBalance, "余额不足提醒"},
	}

	for _, tt := range tests {
		t.Run(string(tt.notifType), func(t *testing.T) {
			got := service.getPushTitle(tt.notifType)
			if got != tt.want {
				t.Errorf("getPushTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotificationService_GetPushBody(t *testing.T) {
	service := &NotificationService{}

	t.Run("OrderConfirm body", func(t *testing.T) {
		data := map[string]interface{}{"OrderID": "ORDER123"}
		body := service.getPushBody(NotificationOrderConfirm, data)
		if !strings.Contains(body, "ORDER123") {
			t.Error("OrderConfirm body should contain order ID")
		}
	})

	t.Run("PaymentSuccess body", func(t *testing.T) {
		data := map[string]interface{}{
			"OrderID": "ORDER123",
			"Amount":  99.99,
		}
		body := service.getPushBody(NotificationPaymentSuccess, data)
		if !strings.Contains(body, "ORDER123") {
			t.Error("PaymentSuccess body should contain order ID")
		}
	})

	t.Run("LowBalance body", func(t *testing.T) {
		data := map[string]interface{}{
			"Balance": 5.50,
		}
		body := service.getPushBody(NotificationLowBalance, data)
		if !strings.Contains(body, "5.5") {
			t.Error("LowBalance body should contain balance")
		}
	})
}

func TestNotificationService_ConvertData(t *testing.T) {
	service := &NotificationService{}

	data := map[string]interface{}{
		"order_id":  "ORDER123",
		"amount":    99.99,
		"is_member": true,
		"count":     5,
	}

	result := service.convertData(data)

	if result["order_id"] != "ORDER123" {
		t.Errorf("convertData() order_id = %v, want ORDER123", result["order_id"])
	}

	if result["amount"] != "99.99" {
		t.Errorf("convertData() amount = %v, want 99.99", result["amount"])
	}

	if result["is_member"] != "true" {
		t.Errorf("convertData() is_member = %v, want true", result["is_member"])
	}

	if result["count"] != "5" {
		t.Errorf("convertData() count = %v, want 5", result["count"])
	}
}

func TestEmailService_New(t *testing.T) {
	config := &EmailConfig{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user@example.com",
		SMTPPassword: "password",
		FromName:     "Test",
		FromEmail:    "test@example.com",
		UseTLS:       true,
	}

	service := NewEmailService(config)

	if service == nil {
		t.Fatal("NewEmailService() returned nil")
	}

	if service.config != config {
		t.Error("Service config should match provided config")
	}
}

func TestPushService_New(t *testing.T) {
	config := &PushConfig{
		FCMServerKey: "test_key",
	}

	service := NewPushService(config)

	if service == nil {
		t.Fatal("NewPushService() returned nil")
	}

	if service.config != config {
		t.Error("Service config should match provided config")
	}
}

func TestNotificationService_New(t *testing.T) {
	emailConfig := &EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
	}
	pushConfig := &PushConfig{
		FCMServerKey: "test_key",
	}

	service := NewNotificationService(emailConfig, pushConfig)

	if service == nil {
		t.Fatal("NewNotificationService() returned nil")
	}

	if service.email == nil {
		t.Error("Service email should not be nil")
	}

	if service.push == nil {
		t.Error("Service push should not be nil")
	}
}
