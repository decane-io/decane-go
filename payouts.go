package decane

import (
	"context"
	"encoding/json"
	"fmt"
)

// PayoutService handles payout operations.
type PayoutService struct {
	http *httpClient
}

// Create requests a payout to an external wallet address. Pass
// WithIdempotencyKey to set the Idempotency-Key header; if omitted a random
// key is generated so automatic retries are safe.
func (s *PayoutService) Create(ctx context.Context, params *CreatePayoutParams, opts ...RequestOption) (*Payout, error) {
	data, err := s.http.postIdempotent(ctx, "/vendor/payout", params, opts)
	if err != nil {
		return nil, err
	}
	var p Payout
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("flow: unmarshal payout: %w", err)
	}
	return &p, nil
}

// CreateBatch creates multiple payouts in a single request. Idempotency works
// as in Create.
func (s *PayoutService) CreateBatch(ctx context.Context, payouts []*CreatePayoutParams, opts ...RequestOption) ([]*Payout, error) {
	body := map[string]interface{}{"payouts": payouts}
	data, err := s.http.postIdempotent(ctx, "/vendor/payout/batch", body, opts)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Payouts []*Payout `json:"payouts"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("flow: unmarshal payouts: %w", err)
	}
	return resp.Payouts, nil
}

// Get retrieves a payout by ID.
func (s *PayoutService) Get(ctx context.Context, id string) (*Payout, error) {
	data, err := s.http.get(ctx, "/vendor/query/payout/"+id, nil)
	if err != nil {
		return nil, err
	}
	var p Payout
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("flow: unmarshal payout: %w", err)
	}
	return &p, nil
}

// List returns a paginated list of payouts.
func (s *PayoutService) List(ctx context.Context, params *ListParams) (*PaginatedList[Payout], error) {
	p := listToMap(params)
	data, err := s.http.get(ctx, "/vendor/query/payouts", p)
	if err != nil {
		return nil, err
	}
	var result PaginatedList[Payout]
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal payouts: %w", err)
	}
	return &result, nil
}
