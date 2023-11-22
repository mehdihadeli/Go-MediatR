package queries

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"cqrsexample/internal/products"
	"cqrsexample/internal/products/features/getting_product_by_id/dtos"
	"cqrsexample/internal/products/repository"
)

type GetProductByIdQueryHandler struct {
	productRepository *repository.InMemoryProductRepository
}

func NewGetProductByIdHandler(productRepository *repository.InMemoryProductRepository) *GetProductByIdQueryHandler {
	return &GetProductByIdQueryHandler{productRepository: productRepository}
}

func (q *GetProductByIdQueryHandler) Handle(ctx context.Context, query *GetProductByIdQuery) (*dtos.GetProductByIdQueryResponse, error) {
	product, err := q.productRepository.GetProductById(ctx, query.ProductID)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("product with id %s not found", query.ProductID))
	}

	productDto := products.MapProductToProductDto(product)

	return &dtos.GetProductByIdQueryResponse{Product: productDto}, nil
}
