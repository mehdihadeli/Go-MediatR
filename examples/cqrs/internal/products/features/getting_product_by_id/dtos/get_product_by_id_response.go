package dtos

import (
	"github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/dtos"
)

type GetProductByIdQueryResponse struct {
	Product *dtos.ProductDto `json:"product"`
}
