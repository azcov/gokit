package doku

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/azcov/gokit/payment"
)

var _ payment.Gateway = (*DOKU)(nil)

const (
	sandboxBase    = "https://api-sandbox.doku.com"
	productionBase = "https://api.doku.com"
)

type DOKU struct {
	clientID  string
	secretKey string
	baseURL   string
	http      *http.Client
}

type Config struct {
	ClientID     string `config:"client_id"`
	SecretKey    string `config:"secret_key"`
	IsProduction bool   `config:"is_production"`
}

func New(cfg Config) *DOKU {
	base := sandboxBase
	if cfg.IsProduction {
		base = productionBase
	}
	return &DOKU{
		clientID:  cfg.ClientID,
		secretKey: cfg.SecretKey,
		baseURL:   base,
		http:      &http.Client{Timeout: 30 * time.Second},
	}
}

type createInvoiceReq struct {
	Order    orderData    `json:"order"`
	Customer customerData `json:"customer"`
	Payment  paymentData  `json:"payment"`
}

type orderData struct {
	InvoiceNumber string     `json:"invoice_number"`
	LineItems     []lineItem `json:"line_items"`
}

type lineItem struct {
	Name  string `json:"name"`
	Price int64  `json:"price"`
	Qty   int    `json:"quantity"`
}

type customerData struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone,omitempty"`
}

type paymentData struct {
	PaymentDueDate int `json:"payment_due_date"` // minutes
}

func (d *DOKU) CreateCharge(ctx context.Context, req payment.ChargeRequest) (*payment.Charge, error) {
	invoiceNum := "INV-" + newID()
	desc := req.Description
	if desc == "" {
		desc = "Payment"
	}

	body := createInvoiceReq{
		Order: orderData{
			InvoiceNumber: invoiceNum,
			LineItems:     []lineItem{{Name: desc, Price: req.Amount, Qty: 1}},
		},
		Customer: customerData{
			Name:  req.Customer.Name,
			Email: req.Customer.Email,
			Phone: req.Customer.Phone,
		},
		Payment: paymentData{PaymentDueDate: 60},
	}

	var resp map[string]any
	if err := d.post(ctx, "/checkout/v1/payment", body, &resp); err != nil {
		return nil, fmt.Errorf("doku: create charge: %w", err)
	}

	paymentURL, _ := resp["payment_url"].(string)
	return &payment.Charge{
		ID:          invoiceNum,
		ExternalID:  invoiceNum,
		Amount:      req.Amount,
		Currency:    payment.IDR,
		Status:      payment.StatusPending,
		Description: desc,
		PaymentURL:  paymentURL,
		CreatedAt:   time.Now(),
		Meta:        resp,
	}, nil
}

func (d *DOKU) GetCharge(ctx context.Context, id string) (*payment.Charge, error) {
	var resp map[string]any
	if err := d.get(ctx, "/orders/v1/status/"+id, &resp); err != nil {
		return nil, fmt.Errorf("doku: get charge: %w", err)
	}

	statusStr, _ := resp["status"].(string)
	return &payment.Charge{
		ID:         id,
		ExternalID: id,
		Currency:   payment.IDR,
		Status:     mapStatus(statusStr),
		CreatedAt:  time.Now(),
		Meta:       resp,
	}, nil
}

func (d *DOKU) Refund(ctx context.Context, req payment.RefundRequest) (*payment.RefundResult, error) {
	body := map[string]any{
		"original_invoice_number": req.ChargeID,
		"amount":                  req.Amount,
		"comment":                 req.Reason,
	}
	var resp map[string]any
	if err := d.post(ctx, "/refunds/v1", body, &resp); err != nil {
		return nil, fmt.Errorf("doku: refund: %w", err)
	}
	refID, _ := resp["refund_id"].(string)
	return &payment.RefundResult{
		ID:       refID,
		ChargeID: req.ChargeID,
		Amount:   req.Amount,
		Status:   payment.StatusPending,
		Meta:     resp,
	}, nil
}

// VerifyWebhook validates DOKU notification signature.
// DOKU signs notifications using SHA-256 of specific fields.
func (d *DOKU) VerifyWebhook(_ context.Context, payload []byte, headers map[string]string) (*payment.WebhookEvent, error) {
	sig := headers["Signature"]

	var notif map[string]any
	if err := json.Unmarshal(payload, &notif); err != nil {
		return nil, fmt.Errorf("doku: parse webhook: %w", err)
	}

	invoiceNum, _ := notif["invoice_number"].(string)
	// DOKU signature: SHA256(client_id + invoice_number + secret_key)
	raw := d.clientID + invoiceNum + d.secretKey
	hash := sha256.Sum256([]byte(raw))
	expected := strings.ToUpper(hex.EncodeToString(hash[:]))
	if sig != "" && expected != strings.ToUpper(sig) {
		return nil, fmt.Errorf("doku: webhook signature mismatch")
	}

	statusStr, _ := notif["status"].(string)
	return &payment.WebhookEvent{
		Type:     statusStr,
		ChargeID: invoiceNum,
		Status:   mapStatus(statusStr),
		Raw:      notif,
	}, nil
}

func (d *DOKU) post(ctx context.Context, path string, body, dest any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", d.clientID)
	req.Header.Set("Request-Id", newID())
	req.Header.Set("Request-Timestamp", time.Now().UTC().Format(time.RFC3339))
	d.signRequest(req, b)
	return d.do(req, dest)
}

func (d *DOKU) get(ctx context.Context, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Client-Id", d.clientID)
	req.Header.Set("Request-Id", newID())
	d.signRequest(req, nil)
	return d.do(req, dest)
}

func (d *DOKU) signRequest(req *http.Request, body []byte) {
	// DOKU signature: SHA256(Client-Id + Request-Id + Request-Timestamp + body)
	parts := d.clientID + req.Header.Get("Request-Id") + req.Header.Get("Request-Timestamp")
	if body != nil {
		parts += string(body)
	}
	hash := sha256.Sum256([]byte(parts + d.secretKey))
	req.Header.Set("Signature", strings.ToUpper(hex.EncodeToString(hash[:])))
}

func (d *DOKU) do(req *http.Request, dest any) error {
	resp, err := d.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var e map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return fmt.Errorf("doku: HTTP %d: %v", resp.StatusCode, e)
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

func mapStatus(s string) payment.Status {
	switch strings.ToUpper(s) {
	case "SUCCESS", "PAID", "SETTLEMENT":
		return payment.StatusSettled
	case "EXPIRED":
		return payment.StatusExpired
	case "FAILED", "DENIED":
		return payment.StatusFailed
	case "CANCELED":
		return payment.StatusCanceled
	default:
		return payment.StatusPending
	}
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
