package razorpay

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/azcov/gokit/payment"
	rzp "github.com/razorpay/razorpay-go"
)

var _ payment.Gateway = (*Razorpay)(nil)

type Razorpay struct {
	client        *rzp.Client
	keySecret     string
	webhookSecret string
}

type Config struct {
	KeyID         string `config:"key_id"`
	KeySecret     string `config:"key_secret"`
	WebhookSecret string `config:"webhook_secret"`
}

func New(cfg Config) *Razorpay {
	return &Razorpay{
		client:        rzp.NewClient(cfg.KeyID, cfg.KeySecret),
		keySecret:     cfg.KeySecret,
		webhookSecret: cfg.WebhookSecret,
	}
}

func (r *Razorpay) CreateCharge(_ context.Context, req payment.ChargeRequest) (*payment.Charge, error) {
	cur := string(req.Currency)
	if cur == "" {
		cur = "INR"
	}
	data := map[string]any{
		"amount":   req.Amount,
		"currency": cur,
		"receipt":  "rcpt-" + newID(),
	}
	if req.Description != "" {
		data["description"] = req.Description
	}
	if req.Customer.Email != "" {
		data["notes"] = map[string]any{"customer_email": req.Customer.Email}
	}

	order, err := r.client.Order.Create(data, nil)
	if err != nil {
		return nil, fmt.Errorf("razorpay: create order: %w", err)
	}

	id, _ := order["id"].(string)
	return &payment.Charge{
		ID:          id,
		ExternalID:  id,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Status:      payment.StatusPending,
		Description: req.Description,
		CreatedAt:   time.Now(),
		Meta:        order,
	}, nil
}

func (r *Razorpay) GetCharge(_ context.Context, id string) (*payment.Charge, error) {
	order, err := r.client.Order.Fetch(id, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("razorpay: fetch order: %w", err)
	}
	amount, _ := order["amount"].(float64)
	currency, _ := order["currency"].(string)
	return &payment.Charge{
		ID:         id,
		ExternalID: id,
		Amount:     int64(amount),
		Currency:   payment.Currency(currency),
		Status:     razorpayStatus(order),
		CreatedAt:  time.Now(),
		Meta:       order,
	}, nil
}

func (r *Razorpay) Refund(_ context.Context, req payment.RefundRequest) (*payment.RefundResult, error) {
	data := map[string]any{"amount": req.Amount}
	if req.Reason != "" {
		data["notes"] = map[string]any{"reason": req.Reason}
	}
	ref, err := r.client.Payment.Refund(req.ChargeID, int(req.Amount), data, nil)
	if err != nil {
		return nil, fmt.Errorf("razorpay: refund: %w", err)
	}
	refID, _ := ref["id"].(string)
	return &payment.RefundResult{
		ID:       refID,
		ChargeID: req.ChargeID,
		Amount:   req.Amount,
		Status:   payment.StatusPending,
		Meta:     ref,
	}, nil
}

// VerifyWebhook validates Razorpay webhook using HMAC-SHA256.
// Expected header: X-Razorpay-Signature
func (r *Razorpay) VerifyWebhook(_ context.Context, payload []byte, headers map[string]string) (*payment.WebhookEvent, error) {
	sig := headers["X-Razorpay-Signature"]
	if sig == "" {
		sig = headers["x-razorpay-signature"]
	}

	mac := hmac.New(sha256.New, []byte(r.webhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return nil, fmt.Errorf("razorpay: webhook signature mismatch")
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("razorpay: parse webhook: %w", err)
	}

	evType, _ := raw["event"].(string)
	payload_, _ := raw["payload"].(map[string]any)
	paymentEntity, _ := payload_["payment"].(map[string]any)
	entity, _ := paymentEntity["entity"].(map[string]any)
	chargeID, _ := entity["order_id"].(string)

	return &payment.WebhookEvent{
		Type:     evType,
		ChargeID: chargeID,
		Status:   razorpayEventStatus(evType),
		Raw:      raw,
	}, nil
}

func razorpayStatus(order map[string]any) payment.Status {
	s, _ := order["status"].(string)
	switch s {
	case "paid":
		return payment.StatusSettled
	case "created", "attempted":
		return payment.StatusPending
	default:
		return payment.StatusFailed
	}
}

func razorpayEventStatus(evType string) payment.Status {
	switch evType {
	case "payment.captured":
		return payment.StatusSettled
	case "payment.failed":
		return payment.StatusFailed
	default:
		return payment.StatusPending
	}
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
