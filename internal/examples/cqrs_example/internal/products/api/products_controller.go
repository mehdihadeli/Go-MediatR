package api

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"

	"cqrsexample/internal/products/features/creating_product/commands"
	dtos3 "cqrsexample/internal/products/features/creating_product/dtos"
	"cqrsexample/internal/products/features/getting_product_by_id/dtos"
	"cqrsexample/internal/products/features/getting_product_by_id/queries"
	"github.com/mehdihadeli/go-mediatr"
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
// @Param CreateProductRequestDto body creatingProductsDtos.CreateProductRequestDto true "Product data"
// @Success 201 {object} creatingProductsDtos.CreateProductResponseDto
// @Router /api/v1/products [post]
func (pc *ProductsController) createProduct() echo.HandlerFunc {

	return func(ctx echo.Context) error {
		request := &dtos3.CreateProductRequestDto{}
		if err := ctx.Bind(request); err != nil {
			return err
		}

		if err := pc.validator.StructCtx(ctx.Request().Context(), request); err != nil {
			return err
		}

		command := commands.NewCreateProductCommand(request.Name, request.Description, request.Price)
		result, err := mediatr.Send[*commands.CreateProductCommand, *dtos3.CreateProductCommandResponse](ctx.Request().Context(), command)

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
// @Success 200 {object} gettingProductByIdDtos.GetProductByIdResponseDto
// @Router /api/v1/products/{id} [get]
func (pc *ProductsController) getProductByID() echo.HandlerFunc {
	return func(ctx echo.Context) error {

		request := &dtos.GetProductByIdRequestDto{}
		if err := ctx.Bind(request); err != nil {
			return err
		}

		query := queries.NewGetProductByIdQuery(request.ProductId)

		if err := pc.validator.StructCtx(ctx.Request().Context(), query); err != nil {
			return err
		}

		queryResult, err := mediatr.Send[*queries.GetProductByIdQuery, *dtos.GetProductByIdQueryResponse](ctx.Request().Context(), query)

		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, queryResult)
	}
}
