package controller

import (
	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
	"github.com/1995parham-teaching/redpanda101/internal/infra/http/request"
	"github.com/1995parham-teaching/redpanda101/internal/infra/producer"
	"github.com/go-fuego/fuego"
)

type Order struct {
	Producer *producer.Producer
}

func (c Order) New(ctx fuego.ContextWithBody[request.Order]) error {
	o, err := ctx.Body()
	if err != nil {
		return fuego.BadRequestError{
			Err: err,
		}
	}

	if err := c.Producer.Produce(ctx.Context(), model.Order{
		ID: 0,
		SrcCurrency: o.SrcCurrency,
		DstCurrency: o.DstCurrency,
		Description: o.Description,
		Channel: o.Channel,
	}); err != nil {
		return fuego.InternalServerError{
			Err: err,
		}
	}

	return nil
}
