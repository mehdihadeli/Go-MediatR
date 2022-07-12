# Golang MediatR
This package is a `Mediator Pattern` implementation in golang, and inspired by great [jbogard/MediatR](https://github.com/jbogard/MediatR) library in .Net.

For decoupling some objects in a system we could use in `Mediator` class for decrease coupling between the classes. Mostly I uses this pattern when I use CQRS in my system.

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

// Registering Command Handler to the mediatR

mediatR.RegisterHandler[*creating_product.CreateProduct, *creating_products_dtos.CreateProductResponseDto](createProductCommandHandler)

```

## Sending Request

``` go
// Sending command to mediatR for routing to the corresponding command handler

command := creating_product.NewCreateProductCommand(request.Name, request.Description, request.Price)
mediatR.Send[*creating_products_dtos.CreateProductResponseDto](ctx.Request().Context(), command)
```

## Using Pipeline Behaviors
TODO