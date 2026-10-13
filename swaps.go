package decane

import (
	"context"
	"encoding/json"
	"fmt"
)

// SwapService handles currency swap operations.
type SwapService struct {
	http *httpClient
}

// Quote gets a price quote for a swap.
func (s *SwapService) Quote(ctx context.Context, params *SwapParams) (*SwapQuote, error) {
	data, err := s.http.post(ctx, "/swaps/quote", params)
	if err != nil {
		return nil, err
	}
	var q SwapQuote
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, fmt.Errorf("flow: unmarshal swap quote: %w", err)
	}
	return &q, nil
}

// Create executes a swap.
func (s *SwapService) Create(ctx context.Context, params *SwapParams) (*Swap, error) {
	data, err := s.http.post(ctx, "/swaps", params)
	if err != nil {
		return nil, err
	}
	var sw Swap
	if err := json.Unmarshal(data, &sw); err != nil {
		return nil, fmt.Errorf("flow: unmarshal swap: %w", err)
	}
	return &sw, nil
}

// Get retrieves a swap by ID.
func (s *SwapService) Get(ctx context.Context, id string) (*Swap, error) {
	data, err := s.http.get(ctx, "/swaps/"+id, nil)
	if err != nil {
		return nil, err
	}
	var sw Swap
	if err := json.Unmarshal(data, &sw); err != nil {
		return nil, fmt.Errorf("flow: unmarshal swap: %w", err)
	}
	return &sw, nil
}

// List returns a paginated list of swaps.
func (s *SwapService) List(ctx context.Context, params *ListParams) (*PaginatedList[Swap], error) {
	p := listToMap(params)
	data, err := s.http.get(ctx, "/swaps", p)
	if err != nil {
		return nil, err
	}
	var result PaginatedList[Swap]
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal swaps: %w", err)
	}
	return &result, nil
}
