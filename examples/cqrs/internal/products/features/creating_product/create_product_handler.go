package creating_product

import (
	"context"
	"mediatR/examples/cqrs/internal/products"
	creating_product_dtos "mediatR/examples/cqrs/internal/products/features/creating_product/dtos"
	"mediatR/examples/cqrs/internal/products/models"
)

type CreateProductHandler struct {
	productRepository *products.InMemoryProductRepository
}

func NewCreateProductHandler(productRepository *products.InMemoryProductRepository) *CreateProductHandler {
	return &CreateProductHandler{productRepository: productRepository}
}

func (c *CreateProductHandler) Handle(ctx context.Context, command *CreateProduct) (*creating_product_dtos.CreateProductResponseDto, error) {

	product := &models.Product{
		ProductID:   command.ProductID,
		Name:        command.Name,
		Description: command.Description,
		Price:       command.Price,
		CreatedAt:   command.CreatedAt,
	}

	createdProduct, err := c.productRepository.CreateProduct(ctx, product)
	if err != nil {
		return nil, err
	}

	response := &creating_product_dtos.CreateProductResponseDto{ProductID: createdProduct.ProductID}

	return response, nil
}
