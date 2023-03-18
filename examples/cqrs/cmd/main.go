package main

import (
	"context"
	"github.com/ehsandavari/go-mediator"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ehsandavari/go-mediator/examples/cqrs/docs"
	productApi "github.com/ehsandavari/go-mediator/examples/cqrs/internal/products/api"
	"github.com/ehsandavari/go-mediator/examples/cqrs/internal/products/features/creating_product/commands"
	creatingProductsDtos "github.com/ehsandavari/go-mediator/examples/cqrs/internal/products/features/creating_product/dtos"
	"github.com/ehsandavari/go-mediator/examples/cqrs/internal/products/features/creating_product/events"
	gettingProductByIdDtos "github.com/ehsandavari/go-mediator/examples/cqrs/internal/products/features/getting_product_by_id/dtos"
	"github.com/ehsandavari/go-mediator/examples/cqrs/internal/products/features/getting_product_by_id/queries"
	"github.com/ehsandavari/go-mediator/examples/cqrs/internal/products/repository"
	"github.com/ehsandavari/go-mediator/examples/cqrs/internal/shared/behaviours"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

//swag init --parseDependency --parseInternal --parseDepth 1 -g ./cmd/main.go

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	echo := echo.New()
	productRepository := repository.NewInMemoryProductRepository()

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Register request handlers to the mediator

	createProductCommandHandler := commands.NewCreateProductCommandHandler(productRepository)
	getByIdQueryHandler := queries.NewGetProductByIdHandler(productRepository)

	err := mediator.RegisterRequestHandler[*commands.CreateProductCommand, *creatingProductsDtos.CreateProductCommandResponse](createProductCommandHandler)
	if err != nil {
		log.Fatal(err)
	}

	err = mediator.RegisterRequestHandler[*queries.GetProductByIdQuery, *gettingProductByIdDtos.GetProductByIdQueryResponse](getByIdQueryHandler)
	if err != nil {
		log.Fatal(err)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Register notification handlers to the mediator
	notificationHandler := events.NewProductCreatedEventHandler()
	err = mediator.RegisterNotificationHandler[*events.ProductCreatedEvent](notificationHandler)
	if err != nil {
		log.Fatal(err)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Register request handlers pipeline to the mediator
	loggerPipeline := &behaviours.RequestLoggerBehaviour{}
	err = mediator.RegisterRequestPipelineBehaviors(loggerPipeline)
	if err != nil {
		log.Fatal(err)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Controllers setup
	controller := productApi.NewProductsController(echo)

	productApi.MapProductsRoutes(echo, controller)

	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Title = "Catalogs Write-Service Api"
	docs.SwaggerInfo.Description = "Catalogs Write-Service Api."

	echo.GET("/swagger/*", echoSwagger.WrapHandler)

	go func() {
		if err := echo.Start(":9080"); err != nil {
			log.Fatal("Error starting Server", err)
		}
	}()

	<-ctx.Done()

	if err := echo.Shutdown(ctx); err != nil {
		log.Fatal("(Shutdown) err", err)
	}
}
