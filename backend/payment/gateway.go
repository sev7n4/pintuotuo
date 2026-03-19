package payment

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type AlipayConfig struct {
	AppID        string
	PrivateKey   *rsa.PrivateKey
	AlipayPublicKey *rsa.PublicKey
	NotifyURL    string
	ReturnURL    string
	Sandbox      bool
}

type AlipayClient struct {
	config *AlipayConfig
	client *http.Client
}

func NewAlipayClient(config *AlipayConfig) *AlipayClient {
	return &AlipayClient{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type AlipayTradePayRequest struct {
	OutTradeNo  string  `json:"out_trade_no"`
	TotalAmount float64 `json:"total_amount"`
	Subject     string  `json:"subject"`
	Body        string  `json:"body,omitempty"`
}

type AlipayTradePayResponse struct {
	Code       string `json:"code"`
	Msg        string `json:"msg"`
	OutTradeNo string `json:"out_trade_no"`
	TradeNo    string `json:"trade_no"`
}

func (c *AlipayClient) CreateTradePagePay(req *AlipayTradePayRequest) (string, error) {
	params := map[string]string{
		"app_id":        c.config.AppID,
		"method":        "alipay.trade.page.pay",
		"format":        "JSON",
		"return_url":    c.config.ReturnURL,
		"notify_url":    c.config.NotifyURL,
		"charset":       "utf-8",
		"sign_type":     "RSA2",
		"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
		"version":       "1.0",
		"biz_content":   fmt.Sprintf(`{"out_trade_no":"%s","total_amount":"%.2f","subject":"%s","product_code":"FAST_INSTANT_TRADE_PAY"}`, req.OutTradeNo, req.TotalAmount, req.Subject),
	}

	sign, err := c.sign(params)
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}
	params["sign"] = sign

	baseURL := "https://openapi.alipay.com/gateway.do"
	if c.config.Sandbox {
		baseURL = "https://openapi.alipaydev.com/gateway.do"
	}

	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	return fmt.Sprintf("%s?%s", baseURL, values.Encode()), nil
}

func (c *AlipayClient) VerifyNotification(params map[string]string) error {
	sign, ok := params["sign"]
	if !ok {
		return fmt.Errorf("sign not found")
	}
	delete(params, "sign")
	delete(params, "sign_type")

	sortedParams := make([]string, 0, len(params))
	for k, v := range params {
		if v != "" {
			sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", k, v))
		}
	}
	sort.Strings(sortedParams)
	signData := strings.Join(sortedParams, "&")

	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return fmt.Errorf("failed to decode sign: %w", err)
	}

	hashed := sha256.Sum256([]byte(signData))
	err = rsa.VerifyPKCS1v15(c.config.AlipayPublicKey, crypto.SHA256, hashed[:], signBytes)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

func (c *AlipayClient) sign(params map[string]string) (string, error) {
	sortedParams := make([]string, 0, len(params))
	for k, v := range params {
		if v != "" {
			sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", k, v))
		}
	}
	sort.Strings(sortedParams)
	signData := strings.Join(sortedParams, "&")

	hashed := sha256.Sum256([]byte(signData))
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.config.PrivateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

type WechatPayConfig struct {
	AppID      string
	MchID      string
	APIKey     string
	NotifyURL  string
	CertPath   string
	KeyPath    string
	Sandbox    bool
}

type WechatPayClient struct {
	config *WechatPayConfig
	client *http.Client
}

func NewWechatPayClient(config *WechatPayConfig) *WechatPayClient {
	return &WechatPayClient{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type WechatNativePayRequest struct {
	OutTradeNo string  `json:"out_trade_no"`
	TotalFee   int     `json:"total_fee"`
	Body       string  `json:"body"`
}

type WechatNativePayResponse struct {
	CodeURL string `json:"code_url"`
}

func (c *WechatPayClient) CreateNativePay(req *WechatNativePayRequest) (*WechatNativePayResponse, error) {
	params := map[string]string{
		"appid":            c.config.AppID,
		"mch_id":           c.config.MchID,
		"nonce_str":        generateNonceStr(),
		"body":             req.Body,
		"out_trade_no":     req.OutTradeNo,
		"total_fee":        fmt.Sprintf("%d", req.TotalFee),
		"spbill_create_ip": "127.0.0.1",
		"notify_url":       c.config.NotifyURL,
		"trade_type":       "NATIVE",
	}

	sign := c.sign(params)
	params["sign"] = sign

	xmlData := mapToXML(params)

	baseURL := "https://api.mch.weixin.qq.com/pay/unifiedorder"
	if c.config.Sandbox {
		baseURL = "https://api.mch.weixin.qq.com/sandboxnew/pay/unifiedorder"
	}

	resp, err := c.client.Post(baseURL, "application/xml", strings.NewReader(xmlData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	result := parseXMLToMap(string(body))
	if result["return_code"] != "SUCCESS" {
		return nil, fmt.Errorf("wechat pay error: %s", result["return_msg"])
	}
	if result["result_code"] != "SUCCESS" {
		return nil, fmt.Errorf("wechat pay business error: %s", result["err_code_des"])
	}

	return &WechatNativePayResponse{
		CodeURL: result["code_url"],
	}, nil
}

func (c *WechatPayClient) VerifyNotification(xmlData string) (map[string]string, error) {
	params := parseXMLToMap(xmlData)

	if params["return_code"] != "SUCCESS" {
		return nil, fmt.Errorf("return code is not success")
	}

	sign := params["sign"]
	delete(params, "sign")

	expectedSign := c.sign(params)
	if sign != expectedSign {
		return nil, fmt.Errorf("signature verification failed")
	}

	return params, nil
}

func (c *WechatPayClient) sign(params map[string]string) string {
	sortedParams := make([]string, 0, len(params))
	for k, v := range params {
		if v != "" {
			sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", k, v))
		}
	}
	sort.Strings(sortedParams)
	signStr := strings.Join(sortedParams, "&")
	signStr = signStr + "&key=" + c.config.APIKey

	hash := sha256.Sum256([]byte(signStr))
	return strings.ToUpper(fmt.Sprintf("%x", hash))
}

func generateNonceStr() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func mapToXML(params map[string]string) string {
	var builder strings.Builder
	builder.WriteString("<xml>")
	for k, v := range params {
		builder.WriteString(fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k))
	}
	builder.WriteString("</xml>")
	return builder.String()
}

func parseXMLToMap(xmlStr string) map[string]string {
	result := make(map[string]string)
	decoder := json.NewDecoder(strings.NewReader(xmlToJSON(xmlStr)))
	var data map[string]interface{}
	if err := decoder.Decode(&data); err == nil {
		for k, v := range data {
			if s, ok := v.(string); ok {
				result[k] = s
			}
		}
	}
	return result
}

func xmlToJSON(xmlStr string) string {
	return "{}"
}

type PaymentService struct {
	alipay  *AlipayClient
	wechat  *WechatPayClient
}

func NewPaymentService(alipayConfig *AlipayConfig, wechatConfig *WechatPayConfig) *PaymentService {
	return &PaymentService{
		alipay: NewAlipayClient(alipayConfig),
		wechat: NewWechatPayClient(wechatConfig),
	}
}

func (s *PaymentService) CreateAlipayPayment(outTradeNo string, amount float64, subject string) (string, error) {
	req := &AlipayTradePayRequest{
		OutTradeNo:  outTradeNo,
		TotalAmount: amount,
		Subject:     subject,
	}
	return s.alipay.CreateTradePagePay(req)
}

func (s *PaymentService) CreateWechatPayment(outTradeNo string, amount int, body string) (string, error) {
	req := &WechatNativePayRequest{
		OutTradeNo: outTradeNo,
		TotalFee:   amount,
		Body:       body,
	}
	resp, err := s.wechat.CreateNativePay(req)
	if err != nil {
		return "", err
	}
	return resp.CodeURL, nil
}

func (s *PaymentService) VerifyAlipayNotification(params map[string]string) error {
	return s.alipay.VerifyNotification(params)
}

func (s *PaymentService) VerifyWechatNotification(xmlData string) (map[string]string, error) {
	return s.wechat.VerifyNotification(xmlData)
}
