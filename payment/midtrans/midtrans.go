package midtrans

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/azcov/gokit/payment"
	mt "github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

var _ payment.Gateway = (*Midtrans)(nil)

type Midtrans struct {
	snap      snap.Client
	core      coreapi.Client
	serverKey string
}

type Config struct {
	ServerKey    string `config:"server_key"`
	IsProduction bool   `config:"is_production"`
}

func New(cfg Config) *Midtrans {
	env := mt.Sandbox
	if cfg.IsProduction {
		env = mt.Production
	}
	m := &Midtrans{serverKey: cfg.ServerKey}
	m.snap.New(cfg.ServerKey, env)
	m.core.New(cfg.ServerKey, env)
	return m
}

func (m *Midtrans) CreateCharge(_ context.Context, req payment.ChargeRequest) (*payment.Charge, error) {
	orderID := "order-" + newID()
	snapReq := &snap.Request{
		TransactionDetails: mt.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: req.Amount,
		},
		CustomerDetail: &mt.CustomerDetails{
			Email: req.Customer.Email,
			FName: req.Customer.Name,
			Phone: req.Customer.Phone,
		},
	}
	if req.SuccessURL != "" {
		snapReq.Callbacks = &snap.Callbacks{Finish: req.SuccessURL}
	}

	resp, err := m.snap.CreateTransaction(snapReq)
	if err != nil {
		return nil, fmt.Errorf("midtrans: %w", err)
	}

	cur := req.Currency
	if cur == "" {
		cur = payment.IDR
	}
	return &payment.Charge{
		ID:          orderID,
		ExternalID:  orderID,
		Amount:      req.Amount,
		Currency:    cur,
		Status:      payment.StatusPending,
		Description: req.Description,
		PaymentURL:  resp.RedirectURL,
		CreatedAt:   time.Now(),
		Meta:        map[string]any{"snap_token": resp.Token},
	}, nil
}

func (m *Midtrans) GetCharge(_ context.Context, id string) (*payment.Charge, error) {
	resp, err := m.core.CheckTransaction(id)
	if err != nil {
		return nil, fmt.Errorf("midtrans: %w", err)
	}
	// GrossAmount is a decimal string like "10000.00"
	gross, _ := strconv.ParseFloat(resp.GrossAmount, 64)
	return &payment.Charge{
		ID:         resp.OrderID,
		ExternalID: resp.TransactionID,
		Amount:     int64(gross),
		Currency:   payment.IDR,
		Status:     mapStatus(resp.TransactionStatus),
		CreatedAt:  time.Now(),
		Meta: map[string]any{
			"payment_type": resp.PaymentType,
			"status_code":  resp.StatusCode,
		},
	}, nil
}

func (m *Midtrans) Refund(_ context.Context, req payment.RefundRequest) (*payment.RefundResult, error) {
	resp, err := m.core.RefundTransaction(req.ChargeID, &coreapi.RefundReq{
		RefundKey: newID(),
		Amount:    req.Amount,
		Reason:    req.Reason,
	})
	if err != nil {
		return nil, fmt.Errorf("midtrans: refund: %w", err)
	}
	return &payment.RefundResult{
		ID:       resp.RefundChargebackUUID,
		ChargeID: req.ChargeID,
		Amount:   req.Amount,
		Status:   payment.StatusPending,
	}, nil
}

// VerifyWebhook validates the Midtrans notification signature.
// Signature = SHA512(order_id + status_code + gross_amount + server_key)
func (m *Midtrans) VerifyWebhook(_ context.Context, payload []byte, _ map[string]string) (*payment.WebhookEvent, error) {
	var notif map[string]any
	if err := json.Unmarshal(payload, &notif); err != nil {
		return nil, fmt.Errorf("midtrans: parse webhook: %w", err)
	}

	orderID, _ := notif["order_id"].(string)
	statusCode, _ := notif["status_code"].(string)
	grossAmount, _ := notif["gross_amount"].(string)
	sigKey, _ := notif["signature_key"].(string)

	raw := orderID + statusCode + grossAmount + m.serverKey
	hash := sha512.Sum512([]byte(raw))
	if hex.EncodeToString(hash[:]) != sigKey {
		return nil, fmt.Errorf("midtrans: webhook signature mismatch")
	}

	txStatus, _ := notif["transaction_status"].(string)
	// Strip decimal from gross_amount for the amount field
	gross, _ := strconv.ParseFloat(strings.TrimSpace(grossAmount), 64)

	return &payment.WebhookEvent{
		Type:     txStatus,
		ChargeID: orderID,
		Status:   mapStatus(txStatus),
		Raw: map[string]any{
			"notification": notif,
			"amount":       int64(gross),
		},
	}, nil
}

func mapStatus(s string) payment.Status {
	switch s {
	case "capture":
		return payment.StatusCapture
	case "settlement":
		return payment.StatusSettled
	case "pending":
		return payment.StatusPending
	case "deny", "failure":
		return payment.StatusFailed
	case "cancel":
		return payment.StatusCanceled
	case "expire":
		return payment.StatusExpired
	default:
		return payment.StatusPending
	}
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
