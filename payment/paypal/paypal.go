package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/azcov/gokit/payment"
	pp "github.com/plutov/paypal/v4"
)

var _ payment.Gateway = (*PayPal)(nil)

type PayPal struct {
	client    *pp.Client
	webhookID string
}

type Config struct {
	ClientID     string `env:"CLIENT_ID" json:"client_id" yaml:"client_id"`
	ClientSecret string `env:"CLIENT_SECRET" json:"client_secret" yaml:"client_secret"`
	WebhookID    string `env:"WEBHOOK_ID" json:"webhook_id" yaml:"webhook_id"`
	IsProduction bool   `env:"IS_PRODUCTION" json:"is_production" yaml:"is_production"`
}

func New(ctx context.Context, cfg Config) (*PayPal, error) {
	base := pp.APIBaseSandBox
	if cfg.IsProduction {
		base = pp.APIBaseLive
	}
	c, err := pp.NewClient(cfg.ClientID, cfg.ClientSecret, base)
	if err != nil {
		return nil, err
	}
	if _, err = c.GetAccessToken(ctx); err != nil {
		return nil, err
	}
	return &PayPal{client: c, webhookID: cfg.WebhookID}, nil
}

func (p *PayPal) CreateCharge(ctx context.Context, req payment.ChargeRequest) (*payment.Charge, error) {
	value := amountString(req.Amount, req.Currency)
	order, err := p.client.CreateOrder(ctx,
		pp.OrderIntentCapture,
		[]pp.PurchaseUnitRequest{{
			Amount:      &pp.PurchaseUnitAmount{Currency: string(req.Currency), Value: value},
			Description: req.Description,
		}},
		nil, // paymentSource — set client-side via JS SDK
		&pp.ApplicationContext{ReturnURL: req.SuccessURL, CancelURL: req.CancelURL},
	)
	if err != nil {
		return nil, fmt.Errorf("paypal: create order: %w", err)
	}

	var paymentURL string
	for _, link := range order.Links {
		if link.Rel == "approve" {
			paymentURL = link.Href
		}
	}

	return &payment.Charge{
		ID:         order.ID,
		ExternalID: order.ID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Status:     paypalStatus(order.Status),
		PaymentURL: paymentURL,
		CreatedAt:  time.Now(),
	}, nil
}

func (p *PayPal) GetCharge(ctx context.Context, id string) (*payment.Charge, error) {
	order, err := p.client.GetOrder(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("paypal: get order: %w", err)
	}
	return &payment.Charge{
		ID:         order.ID,
		ExternalID: order.ID,
		Status:     paypalStatus(order.Status),
		CreatedAt:  time.Now(),
	}, nil
}

func (p *PayPal) Refund(ctx context.Context, req payment.RefundRequest) (*payment.RefundResult, error) {
	value := amountString(req.Amount, payment.USD)
	ref, err := p.client.RefundCapture(ctx, req.ChargeID, pp.RefundCaptureRequest{
		Amount:      &pp.Money{Value: value},
		NoteToPayer: req.Reason,
	})
	if err != nil {
		return nil, fmt.Errorf("paypal: refund: %w", err)
	}
	return &payment.RefundResult{
		ID:       ref.ID,
		ChargeID: req.ChargeID,
		Amount:   req.Amount,
		Status:   payment.StatusPending,
	}, nil
}

// VerifyWebhook validates the PayPal webhook signature via PayPal's verification API.
func (p *PayPal) VerifyWebhook(ctx context.Context, payload []byte, headers map[string]string) (*payment.WebhookEvent, error) {
	// Reconstruct an http.Request so the SDK can read headers and body.
	httpReq := &http.Request{
		Header: make(http.Header),
		Body:   io.NopCloser(bytes.NewReader(payload)),
	}
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := p.client.VerifyWebhookSignature(ctx, httpReq, p.webhookID)
	if err != nil {
		return nil, fmt.Errorf("paypal: webhook verify: %w", err)
	}
	if resp.VerificationStatus != "SUCCESS" {
		return nil, fmt.Errorf("paypal: webhook verification failed: %s", resp.VerificationStatus)
	}

	var raw map[string]any
	_ = json.Unmarshal(payload, &raw)
	evType, _ := raw["event_type"].(string)
	resource, _ := raw["resource"].(map[string]any)
	chargeID, _ := resource["id"].(string)

	return &payment.WebhookEvent{
		Type:     evType,
		ChargeID: chargeID,
		Status:   paypalResourceStatus(resource),
		Raw:      raw,
	}, nil
}

func paypalStatus(s string) payment.Status {
	switch s {
	case "COMPLETED":
		return payment.StatusSettled
	case "VOIDED", "CANCELLED":
		return payment.StatusCanceled
	default:
		return payment.StatusPending
	}
}

func paypalResourceStatus(resource map[string]any) payment.Status {
	s, _ := resource["status"].(string)
	return paypalStatus(s)
}

func amountString(amount int64, currency payment.Currency) string {
	switch currency {
	case payment.IDR, payment.JPY, payment.VND:
		return fmt.Sprintf("%d", amount)
	default:
		return fmt.Sprintf("%.2f", float64(amount)/100)
	}
}
