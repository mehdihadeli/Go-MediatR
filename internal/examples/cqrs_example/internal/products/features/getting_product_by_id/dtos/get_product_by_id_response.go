package dtos

import "cqrsexample/internal/products/dtos"

type GetProductByIdQueryResponse struct {
	Product *dtos.ProductDto `json:"product"`
}
