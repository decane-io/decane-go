package decane

import (
	"context"
	"encoding/json"
	"fmt"
)

// EventService reads the webhook log. Needs the webhooks:manage scope.
type EventService struct {
	http *httpClient
}

// List returns webhook events, newest first (GET /events).
func (s *EventService) List(ctx context.Context, params *ListParams) (*List[Event], error) {
	data, err := s.http.get(ctx, "/events", listQuery(params))
	if err != nil {
		return nil, err
	}
	var l List[Event]
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("decane: unmarshal events: %w", err)
	}
	return &l, nil
}

// Resend sends an event again now (POST /events/{id}/resend).
func (s *EventService) Resend(ctx context.Context, id string) (*Event, error) {
	data, err := s.http.post(ctx, "/events/"+id+"/resend", nil)
	if err != nil {
		return nil, err
	}
	var e Event
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("decane: unmarshal event: %w", err)
	}
	return &e, nil
}
