package dtos

import (
	"github.com/ehsandavari/go-mediator/examples/cqrs/internal/products/dtos"
)

type GetProductByIdQueryResponse struct {
	Product *dtos.ProductDto `json:"product"`
}
