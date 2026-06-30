package orderbook

import (
	"errors"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
)

var (
	ErrInvalidSide      = errors.New("order side must be buy or sell")
	ErrNonPositivePrice = errors.New("order price must be greater than zero")
	ErrNonPositiveQty   = errors.New("order quantity must be greater than zero")
)

// Engine routes each order to the order book for its symbol, lazily creating
// books as new markets appear. It is the in-memory materialised view of the
// orders log: replaying the same orders in the same order always reproduces the
// same books and the same trades.
//
// Engine is not safe for concurrent use; the matcher service drives it from a
// single goroutine per partition, which is also what preserves arrival order.
type Engine struct {
	books map[string]*OrderBook
}

// NewEngine returns an empty engine with no books.
func NewEngine() *Engine {
	return &Engine{books: make(map[string]*OrderBook)}
}

// Submit validates and applies an order, returning the resulting trades and the
// quantity that came to rest in the book.
func (e *Engine) Submit(order model.Order) ([]model.Trade, uint64, error) {
	if !order.Side.Valid() {
		return nil, 0, ErrInvalidSide
	}

	if order.Price == 0 {
		return nil, 0, ErrNonPositivePrice
	}

	if order.Quantity == 0 {
		return nil, 0, ErrNonPositiveQty
	}

	trades, resting := e.book(order.Symbol()).Match(order)

	return trades, resting, nil
}

// Book returns the order book for a symbol and whether it exists yet. The
// returned book is read-only for callers; mutating it bypasses Submit.
func (e *Engine) Book(symbol string) (*OrderBook, bool) {
	b, ok := e.books[symbol]

	return b, ok
}

// Symbols returns the markets the engine currently tracks.
func (e *Engine) Symbols() []string {
	symbols := make([]string, 0, len(e.books))
	for symbol := range e.books {
		symbols = append(symbols, symbol)
	}

	return symbols
}

// book returns the book for a symbol, creating it on first use.
func (e *Engine) book(symbol string) *OrderBook {
	b, ok := e.books[symbol]
	if !ok {
		b = New(symbol)
		e.books[symbol] = b
	}

	return b
}
