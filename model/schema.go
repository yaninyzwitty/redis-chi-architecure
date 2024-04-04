package model

import "github.com/google/uuid"

// product_id UUID PRIMARY KEY,
// name TEXT,
// description TEXT,
// price DECIMAL,
// stock_quantity INT

type Product struct {
	ProductId      uuid.UUID `json:"product_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Price          float64   `json:"price"`
	Stock_quantity int       `json:"stock_quantity"`
}
