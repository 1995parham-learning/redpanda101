package request

import (
	"errors"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
)

var ErrInvalidChannel = errors.New("invalid channel value")

type Order struct {
	SrcCurrency uint64        `json:"src_currency,omitempty"`
	DstCurrency uint64        `json:"dst_currency,omitempty"`
	Description string        `json:"description,omitempty"`
	Channel     model.Channel `json:"channel,omitempty"`
}

func (o Order) Validate() error {
	if !o.Channel.Valid() {
		return ErrInvalidChannel
	}

	return nil
}
