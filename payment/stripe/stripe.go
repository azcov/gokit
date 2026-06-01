package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/azcov/gokit/payment"
	gostripe "github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/client"
	"github.com/stripe/stripe-go/v81/webhook"
)

var _ payment.Gateway = (*Stripe)(nil)

type Stripe struct {
	c             *client.API
	webhookSecret string
}

type Config struct {
	SecretKey     string `config:"secret_key"`
	WebhookSecret string `config:"webhook_secret"`
}

func New(cfg Config) *Stripe {
	sc := &client.API{}
	sc.Init(cfg.SecretKey, nil)
	return &Stripe{c: sc, webhookSecret: cfg.WebhookSecret}
}

func (s *Stripe) CreateCharge(_ context.Context, req payment.ChargeRequest) (*payment.Charge, error) {
	params := &gostripe.PaymentIntentParams{
		Amount:   gostripe.Int64(req.Amount),
		Currency: gostripe.String(string(req.Currency)),
	}
	if req.Description != "" {
		params.Description = gostripe.String(req.Description)
	}
	if req.Customer.Email != "" {
		params.ReceiptEmail = gostripe.String(req.Customer.Email)
	}

	pi, err := s.c.PaymentIntents.New(params)
	if err != nil {
		return nil, err
	}
	return toCharge(pi), nil
}

func (s *Stripe) GetCharge(_ context.Context, id string) (*payment.Charge, error) {
	pi, err := s.c.PaymentIntents.Get(id, nil)
	if err != nil {
		return nil, err
	}
	return toCharge(pi), nil
}

func (s *Stripe) Refund(_ context.Context, req payment.RefundRequest) (*payment.RefundResult, error) {
	params := &gostripe.RefundParams{
		PaymentIntent: gostripe.String(req.ChargeID),
	}
	if req.Amount > 0 {
		params.Amount = gostripe.Int64(req.Amount)
	}
	r, err := s.c.Refunds.New(params)
	if err != nil {
		return nil, err
	}
	return &payment.RefundResult{
		ID:       r.ID,
		ChargeID: req.ChargeID,
		Amount:   r.Amount,
		Status:   payment.StatusPending,
	}, nil
}

func (s *Stripe) VerifyWebhook(_ context.Context, payload []byte, headers map[string]string) (*payment.WebhookEvent, error) {
	sig := headers["Stripe-Signature"]
	event, err := webhook.ConstructEvent(payload, sig, s.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("stripe: webhook signature invalid: %w", err)
	}

	var raw map[string]any
	_ = json.Unmarshal(event.Data.Raw, &raw)

	chargeID, _ := raw["id"].(string)
	return &payment.WebhookEvent{
		Type:     string(event.Type),
		ChargeID: chargeID,
		Status:   stripeStatus(raw),
		Raw:      map[string]any{"type": event.Type, "data": raw},
	}, nil
}

func toCharge(pi *gostripe.PaymentIntent) *payment.Charge {
	c := &payment.Charge{
		ID:         pi.ID,
		ExternalID: pi.ID,
		Amount:     pi.Amount,
		Currency:   payment.Currency(pi.Currency),
		Status:     mapStripeStatus(string(pi.Status)),
		CreatedAt:  time.Unix(pi.Created, 0),
		Meta:       map[string]any{"client_secret": pi.ClientSecret},
	}
	if pi.Description != "" {
		c.Description = pi.Description
	}
	return c
}

func mapStripeStatus(s string) payment.Status {
	switch s {
	case "succeeded":
		return payment.StatusSettled
	case "canceled":
		return payment.StatusCanceled
	case "requires_payment_method", "requires_confirmation", "requires_action", "processing":
		return payment.StatusPending
	default:
		return payment.StatusFailed
	}
}

func stripeStatus(raw map[string]any) payment.Status {
	s, _ := raw["status"].(string)
	return mapStripeStatus(s)
}
