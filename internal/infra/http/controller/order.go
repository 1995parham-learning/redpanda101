package controller

import (
	"net/http"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
	"github.com/1995parham-teaching/redpanda101/internal/infra/http/request"
	"github.com/1995parham-teaching/redpanda101/internal/infra/producer"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type Order struct {
	Producer *producer.Producer
}

func (c Order) New(ctx *echo.Context) error {
	var o request.Order
	if err := ctx.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse request body as an order").Wrap(err)
	}

	if err := o.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	d := model.Order{
		ID:          uuid.New().String(),
		SrcCurrency: o.SrcCurrency,
		DstCurrency: o.DstCurrency,
		Description: o.Description,
		Channel:     o.Channel,
	}

	if err := c.Producer.Produce(ctx.Request().Context(), d); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to publish order into Kafka").Wrap(err)
	}

	return ctx.JSON(http.StatusOK, d)
}

func (c Order) Register(e *echo.Echo) {
	e.POST("/orders/", c.New)
}
