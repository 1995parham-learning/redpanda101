// Package orderbook implements a continuous limit order book (CLOB) with
// price-time priority. It is pure domain logic: no I/O, no clocks, no randomness,
// so the matching behaviour is fully deterministic and unit-testable. The
// matcher service (internal/infra/matcher) feeds it orders replayed from the
// Redpanda log and turns the trades it returns into events.
package orderbook

import (
	"sort"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
)

// restingOrder is an order sitting in the book with its still-open quantity and
// a sequence number capturing arrival order (for time priority within a price).
type restingOrder struct {
	order     model.Order
	remaining uint64
	seq       uint64
}

// OrderBook holds the resting orders for a single symbol. bids and asks are kept
// sorted best-first so the front of each slice is the best price; ties are
// broken by arrival sequence (FIFO), giving price-time priority.
//
// Sorted slices keep the matching logic obvious at the cost of O(n) inserts. A
// production engine would use a price-level tree or heap; for a teaching demo
// clarity wins.
type OrderBook struct {
	symbol string
	bids   []*restingOrder // highest price first
	asks   []*restingOrder // lowest price first
	seq    uint64          // monotonic arrival counter
}

// New returns an empty order book for the given symbol.
func New(symbol string) *OrderBook {
	return &OrderBook{
		symbol: symbol,
		bids:   nil,
		asks:   nil,
		seq:    0,
	}
}

// Match applies an incoming limit order to the book. It crosses the order
// against the opposite side while prices overlap, executing each fill at the
// resting (maker) price, and rests any unfilled remainder. It returns the
// trades generated (without ID/CreatedAt, which the caller stamps) and the
// quantity left resting in the book.
func (b *OrderBook) Match(incoming model.Order) ([]model.Trade, uint64) {
	remaining := incoming.Quantity

	var (
		trades []model.Trade
		makers *[]*restingOrder
	)

	if incoming.Side == model.Buy {
		makers = &b.asks
	} else {
		makers = &b.bids
	}

	for remaining > 0 && len(*makers) > 0 {
		best := (*makers)[0]

		if !crosses(incoming, best.order) {
			break
		}

		qty := min(remaining, best.remaining)

		trades = append(trades, b.newTrade(incoming, best.order, qty))

		remaining -= qty
		best.remaining -= qty

		if best.remaining == 0 {
			*makers = (*makers)[1:]
		}
	}

	if remaining > 0 {
		b.rest(incoming, remaining)
	}

	return trades, remaining
}

// PriceLevel aggregates the open quantity resting at a single price.
type PriceLevel struct {
	Price    uint64 `json:"price"`
	Quantity uint64 `json:"quantity"`
	Orders   int    `json:"orders"`
}

// Snapshot is a read-only view of the top of both sides of the book.
type Snapshot struct {
	Symbol string       `json:"symbol"`
	Bids   []PriceLevel `json:"bids"`
	Asks   []PriceLevel `json:"asks"`
}

// BestBid returns the highest resting buy price and whether one exists.
func (b *OrderBook) BestBid() (uint64, bool) {
	if len(b.bids) == 0 {
		return 0, false
	}

	return b.bids[0].order.Price, true
}

// BestAsk returns the lowest resting sell price and whether one exists.
func (b *OrderBook) BestAsk() (uint64, bool) {
	if len(b.asks) == 0 {
		return 0, false
	}

	return b.asks[0].order.Price, true
}

// Snapshot returns up to depth aggregated price levels per side, best first.
func (b *OrderBook) Snapshot(depth int) Snapshot {
	return Snapshot{
		Symbol: b.symbol,
		Bids:   levels(b.bids, depth),
		Asks:   levels(b.asks, depth),
	}
}

// levels collapses a best-first sorted side into aggregated price levels.
func levels(orders []*restingOrder, depth int) []PriceLevel {
	var out []PriceLevel

	for _, ro := range orders {
		if n := len(out); n > 0 && out[n-1].Price == ro.order.Price {
			out[n-1].Quantity += ro.remaining
			out[n-1].Orders++

			continue
		}

		if len(out) == depth {
			break
		}

		out = append(out, PriceLevel{Price: ro.order.Price, Quantity: ro.remaining, Orders: 1})
	}

	return out
}

// crosses reports whether a taker order and a resting maker order overlap in
// price and can therefore trade.
func crosses(taker, maker model.Order) bool {
	if taker.Side == model.Buy {
		// A buy crosses an ask priced at or below the buy's limit.
		return taker.Price >= maker.Price
	}

	// A sell crosses a bid priced at or above the sell's limit.
	return taker.Price <= maker.Price
}

// newTrade builds a trade between a taker and a maker at the maker's price.
func (b *OrderBook) newTrade(taker, maker model.Order, qty uint64) model.Trade {
	buyID, sellID := taker.ID, maker.ID
	if taker.Side == model.Sell {
		buyID, sellID = maker.ID, taker.ID
	}

	// nolint: exhaustruct // ID and CreatedAt are stamped by the matcher service.
	return model.Trade{
		Symbol:      b.symbol,
		Price:       maker.Price,
		Quantity:    qty,
		BuyOrderID:  buyID,
		SellOrderID: sellID,
		TakerSide:   taker.Side,
	}
}

// rest inserts the unfilled remainder of an order into the correct side of the
// book, keeping the side sorted best-first with FIFO tie-breaking.
func (b *OrderBook) rest(order model.Order, remaining uint64) {
	ro := &restingOrder{order: order, remaining: remaining, seq: b.seq}
	b.seq++

	if order.Side == model.Buy {
		b.bids = insertSorted(b.bids, ro, bidLess)
	} else {
		b.asks = insertSorted(b.asks, ro, askLess)
	}
}

// bidLess reports whether bid a ranks ahead of bid b: higher price first, then
// earlier arrival.
func bidLess(a, b *restingOrder) bool {
	if a.order.Price != b.order.Price {
		return a.order.Price > b.order.Price
	}

	return a.seq < b.seq
}

// askLess reports whether ask a ranks ahead of ask b: lower price first, then
// earlier arrival.
func askLess(a, b *restingOrder) bool {
	if a.order.Price != b.order.Price {
		return a.order.Price < b.order.Price
	}

	return a.seq < b.seq
}

// insertSorted places ro into an already best-first-sorted slice using less to
// decide ordering. Because ro always has the newest seq, equal-price orders sort
// after the resting ones, preserving FIFO.
func insertSorted(orders []*restingOrder, ro *restingOrder, less func(a, b *restingOrder) bool) []*restingOrder {
	idx := sort.Search(len(orders), func(i int) bool {
		return less(ro, orders[i])
	})

	orders = append(orders, nil)
	copy(orders[idx+1:], orders[idx:])
	orders[idx] = ro

	return orders
}
