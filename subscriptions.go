package decane

import (
	"context"
	"encoding/json"
	"fmt"
)

type SubscriptionService struct{ http *httpClient }

type CreatePlanParams struct {
	Name           string  `json:"name"`
	Amount         float64 `json:"amount"`
	Description    string  `json:"description,omitempty"`
	Currency       string  `json:"currency,omitempty"`
	IntervalType   string  `json:"interval_type,omitempty"`
	IntervalCount  int     `json:"interval_count,omitempty"`
	TrialDays      int     `json:"trial_days,omitempty"`
	MaxSubscribers int     `json:"max_subscribers,omitempty"`
	WebhookURL     string  `json:"webhook_url,omitempty"`
}

func (s *SubscriptionService) CreatePlan(ctx context.Context, params *CreatePlanParams) (map[string]interface{}, error) {
	data, err := s.http.post(ctx, "/merchants/me/subscription-plans", params)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal plan: %w", err)
	}
	return result, nil
}

func (s *SubscriptionService) ListPlans(ctx context.Context, params map[string]string) ([]map[string]interface{}, error) {
	data, err := s.http.get(ctx, "/merchants/me/subscription-plans", params)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal plans: %w", err)
	}
	return result, nil
}

func (s *SubscriptionService) UpdatePlan(ctx context.Context, planID string, params map[string]interface{}) (map[string]interface{}, error) {
	data, err := s.http.do(ctx, "PATCH", "/merchants/me/subscription-plans/"+planID, params, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal plan: %w", err)
	}
	return result, nil
}

func (s *SubscriptionService) DeletePlan(ctx context.Context, planID string) error {
	_, err := s.http.do(ctx, "DELETE", "/merchants/me/subscription-plans/"+planID, nil, nil)
	return err
}

func (s *SubscriptionService) ListSubscriptions(ctx context.Context, params map[string]string) ([]map[string]interface{}, error) {
	data, err := s.http.get(ctx, "/merchants/me/subscriptions", params)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal subscriptions: %w", err)
	}
	return result, nil
}

func (s *SubscriptionService) CancelSubscription(ctx context.Context, subID string) (map[string]interface{}, error) {
	data, err := s.http.post(ctx, "/merchants/me/subscriptions/"+subID+"/cancel", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal subscription: %w", err)
	}
	return result, nil
}

func (s *SubscriptionService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	data, err := s.http.get(ctx, "/merchants/me/subscriptions/stats", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal stats: %w", err)
	}
	return result, nil
}

func (s *SubscriptionService) ListPayments(ctx context.Context, params map[string]string) ([]map[string]interface{}, error) {
	data, err := s.http.get(ctx, "/merchants/me/subscription-payments", params)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("flow: unmarshal payments: %w", err)
	}
	return result, nil
}
