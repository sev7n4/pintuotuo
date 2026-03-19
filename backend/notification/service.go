package notification

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromName     string
	FromEmail    string
	UseTLS       bool
}

type EmailService struct {
	config *EmailConfig
}

func NewEmailService(config *EmailConfig) *EmailService {
	return &EmailService{config: config}
}

type EmailMessage struct {
	To      string
	Subject string
	Body    string
	IsHTML  bool
}

func (s *EmailService) Send(msg *EmailMessage) error {
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	headers["To"] = msg.To
	headers["Subject"] = msg.Subject
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	if msg.IsHTML {
		headers["MIME-Version"] = "1.0"
		headers["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}

	var body bytes.Buffer
	for k, v := range headers {
		body.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	body.WriteString("\r\n")
	body.WriteString(msg.Body)

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	if s.config.UseTLS {
		return s.sendWithTLS(addr, auth, s.config.FromEmail, []string{msg.To}, body.Bytes())
	}

	return smtp.SendMail(addr, auth, s.config.FromEmail, []string{msg.To}, body.Bytes())
}

func (s *EmailService) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.config.SMTPHost,
	})
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client, err2 := smtp.NewClient(conn, s.config.SMTPHost)
	if err2 != nil {
		return fmt.Errorf("failed to create client: %w", err2)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	w, err3 := client.Data()
	if err3 != nil {
		return fmt.Errorf("failed to get data writer: %w", err3)
	}

	_, err4 := w.Write(msg)
	if err4 != nil {
		return fmt.Errorf("failed to write message: %w", err4)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return client.Quit()
}

type EmailTemplate struct {
	Welcome     *template.Template
	ResetPassword *template.Template
	OrderConfirm  *template.Template
	PaymentSuccess *template.Template
	LowBalance    *template.Template
}

var emailTemplates *EmailTemplate

func InitEmailTemplates() {
	emailTemplates = &EmailTemplate{
		Welcome: template.Must(template.New("welcome").Parse(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 10px 10px 0 0;">
    <h1 style="margin: 0;">欢迎加入拼脱脱！</h1>
  </div>
  <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
    <p>亲爱的 {{.Name}}，</p>
    <p>感谢您注册拼脱脱！我们很高兴为您提供服务。</p>
    <p>您现在可以：</p>
    <ul>
      <li>购买和出售 AI Token</li>
      <li>参与拼团活动享受优惠</li>
      <li>邀请好友获得返利</li>
      <li>使用我们的 API 代理服务</li>
    </ul>
    <p>如有任何问题，请随时联系我们的客服团队。</p>
    <p>祝您使用愉快！</p>
    <p style="color: #888;">拼脱脱团队</p>
  </div>
</body>
</html>
`)),
		ResetPassword: template.Must(template.New("reset_password").Parse(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: #f5222d; color: white; padding: 30px; border-radius: 10px 10px 0 0;">
    <h1 style="margin: 0;">密码重置</h1>
  </div>
  <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
    <p>您好，</p>
    <p>我们收到了您的密码重置请求。</p>
    <p>您的验证码是：<strong style="font-size: 24px; color: #f5222d;">{{.Code}}</strong></p>
    <p>验证码有效期为 15 分钟，请尽快使用。</p>
    <p>如果您没有请求重置密码，请忽略此邮件。</p>
    <p style="color: #888;">拼脱脱团队</p>
  </div>
</body>
</html>
`)),
		OrderConfirm: template.Must(template.New("order_confirm").Parse(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: #52c41a; color: white; padding: 30px; border-radius: 10px 10px 0 0;">
    <h1 style="margin: 0;">订单确认</h1>
  </div>
  <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
    <p>亲爱的 {{.Name}}，</p>
    <p>您的订单已创建成功！</p>
    <div style="background: white; padding: 20px; border-radius: 5px; margin: 20px 0;">
      <p><strong>订单号：</strong>{{.OrderID}}</p>
      <p><strong>商品：</strong>{{.ProductName}}</p>
      <p><strong>数量：</strong>{{.Quantity}}</p>
      <p><strong>金额：</strong>¥{{.Amount}}</p>
    </div>
    <p>请尽快完成支付，订单将在 30 分钟后自动取消。</p>
    <p style="color: #888;">拼脱脱团队</p>
  </div>
</body>
</html>
`)),
		PaymentSuccess: template.Must(template.New("payment_success").Parse(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: #52c41a; color: white; padding: 30px; border-radius: 10px 10px 0 0;">
    <h1 style="margin: 0;">支付成功</h1>
  </div>
  <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
    <p>亲爱的 {{.Name}}，</p>
    <p>您的支付已成功完成！</p>
    <div style="background: white; padding: 20px; border-radius: 5px; margin: 20px 0;">
      <p><strong>订单号：</strong>{{.OrderID}}</p>
      <p><strong>支付金额：</strong>¥{{.Amount}}</p>
      <p><strong>支付方式：</strong>{{.PayMethod}}</p>
      <p><strong>支付时间：</strong>{{.PaidAt}}</p>
    </div>
    <p>感谢您的购买！</p>
    <p style="color: #888;">拼脱脱团队</p>
  </div>
</body>
</html>
`)),
		LowBalance: template.Must(template.New("low_balance").Parse(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: #faad14; color: white; padding: 30px; border-radius: 10px 10px 0 0;">
    <h1 style="margin: 0;">余额不足提醒</h1>
  </div>
  <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
    <p>亲爱的 {{.Name}}，</p>
    <p>您的账户余额已不足，请及时充值。</p>
    <div style="background: white; padding: 20px; border-radius: 5px; margin: 20px 0;">
      <p><strong>当前余额：</strong>${{.Balance}}</p>
      <p><strong>阈值：</strong>${{.Threshold}}</p>
    </div>
    <p>充值后即可继续使用我们的服务。</p>
    <p style="color: #888;">拼脱脱团队</p>
  </div>
</body>
</html>
`)),
	}
}

func (s *EmailService) SendWelcomeEmail(to, name string) error {
	var body bytes.Buffer
	if err := emailTemplates.Welcome.Execute(&body, map[string]string{"Name": name}); err != nil {
		return err
	}
	return s.Send(&EmailMessage{
		To:      to,
		Subject: "欢迎加入拼脱脱！",
		Body:    body.String(),
		IsHTML:  true,
	})
}

func (s *EmailService) SendResetPasswordEmail(to, code string) error {
	var body bytes.Buffer
	if err := emailTemplates.ResetPassword.Execute(&body, map[string]string{"Code": code}); err != nil {
		return err
	}
	return s.Send(&EmailMessage{
		To:      to,
		Subject: "拼脱脱 - 密码重置验证码",
		Body:    body.String(),
		IsHTML:  true,
	})
}

func (s *EmailService) SendOrderConfirmEmail(to, name string, data map[string]interface{}) error {
	data["Name"] = name
	var body bytes.Buffer
	if err := emailTemplates.OrderConfirm.Execute(&body, data); err != nil {
		return err
	}
	return s.Send(&EmailMessage{
		To:      to,
		Subject: fmt.Sprintf("拼脱脱 - 订单确认 #%v", data["OrderID"]),
		Body:    body.String(),
		IsHTML:  true,
	})
}

func (s *EmailService) SendPaymentSuccessEmail(to, name string, data map[string]interface{}) error {
	data["Name"] = name
	var body bytes.Buffer
	if err := emailTemplates.PaymentSuccess.Execute(&body, data); err != nil {
		return err
	}
	return s.Send(&EmailMessage{
		To:      to,
		Subject: fmt.Sprintf("拼脱脱 - 支付成功 #%v", data["OrderID"]),
		Body:    body.String(),
		IsHTML:  true,
	})
}

func (s *EmailService) SendLowBalanceEmail(to, name string, balance, threshold float64) error {
	var body bytes.Buffer
	if err := emailTemplates.LowBalance.Execute(&body, map[string]interface{}{
		"Name":      name,
		"Balance":   balance,
		"Threshold": threshold,
	}); err != nil {
		return err
	}
	return s.Send(&EmailMessage{
		To:      to,
		Subject: "拼脱脱 - 余额不足提醒",
		Body:    body.String(),
		IsHTML:  true,
	})
}

type PushConfig struct {
	FCMServerKey string
	APNSCertPath string
	APNSKeyPath  string
}

type PushService struct {
	config *PushConfig
	client *http.Client
}

func NewPushService(config *PushConfig) *PushService {
	return &PushService{
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type PushMessage struct {
	Token       string
	Title       string
	Body        string
	Data        map[string]string
	ClickAction string
}

func (s *PushService) SendFCM(msg *PushMessage) error {
	payload := map[string]interface{}{
		"to": msg.Token,
		"notification": map[string]interface{}{
			"title":       msg.Title,
			"body":        msg.Body,
			"click_action": msg.ClickAction,
		},
		"data": msg.Data,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+s.config.FCMServerKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("FCM request failed with status %d", resp.StatusCode)
	}

	return nil
}

type NotificationService struct {
	email *EmailService
	push  *PushService
}

func NewNotificationService(emailConfig *EmailConfig, pushConfig *PushConfig) *NotificationService {
	InitEmailTemplates()
	return &NotificationService{
		email: NewEmailService(emailConfig),
		push:  NewPushService(pushConfig),
	}
}

type NotificationType string

const (
	NotificationWelcome       NotificationType = "welcome"
	NotificationResetPassword NotificationType = "reset_password"
	NotificationOrderConfirm  NotificationType = "order_confirm"
	NotificationPaymentSuccess NotificationType = "payment_success"
	NotificationLowBalance    NotificationType = "low_balance"
)

type NotificationRequest struct {
	Type     NotificationType
	UserID   int
	Email    string
	Name     string
	Data     map[string]interface{}
	PushToken string
}

func (s *NotificationService) Send(req *NotificationRequest) error {
	var err error

	switch req.Type {
	case NotificationWelcome:
		err = s.email.SendWelcomeEmail(req.Email, req.Name)
	case NotificationResetPassword:
		if code, ok := req.Data["code"].(string); ok {
			err = s.email.SendResetPasswordEmail(req.Email, code)
		}
	case NotificationOrderConfirm:
		err = s.email.SendOrderConfirmEmail(req.Email, req.Name, req.Data)
	case NotificationPaymentSuccess:
		err = s.email.SendPaymentSuccessEmail(req.Email, req.Name, req.Data)
	case NotificationLowBalance:
		balance, _ := req.Data["balance"].(float64)
		threshold, _ := req.Data["threshold"].(float64)
		err = s.email.SendLowBalanceEmail(req.Email, req.Name, balance, threshold)
	}

	if req.PushToken != "" && s.push != nil {
		pushErr := s.push.SendFCM(&PushMessage{
			Token: req.PushToken,
			Title: s.getPushTitle(req.Type),
			Body:  s.getPushBody(req.Type, req.Data),
			Data:  s.convertData(req.Data),
		})
		if pushErr != nil && err == nil {
			err = pushErr
		}
	}

	return err
}

func (s *NotificationService) getPushTitle(t NotificationType) string {
	titles := map[NotificationType]string{
		NotificationWelcome:        "欢迎加入拼脱脱！",
		NotificationResetPassword:  "密码重置验证码",
		NotificationOrderConfirm:   "订单创建成功",
		NotificationPaymentSuccess: "支付成功",
		NotificationLowBalance:     "余额不足提醒",
	}
	return titles[t]
}

func (s *NotificationService) getPushBody(t NotificationType, data map[string]interface{}) string {
	switch t {
	case NotificationOrderConfirm:
		return fmt.Sprintf("订单 #%v 已创建，请尽快支付", data["OrderID"])
	case NotificationPaymentSuccess:
		return fmt.Sprintf("订单 #%v 支付成功，金额 ¥%v", data["OrderID"], data["Amount"])
	case NotificationLowBalance:
		return fmt.Sprintf("您的余额已不足 $%v，请及时充值", data["Balance"])
	default:
		return ""
	}
}

func (s *NotificationService) convertData(data map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range data {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

func (s *NotificationService) SendBatchEmails(recipients []string, subject, body string) error {
	for _, to := range recipients {
		if err := s.email.Send(&EmailMessage{
			To:      to,
			Subject: subject,
			Body:    body,
			IsHTML:  strings.Contains(body, "<html"),
		}); err != nil {
			return err
		}
	}
	return nil
}
