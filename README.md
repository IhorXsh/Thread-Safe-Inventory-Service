# Thread-Safe-Inventory-Service

Thread-safe inventory service with HTTP API (`gorilla/mux`), `slog` logging, OpenTelemetry tracing, and Prometheus metrics.

## Structure
- `cmd/inventory-service/main.go` - app bootstrap, tracer init, graceful shutdown.
- `internal/server/server.go` - server wiring, router setup, metrics registration.
- `internal/server/handlers.go` - HTTP handlers for stock and reservation endpoints.
- `internal/server/middleware.go` - observability middleware (tracing + metrics).
- `internal/inventory/types.go` - `Product`, `ReserveItem`, errors, and `Product` getters/setters.
- `internal/inventory/safe_inventory.go` - `SafeInventoryService` with `sync.RWMutex`.
- `internal/inventory/unsafe_inventory.go` - intentionally unsafe service for race demonstration.
- `internal/inventory/inventory_test.go` - concurrency tests (`t.Run`, `errors.Is`).
- `internal/inventory/inventory_edge_test.go` - edge-case tests.
- `REVIEW.md` - race-condition analysis.
- `ANSWERS.md` - conceptual answers.

## Run
```bash
go run ./cmd/inventory-service
```

Optional env var:
- `PORT` (default: `8080`)
- `METRICS_PORT` (default: `9090`)

## HTTP API
- `GET /healthz`
- `GET /stock/{productID}`
- `POST /reserve`
- `POST /reserve-multiple`

Trace ID:
- Request/response logs are correlated with `trace_id` from OpenTelemetry span context.

Response format:
- All handlers return:
```json
{
  "status": "ok|error",
  "error": "optional error message",
  "data": "optional payload"
}
```

Metrics endpoint (separate server):
- `GET /metrics` on `METRICS_PORT` (default `http://localhost:9090/metrics`)

### Request examples
`POST /reserve`
```json
{
  "product_id": "A",
  "quantity": 2
}
```

`POST /reserve-multiple`
```json
{
  "items": [
    {"product_id": "A", "quantity": 2},
    {"product_id": "B", "quantity": 1}
  ]
}
```

## Observability
Prometheus metrics:
- `inventory_http_request_duration_seconds` - request duration histogram.
- `inventory_http_requests_total` - total requests by `method`, `path`, `status`.
- `inventory_http_request_errors_total` - requests with status `>=400`.

Error rate is derived in Prometheus queries, for example:
```promql
sum(rate(inventory_http_request_errors_total[5m])) / sum(rate(inventory_http_requests_total[5m]))
```

Tracing:
- OpenTelemetry spans are created per request in middleware.
- Trace exporter: `stdout` (pretty-print).

## Tests
```bash
go test ./...
```
