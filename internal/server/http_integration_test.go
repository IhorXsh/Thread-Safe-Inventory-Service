package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IhorXsh/Thread-Safe-Inventory-Service/internal/inventory"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

func TestHTTPEndpoints(t *testing.T) {
	t.Run("healthz", func(t *testing.T) {
		ts := newTestServer(t)
		defer ts.Close()

		resp := mustRequest(t, http.MethodGet, ts.URL+"/healthz", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var out Response
		decodeJSONBody(t, resp.Body, &out)
		if out.Status != StatusOK {
			t.Fatalf("expected status ok in body, got %q", out.Status)
		}
	})

	t.Run("request id generated when missing", func(t *testing.T) {
		ts := newTestServer(t)
		defer ts.Close()

		resp := mustRequest(t, http.MethodGet, ts.URL+"/healthz", nil)
		defer resp.Body.Close()

		requestID := resp.Header.Get(requestIDHeader)
		if requestID == "" {
			t.Fatalf("expected %s header in response", requestIDHeader)
		}
		if _, err := uuid.Parse(requestID); err != nil {
			t.Fatalf("expected valid uuid request id, got %q", requestID)
		}
	})

	t.Run("request id from header is preserved", func(t *testing.T) {
		ts := newTestServer(t)
		defer ts.Close()

		incoming := uuid.NewString()
		req, err := http.NewRequest(http.MethodGet, ts.URL+"/healthz", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set(requestIDHeader, incoming)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to execute request: %v", err)
		}
		defer resp.Body.Close()

		if got := resp.Header.Get(requestIDHeader); got != incoming {
			t.Fatalf("expected response request id %q, got %q", incoming, got)
		}
	})

	t.Run("get stock", func(t *testing.T) {
		ts := newTestServer(t)
		defer ts.Close()

		resp := mustRequest(t, http.MethodGet, ts.URL+"/stock/A", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var body map[string]any
		decodeResponseData(t, resp.Body, &body)
		if got := asUint64(t, body["stock"]); got != 10 {
			t.Fatalf("expected stock 10, got %d", got)
		}
	})

	t.Run("reserve success", func(t *testing.T) {
		ts := newTestServer(t)
		defer ts.Close()

		payload := map[string]any{"product_id": "A", "quantity": 2}
		resp := mustJSONRequest(t, http.MethodPost, ts.URL+"/reserve", payload)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
		var reserveOut Response
		decodeJSONBody(t, resp.Body, &reserveOut)
		if reserveOut.Status != StatusOK {
			t.Fatalf("expected reserve status ok, got %q", reserveOut.Status)
		}

		stockResp := mustRequest(t, http.MethodGet, ts.URL+"/stock/A", nil)
		defer stockResp.Body.Close()

		var body map[string]any
		decodeResponseData(t, stockResp.Body, &body)
		if got := asUint64(t, body["stock"]); got != 8 {
			t.Fatalf("expected stock 8 after reserve, got %d", got)
		}
	})

	t.Run("reserve insufficient stock", func(t *testing.T) {
		ts := newTestServer(t)
		defer ts.Close()

		payload := map[string]any{"product_id": "B", "quantity": 100}
		resp := mustJSONRequest(t, http.MethodPost, ts.URL+"/reserve", payload)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", resp.StatusCode)
		}
		var out Response
		decodeJSONBody(t, resp.Body, &out)
		if out.Status != StatusError {
			t.Fatalf("expected status error, got %q", out.Status)
		}
	})

	t.Run("reserve multiple success", func(t *testing.T) {
		ts := newTestServer(t)
		defer ts.Close()

		payload := map[string]any{
			"items": []map[string]any{
				{"product_id": "A", "quantity": 2},
				{"product_id": "B", "quantity": 1},
			},
		}
		resp := mustJSONRequest(t, http.MethodPost, ts.URL+"/reserve-multiple", payload)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
		var out Response
		decodeJSONBody(t, resp.Body, &out)
		if out.Status != StatusOK {
			t.Fatalf("expected status ok, got %q", out.Status)
		}

		stockA := mustRequest(t, http.MethodGet, ts.URL+"/stock/A", nil)
		defer stockA.Body.Close()
		var bodyA map[string]any
		decodeResponseData(t, stockA.Body, &bodyA)
		if got := asUint64(t, bodyA["stock"]); got != 8 {
			t.Fatalf("expected A stock 8, got %d", got)
		}

		stockB := mustRequest(t, http.MethodGet, ts.URL+"/stock/B", nil)
		defer stockB.Body.Close()
		var bodyB map[string]any
		decodeResponseData(t, stockB.Body, &bodyB)
		if got := asUint64(t, bodyB["stock"]); got != 4 {
			t.Fatalf("expected B stock 4, got %d", got)
		}
	})

	t.Run("reserve multiple atomicity", func(t *testing.T) {
		ts := newTestServer(t)
		defer ts.Close()

		payload := map[string]any{
			"items": []map[string]any{
				{"product_id": "A", "quantity": 8},
				{"product_id": "B", "quantity": 8},
			},
		}
		resp := mustJSONRequest(t, http.MethodPost, ts.URL+"/reserve-multiple", payload)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", resp.StatusCode)
		}
		var out Response
		decodeJSONBody(t, resp.Body, &out)
		if out.Status != StatusError {
			t.Fatalf("expected status error, got %q", out.Status)
		}

		stockA := mustRequest(t, http.MethodGet, ts.URL+"/stock/A", nil)
		defer stockA.Body.Close()
		var bodyA map[string]any
		decodeResponseData(t, stockA.Body, &bodyA)
		if got := asUint64(t, bodyA["stock"]); got != 10 {
			t.Fatalf("expected A stock unchanged at 10, got %d", got)
		}

		stockB := mustRequest(t, http.MethodGet, ts.URL+"/stock/B", nil)
		defer stockB.Body.Close()
		var bodyB map[string]any
		decodeResponseData(t, stockB.Body, &bodyB)
		if got := asUint64(t, bodyB["stock"]); got != 5 {
			t.Fatalf("expected B stock unchanged at 5, got %d", got)
		}
	})
}

func decodeResponseData(t *testing.T, r io.Reader, out any) {
	t.Helper()

	var resp Response
	decodeJSONBody(t, r, &resp)
	if resp.Status != StatusOK {
		t.Fatalf("expected response status ok, got %q", resp.Status)
	}

	dataMap, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("unexpected response data type: %T", resp.Data)
	}
	b, err := json.Marshal(dataMap)
	if err != nil {
		t.Fatalf("failed to marshal response data: %v", err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatalf("failed to unmarshal response data: %v", err)
	}
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := inventory.NewSafeInventoryService(map[inventory.ProductID]*inventory.Product{
		"A": inventory.NewProduct("A", "Widget A", 10),
		"B": inventory.NewProduct("B", "Widget B", 5),
	})

	s := New(svc, logger, otel.Tracer("test-http"))
	return httptest.NewServer(s.Handler())
}

func mustRequest(t *testing.T, method, url string, body io.Reader) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}
	return resp
}

func mustJSONRequest(t *testing.T, method, url string, payload any) *http.Response {
	t.Helper()

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}
	return resp
}

func decodeJSONBody(t *testing.T, r io.Reader, out any) {
	t.Helper()

	if err := json.NewDecoder(r).Decode(out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func asUint64(t *testing.T, v any) uint64 {
	t.Helper()

	f, ok := v.(float64)
	if !ok {
		t.Fatalf("unexpected type for numeric field: %T", v)
	}
	return uint64(f)
}
