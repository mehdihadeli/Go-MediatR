package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cqrsexample/docs"
	"cqrsexample/internal/products/api"
	"cqrsexample/internal/products/features/creating_product/commands"
	"cqrsexample/internal/products/features/creating_product/dtos"
	"cqrsexample/internal/products/features/creating_product/events"
	dtos2 "cqrsexample/internal/products/features/getting_product_by_id/dtos"
	"cqrsexample/internal/products/features/getting_product_by_id/queries"
	"cqrsexample/internal/products/repository"
	"cqrsexample/internal/shared/behaviours"
	"github.com/mehdihadeli/go-mediatr"

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

	err := mediatr.RegisterRequestHandler[*commands.CreateProductCommand, *dtos.CreateProductCommandResponse](createProductCommandHandler)
	if err != nil {
		log.Fatal(err)
	}

	err = mediatr.RegisterRequestHandler[*queries.GetProductByIdQuery, *dtos2.GetProductByIdQueryResponse](getByIdQueryHandler)
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
	controller := api.NewProductsController(echo)

	api.MapProductsRoutes(echo, controller)

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
