package decane

import (
	"context"
	"encoding/json"
	"fmt"
)

// BalanceService reads your balances.
type BalanceService struct {
	http *httpClient
}

// Get returns your balances (GET /balances). Needs the wallets:read scope.
func (s *BalanceService) Get(ctx context.Context) (*Balance, error) {
	data, err := s.http.get(ctx, "/balances", nil)
	if err != nil {
		return nil, err
	}
	var b Balance
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("decane: unmarshal balance: %w", err)
	}
	return &b, nil
}

// WalletService lists the named wallets the API key can use.
type WalletService struct {
	http *httpClient
}

// List returns the key's wallets, each with its own balance
// (GET /vendor/query/balances). Needs the wallets:read scope. Use a wallet ID
// as SourceWalletID on payouts and swaps.
func (s *WalletService) List(ctx context.Context) (*WalletList, error) {
	data, err := s.http.get(ctx, "/vendor/query/balances", nil)
	if err != nil {
		return nil, err
	}
	var l WalletList
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("decane: unmarshal wallets: %w", err)
	}
	return &l, nil
}
