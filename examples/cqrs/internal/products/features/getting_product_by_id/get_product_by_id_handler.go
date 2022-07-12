package getting_product_by_id

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	products2 "mediatR/examples/cqrs/internal/products"
	dtos2 "mediatR/examples/cqrs/internal/products/features/getting_product_by_id/dtos"
)

type GetProductByIdHandler struct {
	productRepository *products2.InMemoryProductRepository
}

func NewGetProductByIdHandler(productRepository *products2.InMemoryProductRepository) *GetProductByIdHandler {
	return &GetProductByIdHandler{productRepository: productRepository}
}

func (q *GetProductByIdHandler) Handle(ctx context.Context, query *GetProductById) (*dtos2.GetProductByIdResponseDto, error) {
	product, err := q.productRepository.GetProductById(ctx, query.ProductID)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("product with id %s not found", query.ProductID))
	}

	productDto := products2.MapProductToProductDto(product)

	return &dtos2.GetProductByIdResponseDto{Product: productDto}, nil
}
