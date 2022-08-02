package main

import (
	"context"
	"github.com/mehdihadeli/mediatr"
	"github.com/mehdihadeli/mediatr/examples/cqrs/docs"
	productApi "github.com/mehdihadeli/mediatr/examples/cqrs/internal/products/api"
	"github.com/mehdihadeli/mediatr/examples/cqrs/internal/products/features/creating_product"
	creatingProductsDtos "github.com/mehdihadeli/mediatr/examples/cqrs/internal/products/features/creating_product/dtos"
	"github.com/mehdihadeli/mediatr/examples/cqrs/internal/products/features/getting_product_by_id"
	gettingProductByIdDtos "github.com/mehdihadeli/mediatr/examples/cqrs/internal/products/features/getting_product_by_id/dtos"
	"github.com/mehdihadeli/mediatr/examples/cqrs/internal/products/repository"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

//swag init --parseDependency --parseInternal --parseDepth 1 -g ./cmd/main.go

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	echo := echo.New()
	productRepository := repository.NewInMemoryProductRepository()

	createProductCommandHandler := creatingProduct.NewCreateProductCommandHandler(productRepository)
	getByIdQueryHandler := gettingProductById.NewGetProductByIdHandler(productRepository)

	// Register handlers to the mediatr
	err := mediatr.RegisterRequestHandler[*creatingProduct.CreateProductCommand, *creatingProductsDtos.CreateProductCommandResponse](createProductCommandHandler)
	if err != nil {
		log.Fatal(err)
	}

	err = mediatr.RegisterRequestHandler[*gettingProductById.GetProductByIdQuery, *gettingProductByIdDtos.GetProductByIdQueryResponse](getByIdQueryHandler)
	if err != nil {
		log.Fatal(err)
	}

	controller := productApi.NewProductsController(echo)

	productApi.MapProductsRoutes(echo, controller)

	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Title = "Catalogs Write-Service Api"
	docs.SwaggerInfo.Description = "Catalogs Write-Service Api."

	echo.GET("/swagger/*", echoSwagger.WrapHandler)

	go func() {
		if err := echo.Start(":9080"); err != nil {
			log.Fatalf("Error starting Server: ", err)
		}
	}()

	<-ctx.Done()

	if err := echo.Shutdown(ctx); err != nil {
		log.Fatal("(Shutdown) err: {%v}", err)
	}
}
