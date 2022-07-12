package api

import (
	"mediatR"
	"mediatR/examples/cqrs/internal/products/features/creating_product"
	creating_products_dtos "mediatR/examples/cqrs/internal/products/features/creating_product/dtos"
	"mediatR/examples/cqrs/internal/products/features/getting_product_by_id"
	getting_product_by_id_dtos "mediatR/examples/cqrs/internal/products/features/getting_product_by_id/dtos"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type ProductsController struct {
	echo      *echo.Echo
	validator *validator.Validate
}

func NewProductsController(echo *echo.Echo) *ProductsController {
	return &ProductsController{echo: echo, validator: validator.New()}
}

// CreateProduct
// @Tags Products
// @Summary Create product
// @Description Create new product item
// @Accept json
// @Produce json
// @Param CreateProductRequestDto body creating_products_dtos.CreateProductRequestDto true "Product data"
// @Success 201 {object} creating_products_dtos.CreateProductResponseDto
// @Router /api/v1/products [post]
func (pc *ProductsController) createProduct() echo.HandlerFunc {

	return func(ctx echo.Context) error {
		request := &creating_products_dtos.CreateProductRequestDto{}
		if err := ctx.Bind(request); err != nil {
			return err
		}

		if err := pc.validator.StructCtx(ctx.Request().Context(), request); err != nil {
			return err
		}

		command := creating_product.NewCreateProductCommand(request.Name, request.Description, request.Price)
		result, err := mediatR.Send[*creating_products_dtos.CreateProductResponseDto](ctx.Request().Context(), command)

		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusCreated, result)
	}
}

// GetProductByID
// @Tags Products
// @Summary Get product
// @Description Get product by id
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} getting_product_by_id_dtos.GetProductByIdResponseDto
// @Router /api/v1/products/{id} [get]
func (pc *ProductsController) getProductByID() echo.HandlerFunc {
	return func(ctx echo.Context) error {

		request := &getting_product_by_id_dtos.GetProductByIdRequestDto{}
		if err := ctx.Bind(request); err != nil {
			return err
		}

		query := getting_product_by_id.NewGetProductByIdQuery(request.ProductId)

		if err := pc.validator.StructCtx(ctx.Request().Context(), query); err != nil {
			return err
		}

		queryResult, err := mediatR.Send[*getting_product_by_id_dtos.GetProductByIdResponseDto](ctx.Request().Context(), query)

		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, queryResult)
	}
}
