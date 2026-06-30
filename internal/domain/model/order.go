package model

import "strconv"

// Side indicates whether an order wants to buy or sell the base currency.
type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

// Valid returns true if the side is a known value.
func (s Side) Valid() bool {
	switch s {
	case Buy, Sell:
		return true
	default:
		return false
	}
}

// Opposite returns the side an order must match against.
func (s Side) Opposite() Side {
	if s == Buy {
		return Sell
	}

	return Buy
}

type Channel int

const (
	Unknown            Channel = 0
	Web                Channel = 1
	WebFast            Channel = 2
	WebConvert         Channel = 3
	WebSimple          Channel = 4
	Android            Channel = 11
	AndroidFast        Channel = 12
	AndroidConvert     Channel = 13
	AndroidSimple      Channel = 14
	iOS                Channel = 21
	iOSConvert         Channel = 23
	API                Channel = 31
	APIInternal        Channel = 32
	APIConvert         Channel = 33
	APIInternalOld     Channel = 34
	WebV1              Channel = 41
	WebV2              Channel = 42
	SystemMargin       Channel = 51
	SystemBlock        Channel = 52
	SystemABCLiquidate Channel = 53
	SystemLiquidator   Channel = 54
	Locket             Channel = 61
)

// Valid returns true if the channel is a known valid value.
func (c Channel) Valid() bool {
	switch c {
	case Unknown, Web, WebFast, WebConvert, WebSimple,
		Android, AndroidFast, AndroidConvert, AndroidSimple,
		iOS, iOSConvert,
		API, APIInternal, APIConvert, APIInternalOld,
		WebV1, WebV2,
		SystemMargin, SystemBlock, SystemABCLiquidate, SystemLiquidator,
		Locket:
		return true
	default:
		return false
	}
}

type Order struct {
	ID          string `json:"id,omitempty"`
	SrcCurrency uint64 `json:"src_currency,omitempty"`
	DstCurrency uint64 `json:"dst_currency,omitempty"`
	Side        Side   `json:"side,omitempty"`
	// Price is the quote-currency (DstCurrency) cost of one unit of the base
	// currency (SrcCurrency). Integer to avoid floating-point money bugs.
	Price uint64 `json:"price,omitempty"`
	// Quantity is the amount of base currency (SrcCurrency) to trade.
	Quantity    uint64  `json:"quantity,omitempty"`
	Description string  `json:"description,omitempty"`
	Channel     Channel `json:"channel,omitempty"`
}

// Symbol identifies the market an order belongs to: the base/quote currency
// pair. All orders for a symbol share a single order book and must be processed
// in arrival order, so it also serves as the Kafka partition key.
func (o Order) Symbol() string {
	return strconv.FormatUint(o.SrcCurrency, 10) + "-" + strconv.FormatUint(o.DstCurrency, 10)
}
