// Package payment defines a payment-gateway interface (CreateCharge, GetCharge,
// Refund, VerifyWebhook) with provider-agnostic money, method, and status types.
//
// Implementations: payment/stripe, payment/paypal, payment/midtrans,
// payment/xendit, payment/razorpay, payment/doku, payment/mayar.
package payment
