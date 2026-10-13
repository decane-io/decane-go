package decane

import (
	"context"
	"encoding/json"
	"fmt"
)

// WalletService handles wallet operations.
type WalletService struct {
	http *httpClient
}

// Generate creates a new deposit wallet.
func (s *WalletService) Generate(ctx context.Context, params *GenerateWalletParams) (*Wallet, error) {
	data, err := s.http.post(ctx, "/wallets/generate", params)
	if err != nil {
		return nil, err
	}
	var w Wallet
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("flow: unmarshal wallet: %w", err)
	}
	return &w, nil
}

// Get retrieves a wallet by ID.
func (s *WalletService) Get(ctx context.Context, id string) (*Wallet, error) {
	data, err := s.http.get(ctx, "/wallets/"+id, nil)
	if err != nil {
		return nil, err
	}
	var w Wallet
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("flow: unmarshal wallet: %w", err)
	}
	return &w, nil
}

// List returns a paginated list of wallets.
func (s *WalletService) List(ctx context.Context, params *ListParams) (*PaginatedList[Wallet], error) {
	p := listToMap(params)
	data, err := s.http.get(ctx, "/wallets", p)
	if err != nil {
		return nil, err
	}
	var result PaginatedList[Wallet]
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal wallets: %w", err)
	}
	return &result, nil
}

// Balance retrieves the balance for a specific wallet.
func (s *WalletService) Balance(ctx context.Context, id string) (*Wallet, error) {
	data, err := s.http.get(ctx, "/wallets/"+id+"/balance", nil)
	if err != nil {
		return nil, err
	}
	var w Wallet
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("flow: unmarshal wallet balance: %w", err)
	}
	return &w, nil
}
