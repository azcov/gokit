package mayar

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/azcov/go-mayar-sdk"
	"github.com/azcov/go-mayar-sdk/webhook"
	"github.com/azcov/gokit/payment"
)

var _ payment.Gateway = (*Mayar)(nil)

type Mayar struct {
	c            *mayar.Client
	webhookKey   string
}

type Config struct {
	APIKey        string `config:"api_key"`
	WebhookSecret string `config:"webhook_secret"`
	Sandbox       bool   `config:"sandbox"`
}

func New(cfg Config) *Mayar {
	opts := []mayar.Option{}
	if cfg.Sandbox {
		opts = append(opts, mayar.WithSandbox())
	}
	return &Mayar{
		c:          mayar.New(cfg.APIKey, opts...),
		webhookKey: cfg.WebhookSecret,
	}
}

func (m *Mayar) CreateCharge(ctx context.Context, req payment.ChargeRequest) (*payment.Charge, error) {
	params := mayar.CreatePaymentRequestParams{
		Name:        req.Customer.Name,
		Email:       req.Customer.Email,
		Mobile:      req.Customer.Phone,
		Amount:      req.Amount,
		Description: req.Description,
		RedirectURL: firstNonEmpty(req.SuccessURL, req.CancelURL),
	}
	if req.ExpiresAt != nil {
		params.ExpiredAt = req.ExpiresAt.Format(time.RFC3339)
	}

	pr, err := m.c.PaymentRequest.Create(ctx, params)
	if err != nil {
		return nil, err
	}
	return toCharge(pr, req), nil
}

func (m *Mayar) GetCharge(ctx context.Context, id string) (*payment.Charge, error) {
	pr, err := m.c.PaymentRequest.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return toCharge(pr, payment.ChargeRequest{}), nil
}

func (m *Mayar) Refund(_ context.Context, _ payment.RefundRequest) (*payment.RefundResult, error) {
	return nil, fmt.Errorf("mayar: refunds are not supported by the API; cancel or reopen the payment request instead")
}

func (m *Mayar) VerifyWebhook(_ context.Context, payload []byte, _ map[string]string) (*payment.WebhookEvent, error) {
	event, err := webhook.Parse(payload)
	if err != nil {
		return nil, fmt.Errorf("mayar: webhook parse failed: %w", err)
	}
	return &payment.WebhookEvent{
		Type:     string(event.Event),
		ChargeID: event.Data.ID,
		Status:   mapBoolStatus(event.Data.Status),
		Raw:      rawEvent(event),
	}, nil
}

func toCharge(pr *mayar.PaymentRequest, req payment.ChargeRequest) *payment.Charge {
	c := &payment.Charge{
		ID:          pr.ID,
		ExternalID:  pr.ID,
		Amount:      pr.Amount,
		Currency:    payment.Currency("IDR"),
		Method:      payment.MethodEWallet,
		Status:      mapStringStatus(pr.Status),
		Description: pr.Description,
		Customer: payment.Customer{
			Name:  pr.Name,
			Email: pr.Email,
			Phone: pr.Mobile,
		},
		PaymentURL: pr.Link,
		CreatedAt:  pr.CreatedAt.Time(),
		UpdatedAt:  pr.UpdatedAt.Time(),
		Meta:       map[string]any{"redirect_url": pr.RedirectURL},
	}
	if pr.ExpiredAt != "" {
		if t, err := time.Parse(time.RFC3339, pr.ExpiredAt); err == nil {
			c.ExpiresAt = &t
		}
	}
	if c.PaymentURL == "" {
		c.PaymentURL = pr.RedirectURL
	}
	if c.Description == "" {
		c.Description = req.Description
	}
	return c
}

func mapStringStatus(s string) payment.Status {
	switch s {
	case "paid", "settled", "success":
		return payment.StatusSettled
	case "expired":
		return payment.StatusExpired
	case "canceled", "cancelled":
		return payment.StatusCanceled
	case "failed", "rejected":
		return payment.StatusFailed
	case "refunded":
		return payment.StatusRefunded
	default:
		return payment.StatusPending
	}
}

func mapBoolStatus(paid bool) payment.Status {
	if paid {
		return payment.StatusSettled
	}
	return payment.StatusPending
}

func rawEvent(e *webhook.Event) map[string]any {
	b, _ := json.Marshal(e)
	raw := map[string]any{}
	_ = json.Unmarshal(b, &raw)
	return raw
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
