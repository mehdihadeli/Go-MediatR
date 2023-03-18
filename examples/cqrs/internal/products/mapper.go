package products

import (
	"github.com/ehsandavari/go-mediator/examples/cqrs/internal/products/dtos"
	"github.com/ehsandavari/go-mediator/examples/cqrs/internal/products/models"
)

func MapProductToProductDto(product *models.Product) *dtos.ProductDto {
	return &dtos.ProductDto{
		ProductID:   product.ProductID,
		Name:        product.Name,
		Description: product.Description,
		UpdatedAt:   product.UpdatedAt,
		CreatedAt:   product.CreatedAt,
		Price:       product.Price,
	}
}
