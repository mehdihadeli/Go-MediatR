package dtos

import (
	"github.com/mehdihadeli/Go-MediatR/examples/cqrs/internal/products/dtos"
)

type GetProductByIdResponseDto struct {
	Product *dtos.ProductDto `json:"product"`
}
