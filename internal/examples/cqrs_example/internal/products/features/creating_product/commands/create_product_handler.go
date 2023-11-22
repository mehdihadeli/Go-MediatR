package commands

import (
	"context"
	"fmt"

	"cqrsexample/internal/products/features/creating_product/dtos"
	"cqrsexample/internal/products/features/creating_product/events"
	"cqrsexample/internal/products/models"
	"cqrsexample/internal/products/repository"
	"github.com/mehdihadeli/go-mediatr"
)

type CreateProductCommandHandler struct {
	productRepository *repository.InMemoryProductRepository
}

func NewCreateProductCommandHandler(productRepository *repository.InMemoryProductRepository) *CreateProductCommandHandler {
	return &CreateProductCommandHandler{productRepository: productRepository}
}

func (c *CreateProductCommandHandler) Handle(ctx context.Context, command *CreateProductCommand) (*dtos.CreateProductCommandResponse, error) {
	isLoggerPipelineEnabled := ctx.Value("logger_pipeline").(bool)
	if isLoggerPipelineEnabled {
		fmt.Println("[CreateProductCommandHandler]: logging pipeline is enabled")
	}

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

	response := &dtos.CreateProductCommandResponse{ProductID: createdProduct.ProductID}

	// Publish notification event to the mediatr for dispatching to the notification handlers

	productCreatedEvent := events.NewProductCreatedEvent(createdProduct.ProductID, createdProduct.Name, createdProduct.Description, createdProduct.Price, createdProduct.CreatedAt)
	err = mediatr.Publish[*events.ProductCreatedEvent](ctx, productCreatedEvent)
	if err != nil {
		return nil, err
	}

	return response, nil
}
