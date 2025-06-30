package controller

import (
	"math/rand/v2"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
	"github.com/1995parham-teaching/redpanda101/internal/infra/http/request"
	"github.com/1995parham-teaching/redpanda101/internal/infra/producer"
	"github.com/go-fuego/fuego"
)

type Order struct {
	Producer *producer.Producer
}

func (c Order) New(ctx fuego.ContextWithBody[request.Order]) (*model.Order, error) {
	o, err := ctx.Body()
	if err != nil {
		return nil, fuego.BadRequestError{
			Err:      err,
			Type:     "",
			Title:    "Bad Request",
			Status:   0,
			Detail:   "failed to parse request body as an order",
			Instance: "",
			Errors:   nil,
		}
	}

	d := model.Order{
		ID:          rand.Uint64(), // nolint: gosec
		SrcCurrency: o.SrcCurrency,
		DstCurrency: o.DstCurrency,
		Description: o.Description,
		Channel:     o.Channel,
	}

	err = c.Producer.Produce(ctx.Context(), d)
	if err != nil {
		return nil, fuego.InternalServerError{
			Err:      err,
			Type:     "",
			Title:    "Internal server error",
			Status:   0,
			Detail:   "failed to publish order into Kafka",
			Instance: "",
			Errors:   nil,
		}
	}

	return &d, nil
}

func (c Order) Register(s *fuego.Server) {
	fuego.Post(s, "/orders/", c.New)
}
