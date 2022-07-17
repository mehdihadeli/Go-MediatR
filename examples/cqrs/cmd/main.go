package main

import (
	"context"
	mediatr "github.com/mehdihadeli/Go-MediatR"
	"github.com/mehdihadeli/Go-MediatR/examples/cqrs/docs"
	product_api "github.com/mehdihadeli/Go-MediatR/examples/cqrs/internal/products/api"
	"github.com/mehdihadeli/Go-MediatR/examples/cqrs/internal/products/features/creating_product"
	creating_products_dtos "github.com/mehdihadeli/Go-MediatR/examples/cqrs/internal/products/features/creating_product/dtos"
	"github.com/mehdihadeli/Go-MediatR/examples/cqrs/internal/products/features/getting_product_by_id"
	getting_product_by_id_dtos "github.com/mehdihadeli/Go-MediatR/examples/cqrs/internal/products/features/getting_product_by_id/dtos"
	"github.com/mehdihadeli/Go-MediatR/examples/cqrs/internal/products/repository"
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

	createProductCommandHandler := creating_product.NewCreateProductCommandHandler(productRepository)
	getByIdQueryHandler := getting_product_by_id.NewGetProductByIdHandler(productRepository)

	// Register handlers to the mediatr
	err := mediatr.RegisterHandler[*creating_product.CreateProductCommand, *creating_products_dtos.CreateProductResponseDto](createProductCommandHandler)
	if err != nil {
		log.Fatal(err)
	}

	err = mediatr.RegisterHandler[*getting_product_by_id.GetProductByIdQuery, *getting_product_by_id_dtos.GetProductByIdResponseDto](getByIdQueryHandler)
	if err != nil {
		log.Fatal(err)
	}

	controller := product_api.NewProductsController(echo)

	product_api.MapProductsRoutes(echo, controller)

	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Title = "Catalogs Write-Service Api"
	docs.SwaggerInfo.Description = "Catalogs Write-Service Api."

	echo.GET("/swagger/*", echoSwagger.WrapHandler)

	go func() {
		if err := echo.Start(":9090"); err != nil {
			log.Fatalf("Error starting Server: ", err)
		}
	}()

	<-ctx.Done()

	if err := echo.Shutdown(ctx); err != nil {
		log.Fatal("(Shutdown) err: {%v}", err)
	}
}
