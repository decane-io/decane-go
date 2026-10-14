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
		return nil, fmt.Errorf("decane: unmarshal payout: %w", err)
	}
	return &p, nil
}

// CreateBatch creates up to 100 payouts in one request. Idempotency works as
// in Create.
func (s *PayoutService) CreateBatch(ctx context.Context, params *CreateBatchParams, opts ...RequestOption) (*PayoutBatch, error) {
	data, err := s.http.postIdempotent(ctx, "/vendor/payout/batch", params, opts)
	if err != nil {
		return nil, err
	}
	var b PayoutBatch
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("decane: unmarshal batch: %w", err)
	}
	return &b, nil
}

// Get retrieves a payout by ID (GET /payouts/{id}).
func (s *PayoutService) Get(ctx context.Context, id string) (*Payout, error) {
	data, err := s.http.get(ctx, "/payouts/"+id, nil)
	if err != nil {
		return nil, err
	}
	var p Payout
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("decane: unmarshal payout: %w", err)
	}
	return &p, nil
}

// List returns one page of payouts, newest first (GET /payouts).
func (s *PayoutService) List(ctx context.Context, params *ListParams) (*List[Payout], error) {
	data, err := s.http.get(ctx, "/payouts", listQuery(params))
	if err != nil {
		return nil, err
	}
	var result List[Payout]
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decane: unmarshal payouts: %w", err)
	}
	return &result, nil
}
