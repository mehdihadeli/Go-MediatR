# 🚃 Golang MediatR

[![CI](https://github.com/mehdihadeli/Go-MediatR/actions/workflows/ci.yml/badge.svg?branch=main&style=flat-square)](https://github.com/mehdihadeli/Go-MediatR/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mehdihadeli/Go-MediatR)](https://goreportcard.com/report/github.com/mehdihadeli/Go-MediatR)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.18-61CFDD.svg?style=flat-square)
[![](https://godoc.org/github.com/mehdihadeli/Go-MediatR?status.svg)](https://pkg.go.dev/github.com/mehdihadeli/Go-MediatR)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/mehdihadeli/Go-MediatR/blob/main/LICENCE)

This package is a `Mediator Pattern` implementation in golang, and inspired by great [jbogard/mediatr](https://github.com/jbogard/mediatr) library in .Net.

For decoupling some objects in a system we could use `Mediator` object as an interface, for decrease coupling between the objects. Mostly I uses this pattern when I use CQRS in my system.

## Installation

```bash
go get github.com/mehdihadeli/mediatr
```

## Registering Handlers 

``` go

// Command

type CreateProductCommand struct {
	ProductID   uuid.UUID `validate:"required"`
	Name        string    `validate:"required,gte=0,lte=255"`
	Description string    `validate:"required,gte=0,lte=5000"`
	Price       float64   `validate:"required,gte=0"`
	CreatedAt   time.Time `validate:"required"`
}

// Command Handler

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

// Registering Command Handler to the mediatr

mediatr.RegisterHandler[*creating_product.CreateProduct, *creating_products_dtos.CreateProductResponseDto](createProductCommandHandler)

```

## Sending Request

``` go
// Sending command to mediatr for routing to the corresponding command handler

command := creating_product.NewCreateProductCommand(request.Name, request.Description, request.Price)
mediatr.Send[*creating_products_dtos.CreateProductResponseDto](ctx.Request().Context(), command)
```

## Using Pipeline Behaviors
TODO
