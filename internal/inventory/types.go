package inventory

import "errors"

type ProductID string

type Product struct {
	id    ProductID
	name  string
	stock int
}

type ReserveItem struct {
	ProductID ProductID
	Quantity  int
}

func NewProduct(id ProductID, name string, stock int) *Product {
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

func (p *Product) GetStock() int {
	return p.stock
}

func (p *Product) SetStock(stock int) {
	p.stock = stock
}

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)
