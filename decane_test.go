package decane

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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

// capture records the last request a test server saw.
type capture struct {
	method, path, query string
	body                map[string]interface{}
	header              http.Header
}

func serve(t *testing.T, reply string) (*Client, *capture, func()) {
	t.Helper()
	c := &capture{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.method, c.path, c.query, c.header = r.Method, r.URL.Path, r.URL.RawQuery, r.Header
		c.body = nil
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&c.body)
		}
		w.Write([]byte(reply))
	}))
	client := New("gr_key", WithAPISecret("gr_secret_x"), WithBaseURL(ts.URL))
	return client, c, ts.Close
}

func check(t *testing.T, c *capture, method, path string) {
	t.Helper()
	if c.method != method || c.path != path {
		t.Fatalf("want %s %s, got %s %s", method, path, c.method, c.path)
	}
}

func TestNewClient(t *testing.T) {
	c := New("gr_x", WithAPISecret("s"), WithWebhookSecret("w"))
	if c.Invoices == nil || c.Payouts == nil || c.Balances == nil || c.Wallets == nil || c.Swaps == nil || c.Events == nil || c.Webhooks == nil {
		t.Fatal("expected all services")
	}
	if Version != "0.9.0" {
		t.Fatalf("version %s", Version)
	}
}

func TestAuthHeaders(t *testing.T) {
	client, c, done := serve(t, `{}`)
	defer done()
	client.Balances.Get(context.Background())
	if c.header.Get("X-API-Key") != "gr_key" || c.header.Get("X-API-Secret") != "gr_secret_x" {
		t.Fatalf("auth headers: %v", c.header)
	}
	if ua := c.header.Get("User-Agent"); ua != "decane-go/0.9.0" {
		t.Fatalf("user agent %q", ua)
	}
}

func TestInvoiceCreateGetCancelSimulate(t *testing.T) {
	client, c, done := serve(t, `{"id":"inv_1","status":"created","fiat_amount":"100","fiat_currency":"USD","crypto_amount":"100.000000","received_amount":"0"}`)
	defer done()
	ctx := context.Background()
	inv, err := client.Invoices.Create(ctx, &CreateInvoiceParams{Amount: "100", Currency: "USD", Network: "tron", CryptoCurrency: "USDT"})
	if err != nil {
		t.Fatal(err)
	}
	check(t, c, "POST", "/invoices")
	if c.body["amount"] != "100" || c.body["crypto_currency"] != "USDT" {
		t.Fatalf("body %v", c.body)
	}
	if inv.ID != "inv_1" || inv.Amount != "100" || inv.CryptoAmount != "100.000000" {
		t.Fatalf("invoice %+v", inv)
	}
	client.Invoices.Get(ctx, "inv_1")
	check(t, c, "GET", "/invoices/inv_1")
	client.Invoices.Cancel(ctx, "inv_1")
	check(t, c, "POST", "/invoices/inv_1/cancel")
	client.Invoices.Simulate(ctx, "inv_1", "completed")
	check(t, c, "POST", "/invoices/inv_1/simulate")
	if c.body["outcome"] != "completed" {
		t.Fatalf("body %v", c.body)
	}
}

func TestCursorList(t *testing.T) {
	client, c, done := serve(t, `{"object":"list","data":[{"id":"inv_1"},{"id":"inv_2"}],"has_more":true,"next_cursor":"inv_2"}`)
	defer done()
	l, err := client.Invoices.List(context.Background(), &ListParams{Limit: 2, StartingAfter: "inv_0", Status: "completed"})
	if err != nil {
		t.Fatal(err)
	}
	check(t, c, "GET", "/invoices")
	q, _ := url.ParseQuery(c.query)
	if q.Get("limit") != "2" || q.Get("starting_after") != "inv_0" || q.Get("status") != "completed" || q.Get("page") != "" {
		t.Fatalf("query %s", c.query)
	}
	if len(l.Data) != 2 || !l.HasMore || l.NextCursor != "inv_2" {
		t.Fatalf("list %+v", l)
	}
}

func TestPayouts(t *testing.T) {
	client, c, done := serve(t, `{"id":"po_1","status":"pending","gross_amount":50.5,"net_amount":"50","fee_amount":0.5,"currency":"USDT"}`)
	defer done()
	ctx := context.Background()
	po, err := client.Payouts.Create(ctx, &CreatePayoutParams{
		RecipientAddress: "TXyz", Network: "tron", Currency: "USDT", Amount: "50.5",
		ReferenceID: "ref-1", SourceWalletID: "w-1",
	}, WithIdempotencyKey("k-1"))
	if err != nil {
		t.Fatal(err)
	}
	check(t, c, "POST", "/vendor/payout")
	if c.body["amount"] != "50.5" || c.body["source_wallet_id"] != "w-1" || c.body["recipient_address"] != "TXyz" || c.body["reference_id"] != "ref-1" {
		t.Fatalf("body %v", c.body)
	}
	if c.header.Get("Idempotency-Key") != "k-1" {
		t.Fatal("idempotency key")
	}
	if po.GrossAmount != "50.5" || po.NetAmount != "50" || po.FeeAmount != "0.5" {
		t.Fatalf("payout %+v", po)
	}

	client.Payouts.Get(ctx, "po_1")
	check(t, c, "GET", "/payouts/po_1")
	client.Payouts.List(ctx, &ListParams{Limit: 5})
	check(t, c, "GET", "/payouts")

	client.Payouts.CreateBatch(ctx, &CreateBatchParams{
		ReferenceID: "b",
		Payouts:     []*CreatePayoutParams{{RecipientAddress: "T1", Network: "tron", Currency: "USDT", Amount: "1"}},
	})
	check(t, c, "POST", "/vendor/payout/batch")
	ps, ok := c.body["payouts"].([]interface{})
	if !ok || len(ps) != 1 || c.body["reference_id"] != "b" {
		t.Fatalf("batch body %v", c.body)
	}
}

func TestBalancesAndWallets(t *testing.T) {
	client, c, done := serve(t, `{"object":"balance","pooled":{"tron":{"USDT":"12.5"}},"account":{"BTC":"0.1"},"pending":{}}`)
	defer done()
	b, err := client.Balances.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	check(t, c, "GET", "/balances")
	if b.Pooled["tron"]["USDT"] != "12.5" || b.Account["BTC"] != "0.1" {
		t.Fatalf("balance %+v", b)
	}

	client2, c2, done2 := serve(t, `{"api_key_id":"k","api_key_name":"main","wallets":[{"wallet_id":"w1","name":"Ops","network":"tron","address":"TA","native_symbol":"TRX","native_balance":"3","token_balances":{"USDT":"12.5"}}],"fetched_at":"now"}`)
	defer done2()
	l, err := client2.Wallets.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	check(t, c2, "GET", "/vendor/query/balances")
	if len(l.Wallets) != 1 || l.Wallets[0].ID != "w1" || l.Wallets[0].TokenBalances["USDT"] != "12.5" {
		t.Fatalf("wallets %+v", l)
	}
}

func TestSwaps(t *testing.T) {
	ctx := context.Background()
	client, c, done := serve(t, `{"quote_id":"q1","to_amount":"20.1","from_amount":"0.5","slippage":0.5}`)
	defer done()
	q, err := client.Swaps.Quote(ctx, &SwapQuoteParams{FromNetwork: "bsc", FromCurrency: "BNB", FromAmount: "0.5", ToNetwork: "tron", ToCurrency: "USDT"})
	if err != nil {
		t.Fatal(err)
	}
	check(t, c, "POST", "/vendor/swap/quote")
	for k, want := range map[string]string{"from_network": "bsc", "from_currency": "BNB", "from_amount": "0.5", "to_network": "tron", "to_currency": "USDT"} {
		if c.body[k] != want {
			t.Fatalf("%s = %v, body %v", k, c.body[k], c.body)
		}
	}
	if _, has := c.body["amount"]; has {
		t.Fatal("old amount field sent")
	}
	if q.QuoteID != "q1" || q.ToAmount != "20.1" {
		t.Fatalf("quote %+v", q)
	}

	client.Swaps.Create(ctx, &SwapParams{
		QuoteID: "q1", FromNetwork: "bsc", FromCurrency: "BNB", FromAmount: "0.5", ToNetwork: "tron", ToCurrency: "USDT",
		SourceWalletID: "w1", DestinationAddress: "TXyz",
	})
	check(t, c, "POST", "/vendor/swap")
	if c.body["quote_id"] != "q1" || c.body["source_wallet_id"] != "w1" || c.body["destination_address"] != "TXyz" {
		t.Fatalf("body %v", c.body)
	}
	if _, has := c.body["destination_wallet_id"]; has {
		t.Fatal("empty destination_wallet_id should be omitted")
	}

	client.Swaps.Get(ctx, "s1")
	check(t, c, "GET", "/vendor/swap/s1")
}

func TestSwapListAndDone(t *testing.T) {
	client, c, done := serve(t, `{"swaps":[{"swap_id":"s1","status":"COMPLETED","from_amount":"1"}],"total":7}`)
	defer done()
	l, err := client.Swaps.List(context.Background(), &SwapListParams{Status: "completed", Limit: 10, Offset: 20})
	if err != nil {
		t.Fatal(err)
	}
	check(t, c, "GET", "/vendor/swaps")
	q, _ := url.ParseQuery(c.query)
	if q.Get("limit") != "10" || q.Get("offset") != "20" || q.Get("status") != "completed" {
		t.Fatalf("query %s", c.query)
	}
	if l.Total != 7 || l.Swaps[0].ID != "s1" || !l.Swaps[0].Done() {
		t.Fatalf("list %+v", l)
	}
	if (&Swap{Status: "processing"}).Done() {
		t.Fatal("processing is not done")
	}
}

func TestEvents(t *testing.T) {
	client, c, done := serve(t, `{"object":"list","data":[{"id":"e1","event":"invoice.completed","status":"delivered","attempts":1,"response_status":200,"livemode":true}],"has_more":false,"next_cursor":null}`)
	defer done()
	ctx := context.Background()
	l, err := client.Events.List(ctx, &ListParams{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	check(t, c, "GET", "/events")
	if len(l.Data) != 1 || l.Data[0].Type != "invoice.completed" || *l.Data[0].ResponseStatus != 200 {
		t.Fatalf("events %+v", l)
	}
	if _, err := client.Events.Resend(ctx, "e1"); err != nil {
		t.Fatal(err)
	}
	check(t, c, "POST", "/events/e1/resend")
}

func TestAmountHelpers(t *testing.T) {
	x, y := 0.1, 0.2
	if FormatAmount(x+y) != "0.30000000000000004" || FormatAmount(12.5) != "12.5" || FormatAmount(1e21) != "1000000000000000000000" {
		t.Fatal("FormatAmount")
	}
	var a Amount
	if err := json.Unmarshal([]byte(`12.50`), &a); err != nil || a != "12.50" {
		t.Fatalf("number: %q %v", a, err)
	}
}

func TestErrorTypes(t *testing.T) {
	for status, want := range map[int]string{403: "*decane.PermissionError", 409: "*decane.ConflictError", 422: "*decane.ValidationError"} {
		st := status
		client, _, done := func() (*Client, *capture, func()) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(st)
				w.Write([]byte(`{"detail":"x","error":{"type":"invalid_request_error","code":"bad","message":"bad field","param":"amount"}}`))
			}))
			return New("k", WithBaseURL(ts.URL)), nil, ts.Close
		}()
		_, err := client.Balances.Get(context.Background())
		if got := fmt.Sprintf("%T", err); got != want {
			t.Fatalf("%d: want %s got %s", st, want, got)
		}
		done()
	}
}

func TestWebhookVerify(t *testing.T) {
	secret := "whsec_test_secret"
	payload := []byte(`{"id":"evt_123","event":"invoice.completed","data":{"invoice_id":"inv_123"}}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	client := New("test_key", WithWebhookSecret(secret))

	event, err := client.Webhooks.Verify(payload, signature)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.ID != "evt_123" || event.Type != "invoice.completed" {
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
		w.Write([]byte(`{"detail":"Invalid API key","error":{"type":"authentication_error","code":"unauthorized","message":"Invalid API key","param":null},"request_id":"req_1"}`))
	})
	defer ts.Close()

	_, err := client.Invoices.Get(context.Background(), "inv_123")
	if err == nil {
		t.Fatal("expected error")
	}
	ae, ok := err.(*AuthenticationError)
	if !ok {
		t.Fatalf("expected AuthenticationError, got %T: %v", err, err)
	}
	if ae.Code != "unauthorized" || ae.Type != "authentication_error" || ae.RequestID != "req_1" || ae.Message != "Invalid API key" {
		t.Fatalf("unexpected error fields: %+v", ae)
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

	client.Invoices.Create(ctx, &CreateInvoiceParams{Amount: "1"}, WithIdempotencyKey("order-42"))
	client.Payouts.Create(ctx, &CreatePayoutParams{Amount: "1"}, WithIdempotencyKey("p-1"))
	client.Payouts.CreateBatch(ctx, &CreateBatchParams{}, WithIdempotencyKey("b-1"))
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
	if _, err := client.Invoices.Create(context.Background(), &CreateInvoiceParams{Amount: "1"}); err != nil {
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

func TestConstructEventFromRequest(t *testing.T) {
	secret := "whsec_test"
	body := []byte(`{"id":"evt_1","event":"payout.completed","created":1700000000,"data":{"payout_id":"po_1"}}`)
	svc := NewWebhookService(secret)
	r, _ := http.NewRequest("POST", "/", bytes.NewReader(body))
	r.Header.Set("Decane-Signature", signHeader(secret, body, time.Now().Unix()))
	ev, err := svc.ConstructEventFromRequest(r, 0)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != "payout.completed" || ev.Data["payout_id"] != "po_1" || ev.Created != 1700000000 {
		t.Fatalf("event %+v", ev)
	}
	r2, _ := http.NewRequest("POST", "/", bytes.NewReader(body))
	if _, err := svc.ConstructEventFromRequest(r2, 0); err == nil {
		t.Fatal("expected error without header")
	}
}
