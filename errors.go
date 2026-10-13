package decane

import "fmt"

// DecaneError is the base error type for all API errors.
type DecaneError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Status  int    `json:"status"`
}

func (e *DecaneError) Error() string {
	return fmt.Sprintf("flow: %s (code=%s, status=%d)", e.Message, e.Code, e.Status)
}

// AuthenticationError indicates an invalid or missing API key.
type AuthenticationError struct{ DecaneError }

// NotFoundError indicates the requested resource was not found.
type NotFoundError struct{ DecaneError }

// ValidationError indicates invalid request parameters.
type ValidationError struct {
	DecaneError
	Errors []map[string]string `json:"errors"`
}

// RateLimitError indicates too many requests.
type RateLimitError struct {
	DecaneError
	RetryAfter float64 `json:"retry_after"`
}

// ServerError indicates a server-side error.
type ServerError struct{ DecaneError }

// InvalidSignatureError indicates webhook signature verification failed.
type InvalidSignatureError struct {
	Message string
}

func (e *InvalidSignatureError) Error() string {
	if e.Message == "" {
		return "flow: invalid webhook signature"
	}
	return "flow: " + e.Message
}
