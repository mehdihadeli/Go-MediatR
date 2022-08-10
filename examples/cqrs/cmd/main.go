package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mehdihadeli/go-mediatr"
	"github.com/mehdihadeli/go-mediatr/examples/cqrs/docs"
	productApi "github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/api"
	"github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/features/creating_product/commands"
	creatingProductsDtos "github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/features/creating_product/dtos"
	"github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/features/creating_product/events"
	gettingProductByIdDtos "github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/features/getting_product_by_id/dtos"
	"github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/features/getting_product_by_id/queries"
	"github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/repository"
	"github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/shared/behaviours"

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
	// Register request handlers to the mediatr

	createProductCommandHandler := commands.NewCreateProductCommandHandler(productRepository)
	getByIdQueryHandler := queries.NewGetProductByIdHandler(productRepository)

	err := mediatr.RegisterRequestHandler[*commands.CreateProductCommand, *creatingProductsDtos.CreateProductCommandResponse](createProductCommandHandler)
	if err != nil {
		log.Fatal(err)
	}

	err = mediatr.RegisterRequestHandler[*queries.GetProductByIdQuery, *gettingProductByIdDtos.GetProductByIdQueryResponse](getByIdQueryHandler)
	if err != nil {
		log.Fatal(err)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Register notification handlers to the mediatr
	notificationHandler := events.NewProductCreatedEventHandler()
	err = mediatr.RegisterNotificationHandler[*events.ProductCreatedEvent](notificationHandler)
	if err != nil {
		log.Fatal(err)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Register request handlers pipeline to the mediatr
	loggerPipeline := &behaviours.RequestLoggerBehaviour{}
	err = mediatr.RegisterRequestPipelineBehaviors(loggerPipeline)
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
