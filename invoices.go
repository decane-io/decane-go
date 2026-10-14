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
		return nil, fmt.Errorf("decane: unmarshal invoice: %w", err)
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
		return nil, fmt.Errorf("decane: unmarshal invoice: %w", err)
	}
	return &inv, nil
}

// List returns one page of invoices, newest first of invoices.
func (s *InvoiceService) List(ctx context.Context, params *ListParams) (*List[Invoice], error) {
	data, err := s.http.get(ctx, "/invoices", listQuery(params))
	if err != nil {
		return nil, err
	}
	var result List[Invoice]
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decane: unmarshal invoices: %w", err)
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
		return nil, fmt.Errorf("decane: unmarshal invoice: %w", err)
	}
	return &inv, nil
}

// Simulate settles a test invoice. outcome is "completed", "underpaid" or
// "expired". It works with a test key (gr_test_...) only.
func (s *InvoiceService) Simulate(ctx context.Context, id, outcome string) (*Invoice, error) {
	data, err := s.http.post(ctx, "/invoices/"+id+"/simulate", map[string]string{"outcome": outcome})
	if err != nil {
		return nil, err
	}
	var inv Invoice
	if err := json.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("decane: unmarshal invoice: %w", err)
	}
	return &inv, nil
}

func listQuery(p *ListParams) map[string]string {
	if p == nil {
		return nil
	}
	m := map[string]string{}
	if p.Limit > 0 {
		m["limit"] = strconv.Itoa(p.Limit)
	}
	if p.StartingAfter != "" {
		m["starting_after"] = p.StartingAfter
	}
	if p.Status != "" {
		m["status"] = p.Status
	}
	return m
}
