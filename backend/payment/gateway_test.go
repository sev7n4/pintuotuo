package payment

import (
	"testing"
)

func TestGenerateNonceStr(t *testing.T) {
	nonce1 := generateNonceStr()
	nonce2 := generateNonceStr()

	if len(nonce1) == 0 {
		t.Error("generateNonceStr() returned empty string")
	}

	if nonce1 == nonce2 {
		t.Error("generateNonceStr() should return unique values")
	}
}

func TestMapToXML(t *testing.T) {
	params := map[string]string{
		"appid":     "test_app_id",
		"mch_id":    "test_mch_id",
		"nonce_str": "abc123",
	}

	xml := mapToXML(params)

	if xml == "" {
		t.Error("mapToXML() returned empty string")
	}

	if !contains(xml, "<xml>") || !contains(xml, "</xml>") {
		t.Error("mapToXML() should contain xml tags")
	}

	if !contains(xml, "test_app_id") {
		t.Error("mapToXML() should contain the values")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestAlipayClient_Sign(t *testing.T) {
	t.Skip("Requires RSA key pair for testing")
}

func TestWechatPayClient_Sign(t *testing.T) {
	config := &WechatPayConfig{
		AppID:   "test_app_id",
		MchID:   "test_mch_id",
		APIKey:  "test_api_key_32_characters_long",
		Sandbox: true,
	}

	client := NewWechatPayClient(config)

	params := map[string]string{
		"appid":     "test_app_id",
		"mch_id":    "test_mch_id",
		"nonce_str": "abc123",
	}

	sign := client.sign(params)

	if sign == "" {
		t.Error("sign() returned empty string")
	}

	if len(sign) != 64 {
		t.Errorf("sign() should return 64 character SHA256 hash, got %d", len(sign))
	}
}

func TestPaymentService_CreateAlipayPayment(t *testing.T) {
	t.Skip("Requires RSA key pair for signing - tested in integration tests")
}

func TestPaymentService_CreateWechatPayment(t *testing.T) {
	service := NewPaymentService(
		nil,
		&WechatPayConfig{
			AppID:   "test_app_id",
			MchID:   "test_mch_id",
			APIKey:  "test_api_key_32_characters_long",
			Sandbox: true,
		},
	)

	_, err := service.CreateWechatPayment("ORDER123", 10000, "Test Order")
	if err == nil {
		t.Error("CreateWechatPayment() should fail without real API")
	}
}

func TestAlipayConfig_Sandbox(t *testing.T) {
	sandboxConfig := &AlipayConfig{
		AppID:   "test",
		Sandbox: true,
	}

	if !sandboxConfig.Sandbox {
		t.Error("Sandbox should be true")
	}

	prodConfig := &AlipayConfig{
		AppID:   "test",
		Sandbox: false,
	}

	if prodConfig.Sandbox {
		t.Error("Sandbox should be false")
	}
}

func TestWechatPayConfig_Sandbox(t *testing.T) {
	sandboxConfig := &WechatPayConfig{
		AppID:   "test",
		MchID:   "test",
		APIKey:  "test_api_key_32_characters_long",
		Sandbox: true,
	}

	if !sandboxConfig.Sandbox {
		t.Error("Sandbox should be true")
	}
}
