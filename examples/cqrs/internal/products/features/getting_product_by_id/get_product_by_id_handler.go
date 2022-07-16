package getting_product_by_id

import (
	"context"
	"fmt"
	"mediatr/examples/cqrs/internal/products"
	getting_product_by_id_dtos "mediatr/examples/cqrs/internal/products/features/getting_product_by_id/dtos"
	"mediatr/examples/cqrs/internal/products/repository"

	"github.com/pkg/errors"
)

type GetProductByIdQueryHandler struct {
	productRepository *repository.InMemoryProductRepository
}

func NewGetProductByIdHandler(productRepository *repository.InMemoryProductRepository) *GetProductByIdQueryHandler {
	return &GetProductByIdQueryHandler{productRepository: productRepository}
}

func (q *GetProductByIdQueryHandler) Handle(ctx context.Context, query *GetProductByIdQuery) (*getting_product_by_id_dtos.GetProductByIdResponseDto, error) {
	product, err := q.productRepository.GetProductById(ctx, query.ProductID)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("product with id %s not found", query.ProductID))
	}

	productDto := products.MapProductToProductDto(product)

	return &getting_product_by_id_dtos.GetProductByIdResponseDto{Product: productDto}, nil
}
