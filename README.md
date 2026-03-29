# Inventory Service (Thread-Safe)

Thread-safe inventory management service with atomic stock reservations and concurrent tests.

## Contents
- `inventory.go` – `SafeInventoryService` using `sync.RWMutex`.
- `inventory_test.go` – concurrency tests for oversell protection and atomic multi-reserve.
- `REVIEW.md` – race condition analysis of the original buggy code.
- `ANSWERS.md` – answers to the conceptual questions.

## Run tests
```bash
go test -race
```
