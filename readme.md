# 🚃 Golang MediatR

[![CI](https://github.com/mehdihadeli/Go-MediatR/actions/workflows/ci.yml/badge.svg?branch=main&style=flat-square)](https://github.com/mehdihadeli/Go-MediatR/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mehdihadeli/Go-MediatR)](https://goreportcard.com/report/github.com/mehdihadeli/Go-MediatR)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.18-61CFDD.svg?style=flat-square)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/mehdihadeli/Go-MediatR/blob/main/LICENCE)

> This package is a `Mediator Pattern` implementation in golang, and inspired by great [jbogard/mediatr](https://github.com/jbogard/mediatr) library in .Net.

For decoupling some objects in a system we could use `Mediator` object as an interface, for decrease coupling between the objects. Mostly I uses this pattern when I use CQRS in my system.

There are some samples for using this package [here](./examples).

## Installation

```bash
go get github.com/mehdihadeli/mediatr
```

## Strategies
Mediatr has two strategies for dispatching messages:

1. `Request/Response` messages, dispatched to a `single handler`.
2. `Notification` messages, dispatched to all (multiple) `handlers` and they don't have any response.

## Request/Response Strategy
The `request/response` message, has just `one handler`, and can handle both command and query scenarios in [CQRS Pattern](https://martinfowler.com/bliki/CQRS.html).

### Creating a Request/Response Message

For creating a request (command or query) that has just `one handler`, we could create a command message or query message as a `request` like this:

```go
// Command (Request)
type CreateProductCommand struct {
    ProductID   uuid.UUID `validate:"required"`
    Name        string    `validate:"required,gte=0,lte=255"`
    Description string    `validate:"required,gte=0,lte=5000"`
    Price       float64   `validate:"required,gte=0"`
    CreatedAt   time.Time `validate:"required"`
}

// Query (Request)
type GetProdctByIdQuery struct {
    ProductID uuid.UUID `validate:"required"`
}
```
And for response of these requests, we could create response messages as a `response` like this:

```go
// Command (Response)
type CreateProductCommandResponse struct {
    ProductID uuid.UUID `json:"productId"`
}

// Query (Response)
type GetProdctByIdQueryResponse struct {
    ProductID   uuid.UUID `json:"productId"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Price       float64   `json:"price"`
    CreatedAt   time.Time `json:"createdAt"`
}
```

### Creating Request Handler

For handling our requests, we should create a `single request handler` for each request. Each handler should implement the `RequestHandler` interface. 
```go
type RequestHandler[TRequest any, TResponse any] interface {
	Handle(ctx context.Context, request TRequest) (TResponse, error)
}
```

Here we Create `request handler` (command handler and query handler) for our requests, that implements above interface:

``` go
// Command Handler
type CreateProductCommandHandler struct {
	productRepository *repository.InMemoryProductRepository
}

func NewCreateProductCommandHandler(productRepository *repository.InMemoryProductRepository) *CreateProductCommandHandler {
	return &CreateProductCommandHandler{productRepository: productRepository}
}

func (c *CreateProductCommandHandler) Handle(ctx context.Context, command *CreateProductCommand) (*creatingProductDtos.CreateProductCommandResponse, error) {

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

	response := &creatingProductDtos.CreateProductCommandResponse{ProductID: createdProduct.ProductID}

	return response, nil
}
```

```go
// Query Handler
type GetProductByIdQueryHandler struct {
    productRepository *repository.InMemoryProductRepository
}

func NewGetProductByIdQueryHandler(productRepository *repository.InMemoryProductRepository) *GetProductByIdQueryHandler {
    return &GetProductByIdQueryHandler{productRepository: productRepository}
}

func (c *GetProductByIdQueryHandler) Handle(ctx context.Context, query *GetProductByIdQuery) (*gettingProductDtos.GetProdctByIdQueryResponse, error) {

    product, err := c.productRepository.GetProductById(ctx, query.ProductID)
    if err != nil {
        return nil, err
    }

    response := &gettingProductDtos.GetProdctByIdQueryResponse{
        ProductID:   product.ProductID,
        Name:        product.Name,
        Description: product.Description,
        Price:       product.Price,
        CreatedAt:   product.CreatedAt,
    }

    return response, nil
}
```

> Note: In the cases we don't need a response from our request handler, we can use `Unit` type, that actually is an empty struct:.

### Registering Request Handler to the MediatR
Before `sending` or `dispatching` our requests, we should `register` our request handlers to the MediatR.

Here we register our request handlers (command handler and query handler) to the MediatR:
```go
// Registering `createProductCommandHandler` request handler for `CreateProductCommand` request to the MediatR
mediatr.RegisterHandler[*creatingProduct.CreateProductCommand, *creatingProductsDtos.CreateProductCommandResponse](createProductCommandHandler)

// Registering `getProductByIdQueryHandler` request handler for `GetProductByIdQuery` request to the MediatR
mediatr.RegisterHandler[*gettingProduct.GetProductByIdQuery, *gettingProductDtos.GetProdctByIdQueryResponse](getProductByIdQueryHandler)
```

### Sending Request to the MediatR

Finally, send a message through the mediator.

Here we send our requests to the MediatR for dispatching them to the request handlers (command handler and query handler):
``` go
// Sending `CreateProductCommand` request to mediatr for dispatching to the `CreateProductCommandHandler` request handler
command := &CreateProductCommand{
    ProductID:   uuid.NewV4(),
    Name:        request.name,
    Description: request.description,
    Price:       request.price,
    CreatedAt:   time.Now(),
}

mediatr.Send[*CreateProductCommand, *creatingProductsDtos.CreateProductCommandResponse](ctx, command)
```

```go
// Sending `GetProductByIdQuery` request to mediatr for dispatching to the `GetProductByIdQueryHandler` request handler
query := &GetProdctByIdQuery{
    ProductID:   uuid.NewV4()
}

mediatr.Send[*GetProductByIdQuery, *gettingProductsDtos.GetProductByIdQueryResponse](ctx, query)
```

## Notification Strategy

The `notification` message, can have `multiple handlers` and doesn't have any response, and it can handle an [event notification](https://martinfowler.com/articles/201701-event-driven.html) or notification in event driven architecture.

### Creating a Notification Message

For creating a notification (event), that has multiple `handlers` and doesn't have any response, we could create an event notification as a `notification` like this:

```go
// Event (Notification)
type ProductCreatedEvent struct {
    ProductID uuid.UUID   `json:"productId"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Price       float64   `json:"price"`
    CreatedAt   time.Time `json:"createdAt"`
}
```
This event doesn't have any response.

### Creating Notification Handlers

For handling our notification, we can create `multiple notification handlers` for each notification event. Each handler should implement the `NotificationHandler` interface.
```go
type NotificationHandler[TNotification any] interface {
    Handle(ctx context.Context, notification TNotification) error
}
```

Here we Create multiple `notification event handler` for our notification, that implements above interface:

```go
// Notification Event Handler1
type ProductCreatedEventHandler1 struct {
}

func (c *ProductCreatedEventHandler1) Handle(ctx context.Context, event *ProductCreatedEvent) error {
//Do something with the event here !
    return nil
}
```

```go
// Notification Event Handler2
type ProductCreatedEventHandler2 struct {
}

func (c *ProductCreatedEventHandler2) Handle(ctx context.Context, event *ProductCreatedEvent) error {
//Do something with the event here !
    return nil
}
```

### Registering Notification Handlers to the MediatR
Before `publishing` our notifications, we should `register` our notification handlers to the MediatR.

Here we register our notification handlers to the MediatR:
```go
// Registering `notificationHandler1`, `notificationHandler2` notification handler for `ProductCreatedEvent` notification event to the MediatR
notificationHandler1 := &ProductCreatedEventHandler1{}
notificationHandler2 := &ProductCreatedEventHandler2{}

mediatr.RegisterNotificationHandlers[*events.ProductCreatedEvent](notificationHandler1, notificationHandler2)
```

### Publishing Notification to the MediatR
Finally, publish a notification event through the mediator.

Here we publish our notification to the MediatR for dispatching them to the notification handlers:
``` go
// Publishing `ProductCreatedEvent` notification to mediatr for dispatching to the `ProductCreatedEventHandler1`, `ProductCreatedEventHandler2` notification handlers
productCreatedEvent := 	&ProductCreatedEvent {
    ProductID:   createdProduct.ProductID,
    Name:        createdProduct.Name,
    Price:       createdProduct.Price,
    CreatedAt:   createdProduct.CreatedAt,
    Description: createdProduct.Description,
}
	
mediatr.Publish[*events.ProductCreatedEvent](ctx, productCreatedEvent)
```

## Using Pipeline Behaviors
TODO
