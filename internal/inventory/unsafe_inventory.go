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

func (s *UnsafeInventoryService) Reserve(item ReserveItem) error {
	if item.Quantity == 0 {
		s.logger.Warn("unsafe reserve failed: invalid quantity", "product_id", item.ProductID, "quantity", item.Quantity)
		return ErrInvalidQuantity
	}

	product := s.products[item.ProductID]
	if product == nil {
		s.logger.Warn("unsafe reserve failed: product not found", "product_id", item.ProductID, "quantity", item.Quantity)
		return ErrProductNotFound
	}

	if product.GetStock() < item.Quantity {
		s.logger.Warn("unsafe reserve failed: insufficient stock", "product_id", item.ProductID, "quantity", item.Quantity, "stock", product.GetStock())
		return ErrInsufficientStock
	}

	product.SetStock(product.GetStock() - item.Quantity)
	s.logger.Info("unsafe reserve successful", "product_id", item.ProductID, "quantity", item.Quantity, "stock", product.GetStock())
	return nil
}
