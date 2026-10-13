package decane

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DefaultWebhookTolerance is the default maximum age of a Decane-Signature timestamp.
const DefaultWebhookTolerance = 5 * time.Minute

// WebhookService handles webhook signature verification.
type WebhookService struct {
	secret string
}

// NewWebhookService creates a new WebhookService with the given signing secret.
func NewWebhookService(secret string) *WebhookService {
	return &WebhookService{secret: secret}
}

// Verify checks the legacy X-Decane-Signature header and returns the parsed event.
//
// Deprecated: the legacy signature has no timestamp, so it allows replays.
// Use ConstructEvent or VerifySignature with the Decane-Signature header.
func (s *WebhookService) Verify(payload []byte, signature string) (*WebhookEvent, error) {
	if s.secret == "" {
		return nil, &InvalidSignatureError{Message: "webhook secret not configured"}
	}

	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return nil, &InvalidSignatureError{}
	}

	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("flow: unmarshal webhook event: %w", err)
	}
	return &event, nil
}

// IsValid checks if a legacy webhook signature is valid without parsing the event.
//
// Deprecated: use VerifySignature with the Decane-Signature header.
func (s *WebhookService) IsValid(payload []byte, signature string) bool {
	if s.secret == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// VerifyRequest reads and verifies a webhook from an http.Request.
// It reads the signature from the X-Decane-Signature header, falling back to
// X-Granthe-Signature and the legacy X-Flow-Signature headers for backward compatibility.
func (s *WebhookService) VerifyRequest(r *http.Request) (*WebhookEvent, error) {
	signature := r.Header.Get("X-Decane-Signature")
	if signature == "" {
		signature = r.Header.Get("X-Granthe-Signature")
	}
	if signature == "" {
		signature = r.Header.Get("X-Flow-Signature")
	}
	if signature == "" {
		return nil, &InvalidSignatureError{Message: "missing X-Decane-Signature header"}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("flow: read request body: %w", err)
	}
	defer r.Body.Close()

	return s.Verify(body, signature)
}

// VerifySignature checks a Decane-Signature header ("t=<unix seconds>,v1=<hex>").
// The signed message is "<t>." followed by the raw body. It returns nil when
// valid, or an *InvalidSignatureError when the header is malformed, the
// timestamp is more than tolerance from now, or no v1 value matches.
// A tolerance of zero or less uses DefaultWebhookTolerance (5 minutes).
func (s *WebhookService) VerifySignature(payload []byte, header string, tolerance time.Duration) error {
	if s.secret == "" {
		return &InvalidSignatureError{Message: "webhook secret not configured"}
	}
	if tolerance <= 0 {
		tolerance = DefaultWebhookTolerance
	}

	var ts int64
	haveTS := false
	var candidates []string
	for _, part := range strings.Split(header, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		switch k {
		case "t":
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return &InvalidSignatureError{Message: "invalid Decane-Signature timestamp"}
			}
			ts, haveTS = n, true
		case "v1":
			candidates = append(candidates, v)
		}
	}
	if !haveTS || len(candidates) == 0 {
		return &InvalidSignatureError{Message: "malformed Decane-Signature header"}
	}

	age := time.Since(time.Unix(ts, 0))
	if age < 0 {
		age = -age
	}
	if age > tolerance {
		return &InvalidSignatureError{Message: "Decane-Signature timestamp outside tolerance"}
	}

	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10) + "."))
	mac.Write(payload)
	expected := []byte(hex.EncodeToString(mac.Sum(nil)))

	matched := false
	for _, c := range candidates {
		if hmac.Equal(expected, []byte(c)) {
			matched = true
		}
	}
	if !matched {
		return &InvalidSignatureError{}
	}
	return nil
}

// ConstructEvent verifies a Decane-Signature header and returns the parsed event.
func (s *WebhookService) ConstructEvent(payload []byte, header string, tolerance time.Duration) (*WebhookEvent, error) {
	if err := s.VerifySignature(payload, header, tolerance); err != nil {
		return nil, err
	}
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("flow: unmarshal webhook event: %w", err)
	}
	return &event, nil
}

// ConstructEventFromRequest reads the body and Decane-Signature header (or the older Granthe-Signature) of an
// http.Request, verifies them, and returns the parsed event. Pass 0 for the
// default tolerance.
func (s *WebhookService) ConstructEventFromRequest(r *http.Request, tolerance time.Duration) (*WebhookEvent, error) {
	header := r.Header.Get("Decane-Signature")
	if header == "" {
		header = r.Header.Get("Granthe-Signature")
	}
	if header == "" {
		return nil, &InvalidSignatureError{Message: "missing Decane-Signature header"}
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("flow: read request body: %w", err)
	}
	defer r.Body.Close()
	return s.ConstructEvent(body, header, tolerance)
}
