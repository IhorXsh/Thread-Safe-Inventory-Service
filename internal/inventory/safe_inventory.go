package inventory

import (
	"sync"
)

type SafeInventoryService struct {
	mu       sync.RWMutex
	products map[ProductID]*Product
}

func NewSafeInventoryService(products map[ProductID]*Product) *SafeInventoryService {
	return &SafeInventoryService{products: products}
}

func (s *SafeInventoryService) GetStock(productID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product := s.products[ProductID(productID)]
	if product == nil {
		return 0
	}
	return product.GetStock()
}

func (s *SafeInventoryService) Reserve(productID string, quantity int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	product := s.products[ProductID(productID)]
	if product == nil {
		return ErrProductNotFound
	}

	if product.GetStock() < quantity {
		return ErrInsufficientStock
	}

	product.SetStock(product.GetStock() - quantity)
	return nil
}

func (s *SafeInventoryService) ReserveMultiple(items []ReserveItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
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
