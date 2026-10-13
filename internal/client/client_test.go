package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MoomooOpen/moomoo-cli/internal/reporting"
)

// ── OAuth mode headers ──────────────────────────────────────────────────────

func TestSetHeaders_OAuthMode(t *testing.T) {
	c := New("https://api.example.com",
		WithToken("my-token"),
		WithClientName("test-client"),
		WithVersion("1.0.0"),
	)

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/v1/test", nil)
	c.setHeaders(req, context.Background(), nil)

	tests := []struct {
		header string
		want   string
	}{
		{headerUserAgent, "moomoo/1.0.0"},
		{headerClientName, "test-client"},
		{headerAuth, "Bearer my-token"},
	}
	for _, tt := range tests {
		got := req.Header.Get(tt.header)
		if got != tt.want {
			t.Errorf("header %s = %q, want %q", tt.header, got, tt.want)
		}
	}

}

func TestSetHeaders_OAuthMode_NoToken(t *testing.T) {
	c := New("https://api.example.com",
		WithToken(""),
	)

	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/v1/test", nil)
	c.setHeaders(req, context.Background(), nil)

	if auth := req.Header.Get(headerAuth); len(auth) > 0 {
		t.Errorf("Authorization should be empty when token is empty, got %q", auth)
	}
}

func TestSetHeaders_ContentType_WithBody(t *testing.T) {
	c := New("https://api.example.com")
	req, _ := http.NewRequest(http.MethodPost, "https://api.example.com/v1/order", nil)
	c.setHeaders(req, context.Background(), []byte(`{"qty":1}`))

	if ct := req.Header.Get(headerContentType); ct != contentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", ct, contentTypeJSON)
	}
}

func TestSetHeaders_ContentType_NoBody(t *testing.T) {
	c := New("https://api.example.com")
	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/v1/quote", nil)
	c.setHeaders(req, context.Background(), nil)

	if ct := req.Header.Get(headerContentType); len(ct) > 0 {
		t.Errorf("Content-Type should be absent for GET without body, got %q", ct)
	}
}

func TestSetHeaders_CliCmdFromContext(t *testing.T) {
	c := New("https://api.example.com")
	ctx := context.WithValue(context.Background(), CtxKeyCommand, "moomoo kline")
	req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/v1/kline", nil)
	c.setHeaders(req, ctx, nil)

	if cmd := req.Header.Get(headerCliCmd); cmd != "moomoo kline" {
		t.Errorf("X-Cli-Cmd = %q, want %q", cmd, "moomoo kline")
	}
}

func TestCallAddsReportSourceForRequiredEndpoints(t *testing.T) {
	tests := []struct {
		endpoint string
		opts     CallOpts
	}{
		{endpoint: "trade.order.place", opts: CallOpts{PathParams: map[string]string{"acc_id": "123"}}},
		{endpoint: "quote.cur_kline", opts: CallOpts{PathParams: map[string]string{"symbol": "US.AAPL"}}},
		{endpoint: "sim_trade.input_order", opts: CallOpts{PathParams: map[string]string{"acc_id": "SIM-1"}}},
	}
	for _, test := range tests {
		t.Run(test.endpoint, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get(headerReportSource); got != "source-1" {
					t.Fatalf("%s = %q, want source-1", headerReportSource, got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			c := New(srv.URL, WithReportSource("source-1"))
			if err := c.Call(context.Background(), test.endpoint, test.opts, nil); err != nil {
				t.Fatalf("Call() error = %v", err)
			}
		})
	}
}

func TestCallAddsReportSourceForEveryRegisteredEndpoint(t *testing.T) {
	for endpoint, ep := range Endpoints {
		t.Run(endpoint, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get(headerReportSource); got != "source-1" {
					t.Fatalf("%s = %q, want source-1", headerReportSource, got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			c := New(srv.URL, WithReportSource("source-1"))
			err := c.Call(context.Background(), endpoint, CallOpts{
				PathParams: placeholderPathParams(ep.Path),
			}, nil)
			if err != nil {
				t.Fatalf("Call() error = %v", err)
			}
		})
	}
}

func placeholderPathParams(path string) map[string]string {
	params := map[string]string{}
	for _, part := range strings.Split(path, "/") {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			params[strings.Trim(part, "{}")] = "value-1"
		}
	}
	return params
}

func TestCallAddsReportSourceForReadOnlyEndpoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(headerReportSource); got != "source-1" {
			t.Fatalf("%s = %q, want source-1", headerReportSource, got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := New(srv.URL, WithReportSource("source-1"))
	if err := c.Call(context.Background(), "trade.order.open", CallOpts{
		PathParams: map[string]string{"acc_id": "123"},
	}, nil); err != nil {
		t.Fatalf("Call() error = %v", err)
	}
}

func TestRegisteredEndpointsHaveReportingRoutes(t *testing.T) {
	for endpoint, ep := range Endpoints {
		route := reporting.MatchToolRoute(ep.Method, ep.Path)
		if route.ID == "" {
			t.Fatalf("%s %s (%s) has no reporting route", ep.Method, ep.Path, endpoint)
		}
	}
}

// ── Token management ────────────────────────────────────────────────────────

func TestSetToken_ReadToken(t *testing.T) {
	c := New("https://api.example.com", WithToken("initial"))
	if tok := c.readToken(); tok != "initial" {
		t.Errorf("initial token = %q, want %q", tok, "initial")
	}

	c.SetToken("updated")
	if tok := c.readToken(); tok != "updated" {
		t.Errorf("updated token = %q, want %q", tok, "updated")
	}
}

// ── Constructor ─────────────────────────────────────────────────────────────

func TestNew_Defaults(t *testing.T) {
	c := New("https://api.example.com/")
	if c.base != "https://api.example.com" {
		t.Errorf("base = %q, want trailing slash trimmed", c.base)
	}
	if c.version != "dev" {
		t.Errorf("version = %q, want %q", c.version, "dev")
	}
}

// ── buildURL ────────────────────────────────────────────────────────────────

func TestBuildURL_NoQuery(t *testing.T) {
	c := New("https://api.example.com")
	got := c.buildURL("/v1/quote", "")
	if got != "https://api.example.com/v1/quote" {
		t.Errorf("url = %q", got)
	}
}

func TestBuildURL_WithQuery(t *testing.T) {
	c := New("https://api.example.com")
	got := c.buildURL("/v1/quote", "symbol=700.HK")
	if got != "https://api.example.com/v1/quote?symbol=700.HK" {
		t.Errorf("url = %q", got)
	}
}

// ── resolvePathParams ───────────────────────────────────────────────────────

func TestResolvePathParams(t *testing.T) {
	got := resolvePathParams("/v1/orders/{id}/cancel", map[string]string{"id": "12345"})
	if got != "/v1/orders/12345/cancel" {
		t.Errorf("got %q", got)
	}
}

func TestResolvePathParams_SpecialChars(t *testing.T) {
	got := resolvePathParams("/v1/{symbol}", map[string]string{"symbol": "BTC/USD"})
	if got != "/v1/BTC%2FUSD" {
		t.Errorf("got %q, expected URL-encoded slash", got)
	}
}

// ── End-to-end with httptest ────────────────────────────────────────────────

func TestDo_OAuthMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want Bearer token", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer srv.Close()

	c := New(srv.URL, WithToken("test-token"))

	var result map[string]string
	err := c.Do(context.Background(), http.MethodGet, "/v1/test", nil, nil, &result)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("status = %q", result["status"])
	}
}

func TestDo_PostWithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q", ct)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"order_id":"123"}`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	body := map[string]any{"symbol": "700.HK", "qty": 100}
	var result map[string]string
	err := c.Do(context.Background(), http.MethodPost, "/v1/order", nil, body, &result)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if result["order_id"] != "123" {
		t.Errorf("order_id = %q", result["order_id"])
	}
}

type captureReporter struct {
	events []reporting.Event
}

func (r *captureReporter) Report(event reporting.Event) {
	r.events = append(r.events, event)
}

func TestCallReportsRequestEvent(t *testing.T) {
	var reportSourceHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reportSourceHeader = r.Header.Get(headerReportSource)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"items":[]}}`))
	}))
	defer srv.Close()

	capture := &captureReporter{}
	c := New(srv.URL,
		WithReportSource("6"),
		WithReportMetadata("moomoo", "codex"),
		WithRequestReporter(capture),
		WithVersion("v1.2.3"),
	)

	ctx := context.WithValue(context.Background(), CtxKeyCommand, "moomoo quote")
	var result map[string]any
	err := c.Call(ctx, "quote.stock_quote", CallOpts{
		Body: map[string]any{"code_list": []string{"US.AAPL"}},
	}, &result)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if reportSourceHeader != "6" {
		t.Fatalf("%s = %q, want 6", headerReportSource, reportSourceHeader)
	}
	if len(capture.events) != 1 {
		t.Fatalf("captured %d events, want 1", len(capture.events))
	}
	event := capture.events[0]
	if event.Source != 6 || event.AgentName != "codex" || event.Platform != "moomoo" {
		t.Fatalf("unexpected source metadata: %+v", event)
	}
	if event.ToolID != "quote_stock_quote" {
		t.Fatalf("tool id = %q, want quote_stock_quote", event.ToolID)
	}
	if event.APIPathTemplate != "/api/v1.0/quote/stock-quote" {
		t.Fatalf("path template = %q", event.APIPathTemplate)
	}
	if event.TraceID == "" {
		t.Fatal("trace id should not be empty")
	}
	if event.Version != "v1.2.3" {
		t.Fatalf("version = %q, want v1.2.3", event.Version)
	}
}

func TestCallReportsSimTradeRequestEvent(t *testing.T) {
	var reportSourceHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reportSourceHeader = r.Header.Get(headerReportSource)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"order_id":"SIM-OID-1"}}`))
	}))
	defer srv.Close()

	capture := &captureReporter{}
	c := New(srv.URL,
		WithReportSource("8"),
		WithReportMetadata("moomoo", "codex"),
		WithRequestReporter(capture),
		WithVersion("v1.2.3"),
	)

	ctx := context.WithValue(context.Background(), CtxKeyCommand, "moomoo sim order buy")
	err := c.Call(ctx, "sim_trade.input_order", CallOpts{
		PathParams: map[string]string{"acc_id": "SIM-1"},
		Body:       map[string]any{"symbol": "US.AAPL", "qty": "1"},
	}, nil)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if reportSourceHeader != "8" {
		t.Fatalf("%s = %q, want 8", headerReportSource, reportSourceHeader)
	}
	if len(capture.events) != 1 {
		t.Fatalf("captured %d events, want 1", len(capture.events))
	}
	event := capture.events[0]
	if event.Source != 8 {
		t.Fatalf("source = %d, want 8", event.Source)
	}
	if event.Method != http.MethodPost {
		t.Fatalf("method = %q, want POST", event.Method)
	}
	if event.ToolID != "sim_trade_input_order" {
		t.Fatalf("tool id = %q, want sim_trade_input_order", event.ToolID)
	}
	if event.APIPath != "/api/v1.0/sim-trade/SIM-1/orders" {
		t.Fatalf("api path = %q, want /api/v1.0/sim-trade/SIM-1/orders", event.APIPath)
	}
	if event.APIPathTemplate != "/api/v1.0/sim-trade/{acc_id}/orders" {
		t.Fatalf("path template = %q", event.APIPathTemplate)
	}
	if !strings.Contains(event.RequestBody, `"symbol":"US.AAPL"`) {
		t.Fatalf("request body = %q, want symbol", event.RequestBody)
	}
	if !strings.Contains(event.ResponseBody, `"order_id":"SIM-OID-1"`) {
		t.Fatalf("response body = %q, want order id", event.ResponseBody)
	}
}

type capturedReportRequest struct {
	Event reporting.Event `json:"event"`
}

func TestCallReportsThroughAsyncReporter(t *testing.T) {
	received := make(chan capturedReportRequest, 1)
	reportSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer oauth-token" {
			t.Fatalf("Authorization = %q, want Bearer oauth-token", auth)
		}
		var req capturedReportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode report request: %v", err)
		}
		received <- req
		w.WriteHeader(http.StatusOK)
	}))
	defer reportSrv.Close()

	businessSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"order_id":"OID-1","price":"123"}}`))
	}))
	defer businessSrv.Close()

	reporter := reporting.NewReporter(reporting.Options{
		Endpoint: reportSrv.URL,
		Token:    "oauth-token",
	})
	c := New(businessSrv.URL,
		WithToken("oauth-token"),
		WithReportSource("6"),
		WithReportMetadata("moomoo", "codex"),
		WithRequestReporter(reporter),
		WithVersion("v1.2.3"),
	)

	ctx := context.WithValue(context.Background(), CtxKeyCommand, "moomoo order place")
	var result map[string]any
	err := c.Call(ctx, "trade.order.place", CallOpts{
		PathParams: map[string]string{"acc_id": "123"},
		Body:       map[string]any{"symbol": "US.AAPL", "qty": 1},
	}, &result)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}

	closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := reporter.Close(closeCtx); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	select {
	case req := <-received:
		event := req.Event
		if event.Source != 6 || event.AgentName != "codex" || event.Platform != "moomoo" {
			t.Fatalf("unexpected source metadata: %+v", event)
		}
		if event.ToolID != "trading_order_place" {
			t.Fatalf("tool id = %q, want trading_order_place", event.ToolID)
		}
		if event.APIPath != "/api/v1.0/accounts/123/orders" {
			t.Fatalf("api path = %q", event.APIPath)
		}
		if event.APIPathTemplate != "/api/v1.0/accounts/{acc_id}/orders" {
			t.Fatalf("path template = %q", event.APIPathTemplate)
		}
		if event.OrderID != "OID-1" {
			t.Fatalf("order id = %q, want OID-1", event.OrderID)
		}
		if event.TraceID == "" {
			t.Fatal("trace id should not be empty")
		}
	default:
		t.Fatal("expected async report request")
	}
}

// ── marshalBody ─────────────────────────────────────────────────────────────

func TestMarshalBody_Nil(t *testing.T) {
	b, err := marshalBody(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b != nil {
		t.Errorf("expected nil, got %v", b)
	}
}

func TestMarshalBody_Struct(t *testing.T) {
	body := struct {
		Name string `json:"name"`
	}{Name: "test"}
	b, err := marshalBody(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != `{"name":"test"}` {
		t.Errorf("got %q", string(b))
	}
}

// ── commandFromCtx ──────────────────────────────────────────────────────────

func TestCommandFromCtx_Set(t *testing.T) {
	ctx := context.WithValue(context.Background(), CtxKeyCommand, "moomoo quote")
	if cmd := commandFromCtx(ctx); cmd != "moomoo quote" {
		t.Errorf("got %q", cmd)
	}
}

func TestCommandFromCtx_NotSet(t *testing.T) {
	if cmd := commandFromCtx(context.Background()); cmd != "" {
		t.Errorf("got %q, want empty", cmd)
	}
}
