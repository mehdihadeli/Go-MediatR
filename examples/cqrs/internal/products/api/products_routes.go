package api

import "github.com/labstack/echo/v4"

func MapProductsRoutes(echo *echo.Echo, controller *ProductsController) {
	v1 := echo.Group("/api/v1")
	products := v1.Group("/products")

	products.POST("", controller.createProduct())
	products.GET("/:id", controller.getProductByID())
}
