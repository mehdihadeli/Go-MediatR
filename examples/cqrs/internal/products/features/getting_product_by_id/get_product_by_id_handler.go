package getting_product_by_id

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"mediatR/examples/cqrs/internal/products"
	getting_product_by_id_dtos "mediatR/examples/cqrs/internal/products/features/getting_product_by_id/dtos"
	"mediatR/examples/cqrs/internal/products/repository"
)

type GetProductByIdHandler struct {
	productRepository *repository.InMemoryProductRepository
}

func NewGetProductByIdHandler(productRepository *repository.InMemoryProductRepository) *GetProductByIdHandler {
	return &GetProductByIdHandler{productRepository: productRepository}
}

func (q *GetProductByIdHandler) Handle(ctx context.Context, query *GetProductById) (*getting_product_by_id_dtos.GetProductByIdResponseDto, error) {
	product, err := q.productRepository.GetProductById(ctx, query.ProductID)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("product with id %s not found", query.ProductID))
	}

	productDto := products.MapProductToProductDto(product)

	return &getting_product_by_id_dtos.GetProductByIdResponseDto{Product: productDto}, nil
}
