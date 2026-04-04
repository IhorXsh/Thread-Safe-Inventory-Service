# Thread-Safe-Inventory-Service

Thread-safe inventory management service with atomic stock reservations, a simple CLI, and concurrent tests.

## Structure
- `cmd/inventory-service/main.go` – CLI entrypoint.
- `internal/inventory/inventory.go` – `SafeInventoryService` using `sync.RWMutex`.
- `internal/inventory/inventory_test.go` – concurrency tests for oversell protection and atomic multi-reserve.
- `REVIEW.md` – race condition analysis of the original buggy code.
- `ANSWERS.md` – answers to the conceptual questions.

## Run tests
```bash
go test -race ./...
```

## Run CLI
```bash
go run ./cmd/inventory-service get A
```
