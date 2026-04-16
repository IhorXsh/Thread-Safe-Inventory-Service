package inventory

import "errors"

type ProductID string

type Product struct {
	id    ProductID
	name  string
	stock uint64
}

type ReserveItem struct {
	ProductID ProductID
	Quantity  uint64
}

func NewProduct(id ProductID, name string, stock uint64) *Product {
	return &Product{
		id:    id,
		name:  name,
		stock: stock,
	}
}

func (p *Product) GetID() ProductID {
	return p.id
}

func (p *Product) SetID(id ProductID) {
	p.id = id
}

func (p *Product) GetName() string {
	return p.name
}

func (p *Product) SetName(name string) {
	p.name = name
}

func (p *Product) GetStock() uint64 {
	return p.stock
}

func (p *Product) SetStock(stock uint64) {
	p.stock = stock
}

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidQuantity   = errors.New("invalid quantity")
)
