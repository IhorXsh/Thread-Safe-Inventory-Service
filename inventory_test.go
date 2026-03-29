package inventory

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestReserve_ConcurrentOversell(t *testing.T) {
	svc := NewSafeInventoryService(map[string]*Product{
		"p1": {ID: "p1", Name: "Widget", Stock: 100},
	})

	const goroutines = 200
	var wg sync.WaitGroup
	wg.Add(goroutines)

	var success int32
	var failed int32

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			err := svc.Reserve("p1", 1)
			if err == nil {
				atomic.AddInt32(&success, 1)
				return
			}
			if err == ErrInsufficientStock {
				atomic.AddInt32(&failed, 1)
				return
			}
			t.Errorf("unexpected error: %v", err)
		}()
	}

	wg.Wait()

	if success != 100 {
		t.Fatalf("expected 100 successes, got %d", success)
	}
	if failed != 100 {
		t.Fatalf("expected 100 failures, got %d", failed)
	}
	if got := svc.GetStock("p1"); got != 0 {
		t.Fatalf("expected stock to be 0, got %d", got)
	}
}

func TestReserveMultiple_Atomicity(t *testing.T) {
	svc := NewSafeInventoryService(map[string]*Product{
		"A": {ID: "A", Name: "A", Stock: 10},
		"B": {ID: "B", Name: "B", Stock: 5},
	})

	err := svc.ReserveMultiple([]ReserveItem{
		{ProductID: "A", Quantity: 8},
		{ProductID: "B", Quantity: 8},
	})
	if err != ErrInsufficientStock {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}

	if got := svc.GetStock("A"); got != 10 {
		t.Fatalf("expected A stock to remain 10, got %d", got)
	}
	if got := svc.GetStock("B"); got != 5 {
		t.Fatalf("expected B stock to remain 5, got %d", got)
	}
}
