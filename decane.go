// Package decane is the official Go SDK for the Decane API.
//
//	client := decane.New("gr_...",
//	    decane.WithAPISecret("gr_secret_..."),
//	    decane.WithWebhookSecret("..."))
//
//	invoice, err := client.Invoices.Create(ctx, &decane.CreateInvoiceParams{
//	    Amount:         "100",
//	    Currency:       "USD",
//	    Network:        "tron",
//	    CryptoCurrency: "USDT",
//	})
package decane

import (
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://api.decane.io/api/v1"
	DefaultTimeout = 30 * time.Second
	Version        = "0.9.0"
)

func lower(s string) string { return strings.ToLower(s) }

// Client is the Decane API client.
type Client struct {
	Invoices *InvoiceService
	Payouts  *PayoutService
	Balances *BalanceService
	Wallets  *WalletService
	Swaps    *SwapService
	Events   *EventService
	Webhooks *WebhookService

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
	return func(c *Client) {
		c.http.timeout = d
		c.http.client.Timeout = d
	}
}

// WithAPISecret sets the API secret (gr_secret_...). Every request needs it.
func WithAPISecret(secret string) Option {
	return func(c *Client) { c.http.apiSecret = secret }
}

// WithWebhookSecret sets the secret used to verify webhook signatures.
func WithWebhookSecret(secret string) Option {
	return func(c *Client) { c.Webhooks = NewWebhookService(secret) }
}

// New creates a client. The API key and secret are both required by the API.
func New(apiKey string, opts ...Option) *Client {
	h := newHTTPClient(apiKey, "", DefaultBaseURL, DefaultTimeout)

	c := &Client{http: h, Webhooks: NewWebhookService("")}
	for _, opt := range opts {
		opt(c)
	}

	c.Invoices = &InvoiceService{http: h}
	c.Payouts = &PayoutService{http: h}
	c.Balances = &BalanceService{http: h}
	c.Wallets = &WalletService{http: h}
	c.Swaps = &SwapService{http: h}
	c.Events = &EventService{http: h}
	return c
}
