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

	t.Run("get stock", func(t *testing.T) {
		ts := newTestServer(t)
		defer ts.Close()

		resp := mustRequest(t, http.MethodGet, ts.URL+"/stock/A", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		data := decodeStockData(t, resp.Body)
		if got := data.Stock; got != 10 {
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

		data := decodeStockData(t, stockResp.Body)
		if got := data.Stock; got != 8 {
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
		dataA := decodeStockData(t, stockA.Body)
		if got := dataA.Stock; got != 8 {
			t.Fatalf("expected A stock 8, got %d", got)
		}

		stockB := mustRequest(t, http.MethodGet, ts.URL+"/stock/B", nil)
		defer stockB.Body.Close()
		dataB := decodeStockData(t, stockB.Body)
		if got := dataB.Stock; got != 4 {
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
		dataA := decodeStockData(t, stockA.Body)
		if got := dataA.Stock; got != 10 {
			t.Fatalf("expected A stock unchanged at 10, got %d", got)
		}

		stockB := mustRequest(t, http.MethodGet, ts.URL+"/stock/B", nil)
		defer stockB.Body.Close()
		dataB := decodeStockData(t, stockB.Body)
		if got := dataB.Stock; got != 5 {
			t.Fatalf("expected B stock unchanged at 5, got %d", got)
		}
	})
}

func decodeStockData(t *testing.T, r io.Reader) *StockData {
	t.Helper()

	var resp Response
	decodeJSONBody(t, r, &resp)
	if resp.Status != StatusOK {
		t.Fatalf("expected response status ok, got %q", resp.Status)
	}
	if resp.Data == nil {
		t.Fatal("expected stock data in response, got nil")
	}
	return resp.Data
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
