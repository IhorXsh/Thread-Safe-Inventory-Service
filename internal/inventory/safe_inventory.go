package inventory

import (
	"sync"
)

type SafeInventoryService struct {
	mu       sync.RWMutex
	products map[ProductID]*Product
}

func NewSafeInventoryService(products map[ProductID]*Product) *SafeInventoryService {
	return &SafeInventoryService{
		products: products,
	}
}

func (s *SafeInventoryService) GetStock(productID string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product := s.products[ProductID(productID)]
	if product == nil {
		return 0
	}
	return product.GetStock()
}

func (s *SafeInventoryService) Reserve(item ReserveItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item.Quantity == 0 {
		return ErrInvalidQuantity
	}

	product := s.products[ProductID(item.ProductID)]
	if product == nil {
		return ErrProductNotFound
	}

	if product.GetStock() < item.Quantity {
		return ErrInsufficientStock
	}

	product.SetStock(product.GetStock() - item.Quantity)
	return nil
}

func (s *SafeInventoryService) ReserveMultiple(items []ReserveItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
		if item.Quantity == 0 {
			return ErrInvalidQuantity
		}
		product := s.products[item.ProductID]
		if product == nil {
			return ErrProductNotFound
		}
		if product.GetStock() < item.Quantity {
			return ErrInsufficientStock
		}
	}

	for _, item := range items {
		product := s.products[item.ProductID]
		product.SetStock(product.GetStock() - item.Quantity)
	}

	return nil
}
