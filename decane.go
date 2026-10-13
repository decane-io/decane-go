// Package decane provides the official Go SDK for Decane —
// institutional liquidity rail for crypto settlement.
//
// Usage:
//
//	client := decane.New("gr_...", decane.WithWebhookSecret("gr_whsec_..."))
//
//	invoice, err := client.Invoices.Create(ctx, &decane.CreateInvoiceParams{
//	    Amount:   100,
//	    Currency: "USDT",
//	    Network:  "tron",
//	})
package decane

import "time"

const (
	DefaultBaseURL = "https://api.decane.io/api/v1"
	DefaultTimeout = 30 * time.Second
	Version        = "0.8.0"
)

// Client is the Decane API client.
type Client struct {
	Invoices      *InvoiceService
	Payouts       *PayoutService
	Wallets       *WalletService
	Swaps         *SwapService
	Webhooks      *WebhookService
	PaymentLinks  *PaymentLinkService
	Subscriptions *SubscriptionService

	http *httpClient
}

// Option configures the client.
type Option func(*Client)

// WithBaseURL sets a custom API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.http.baseURL = url }
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.http.timeout = d }
}

// WithAPISecret sets the API secret for authenticated endpoints.
func WithAPISecret(secret string) Option {
	return func(c *Client) { c.http.apiSecret = secret }
}

// WithWebhookSecret sets the webhook signing secret for signature verification.
func WithWebhookSecret(secret string) Option {
	return func(c *Client) { c.Webhooks = NewWebhookService(secret) }
}

// New creates a new Decane client.
func New(apiKey string, opts ...Option) *Client {
	h := newHTTPClient(apiKey, "", DefaultBaseURL, DefaultTimeout)

	c := &Client{
		http:     h,
		Webhooks: NewWebhookService(""),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Invoices = &InvoiceService{http: h}
	c.Payouts = &PayoutService{http: h}
	c.Wallets = &WalletService{http: h}
	c.Swaps = &SwapService{http: h}
	c.PaymentLinks = &PaymentLinkService{http: h}
	c.Subscriptions = &SubscriptionService{http: h}

	return c
}
