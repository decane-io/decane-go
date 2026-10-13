package decane

import (
	"context"
	"encoding/json"
	"fmt"
)

type PaymentLinkService struct{ http *httpClient }

type CreatePaymentLinkParams struct {
	Title          string  `json:"title"`
	Description    string  `json:"description,omitempty"`
	Amount         float64 `json:"amount,omitempty"`
	Currency       string  `json:"currency,omitempty"`
	CryptoCurrency string  `json:"crypto_currency,omitempty"`
	Network        string  `json:"network,omitempty"`
	RedirectURL    string  `json:"redirect_url,omitempty"`
}

func (s *PaymentLinkService) Create(ctx context.Context, params *CreatePaymentLinkParams) (map[string]interface{}, error) {
	data, err := s.http.post(ctx, "/merchants/me/payment-links", params)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal payment link: %w", err)
	}
	return result, nil
}

func (s *PaymentLinkService) List(ctx context.Context, params map[string]string) ([]map[string]interface{}, error) {
	data, err := s.http.get(ctx, "/merchants/me/payment-links", params)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal payment links: %w", err)
	}
	return result, nil
}

func (s *PaymentLinkService) Update(ctx context.Context, linkID string, params map[string]interface{}) (map[string]interface{}, error) {
	data, err := s.http.do(ctx, "PATCH", "/merchants/me/payment-links/"+linkID, params, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal payment link: %w", err)
	}
	return result, nil
}

func (s *PaymentLinkService) Delete(ctx context.Context, linkID string) error {
	_, err := s.http.do(ctx, "DELETE", "/merchants/me/payment-links/"+linkID, nil, nil)
	return err
}
