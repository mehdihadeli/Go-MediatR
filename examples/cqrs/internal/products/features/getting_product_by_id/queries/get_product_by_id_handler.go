package queries

import (
	"context"
	"fmt"
	"github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products"
	gettingProductByIdDtos "github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/features/getting_product_by_id/dtos"
	"github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/repository"

	"github.com/pkg/errors"
)

type GetProductByIdQueryHandler struct {
	productRepository *repository.InMemoryProductRepository
}

func NewGetProductByIdHandler(productRepository *repository.InMemoryProductRepository) *GetProductByIdQueryHandler {
	return &GetProductByIdQueryHandler{productRepository: productRepository}
}

func (q *GetProductByIdQueryHandler) Handle(ctx context.Context, query *GetProductByIdQuery) (*gettingProductByIdDtos.GetProductByIdQueryResponse, error) {
	product, err := q.productRepository.GetProductById(ctx, query.ProductID)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("product with id %s not found", query.ProductID))
	}

	productDto := products.MapProductToProductDto(product)

	return &gettingProductByIdDtos.GetProductByIdQueryResponse{Product: productDto}, nil
}
