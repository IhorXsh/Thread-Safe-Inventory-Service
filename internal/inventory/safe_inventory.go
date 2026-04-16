package inventory

import (
	"log/slog"
	"sync"
)

type SafeInventoryService struct {
	mu       sync.RWMutex
	products map[ProductID]*Product
	logger   *slog.Logger
}

func NewSafeInventoryService(products map[ProductID]*Product) *SafeInventoryService {
	return NewSafeInventoryServiceWithLogger(products, slog.Default())
}

func NewSafeInventoryServiceWithLogger(products map[ProductID]*Product, logger *slog.Logger) *SafeInventoryService {
	if logger == nil {
		logger = slog.Default()
	}
	return &SafeInventoryService{
		products: products,
		logger:   logger,
	}
}

func (s *SafeInventoryService) GetStock(productID string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product := s.products[ProductID(productID)]
	if product == nil {
		s.logger.Warn("get stock failed: product not found", "product_id", productID)
		return 0
	}
	stock := product.GetStock()
	s.logger.Debug("get stock", "product_id", productID, "stock", stock)
	return stock
}

func (s *SafeInventoryService) Reserve(productID string, quantity uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if quantity == 0 {
		s.logger.Warn("reserve failed: invalid quantity", "product_id", productID, "quantity", quantity)
		return ErrInvalidQuantity
	}

	product := s.products[ProductID(productID)]
	if product == nil {
		s.logger.Warn("reserve failed: product not found", "product_id", productID, "quantity", quantity)
		return ErrProductNotFound
	}

	if product.GetStock() < quantity {
		s.logger.Warn("reserve failed: insufficient stock", "product_id", productID, "quantity", quantity, "stock", product.GetStock())
		return ErrInsufficientStock
	}

	product.SetStock(product.GetStock() - quantity)
	s.logger.Info("reserve successful", "product_id", productID, "quantity", quantity, "stock", product.GetStock())
	return nil
}

func (s *SafeInventoryService) ReserveMultiple(items []ReserveItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
		if item.Quantity == 0 {
			s.logger.Warn("reserve multiple failed: invalid quantity", "product_id", item.ProductID, "quantity", item.Quantity)
			return ErrInvalidQuantity
		}
		product := s.products[item.ProductID]
		if product == nil {
			s.logger.Warn("reserve multiple failed: product not found", "product_id", item.ProductID)
			return ErrProductNotFound
		}
		if product.GetStock() < item.Quantity {
			s.logger.Warn("reserve multiple failed: insufficient stock", "product_id", item.ProductID, "quantity", item.Quantity, "stock", product.GetStock())
			return ErrInsufficientStock
		}
	}

	for _, item := range items {
		product := s.products[item.ProductID]
		product.SetStock(product.GetStock() - item.Quantity)
	}
	s.logger.Info("reserve multiple successful", "items_count", len(items))

	return nil
}
