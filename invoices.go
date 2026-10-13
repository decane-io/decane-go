package decane

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// InvoiceService handles invoice operations.
type InvoiceService struct {
	http *httpClient
}

// Create creates a new payment invoice. Pass WithIdempotencyKey to set the
// Idempotency-Key header; if omitted a random key is generated so automatic
// retries are safe.
func (s *InvoiceService) Create(ctx context.Context, params *CreateInvoiceParams, opts ...RequestOption) (*Invoice, error) {
	data, err := s.http.postIdempotent(ctx, "/invoices", params, opts)
	if err != nil {
		return nil, err
	}
	var inv Invoice
	if err := json.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("flow: unmarshal invoice: %w", err)
	}
	return &inv, nil
}

// Get retrieves an invoice by ID.
func (s *InvoiceService) Get(ctx context.Context, id string) (*Invoice, error) {
	data, err := s.http.get(ctx, "/invoices/"+id, nil)
	if err != nil {
		return nil, err
	}
	var inv Invoice
	if err := json.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("flow: unmarshal invoice: %w", err)
	}
	return &inv, nil
}

// List returns a paginated list of invoices.
func (s *InvoiceService) List(ctx context.Context, params *ListParams) (*PaginatedList[Invoice], error) {
	p := listToMap(params)
	data, err := s.http.get(ctx, "/invoices", p)
	if err != nil {
		return nil, err
	}
	var result PaginatedList[Invoice]
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal invoices: %w", err)
	}
	return &result, nil
}

// Cancel cancels a pending invoice.
func (s *InvoiceService) Cancel(ctx context.Context, id string) (*Invoice, error) {
	data, err := s.http.post(ctx, "/invoices/"+id+"/cancel", nil)
	if err != nil {
		return nil, err
	}
	var inv Invoice
	if err := json.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("flow: unmarshal invoice: %w", err)
	}
	return &inv, nil
}

func listToMap(p *ListParams) map[string]string {
	if p == nil {
		return nil
	}
	m := map[string]string{}
	if p.Page > 0 {
		m["page"] = strconv.Itoa(p.Page)
	}
	if p.Limit > 0 {
		m["limit"] = strconv.Itoa(p.Limit)
	}
	if p.Status != "" {
		m["status"] = p.Status
	}
	if p.Currency != "" {
		m["currency"] = p.Currency
	}
	if p.Network != "" {
		m["network"] = p.Network
	}
	return m
}
