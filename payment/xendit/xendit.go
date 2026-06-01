package xendit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/azcov/gokit/payment"
	xnd "github.com/xendit/xendit-go"
	"github.com/xendit/xendit-go/invoice"
)

var _ payment.Gateway = (*Xendit)(nil)

type Xendit struct {
	callbackToken string
}

type Config struct {
	SecretKey     string `config:"secret_key"`
	CallbackToken string `config:"callback_token"`
}

func New(cfg Config) *Xendit {
	xnd.Opt.SecretKey = cfg.SecretKey
	return &Xendit{callbackToken: cfg.CallbackToken}
}

func (x *Xendit) CreateCharge(_ context.Context, req payment.ChargeRequest) (*payment.Charge, error) {
	externalID := "xnd-" + newID()
	cur := string(req.Currency)
	if cur == "" {
		cur = "IDR"
	}

	params := &invoice.CreateParams{
		ExternalID:  externalID,
		Amount:      float64(req.Amount),
		Description: req.Description,
		Currency:    cur,
	}
	if req.Customer.Email != "" {
		params.PayerEmail = req.Customer.Email
	}
	if req.SuccessURL != "" {
		params.SuccessRedirectURL = req.SuccessURL
	}
	if req.FailureURL != "" {
		params.FailureRedirectURL = req.FailureURL
	}

	inv, err := invoice.Create(params)
	if err != nil {
		return nil, fmt.Errorf("xendit: create invoice: %w", err)
	}

	c := &payment.Charge{
		ID:          inv.ExternalID,
		ExternalID:  inv.ID,
		Amount:      int64(inv.Amount),
		Currency:    payment.Currency(inv.Currency),
		Status:      mapStatus(inv.Status),
		Description: inv.Description,
		PaymentURL:  inv.InvoiceURL,
		ExpiresAt:   inv.ExpiryDate, // already *time.Time
		CreatedAt:   time.Now(),
	}
	if inv.Created != nil {
		c.CreatedAt = *inv.Created
	}
	return c, nil
}

func (x *Xendit) GetCharge(_ context.Context, id string) (*payment.Charge, error) {
	inv, err := invoice.Get(&invoice.GetParams{ID: id})
	if err != nil {
		return nil, fmt.Errorf("xendit: get invoice: %w", err)
	}
	c := &payment.Charge{
		ID:         inv.ExternalID,
		ExternalID: inv.ID,
		Amount:     int64(inv.Amount),
		Currency:   payment.Currency(inv.Currency),
		Status:     mapStatus(inv.Status),
		PaymentURL: inv.InvoiceURL,
		ExpiresAt:  inv.ExpiryDate,
		CreatedAt:  time.Now(),
	}
	if inv.Created != nil {
		c.CreatedAt = *inv.Created
	}
	return c, nil
}

func (x *Xendit) Refund(_ context.Context, _ payment.RefundRequest) (*payment.RefundResult, error) {
	return nil, fmt.Errorf("xendit: invoice refunds must be initiated via the Xendit dashboard or payment-method refund API")
}

// VerifyWebhook validates Xendit's X-CALLBACK-TOKEN header.
func (x *Xendit) VerifyWebhook(_ context.Context, payload []byte, headers map[string]string) (*payment.WebhookEvent, error) {
	token := headers["x-callback-token"]
	if token == "" {
		token = headers["X-CALLBACK-TOKEN"]
	}
	if x.callbackToken != "" && token != x.callbackToken {
		return nil, fmt.Errorf("xendit: invalid callback token")
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("xendit: parse webhook: %w", err)
	}

	externalID, _ := raw["external_id"].(string)
	status, _ := raw["status"].(string)

	return &payment.WebhookEvent{
		Type:     status,
		ChargeID: externalID,
		Status:   mapStatus(status),
		Raw:      raw,
	}, nil
}

func mapStatus(s string) payment.Status {
	switch s {
	case "PAID", "SETTLED":
		return payment.StatusSettled
	case "EXPIRED":
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
