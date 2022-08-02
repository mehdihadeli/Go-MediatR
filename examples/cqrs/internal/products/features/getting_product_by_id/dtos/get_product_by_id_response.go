package dtos

import (
	"github.com/mehdihadeli/mediatr/examples/cqrs/internal/products/dtos"
)

type GetProductByIdQueryResponse struct {
	Product *dtos.ProductDto `json:"product"`
}
