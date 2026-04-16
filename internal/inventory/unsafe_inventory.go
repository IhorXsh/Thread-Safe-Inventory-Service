package inventory

import "log/slog"

type UnsafeInventoryService struct {
	products map[ProductID]*Product
	logger   *slog.Logger
}

func NewUnsafeInventoryService(products map[ProductID]*Product) *UnsafeInventoryService {
	return NewUnsafeInventoryServiceWithLogger(products, slog.Default())
}

func NewUnsafeInventoryServiceWithLogger(products map[ProductID]*Product, logger *slog.Logger) *UnsafeInventoryService {
	if logger == nil {
		logger = slog.Default()
	}
	return &UnsafeInventoryService{
		products: products,
		logger:   logger,
	}
}

func (s *UnsafeInventoryService) GetStock(productID string) uint64 {
	product := s.products[ProductID(productID)]
	if product == nil {
		s.logger.Warn("unsafe get stock failed: product not found", "product_id", productID)
		return 0
	}
	stock := product.GetStock()
	s.logger.Debug("unsafe get stock", "product_id", productID, "stock", stock)
	return stock
}

func (s *UnsafeInventoryService) Reserve(productID string, quantity uint64) error {
	if quantity == 0 {
		s.logger.Warn("unsafe reserve failed: invalid quantity", "product_id", productID, "quantity", quantity)
		return ErrInvalidQuantity
	}

	product := s.products[ProductID(productID)]
	if product == nil {
		s.logger.Warn("unsafe reserve failed: product not found", "product_id", productID, "quantity", quantity)
		return ErrProductNotFound
	}

	if product.GetStock() < quantity {
		s.logger.Warn("unsafe reserve failed: insufficient stock", "product_id", productID, "quantity", quantity, "stock", product.GetStock())
		return ErrInsufficientStock
	}

	product.SetStock(product.GetStock() - quantity)
	s.logger.Info("unsafe reserve successful", "product_id", productID, "quantity", quantity, "stock", product.GetStock())
	return nil
}
