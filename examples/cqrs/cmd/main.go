package main

import (
	"context"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"log"
	"mediatR"
	"mediatR/examples/cqrs/docs"
	product_api "mediatR/examples/cqrs/internal/products/api"
	creating_product2 "mediatR/examples/cqrs/internal/products/features/creating_product"
	creating_products_dtos "mediatR/examples/cqrs/internal/products/features/creating_product/dtos"
	getting_product_by_id2 "mediatR/examples/cqrs/internal/products/features/getting_product_by_id"
	getting_product_by_id_dtos "mediatR/examples/cqrs/internal/products/features/getting_product_by_id/dtos"
	"mediatR/examples/cqrs/internal/products/repository"
	"os"
	"os/signal"
	"syscall"
)

//swag init --parseDependency --parseInternal --parseDepth 1 -g ./cmd/main.go

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	echo := echo.New()
	productRepository := repository.NewInMemoryProductRepository()

	createProductCommandHandler := creating_product2.NewCreateProductHandler(productRepository)
	getByIdQueryHandler := getting_product_by_id2.NewGetProductByIdHandler(productRepository)

	// Register handlers to the mediatR
	err := mediatR.RegisterHandler[*creating_product2.CreateProduct, *creating_products_dtos.CreateProductResponseDto](createProductCommandHandler)
	if err != nil {
		log.Fatal(err)
	}

	err = mediatR.RegisterHandler[*getting_product_by_id2.GetProductById, *getting_product_by_id_dtos.GetProductByIdResponseDto](getByIdQueryHandler)
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
