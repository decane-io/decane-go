package decane

import "encoding/json"

// Invoice represents a payment invoice.
type Invoice struct {
	ID                    string                 `json:"id"`
	MerchantID            string                 `json:"merchant_id"`
	Amount                float64                `json:"fiat_amount,string"`
	Currency              string                 `json:"fiat_currency"`
	Status                string                 `json:"status"`
	Network               string                 `json:"network,omitempty"`
	CryptoCurrency        string                 `json:"crypto_currency,omitempty"`
	CryptoAmount          float64                `json:"crypto_amount,string,omitempty"`
	WalletAddress         string                 `json:"deposit_address,omitempty"`
	CustomerEmail         string                 `json:"customer_email,omitempty"`
	CustomerName          string                 `json:"customer_name,omitempty"`
	MerchantOrderID       string                 `json:"merchant_order_id,omitempty"`
	CallbackURL           string                 `json:"callback_url,omitempty"`
	ReturnURL             string                 `json:"return_url,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	PaymentURL            string                 `json:"payment_url,omitempty"`
	TxHash                string                 `json:"tx_hash,omitempty"`
	PaidAmount            float64                `json:"received_amount,string,omitempty"`
	Confirmations         int                    `json:"confirmations,omitempty"`
	RequiredConfirmations int                    `json:"required_confirmations,omitempty"`
	ExpiresAt             string                 `json:"expires_at,omitempty"`
	CreatedAt             string                 `json:"created_at,omitempty"`
	UpdatedAt             string                 `json:"updated_at,omitempty"`
}

// CreateInvoiceParams are parameters for creating an invoice.
// Omit Network and CryptoCurrency for a flex invoice (customer picks on payment page).
type CreateInvoiceParams struct {
	Amount          float64                `json:"amount"`
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

// Payout represents a vendor payout.
type Payout struct {
	ID               string  `json:"id"`
	Network          string  `json:"network"`
	Currency         string  `json:"currency"`
	Status           string  `json:"status"`
	GrossAmount      float64 `json:"gross_amount,string"`
	NetAmount        float64 `json:"net_amount,string"`
	FeeAmount        float64 `json:"fee_amount,string"`
	NMCAmount        float64 `json:"nmc_amount,string,omitempty"`
	RecipientAddress string  `json:"recipient_address"`
	ReferenceID      string  `json:"reference_id,omitempty"`
	BatchID          string  `json:"batch_id,omitempty"`
	TxHash           string  `json:"tx_hash,omitempty"`
	ErrorMessage     string  `json:"error_message,omitempty"`
	CreatedAt        string  `json:"created_at,omitempty"`
	CompletedAt      string  `json:"completed_at,omitempty"`
}

// CreatePayoutParams are parameters for creating a payout.
type CreatePayoutParams struct {
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	Network          string  `json:"network"`
	RecipientAddress string  `json:"recipient_address"`
	ReferenceID      string  `json:"reference_id,omitempty"`
}

// Wallet represents a deposit wallet.
type Wallet struct {
	ID        string  `json:"id"`
	Address   string  `json:"address"`
	Network   string  `json:"network"`
	Currency  string  `json:"currency"`
	Balance   float64 `json:"balance,string"`
	Label     string  `json:"label,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
}

// GenerateWalletParams are parameters for generating a wallet.
type GenerateWalletParams struct {
	Network  string `json:"network"`
	Currency string `json:"currency"`
	Label    string `json:"label,omitempty"`
}

// SwapQuote represents a swap price quote.
type SwapQuote struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	FromAmount   float64 `json:"from_amount,string"`
	ToAmount     float64 `json:"to_amount,string"`
	Rate         float64 `json:"rate,string"`
	ExpiresAt    string  `json:"expires_at,omitempty"`
}

// Swap represents a swap.
type Swap struct {
	ID           string  `json:"id"`
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	FromAmount   float64 `json:"from_amount,string"`
	ToAmount     float64 `json:"to_amount,string"`
	Rate         float64 `json:"rate,string"`
	Status       string  `json:"status"`
	TxHash       string  `json:"tx_hash,omitempty"`
	CreatedAt    string  `json:"created_at,omitempty"`
}

// SwapParams are parameters for creating a swap.
type SwapParams struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Amount       float64 `json:"amount"`
}

// WebhookEvent represents a verified webhook event.
type WebhookEvent struct {
	ID        string                 `json:"id,omitempty"`
	Type      string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt string                 `json:"timestamp,omitempty"`
}

// UnmarshalJSON accepts "timestamp" as either a string or a number (unix
// seconds), since the server sends both depending on the event.
func (e *WebhookEvent) UnmarshalJSON(b []byte) error {
	var raw struct {
		ID        string                 `json:"id"`
		Type      string                 `json:"event"`
		Data      map[string]interface{} `json:"data"`
		Timestamp json.RawMessage        `json:"timestamp"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	e.ID, e.Type, e.Data = raw.ID, raw.Type, raw.Data
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

// ListParams are common parameters for list endpoints.
type ListParams struct {
	Page     int    `json:"page,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Status   string `json:"status,omitempty"`
	Currency string `json:"currency,omitempty"`
	Network  string `json:"network,omitempty"`
}

// PaginatedList is a paginated response.
type PaginatedList[T any] struct {
	Items   []T  `json:"items"`
	Total   int  `json:"total"`
	Page    int  `json:"page"`
	PerPage int  `json:"per_page"`
	HasMore bool `json:"has_more"`
}
