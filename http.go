package decane

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const maxRetries = 3

type httpClient struct {
	baseURL   string
	apiKey    string
	apiSecret string
	timeout   time.Duration
	client    *http.Client
}

func newHTTPClient(apiKey, apiSecret, baseURL string, timeout time.Duration) *httpClient {
	return &httpClient{
		baseURL:   baseURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		timeout:   timeout,
		client:    &http.Client{Timeout: timeout},
	}
}

func (h *httpClient) do(ctx context.Context, method, path string, body interface{}, params map[string]string) ([]byte, error) {
	return h.doHeaders(ctx, method, path, body, params, nil)
}

func (h *httpClient) doHeaders(ctx context.Context, method, path string, body interface{}, params map[string]string, headers map[string]string) ([]byte, error) {
	fullURL := h.baseURL + path

	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			if v != "" {
				q.Set(k, v)
			}
		}
		if qs := q.Encode(); qs != "" {
			fullURL += "?" + qs
		}
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("flow: marshal error: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("flow: request error: %w", err)
		}
		req.Header.Set("X-API-Key", h.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "decane-go/"+Version)
		if h.apiSecret != "" {
			req.Header.Set("X-API-Secret", h.apiSecret)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		// Reset body reader for retries
		if body != nil {
			data, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(data)
			req.Body = io.NopCloser(bodyReader)
		}

		resp, err := h.client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(1<<attempt) * 500 * time.Millisecond)
				continue
			}
			return nil, &DecaneError{Message: err.Error(), Code: "network_error"}
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("flow: read body error: %w", err)
		}

		if shouldRetry(resp.StatusCode) && attempt < maxRetries-1 {
			wait := time.Duration(1<<attempt) * 500 * time.Millisecond
			if resp.StatusCode == 429 {
				if ra := resp.Header.Get("Retry-After"); ra != "" {
					if secs, err := strconv.ParseFloat(ra, 64); err == nil {
						wait = time.Duration(secs * float64(time.Second))
					}
				}
			}
			time.Sleep(wait)
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, parseError(resp.StatusCode, respBody, resp.Header)
		}

		if resp.StatusCode == 204 {
			return nil, nil
		}

		return respBody, nil
	}

	if lastErr != nil {
		return nil, &DecaneError{Message: lastErr.Error(), Code: "network_error"}
	}
	return nil, &DecaneError{Message: "max retries exceeded", Code: "network_error"}
}

func (h *httpClient) get(ctx context.Context, path string, params map[string]string) ([]byte, error) {
	return h.do(ctx, http.MethodGet, path, nil, params)
}

func (h *httpClient) post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return h.do(ctx, http.MethodPost, path, body, nil)
}

// postIdempotent POSTs with an Idempotency-Key header. The SDK retries
// automatically, so if the caller gave no key a random UUID v4 is generated
// once and reused on every attempt.
func (h *httpClient) postIdempotent(ctx context.Context, path string, body interface{}, opts []RequestOption) ([]byte, error) {
	var ro requestOptions
	for _, o := range opts {
		o(&ro)
	}
	key := ro.idempotencyKey
	if key == "" {
		key = newUUID()
	}
	return h.doHeaders(ctx, http.MethodPost, path, body, nil, map[string]string{"Idempotency-Key": key})
}

type requestOptions struct {
	idempotencyKey string
}

// RequestOption customises a single create call.
type RequestOption func(*requestOptions)

// WithIdempotencyKey sets the Idempotency-Key header (1-255 characters:
// letters, digits, . _ : -). Retrying with the same key and body returns the
// first response. If you do not pass one, a random key is generated.
func WithIdempotencyKey(key string) RequestOption {
	return func(o *requestOptions) { o.idempotencyKey = key }
}

func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("decane: crypto/rand failed: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]
}

func shouldRetry(status int) bool {
	return status == 429 || status == 500 || status == 502 || status == 503 || status == 504
}

func parseError(status int, body []byte, headers http.Header) error {
	var errBody struct {
		Detail  string              `json:"detail"`
		Message string              `json:"message"`
		Error   string              `json:"error"`
		Code    string              `json:"code"`
		Errors  []map[string]string `json:"errors"`
	}
	_ = json.Unmarshal(body, &errBody)

	msg := errBody.Detail
	if msg == "" {
		msg = errBody.Message
	}
	if msg == "" {
		msg = errBody.Error
	}
	if msg == "" {
		msg = string(body)
	}

	base := DecaneError{Message: msg, Code: errBody.Code, Status: status}

	switch status {
	case 401:
		return &AuthenticationError{DecaneError: base}
	case 404:
		return &NotFoundError{DecaneError: base}
	case 422:
		return &ValidationError{DecaneError: base, Errors: errBody.Errors}
	case 429:
		var retryAfter float64
		if ra := headers.Get("Retry-After"); ra != "" {
			retryAfter, _ = strconv.ParseFloat(ra, 64)
		}
		return &RateLimitError{DecaneError: base, RetryAfter: retryAfter}
	default:
		if status >= 500 {
			return &ServerError{DecaneError: base}
		}
		return &base
	}
}
