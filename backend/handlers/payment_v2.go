package handlers

import (
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/payment"
)

var paymentService *payment.PaymentService

func parsePrivateKey(keyStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte("-----BEGIN RSA PRIVATE KEY-----\n" + keyStr + "\n-----END RSA PRIVATE KEY-----"))
	if block == nil {
		block, _ = pem.Decode([]byte("-----BEGIN PRIVATE KEY-----\n" + keyStr + "\n-----END PRIVATE KEY-----"))
	}
	if block == nil {
		return nil, fmt.Errorf("failed to parse private key PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return rsaKey, nil
}

func parsePublicKey(keyStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\n" + keyStr + "\n-----END PUBLIC KEY-----"))
	if block == nil {
		return nil, fmt.Errorf("failed to parse public key PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaPub, nil
}

func InitPaymentService() {
	appID := os.Getenv("ALIPAY_APP_ID")
	privateKeyStr := os.Getenv("ALIPAY_PRIVATE_KEY")
	publicKeyStr := os.Getenv("ALIPAY_PUBLIC_KEY")

	if appID == "" || privateKeyStr == "" || publicKeyStr == "" {
		log.Println("[Payment] Alipay config not complete, using mock mode")
		paymentService = payment.NewPaymentService(
			&payment.AlipayConfig{
				AppID:   "",
				Sandbox: true,
			},
			&payment.WechatPayConfig{
				AppID:   "",
				MchID:   "",
				Sandbox: true,
			},
		)
		return
	}

	privateKey, err := parsePrivateKey(strings.ReplaceAll(privateKeyStr, " ", ""))
	if err != nil {
		log.Printf("[Payment] Failed to parse private key: %v", err)
		return
	}

	publicKey, err := parsePublicKey(strings.ReplaceAll(publicKeyStr, " ", ""))
	if err != nil {
		log.Printf("[Payment] Failed to parse public key: %v", err)
		return
	}

	paymentService = payment.NewPaymentService(
		&payment.AlipayConfig{
			AppID:           appID,
			PrivateKey:      privateKey,
			AlipayPublicKey: publicKey,
			ReturnURL:       "http://119.29.173.89/orders",
			NotifyURL:       "http://119.29.173.89:8080/api/v1/payments/webhooks/alipay",
			Sandbox:         true,
		},
		&payment.WechatPayConfig{
			AppID:   os.Getenv("WECHAT_APP_ID"),
			MchID:   os.Getenv("WECHAT_MCH_ID"),
			Sandbox: true,
		},
	)
	log.Printf("[Payment] Alipay initialized with AppID: %s", appID)
}

func IsPaymentServiceInitialized() bool {
	return paymentService != nil
}

type CreatePaymentRequest struct {
	OrderID   int     `json:"order_id" binding:"required"`
	PayMethod string  `json:"pay_method" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
}

type CreatePaymentResponse struct {
	PaymentID  int    `json:"payment_id"`
	PayURL     string `json:"pay_url"`
	QRCodeURL  string `json:"qrcode_url,omitempty"`
	OutTradeNo string `json:"out_trade_no"`
}

func CreatePayment(c *gin.Context) {
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

	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var order models.Order
	err := db.QueryRow(
		"SELECT id, user_id, total_price, status FROM orders WHERE id = $1",
		req.OrderID,
	).Scan(&order.ID, &order.UserID, &order.TotalPrice, &order.Status)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrOrderNotFound)
		return
	}

	if order.UserID != userIDInt {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	if order.Status != paymentStatusPending {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"ORDER_NOT_PENDING",
			"Order is not in pending status",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	outTradeNo := fmt.Sprintf("PT%d%d", time.Now().UnixNano(), req.OrderID)

	var paymentID int
	err = db.QueryRow(
		`INSERT INTO payments (order_id, user_id, amount, pay_method, out_trade_no, status) 
		 VALUES ($1, $2, $3, $4, $5, 'pending') 
		 RETURNING id`,
		req.OrderID, userIDInt, req.Amount, req.PayMethod, outTradeNo,
	).Scan(&paymentID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PAYMENT_CREATE_FAILED",
			"Failed to create payment record",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	var payURL, qrCodeURL string

	log.Printf("[CreatePayment] Creating payment with method: %s, paymentService: %v", req.PayMethod, paymentService)
	if paymentService == nil {
		log.Printf("[CreatePayment] ERROR: paymentService is nil!")
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PAYMENT_SERVICE_NOT_INITIALIZED",
			"Payment service is not initialized",
			http.StatusInternalServerError,
			nil,
		))
		return
	}
	switch strings.ToLower(req.PayMethod) {
	case "alipay":
		payURL, err = paymentService.CreateAlipayPayment(outTradeNo, req.Amount, fmt.Sprintf("订单%d", req.OrderID))
		if err != nil {
			log.Printf("[CreatePayment] Alipay create error: %v", err)
			payURL = fmt.Sprintf("https://mock-payment.example.com/pay?out_trade_no=%s&amount=%.2f", outTradeNo, req.Amount)
			log.Printf("[CreatePayment] Using mock payment URL: %s", payURL)
		}
	case "wechat":
		qrCodeURL, err = paymentService.CreateWechatPayment(outTradeNo, int(req.Amount*100), fmt.Sprintf("订单%d", req.OrderID))
		if err != nil {
			log.Printf("[CreatePayment] Wechat create error: %v", err)
			qrCodeURL = fmt.Sprintf("weixin://wxpay/bizpayurl?pr=%s", outTradeNo)
			log.Printf("[CreatePayment] Using mock wechat QR code URL: %s", qrCodeURL)
		}
	case "balance":
		engine := billing.GetBillingEngine()
		balance, err := engine.GetBalance(userIDInt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"BALANCE_CHECK_FAILED",
				"Failed to check balance",
				http.StatusInternalServerError,
				err,
			))
			return
		}

		if balance < req.Amount {
			middleware.RespondWithError(c, apperrors.ErrInsufficientBalance)
			return
		}

		err = processBalancePayment(db, paymentID, req.OrderID, userIDInt, req.Amount)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"BALANCE_PAYMENT_FAILED",
				err.Error(),
				http.StatusInternalServerError,
				err,
			))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"id":           paymentID,
				"order_id":     req.OrderID,
				"amount":       req.Amount,
				"pay_method":   req.PayMethod,
				"status":       "success",
				"out_trade_no": outTradeNo,
			},
			"message": "Payment completed successfully",
		})
		return
	default:
		middleware.RespondWithError(c, apperrors.NewAppError(
			"UNSUPPORTED_PAY_METHOD",
			"Unsupported payment method",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           paymentID,
			"order_id":     req.OrderID,
			"amount":       req.Amount,
			"pay_method":   req.PayMethod,
			"status":       "pending",
			"pay_url":      payURL,
			"qrcode_url":   qrCodeURL,
			"out_trade_no": outTradeNo,
		},
		"message": "Payment initiated successfully",
	})
}

func processBalancePayment(db *sql.DB, paymentID, orderID, userID int, amount float64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"UPDATE payments SET status = 'success', paid_at = $1 WHERE id = $2",
		time.Now(), paymentID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"UPDATE orders SET status = 'paid' WHERE id = $1",
		orderID,
	)
	if err != nil {
		return err
	}

	engine := billing.GetBillingEngine()
	if err := engine.DeductBalance(userID, amount, fmt.Sprintf("订单支付 #%d", orderID), ""); err != nil {
		return err
	}

	return tx.Commit()
}

func AlipayNotify(c *gin.Context) {
	log.Printf("[AlipayNotify] Received webhook callback")
	if err := c.Request.ParseForm(); err != nil {
		log.Printf("[AlipayNotify] ParseForm error: %v", err)
		c.String(http.StatusBadRequest, "fail")
		return
	}

	params := make(map[string]string)
	for k, v := range c.Request.Form {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	log.Printf("[AlipayNotify] Params: trade_status=%s, out_trade_no=%s", params["trade_status"], params["out_trade_no"])

	if err := paymentService.VerifyAlipayNotification(params); err != nil {
		log.Printf("[AlipayNotify] Verify error: %v", err)
		c.String(http.StatusBadRequest, "fail")
		return
	}

	log.Printf("[AlipayNotify] Verification passed, trade_status=%s", params["trade_status"])

	if params["trade_status"] != "TRADE_SUCCESS" && params["trade_status"] != "TRADE_FINISHED" {
		log.Printf("[AlipayNotify] Trade status not success, returning success")
		c.String(http.StatusOK, "success")
		return
	}

	outTradeNo := params["out_trade_no"]
	tradeNo := params["trade_no"]
	log.Printf("[AlipayNotify] Processing payment: out_trade_no=%s, trade_no=%s", outTradeNo, tradeNo)

	db := config.GetDB()
	if db == nil {
		log.Printf("[AlipayNotify] DB is nil!")
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	var paymentID, orderID, userID int
	var amount float64
	err := db.QueryRow(
		"SELECT id, order_id, user_id, amount FROM payments WHERE out_trade_no = $1",
		outTradeNo,
	).Scan(&paymentID, &orderID, &userID, &amount)
	if err != nil {
		log.Printf("[AlipayNotify] Payment not found: %v", err)
		c.String(http.StatusOK, "success")
		return
	}
	log.Printf("[AlipayNotify] Found payment: id=%d, order_id=%d, amount=%.2f", paymentID, orderID, amount)

	tx, err := db.Begin()
	if err != nil {
		log.Printf("[AlipayNotify] Begin transaction error: %v", err)
		c.String(http.StatusInternalServerError, "fail")
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"UPDATE payments SET status = 'success', transaction_id = $1, paid_at = $2 WHERE id = $3",
		tradeNo, time.Now(), paymentID,
	)
	if err != nil {
		log.Printf("[AlipayNotify] Update payment error: %v", err)
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	_, err = tx.Exec(
		"UPDATE orders SET status = 'paid' WHERE id = $1",
		orderID,
	)
	if err != nil {
		log.Printf("[AlipayNotify] Update order error: %v", err)
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[AlipayNotify] Commit error: %v", err)
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	log.Printf("[AlipayNotify] Payment completed successfully: order_id=%d", orderID)
	c.String(http.StatusOK, "success")
}

func WechatNotify(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusBadRequest, "<xml><return_code><![CDATA[FAIL]]></return_code></xml>")
		return
	}

	params, err := paymentService.VerifyWechatNotification(string(body))
	if err != nil {
		c.String(http.StatusBadRequest, "<xml><return_code><![CDATA[FAIL]]></return_code></xml>")
		return
	}

	if params["result_code"] != "SUCCESS" {
		c.String(http.StatusOK, "<xml><return_code><![CDATA[SUCCESS]]></return_code></xml>")
		return
	}

	outTradeNo := params["out_trade_no"]
	transactionID := params["transaction_id"]

	db := config.GetDB()
	if db == nil {
		c.String(http.StatusInternalServerError, "<xml><return_code><![CDATA[FAIL]]></return_code></xml>")
		return
	}

	var paymentID, orderID int
	err = db.QueryRow(
		"SELECT id, order_id FROM payments WHERE out_trade_no = $1",
		outTradeNo,
	).Scan(&paymentID, &orderID)
	if err != nil {
		c.String(http.StatusOK, "<xml><return_code><![CDATA[SUCCESS]]></return_code></xml>")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.String(http.StatusInternalServerError, "<xml><return_code><![CDATA[FAIL]]></return_code></xml>")
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"UPDATE payments SET status = 'success', transaction_id = $1, paid_at = $2 WHERE id = $3",
		transactionID, time.Now(), paymentID,
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "<xml><return_code><![CDATA[FAIL]]></return_code></xml>")
		return
	}

	_, err = tx.Exec(
		"UPDATE orders SET status = 'paid' WHERE id = $1",
		orderID,
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "<xml><return_code><![CDATA[FAIL]]></return_code></xml>")
		return
	}

	if err := tx.Commit(); err != nil {
		c.String(http.StatusInternalServerError, "<xml><return_code><![CDATA[FAIL]]></return_code></xml>")
		return
	}

	c.String(http.StatusOK, "<xml><return_code><![CDATA[SUCCESS]]></return_code></xml>")
}

func GetPaymentStatus(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := strconv.Atoi(paymentIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var payment models.Payment
	err = db.QueryRow(
		"SELECT id, order_id, user_id, amount, pay_method, out_trade_no, transaction_id, status, paid_at, created_at FROM payments WHERE id = $1",
		paymentID,
	).Scan(&payment.ID, &payment.OrderID, &payment.UserID, &payment.Amount, &payment.PayMethod, &payment.OutTradeNo, &payment.TransactionID, &payment.Status, &payment.PaidAt, &payment.CreatedAt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PAYMENT_NOT_FOUND",
			"Payment not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, payment)
}
