package server

type ProductID string

type ReserveRequest struct {
	ProductID ProductID `json:"product_id"`
	Quantity  uint64    `json:"quantity"`
}

type ReserveMultipleRequest struct {
	Items []ReserveRequest `json:"items"`
}
