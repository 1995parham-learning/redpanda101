package model

import "time"

// Trade is the result of matching an incoming (taker) order against a resting
// (maker) order in the book. It always executes at the maker's price.
type Trade struct {
	ID     string `json:"id,omitempty"`
	Symbol string `json:"symbol,omitempty"`
	// Price is the maker order's price (price-time priority gives the resting
	// order price improvement to the taker).
	Price    uint64 `json:"price,omitempty"`
	Quantity uint64 `json:"quantity,omitempty"`
	// BuyOrderID and SellOrderID are the two orders that crossed, regardless of
	// which side was the taker.
	BuyOrderID  string `json:"buy_order_id,omitempty"`
	SellOrderID string `json:"sell_order_id,omitempty"`
	// TakerSide is the side of the incoming order that triggered the match.
	TakerSide Side      `json:"taker_side,omitempty"`
	CreatedAt time.Time `json:"created_at,omitzero"`
}
