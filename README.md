# Decane Go SDK

> **Renamed from granthe-go.** The old module was `github.com/Granthe-io/granthe-go`. The API and webhook formats are unchanged, and the server still sends `Granthe-Signature` for older integrations. This SDK reads `Decane-Signature` first and falls back to `Granthe-Signature`.

Official Go SDK for [Decane](https://decane.io) — institutional liquidity rail for crypto settlement.

## Installation

```bash
go get github.com/Decane-io/decane-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    decane "github.com/Decane-io/decane-go"
)

func main() {
    client := decane.New("gr_...")

    // Create an invoice
    invoice, err := client.Invoices.Create(context.Background(), &decane.CreateInvoiceParams{
        Amount:   100,
        CryptoCurrency: "USDT",
        Network:  "tron",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Invoice:", invoice.ID, invoice.PaymentURL)
}
```

## Resources

### Invoices

```go
// Create
inv, _ := client.Invoices.Create(ctx, &decane.CreateInvoiceParams{
    Amount: 100, CryptoCurrency: "USDT", Network: "tron",
    Description: "Order #123",
    CallbackURL: "https://example.com/webhook",
}, decane.WithIdempotencyKey("order-123")) // optional, safe to retry

// Get
inv, _ := client.Invoices.Get(ctx, "inv_123")

// List
list, _ := client.Invoices.List(ctx, &decane.ListParams{Page: 1, Limit: 20})

// Cancel
inv, _ := client.Invoices.Cancel(ctx, "inv_123")
```

### Payouts

```go
// Create
payout, _ := client.Payouts.Create(ctx, &decane.CreatePayoutParams{
    Amount: 50, Currency: "USDT", Network: "tron",
    RecipientAddress: "TXyz...",
}, decane.WithIdempotencyKey("payout-123")) // optional, safe to retry

// Batch create
payouts, _ := client.Payouts.CreateBatch(ctx, []*decane.CreatePayoutParams{
    {Amount: 50, Currency: "USDT", Network: "tron", RecipientAddress: "TXyz..."},
    {Amount: 25, Currency: "USDT", Network: "tron", RecipientAddress: "TAbc..."},
})

// Get & List
payout, _ := client.Payouts.Get(ctx, "po_123")
list, _ := client.Payouts.List(ctx, &decane.ListParams{Status: "completed"})
```

### Wallets

```go
// Generate
wallet, _ := client.Wallets.Generate(ctx, &decane.GenerateWalletParams{
    Network: "tron", Currency: "USDT", Label: "deposits",
})

// Get, List, Balance
wallet, _ := client.Wallets.Get(ctx, "w_123")
list, _ := client.Wallets.List(ctx, &decane.ListParams{Network: "tron"})
wallet, _ := client.Wallets.Balance(ctx, "w_123")
```

### Swaps

```go
// Get quote
quote, _ := client.Swaps.Quote(ctx, &decane.SwapParams{
    FromCurrency: "USDT", ToCurrency: "BTC", Amount: 1000,
})

// Execute swap
swap, _ := client.Swaps.Create(ctx, &decane.SwapParams{
    FromCurrency: "USDT", ToCurrency: "BTC", Amount: 1000,
})

// Get & List
swap, _ := client.Swaps.Get(ctx, "sw_123")
list, _ := client.Swaps.List(ctx, nil)
```

### Webhooks

```go
// With webhook secret
client := decane.New("gr_...", decane.WithWebhookSecret("gr_whsec_..."))

// Verify in HTTP handler (recommended: Decane-Signature, 5 minute tolerance)
func webhookHandler(w http.ResponseWriter, r *http.Request) {
    event, err := client.Webhooks.ConstructEventFromRequest(r, 0)
    if err != nil {
        http.Error(w, "Invalid signature", 401)
        return
    }
    switch event.Type {
    case "invoice.paid":
        // Handle payment
    case "payout.completed":
        // Handle payout
    }
    w.WriteHeader(200)
}

// Manual verification with the Decane-Signature header value (t=...,v1=...)
event, err := client.Webhooks.ConstructEvent(payload, header, 5*time.Minute)

// Just check it
if err := client.Webhooks.VerifySignature(payload, header, 0); err == nil {
    // Process
}
```

`Verify`, `IsValid` and `VerifyRequest` check the old `X-Decane-Signature` header. They still work but are deprecated, because they have no timestamp check. `X-Decane-Delivery` is the same on every retry, so use it to drop duplicates.

## Error Handling

```go
inv, err := client.Invoices.Get(ctx, "inv_123")
if err != nil {
    switch e := err.(type) {
    case *decane.AuthenticationError:
        // Invalid API key
    case *decane.NotFoundError:
        // Resource not found
    case *decane.ValidationError:
        // Invalid parameters — check e.Errors
    case *decane.RateLimitError:
        // Too many requests — check e.RetryAfter
    case *decane.ServerError:
        // Server error — retry later
    default:
        // Other error
    }
}
```

## Configuration

```go
client := decane.New("gr_...",
    decane.WithBaseURL("https://custom-api.example.com"),
    decane.WithTimeout(60 * time.Second),
    decane.WithWebhookSecret("gr_whsec_..."),
)
```

## License

MIT
