package inventory

type UnsafeInventoryService struct {
	products map[ProductID]*Product
}

func NewUnsafeInventoryService(products map[ProductID]*Product) *UnsafeInventoryService {
	return &UnsafeInventoryService{
		products: products,
	}
}

func (s *UnsafeInventoryService) GetStock(productID string) uint64 {
	product := s.products[ProductID(productID)]
	if product == nil {
		return 0
	}
	return product.GetStock()
}

func (s *UnsafeInventoryService) Reserve(item ReserveItem) error {
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

	product.SetStock(product.GetStock() - item.Quantity)
	return nil
}
