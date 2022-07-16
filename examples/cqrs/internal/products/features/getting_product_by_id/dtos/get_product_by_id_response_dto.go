package dtos

import (
	"mediatr/examples/cqrs/internal/products/dtos"
)

type GetProductByIdResponseDto struct {
	Product *dtos.ProductDto `json:"product"`
}
