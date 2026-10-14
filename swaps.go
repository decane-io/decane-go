package decane

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// SwapService handles swaps. Needs the swaps:execute scope.
type SwapService struct {
	http *httpClient
}

// Quote gets a price quote (POST /vendor/swap/quote). It commits to nothing.
func (s *SwapService) Quote(ctx context.Context, params *SwapQuoteParams) (*SwapQuote, error) {
	data, err := s.http.post(ctx, "/vendor/swap/quote", params)
	if err != nil {
		return nil, err
	}
	var q SwapQuote
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, fmt.Errorf("decane: unmarshal swap quote: %w", err)
	}
	return &q, nil
}

// Create executes a swap (POST /vendor/swap). Poll Get until Done.
func (s *SwapService) Create(ctx context.Context, params *SwapParams) (*Swap, error) {
	data, err := s.http.post(ctx, "/vendor/swap", params)
	if err != nil {
		return nil, err
	}
	var sw Swap
	if err := json.Unmarshal(data, &sw); err != nil {
		return nil, fmt.Errorf("decane: unmarshal swap: %w", err)
	}
	return &sw, nil
}

// Get retrieves a swap by ID (GET /vendor/swap/{id}).
func (s *SwapService) Get(ctx context.Context, id string) (*Swap, error) {
	data, err := s.http.get(ctx, "/vendor/swap/"+id, nil)
	if err != nil {
		return nil, err
	}
	var sw Swap
	if err := json.Unmarshal(data, &sw); err != nil {
		return nil, fmt.Errorf("decane: unmarshal swap: %w", err)
	}
	return &sw, nil
}

// List returns swaps made with API keys (GET /vendor/swaps). Unlike the other
// lists it pages by Limit (1 to 200) and Offset, and returns a total.
func (s *SwapService) List(ctx context.Context, params *SwapListParams) (*SwapList, error) {
	q := map[string]string{}
	if params != nil {
		if params.Status != "" {
			q["status"] = params.Status
		}
		if params.FromCurrency != "" {
			q["from_currency"] = params.FromCurrency
		}
		if params.ToCurrency != "" {
			q["to_currency"] = params.ToCurrency
		}
		if params.Limit > 0 {
			q["limit"] = strconv.Itoa(params.Limit)
		}
		if params.Offset > 0 {
			q["offset"] = strconv.Itoa(params.Offset)
		}
	}
	data, err := s.http.get(ctx, "/vendor/swaps", q)
	if err != nil {
		return nil, err
	}
	var l SwapList
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("decane: unmarshal swaps: %w", err)
	}
	return &l, nil
}
