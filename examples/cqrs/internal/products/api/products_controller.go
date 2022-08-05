package api

import (
	"github.com/mehdihadeli/go-mediatr"
	"github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/features/creating_product/commands"
	creatingProductsDtos "github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/features/creating_product/dtos"
	gettingProductByIdDtos "github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/features/getting_product_by_id/dtos"
	"github.com/mehdihadeli/go-mediatr/examples/cqrs/internal/products/features/getting_product_by_id/queries"
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
// @Param CreateProductRequestDto body creatingProductsDtos.CreateProductRequestDto true "Product data"
// @Success 201 {object} creatingProductsDtos.CreateProductResponseDto
// @Router /api/v1/products [post]
func (pc *ProductsController) createProduct() echo.HandlerFunc {

	return func(ctx echo.Context) error {
		request := &creatingProductsDtos.CreateProductRequestDto{}
		if err := ctx.Bind(request); err != nil {
			return err
		}

		if err := pc.validator.StructCtx(ctx.Request().Context(), request); err != nil {
			return err
		}

		command := commands.NewCreateProductCommand(request.Name, request.Description, request.Price)
		result, err := mediatr.Send[*commands.CreateProductCommand, *creatingProductsDtos.CreateProductCommandResponse](ctx.Request().Context(), command)

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

		request := &gettingProductByIdDtos.GetProductByIdRequestDto{}
		if err := ctx.Bind(request); err != nil {
			return err
		}

		query := queries.NewGetProductByIdQuery(request.ProductId)

		if err := pc.validator.StructCtx(ctx.Request().Context(), query); err != nil {
			return err
		}

		queryResult, err := mediatr.Send[*queries.GetProductByIdQuery, *gettingProductByIdDtos.GetProductByIdQueryResponse](ctx.Request().Context(), query)

		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, queryResult)
	}
}
