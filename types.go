package decane

import "encoding/json"

// List is a cursor-paginated response, newest first. Pass NextCursor as
// ListParams.StartingAfter to fetch the next page while HasMore is true.
type List[T any] struct {
	Object     string `json:"object"`
	Data       []T    `json:"data"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

// ListParams are the query parameters of a cursor list. Limit is 1 to 100
// (the server default is 20). Status filters by status.
type ListParams struct {
	Limit         int
	StartingAfter string
	Status        string
}

// Invoice is a payment invoice. Money fields are decimal strings.
type Invoice struct {
	ID                    string                 `json:"id"`
	MerchantID            string                 `json:"merchant_id,omitempty"`
	Amount                Amount                 `json:"fiat_amount"`
	Currency              string                 `json:"fiat_currency"`
	Status                string                 `json:"status"`
	Network               string                 `json:"network,omitempty"`
	CryptoCurrency        string                 `json:"crypto_currency,omitempty"`
	CryptoAmount          Amount                 `json:"crypto_amount,omitempty"`
	DepositAddress        string                 `json:"deposit_address,omitempty"`
	CustomerEmail         string                 `json:"customer_email,omitempty"`
	CustomerName          string                 `json:"customer_name,omitempty"`
	MerchantOrderID       string                 `json:"merchant_order_id,omitempty"`
	ReturnURL             string                 `json:"return_url,omitempty"`
	ExtraData             map[string]interface{} `json:"extra_data,omitempty"`
	PaymentURL            string                 `json:"payment_url,omitempty"`
	TxHash                string                 `json:"tx_hash,omitempty"`
	ReceivedAmount        Amount                 `json:"received_amount,omitempty"`
	Confirmations         int                    `json:"confirmations,omitempty"`
	RequiredConfirmations int                    `json:"required_confirmations,omitempty"`
	ExpiresAt             string                 `json:"expires_at,omitempty"`
	CreatedAt             string                 `json:"created_at,omitempty"`
	UpdatedAt             string                 `json:"updated_at,omitempty"`
}

// CreateInvoiceParams are the parameters for creating an invoice. Amount is
// the fiat price as a decimal string. Omit Network and CryptoCurrency for a
// flex invoice the customer finishes on the hosted page.
type CreateInvoiceParams struct {
	Amount          string                 `json:"amount"`
	Currency        string                 `json:"currency,omitempty"`
	Network         string                 `json:"network,omitempty"`
	CryptoCurrency  string                 `json:"crypto_currency,omitempty"`
	MerchantOrderID string                 `json:"merchant_order_id,omitempty"`
	CustomerEmail   string                 `json:"customer_email,omitempty"`
	CustomerName    string                 `json:"customer_name,omitempty"`
	CallbackURL     string                 `json:"callback_url,omitempty"`
	ReturnURL       string                 `json:"return_url,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Payout is a payout to an external address.
type Payout struct {
	ID               string `json:"id"`
	Network          string `json:"network"`
	Currency         string `json:"currency"`
	Status           string `json:"status"`
	GrossAmount      Amount `json:"gross_amount"`
	NetAmount        Amount `json:"net_amount"`
	FeeAmount        Amount `json:"fee_amount"`
	RecipientAddress string `json:"recipient_address"`
	ReferenceID      string `json:"reference_id,omitempty"`
	BatchID          string `json:"batch_id,omitempty"`
	TxHash           string `json:"tx_hash,omitempty"`
	ErrorMessage     string `json:"error_message,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	CompletedAt      string `json:"completed_at,omitempty"`
}

// CreatePayoutParams are the parameters for one payout. Amount is a decimal
// string; the recipient gets the amount minus fees. SourceWalletID picks which
// of your named wallets pays (an ID from Wallets.List); leave it empty to use
// the key's own wallet.
type CreatePayoutParams struct {
	RecipientAddress string `json:"recipient_address"`
	Network          string `json:"network"`
	Currency         string `json:"currency"`
	Amount           string `json:"amount"`
	ReferenceID      string `json:"reference_id,omitempty"`
	SourceWalletID   string `json:"source_wallet_id,omitempty"`
}

// CreateBatchParams are the parameters for a batch of up to 100 payouts.
type CreateBatchParams struct {
	Payouts     []*CreatePayoutParams `json:"payouts"`
	ReferenceID string                `json:"reference_id,omitempty"`
}

// PayoutBatch is the result of CreateBatch.
type PayoutBatch struct {
	ID             string    `json:"id"`
	ReferenceID    string    `json:"reference_id,omitempty"`
	Status         string    `json:"status"`
	TotalPayouts   int       `json:"total_payouts"`
	CompletedCount int       `json:"completed_count"`
	FailedCount    int       `json:"failed_count"`
	TotalGross     Amount    `json:"total_gross"`
	TotalFees      Amount    `json:"total_fees"`
	TotalNet       Amount    `json:"total_net"`
	CreatedAt      string    `json:"created_at,omitempty"`
	CompletedAt    string    `json:"completed_at,omitempty"`
	Payouts        []*Payout `json:"payouts"`
}

// Balance is the response of GET /balances. Pooled is network then coin then
// amount. Account is coin then amount. Pending holds amounts on hold.
type Balance struct {
	Object   string                       `json:"object"`
	Livemode *bool                        `json:"livemode,omitempty"`
	Pooled   map[string]map[string]Amount `json:"pooled"`
	Account  map[string]Amount            `json:"account"`
	Pending  map[string]Amount            `json:"pending"`
}

// Wallet is one named wallet visible to the key, with its own balance.
type Wallet struct {
	ID            string            `json:"wallet_id"`
	Name          string            `json:"name"`
	Network       string            `json:"network"`
	Address       string            `json:"address"`
	NativeSymbol  string            `json:"native_symbol"`
	NativeBalance Amount            `json:"native_balance"`
	TokenBalances map[string]Amount `json:"token_balances"`
	AccruedFees   map[string]Amount `json:"accrued_fees,omitempty"`
}

// WalletList is the response of GET /vendor/query/balances.
type WalletList struct {
	APIKeyID   string   `json:"api_key_id"`
	APIKeyName string   `json:"api_key_name"`
	Wallets    []Wallet `json:"wallets"`
	FetchedAt  string   `json:"fetched_at"`
}

// SwapQuoteParams are the parameters for a swap quote. FromAmount is a decimal
// string. Slippage is 0.1 to 5.0 percent; zero uses the server default (0.5).
type SwapQuoteParams struct {
	FromNetwork  string  `json:"from_network"`
	FromCurrency string  `json:"from_currency"`
	FromAmount   string  `json:"from_amount"`
	ToNetwork    string  `json:"to_network"`
	ToCurrency   string  `json:"to_currency"`
	Slippage     float64 `json:"slippage,omitempty"`
}

// SwapParams are the parameters for executing a swap. QuoteID is optional.
// SourceWalletID and DestinationWalletID default to the key's own wallets.
// DestinationAddress sends the output to an external address instead and
// cannot be combined with DestinationWalletID.
type SwapParams struct {
	QuoteID             string  `json:"quote_id,omitempty"`
	FromNetwork         string  `json:"from_network"`
	FromCurrency        string  `json:"from_currency"`
	FromAmount          string  `json:"from_amount"`
	ToNetwork           string  `json:"to_network"`
	ToCurrency          string  `json:"to_currency"`
	SourceWalletID      string  `json:"source_wallet_id,omitempty"`
	DestinationWalletID string  `json:"destination_wallet_id,omitempty"`
	DestinationAddress  string  `json:"destination_address,omitempty"`
	Slippage            float64 `json:"slippage,omitempty"`
}

// SwapQuote is a price quote, valid for about 60 seconds.
type SwapQuote struct {
	QuoteID         string  `json:"quote_id"`
	Provider        string  `json:"provider"`
	FromCurrency    string  `json:"from_currency"`
	FromNetwork     string  `json:"from_network"`
	FromAmount      Amount  `json:"from_amount"`
	ToCurrency      string  `json:"to_currency"`
	ToNetwork       string  `json:"to_network"`
	ToAmount        Amount  `json:"to_amount"`
	ExchangeRate    Amount  `json:"exchange_rate"`
	PriceImpact     Amount  `json:"price_impact,omitempty"`
	GasEstimate     Amount  `json:"gas_estimate,omitempty"`
	Slippage        float64 `json:"slippage"`
	IsCrossChain    bool    `json:"is_cross_chain"`
	PlatformFee     Amount  `json:"platform_fee"`
	PlatformFeeRate Amount  `json:"platform_fee_rate"`
	ExpiresAt       string  `json:"expires_at"`
}

// Swap is a swap. Poll Swaps.Get until Status is "completed" or "failed"
// (compare case-insensitively, or use Done).
type Swap struct {
	ID                  string `json:"swap_id"`
	Status              string `json:"status"`
	FromCurrency        string `json:"from_currency"`
	FromNetwork         string `json:"from_network"`
	FromAmount          Amount `json:"from_amount"`
	ToCurrency          string `json:"to_currency"`
	ToNetwork           string `json:"to_network"`
	ToAmount            Amount `json:"to_amount,omitempty"`
	ExchangeRate        Amount `json:"exchange_rate,omitempty"`
	PlatformFee         Amount `json:"platform_fee"`
	SourceWalletID      string `json:"source_wallet_id"`
	DestinationWalletID string `json:"destination_wallet_id,omitempty"`
	FromTxHash          string `json:"from_tx_hash,omitempty"`
	ToTxHash            string `json:"to_tx_hash,omitempty"`
	Provider            string `json:"provider,omitempty"`
	ErrorMessage        string `json:"error_message,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

// Done reports whether the swap reached a final state.
func (s *Swap) Done() bool {
	switch lower(s.Status) {
	case "completed", "failed", "refunded", "cancelled":
		return true
	}
	return false
}

// SwapListParams filter GET /vendor/swaps, which pages by offset (limit 1 to 200).
type SwapListParams struct {
	Status       string
	FromCurrency string
	ToCurrency   string
	Limit        int
	Offset       int
}

// SwapList is the response of GET /vendor/swaps.
type SwapList struct {
	Swaps []Swap `json:"swaps"`
	Total int    `json:"total"`
}

// Event is one entry of the webhook log: what was sent, and what came back.
type Event struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"event"`
	Status         string                 `json:"status"`
	URL            string                 `json:"url"`
	Attempts       int                    `json:"attempts"`
	ResponseStatus *int                   `json:"response_status"`
	Error          string                 `json:"error,omitempty"`
	InvoiceID      string                 `json:"invoice_id,omitempty"`
	PayoutID       string                 `json:"payout_id,omitempty"`
	Livemode       bool                   `json:"livemode"`
	CreatedAt      string                 `json:"created_at,omitempty"`
	LastAttemptAt  string                 `json:"last_attempt_at,omitempty"`
	NextAttemptAt  string                 `json:"next_attempt_at,omitempty"`
	DeliveredAt    string                 `json:"delivered_at,omitempty"`
	Payload        map[string]interface{} `json:"payload,omitempty"`
}

// WebhookEvent is a verified webhook delivery. Type is the event name, for
// example "invoice.completed" or "payout.failed". Data carries invoice_id or
// payout_id and the other fields of the object.
type WebhookEvent struct {
	ID        string                 `json:"id,omitempty"`
	Type      string                 `json:"event"`
	Created   int64                  `json:"created,omitempty"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt string                 `json:"timestamp,omitempty"`
}

// UnmarshalJSON accepts "timestamp" as a string or a number, and falls back to
// "type" when "event" is absent.
func (e *WebhookEvent) UnmarshalJSON(b []byte) error {
	var raw struct {
		ID        string                 `json:"id"`
		Event     string                 `json:"event"`
		Type      string                 `json:"type"`
		Created   int64                  `json:"created"`
		Data      map[string]interface{} `json:"data"`
		Timestamp json.RawMessage        `json:"timestamp"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	e.ID, e.Created, e.Data = raw.ID, raw.Created, raw.Data
	e.Type = raw.Event
	if e.Type == "" {
		e.Type = raw.Type
	}
	e.CreatedAt = ""
	if len(raw.Timestamp) > 0 && string(raw.Timestamp) != "null" {
		var s string
		if err := json.Unmarshal(raw.Timestamp, &s); err == nil {
			e.CreatedAt = s
		} else {
			e.CreatedAt = string(raw.Timestamp)
		}
	}
	return nil
}
