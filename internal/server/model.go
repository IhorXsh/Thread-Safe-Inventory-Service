package server

type ProductID string

type ReserveRequest struct {
	ProductID ProductID `json:"product_id"`
	Quantity  uint64    `json:"quantity"`
}

type ReserveMultipleRequest struct {
	Items []ReserveRequest `json:"items"`
}

type ResponseStatus string

const (
	StatusOK    ResponseStatus = "ok"
	StatusError ResponseStatus = "error"
)

type Response struct {
	Status ResponseStatus `json:"status"`
	Error  string         `json:"error,omitempty"`
	Data   *StockData     `json:"data,omitempty"`
}

type StockData struct {
	ProductID ProductID `json:"product_id"`
	Stock     uint64    `json:"stock"`
}
