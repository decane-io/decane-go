package decane

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(handler)
	client := New("test_key", WithBaseURL(ts.URL))
	return client, ts
}

func TestNewClient(t *testing.T) {
	c := New("vf_live_test123")
	if c.Invoices == nil || c.Payouts == nil || c.Wallets == nil || c.Swaps == nil || c.Webhooks == nil {
		t.Fatal("expected all services to be initialized")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	c := New("vf_live_test123",
		WithBaseURL("https://custom.api.com"),
		WithWebhookSecret("whsec_test"),
	)
	if c.http.baseURL != "https://custom.api.com" {
		t.Fatalf("expected custom base URL, got %s", c.http.baseURL)
	}
}

func TestCreateInvoice(t *testing.T) {
	client, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/invoices" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "test_key" {
			t.Fatal("missing API key header")
		}
		json.NewEncoder(w).Encode(Invoice{
			ID: "inv_123", Amount: 100, Currency: "USDT", Network: "tron", Status: "pending",
		})
	})
	defer ts.Close()

	inv, err := client.Invoices.Create(context.Background(), &CreateInvoiceParams{
		Amount: 100, Currency: "USDT", Network: "tron",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.ID != "inv_123" {
		t.Fatalf("expected inv_123, got %s", inv.ID)
	}
}

func TestGetInvoice(t *testing.T) {
	client, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/invoices/inv_123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Invoice{ID: "inv_123", Status: "paid"})
	})
	defer ts.Close()

	inv, err := client.Invoices.Get(context.Background(), "inv_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != "paid" {
		t.Fatalf("expected paid, got %s", inv.Status)
	}
}

func TestListInvoices(t *testing.T) {
	client, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "1" || r.URL.Query().Get("limit") != "10" {
			t.Fatal("missing query params")
		}
		json.NewEncoder(w).Encode(PaginatedList[Invoice]{
			Items: []Invoice{{ID: "inv_1"}, {ID: "inv_2"}}, Total: 2, Page: 1, PerPage: 10,
		})
	})
	defer ts.Close()

	list, err := client.Invoices.List(context.Background(), &ListParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list.Items))
	}
}

func TestCreatePayout(t *testing.T) {
	client, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Payout{
			ID: "po_123", GrossAmount: 50, Currency: "USDT", Status: "pending",
		})
	})
	defer ts.Close()

	po, err := client.Payouts.Create(context.Background(), &CreatePayoutParams{
		Amount: 50, Currency: "USDT", Network: "tron", RecipientAddress: "T...",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if po.ID != "po_123" {
		t.Fatalf("expected po_123, got %s", po.ID)
	}
}

func TestGenerateWallet(t *testing.T) {
	client, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Wallet{
			ID: "w_123", Address: "T...", Network: "tron", Currency: "USDT",
		})
	})
	defer ts.Close()

	w, err := client.Wallets.Generate(context.Background(), &GenerateWalletParams{
		Network: "tron", Currency: "USDT",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.ID != "w_123" {
		t.Fatalf("expected w_123, got %s", w.ID)
	}
}

func TestSwapQuote(t *testing.T) {
	client, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SwapQuote{
			FromCurrency: "USDT", ToCurrency: "BTC", FromAmount: 1000, ToAmount: 0.015, Rate: 0.000015,
		})
	})
	defer ts.Close()

	q, err := client.Swaps.Quote(context.Background(), &SwapParams{
		FromCurrency: "USDT", ToCurrency: "BTC", Amount: 1000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Rate != 0.000015 {
		t.Fatalf("expected rate 0.000015, got %f", q.Rate)
	}
}

func TestWebhookVerify(t *testing.T) {
	secret := "whsec_test_secret"
	payload := []byte(`{"id":"evt_123","event":"invoice.paid","data":{"invoice_id":"inv_123"}}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	client := New("test_key", WithWebhookSecret(secret))

	event, err := client.Webhooks.Verify(payload, signature)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.ID != "evt_123" || event.Type != "invoice.paid" {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestWebhookInvalidSignature(t *testing.T) {
	client := New("test_key", WithWebhookSecret("whsec_secret"))
	_, err := client.Webhooks.Verify([]byte(`{}`), "bad_signature")
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
	if _, ok := err.(*InvalidSignatureError); !ok {
		t.Fatalf("expected InvalidSignatureError, got %T", err)
	}
}

func TestWebhookIsValid(t *testing.T) {
	secret := "whsec_test"
	payload := []byte(`{"test":true}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	client := New("test_key", WithWebhookSecret(secret))
	if !client.Webhooks.IsValid(payload, sig) {
		t.Fatal("expected valid signature")
	}
	if client.Webhooks.IsValid(payload, "wrong") {
		t.Fatal("expected invalid signature")
	}
}

func TestErrorParsing(t *testing.T) {
	client, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid API key"})
	})
	defer ts.Close()

	_, err := client.Invoices.Get(context.Background(), "inv_123")
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Fatalf("expected AuthenticationError, got %T: %v", err, err)
	}
}

func TestNotFoundError(t *testing.T) {
	client, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
	})
	defer ts.Close()

	_, err := client.Invoices.Get(context.Background(), "inv_nonexist")
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}

func TestRateLimitError(t *testing.T) {
	attempt := 0
	client, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.Header().Set("Retry-After", "0.1")
		w.WriteHeader(429)
		json.NewEncoder(w).Encode(map[string]string{"message": "Too many requests"})
	})
	defer ts.Close()

	_, err := client.Invoices.Get(context.Background(), "inv_123")
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*RateLimitError); !ok {
		t.Fatalf("expected RateLimitError, got %T", err)
	}
	if attempt < 2 {
		t.Fatalf("expected retries, got %d attempts", attempt)
	}
}

func signHeader(secret string, payload []byte, ts int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10) + "."))
	mac.Write(payload)
	return "t=" + strconv.FormatInt(ts, 10) + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature(t *testing.T) {
	secret := "whsec_test"
	payload := []byte(`{"id":"evt_1","event":"invoice.completed","created":1700000000,"data":{"id":"inv_1"}}`)
	client := New("test_key", WithWebhookSecret(secret))
	now := time.Now().Unix()

	h := signHeader(secret, payload, now)
	if err := client.Webhooks.VerifySignature(payload, h, 0); err != nil {
		t.Fatalf("expected valid: %v", err)
	}
	spaced := "t=" + strconv.FormatInt(now, 10) + ", v1=" + h[len("t="+strconv.FormatInt(now, 10)+",v1="):] + ", v0=zz"
	if err := client.Webhooks.VerifySignature(payload, spaced, 0); err != nil {
		t.Fatalf("expected valid with spaces and unknown part: %v", err)
	}

	// Fixed timestamp with a huge tolerance.
	if err := client.Webhooks.VerifySignature(payload, signHeader(secret, payload, 1700000000), 100000*time.Hour); err != nil {
		t.Fatalf("expected fixed vector valid: %v", err)
	}

	cases := map[string]struct {
		body   []byte
		header string
	}{
		"wrong secret": {payload, signHeader("other", payload, now)},
		"tampered":     {append([]byte(" "), payload...), h},
		"old":          {payload, signHeader(secret, payload, now-301)},
		"garbage":      {payload, "garbage"},
	}
	for name, c := range cases {
		err := client.Webhooks.VerifySignature(c.body, c.header, 0)
		if _, ok := err.(*InvalidSignatureError); !ok {
			t.Errorf("%s: expected InvalidSignatureError, got %v", name, err)
		}
	}
	if err := client.Webhooks.VerifySignature(payload, signHeader(secret, payload, now-301), 10*time.Minute); err != nil {
		t.Fatalf("expected valid with wider tolerance: %v", err)
	}
}

func TestConstructEvent(t *testing.T) {
	secret := "whsec_test"
	payload := []byte(`{"id":"evt_1","event":"invoice.completed","data":{"id":"inv_1"}}`)
	client := New("test_key", WithWebhookSecret(secret))
	event, err := client.Webhooks.ConstructEvent(payload, signHeader(secret, payload, time.Now().Unix()), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.ID != "evt_1" || event.Type != "invoice.completed" {
		t.Fatalf("unexpected event: %+v", event)
	}
	if _, err := client.Webhooks.ConstructEvent(payload, "t=1,v1=00", 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestIdempotencyKeySent(t *testing.T) {
	var got []string
	client, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Header.Get("Idempotency-Key"))
		w.Write([]byte(`{"id":"x","payouts":[]}`))
	})
	defer ts.Close()
	ctx := context.Background()

	client.Invoices.Create(ctx, &CreateInvoiceParams{Amount: 1}, WithIdempotencyKey("order-42"))
	client.Payouts.Create(ctx, &CreatePayoutParams{Amount: 1}, WithIdempotencyKey("p-1"))
	client.Payouts.CreateBatch(ctx, nil, WithIdempotencyKey("b-1"))
	for i, want := range []string{"order-42", "p-1", "b-1"} {
		if got[i] != want {
			t.Fatalf("call %d: want %q got %q", i, want, got[i])
		}
	}
}

func TestIdempotencyKeyGeneratedAndStableAcrossRetries(t *testing.T) {
	var keys []string
	client, ts := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		if len(keys) == 1 {
			w.WriteHeader(503)
			return
		}
		w.Write([]byte(`{"id":"x"}`))
	})
	defer ts.Close()
	if _, err := client.Invoices.Create(context.Background(), &CreateInvoiceParams{Amount: 1}); err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 || keys[0] != keys[1] || len(keys[0]) != 36 {
		t.Fatalf("expected one stable uuid across retries, got %v", keys)
	}
}

func TestConstructEventNumericTimestamp(t *testing.T) {
	secret := "whsec_test"
	payload := []byte(`{"id":"evt_1","event":"payout.completed","created":1700000000,"timestamp":1700000000,"data":{}}`)
	client := New("test_key", WithWebhookSecret(secret))
	event, err := client.Webhooks.ConstructEvent(payload, signHeader(secret, payload, time.Now().Unix()), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.CreatedAt != "1700000000" {
		t.Fatalf("unexpected CreatedAt %q", event.CreatedAt)
	}
}

func TestLegacyGrantheHeaderFallback(t *testing.T) {
	secret := "gr_whsec_test"
	body := []byte(`{"type":"invoice.paid"}`)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "."))
	mac.Write(body)
	hdr := "t=" + ts + ",v1=" + hex.EncodeToString(mac.Sum(nil))
	svc := NewWebhookService(secret)
	for _, name := range []string{"Decane-Signature", "Granthe-Signature"} {
		r, _ := http.NewRequest("POST", "/", bytes.NewReader(body))
		r.Header.Set(name, hdr)
		if _, err := svc.ConstructEventFromRequest(r, 0); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
}
