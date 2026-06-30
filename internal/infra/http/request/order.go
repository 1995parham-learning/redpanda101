package request

import (
	"errors"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
)

var (
	ErrInvalidChannel   = errors.New("invalid channel value")
	ErrInvalidSide      = errors.New("side must be \"buy\" or \"sell\"")
	ErrNonPositivePrice = errors.New("price must be greater than zero")
	ErrNonPositiveQty   = errors.New("quantity must be greater than zero")
)

type Order struct {
	SrcCurrency uint64        `json:"src_currency,omitempty"`
	DstCurrency uint64        `json:"dst_currency,omitempty"`
	Side        model.Side    `json:"side,omitempty"`
	Price       uint64        `json:"price,omitempty"`
	Quantity    uint64        `json:"quantity,omitempty"`
	Description string        `json:"description,omitempty"`
	Channel     model.Channel `json:"channel,omitempty"`
}

func (o Order) Validate() error {
	if !o.Channel.Valid() {
		return ErrInvalidChannel
	}

	if !o.Side.Valid() {
		return ErrInvalidSide
	}

	if o.Price == 0 {
		return ErrNonPositivePrice
	}

	if o.Quantity == 0 {
		return ErrNonPositiveQty
	}

	return nil
}
