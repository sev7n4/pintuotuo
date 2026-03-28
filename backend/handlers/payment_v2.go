package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
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

func InitPaymentService(alipayAppID string, wechatAppID string, wechatMchID string) {
	paymentService = payment.NewPaymentService(
		&payment.AlipayConfig{
			AppID:   alipayAppID,
			Sandbox: true,
		},
		&payment.WechatPayConfig{
			AppID:   wechatAppID,
			MchID:   wechatMchID,
			Sandbox: true,
		},
	)
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
		log.Printf("[CreatePayment] Bind error: %v, userID: %d", err, userIDInt)
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	log.Printf("[CreatePayment] Request: order_id=%d, pay_method=%s, amount=%.2f, userID=%d", req.OrderID, req.PayMethod, req.Amount, userIDInt)

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
		log.Printf("[CreatePayment] Order query error: %v, order_id: %d", err, req.OrderID)
		middleware.RespondWithError(c, apperrors.ErrOrderNotFound)
		return
	}
	log.Printf("[CreatePayment] Order found: id=%d, user_id=%d, status=%s", order.ID, order.UserID, order.Status)

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
		log.Printf("[CreatePayment] Payment insert error: %v", err)
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PAYMENT_CREATE_FAILED",
			"Failed to create payment record",
			http.StatusInternalServerError,
			err,
		))
		return
	}
	log.Printf("[CreatePayment] Payment created: id=%d, out_trade_no=%s", paymentID, outTradeNo)

	var payURL, qrCodeURL string

	switch strings.ToLower(req.PayMethod) {
	case "alipay":
		payURL, err = paymentService.CreateAlipayPayment(outTradeNo, req.Amount, fmt.Sprintf("订单%d", req.OrderID))
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"ALIPAY_CREATE_FAILED",
				"Failed to create Alipay payment",
				http.StatusInternalServerError,
				err,
			))
			return
		}
	case "wechat":
		qrCodeURL, err = paymentService.CreateWechatPayment(outTradeNo, int(req.Amount*100), fmt.Sprintf("订单%d", req.OrderID))
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"WECHAT_CREATE_FAILED",
				"Failed to create WeChat payment",
				http.StatusInternalServerError,
				err,
			))
			return
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

		c.JSON(http.StatusOK, CreatePaymentResponse{
			PaymentID:  paymentID,
			OutTradeNo: outTradeNo,
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

	c.JSON(http.StatusOK, CreatePaymentResponse{
		PaymentID:  paymentID,
		PayURL:     payURL,
		QRCodeURL:  qrCodeURL,
		OutTradeNo: outTradeNo,
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
		"UPDATE orders SET status = 'paid', paid_at = $1 WHERE id = $2",
		time.Now(), orderID,
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
	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	params := make(map[string]string)
	for k, v := range c.Request.Form {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	if err := paymentService.VerifyAlipayNotification(params); err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	if params["trade_status"] != "TRADE_SUCCESS" && params["trade_status"] != "TRADE_FINISHED" {
		c.String(http.StatusOK, "success")
		return
	}

	outTradeNo := params["out_trade_no"]
	tradeNo := params["trade_no"]

	db := config.GetDB()
	if db == nil {
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
		c.String(http.StatusOK, "success")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.String(http.StatusInternalServerError, "fail")
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"UPDATE payments SET status = 'success', transaction_id = $1, paid_at = $2 WHERE id = $3",
		tradeNo, time.Now(), paymentID,
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	_, err = tx.Exec(
		"UPDATE orders SET status = 'paid', paid_at = $1 WHERE id = $2",
		time.Now(), orderID,
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	if err := tx.Commit(); err != nil {
		c.String(http.StatusInternalServerError, "fail")
		return
	}

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
		"UPDATE orders SET status = 'paid', paid_at = $1 WHERE id = $2",
		time.Now(), orderID,
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
