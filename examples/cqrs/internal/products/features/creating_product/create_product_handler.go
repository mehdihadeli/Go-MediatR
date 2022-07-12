package creating_product

import (
	"context"
	creating_product_dtos "mediatR/examples/cqrs/internal/products/features/creating_product/dtos"
	"mediatR/examples/cqrs/internal/products/models"
	"mediatR/examples/cqrs/internal/products/repository"
)

type CreateProductCommandHandler struct {
	productRepository *repository.InMemoryProductRepository
}

func NewCreateProductCommandHandler(productRepository *repository.InMemoryProductRepository) *CreateProductCommandHandler {
	return &CreateProductCommandHandler{productRepository: productRepository}
}

func (c *CreateProductCommandHandler) Handle(ctx context.Context, command *CreateProductCommand) (*creating_product_dtos.CreateProductResponseDto, error) {

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
