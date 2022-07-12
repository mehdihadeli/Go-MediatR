package dtos

import (
	"mediatR/examples/cqrs/internal/products/dtos"
)

type GetProductByIdResponseDto struct {
	Product *dtos.ProductDto `json:"product"`
}
