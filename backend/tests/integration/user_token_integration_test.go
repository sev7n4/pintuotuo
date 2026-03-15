package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/services/order"
	"github.com/pintuotuo/backend/services/payment"
	"github.com/pintuotuo/backend/services/token"
	"github.com/pintuotuo/backend/services/user"
)

// TestUserTokenIntegration tests the complete user → token → payment flow
func TestUserTokenIntegration(t *testing.T) {
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	ctx := context.Background()

	t.Run("User registration initializes token balance", func(t *testing.T) {
		// Register new user
		req := &user.RegisterRequest{
			Email:    "tokentest@example.com",
			Password: "TestPass123!",
			Name:     "Token Test User",
		}

		newUser, err := ts.UserService.RegisterUser(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, newUser)

		// Verify token balance was initialized
		balance, err := ts.TokenService.GetBalance(ctx, newUser.ID)
		require.NoError(t, err)
		assert.NotNil(t, balance)
		assert.Equal(t, 0.0, balance.Balance)
		assert.Equal(t, newUser.ID, balance.UserID)
	})

	t.Run("Complete payment → token recharge flow", func(t *testing.T) {
		// Create user
		userId := GenerateUniqueID()
		newUser, err := ts.UserService.RegisterUser(ctx, &user.RegisterRequest{
			Email:    "payment@example.com",
			Password: "TestPass123!",
			Name:     "Payment Test User",
		})
		require.NoError(t, err)

		// Create product
		productId := SeedTestProduct(t, ts.DB, GenerateUniqueID())

		// Create order
		orderReq := &order.CreateOrderRequest{
			ProductID: productId,
			Quantity:  2,
		}
		createdOrder, err := ts.OrderService.CreateOrder(ctx, newUser.ID, orderReq)
		require.NoError(t, err)

		// Get initial token balance
		initialBalance, err := ts.TokenService.GetBalance(ctx, newUser.ID)
		require.NoError(t, err)
		assert.Equal(t, 0.0, initialBalance.Balance)

		// Initiate payment
		paymentReq := &payment.InitiatePaymentRequest{
			OrderID:       createdOrder.ID,
			PaymentMethod: "alipay",
		}
		paymentRecord, err := ts.PaymentService.InitiatePayment(ctx, newUser.ID, paymentReq)
		require.NoError(t, err)
		assert.Equal(t, "pending", paymentRecord.Status)

		// Simulate successful Alipay callback
		alipayCallback := &payment.AlipayCallback{
			OutTradeNo:  string(rune(paymentRecord.ID)),
			TradeNo:     "2024031500001234567",
			TotalAmount: paymentRecord.Amount,
			TradeStatus: "TRADE_SUCCESS",
			Timestamp:   "2024-03-15T10:00:00Z",
			Sign:        "mock_signature",
		}

		updatedPayment, err := ts.PaymentService.HandleAlipayCallback(ctx, alipayCallback)
		require.NoError(t, err)
		assert.Equal(t, "success", updatedPayment.Status)

		// Verify token balance increased
		finalBalance, err := ts.TokenService.GetBalance(ctx, newUser.ID)
		require.NoError(t, err)
		assert.Greater(t, finalBalance.Balance, initialBalance.Balance)
		assert.Equal(t, paymentRecord.Amount, finalBalance.Balance)
		assert.Equal(t, paymentRecord.Amount, finalBalance.TotalEarned)

		// Verify transaction was recorded
		transactions, err := ts.TokenService.GetTransactions(ctx, newUser.ID, nil)
		require.NoError(t, err)
		assert.Len(t, transactions, 1)
		assert.Equal(t, "recharge", transactions[0].Type)
		assert.Equal(t, paymentRecord.Amount, transactions[0].Amount)
	})

	t.Run("Token transfer between users", func(t *testing.T) {
		// Create two users
		user1, err := ts.UserService.RegisterUser(ctx, &user.RegisterRequest{
			Email:    "transfer1@example.com",
			Password: "TestPass123!",
			Name:     "Transfer User 1",
		})
		require.NoError(t, err)

		user2, err := ts.UserService.RegisterUser(ctx, &user.RegisterRequest{
			Email:    "transfer2@example.com",
			Password: "TestPass123!",
			Name:     "Transfer User 2",
		})
		require.NoError(t, err)

		// Recharge user1
		rechargeReq := &token.RechargeTokensRequest{
			UserID: user1.ID,
			Amount: 100.0,
			Reason: "Test recharge",
		}
		_, err = ts.TokenService.RechargeTokens(ctx, rechargeReq)
		require.NoError(t, err)

		// Verify user1 balance
		balance1Before, err := ts.TokenService.GetBalance(ctx, user1.ID)
		require.NoError(t, err)
		assert.Equal(t, 100.0, balance1Before.Balance)

		// Transfer from user1 to user2
		transferReq := &token.TransferTokensRequest{
			SenderID:    user1.ID,
			RecipientID: user2.ID,
			Amount:      30.0,
		}
		err = ts.TokenService.TransferTokens(ctx, transferReq)
		require.NoError(t, err)

		// Verify final balances
		balance1After, err := ts.TokenService.GetBalance(ctx, user1.ID)
		require.NoError(t, err)
		assert.Equal(t, 70.0, balance1After.Balance)

		balance2After, err := ts.TokenService.GetBalance(ctx, user2.ID)
		require.NoError(t, err)
		assert.Equal(t, 30.0, balance2After.Balance)

		// Verify transactions
		user1Txs, err := ts.TokenService.GetTransactions(ctx, user1.ID, nil)
		require.NoError(t, err)
		assert.Len(t, user1Txs, 2) // recharge + transfer_out

		user2Txs, err := ts.TokenService.GetTransactions(ctx, user2.ID, nil)
		require.NoError(t, err)
		assert.Len(t, user2Txs, 1) // transfer_in
	})
}

// Add TestServices fields for token service
func init() {
	// This is handled in helpers.go TestServices struct
}
