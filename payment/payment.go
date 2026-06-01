package payment

import (
	"context"
	"time"
)

// Currency codes (ISO 4217).
type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	GBP Currency = "GBP"
	AUD Currency = "AUD"
	SGD Currency = "SGD"
	MYR Currency = "MYR"
	PHP Currency = "PHP"
	THB Currency = "THB"
	IDR Currency = "IDR"
	INR Currency = "INR"
	JPY Currency = "JPY"
	VND Currency = "VND"
)

type Method string

const (
	MethodCard           Method = "card"
	MethodBankTransfer   Method = "bank_transfer"
	MethodVirtualAccount Method = "virtual_account"
	MethodQRIS           Method = "qris"
	MethodEWallet        Method = "ewallet"
	MethodPayLater       Method = "pay_later"
	MethodCash           Method = "cash"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusCapture  Status = "capture"
	StatusSettled  Status = "settled"
	StatusFailed   Status = "failed"
	StatusExpired  Status = "expired"
	StatusCanceled Status = "canceled"
	StatusRefunded Status = "refunded"
)

type Customer struct {
	ID    string
	Name  string
	Email string
	Phone string
}

// ChargeRequest is provider-agnostic input for initiating a payment.
// Amount is always in the smallest currency unit (cents, rupiah, paise, etc.).
type ChargeRequest struct {
	Amount      int64
	Currency    Currency
	Method      Method
	Description string
	Customer    Customer
	// Token is a client-side payment method token (e.g. Stripe card token).
	Token string
	// Redirect URLs for hosted payment pages (PayPal, Midtrans Snap, Xendit Invoice).
	SuccessURL string
	FailureURL string
	CancelURL  string
	ExpiresAt  *time.Time
	Meta       map[string]any
}

// Charge is the result of a CreateCharge call.
type Charge struct {
	ID                   string
	ExternalID           string
	Amount               int64
	Currency             Currency
	Method               Method
	Status               Status
	Description          string
	Customer             Customer
	PaymentURL           string
	VirtualAccountNumber string
	VirtualAccountBank   string
	QRCode               string
	ExpiresAt            *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Meta                 map[string]any
}

type RefundRequest struct {
	ChargeID string
	Amount   int64
	Reason   string
}

type RefundResult struct {
	ID       string
	ChargeID string
	Amount   int64
	Status   Status
	Meta     map[string]any
}

type WebhookEvent struct {
	Type     string
	ChargeID string
	Status   Status
	Raw      map[string]any
}

type Gateway interface {
	CreateCharge(ctx context.Context, req ChargeRequest) (*Charge, error)
	GetCharge(ctx context.Context, id string) (*Charge, error)
	Refund(ctx context.Context, req RefundRequest) (*RefundResult, error)
	VerifyWebhook(ctx context.Context, payload []byte, headers map[string]string) (*WebhookEvent, error)
}
