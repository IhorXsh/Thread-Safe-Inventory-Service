package inventory

import (
	"testing"
)

func TestSafeInventory_EdgeCases(t *testing.T) {
	t.Run("reserve with empty product id", func(t *testing.T) {
		svc := NewSafeInventoryService(map[ProductID]*Product{
			"A": NewProduct("A", "A", 10),
		})

		err := svc.Reserve(ReserveItem{ProductID: "", Quantity: 1})
		if !IsProductNotFound(err) {
			t.Fatalf("expected product not found error, got %v", err)
		}
	})

	t.Run("reserve with zero quantity", func(t *testing.T) {
		svc := NewSafeInventoryService(map[ProductID]*Product{
			"A": NewProduct("A", "A", 10),
		})

		err := svc.Reserve(ReserveItem{ProductID: "A", Quantity: 0})
		if !IsInvalidQuantity(err) {
			t.Fatalf("expected invalid quantity error, got %v", err)
		}
	})

	t.Run("reserve multiple with empty product id", func(t *testing.T) {
		svc := NewSafeInventoryService(map[ProductID]*Product{
			"A": NewProduct("A", "A", 10),
		})

		err := svc.ReserveMultiple([]ReserveItem{
			{ProductID: "", Quantity: 1},
		})
		if !IsProductNotFound(err) {
			t.Fatalf("expected product not found error, got %v", err)
		}
	})

	t.Run("reserve multiple with zero quantity", func(t *testing.T) {
		svc := NewSafeInventoryService(map[ProductID]*Product{
			"A": NewProduct("A", "A", 10),
		})

		err := svc.ReserveMultiple([]ReserveItem{
			{ProductID: "A", Quantity: 0},
		})
		if !IsInvalidQuantity(err) {
			t.Fatalf("expected invalid quantity error, got %v", err)
		}
	})

	t.Run("get stock for missing product", func(t *testing.T) {
		svc := NewSafeInventoryService(map[ProductID]*Product{
			"A": NewProduct("A", "A", 10),
		})

		got := svc.GetStock("missing")
		if got != 0 {
			t.Fatalf("expected stock 0 for missing product, got %d", got)
		}
	})
}
