package inventory

type UnsafeInventoryService struct {
	products map[ProductID]*Product
}

func NewUnsafeInventoryService(products map[ProductID]*Product) *UnsafeInventoryService {
	return &UnsafeInventoryService{products: products}
}

func (s *UnsafeInventoryService) GetStock(productID string) uint64 {
	product := s.products[ProductID(productID)]
	if product == nil {
		return 0
	}
	return product.GetStock()
}

func (s *UnsafeInventoryService) Reserve(productID string, quantity uint64) error {

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
