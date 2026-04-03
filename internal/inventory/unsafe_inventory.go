package inventory

type UnsafeInventoryService struct {
	products map[string]*Product
}

func NewUnsafeInventoryService(products map[string]*Product) *UnsafeInventoryService {
	return &UnsafeInventoryService{products: products}
}

func (s *UnsafeInventoryService) GetStock(productID string) int {
	product := s.products[productID]
	if product == nil {
		return 0
	}
	return product.Stock
}

func (s *UnsafeInventoryService) Reserve(productID string, quantity int) error {

	product := s.products[productID]
	if product == nil {
		return ErrProductNotFound
	}

	if product.Stock < quantity {
		return ErrInsufficientStock
	}

	product.Stock -= quantity
	return nil
}
