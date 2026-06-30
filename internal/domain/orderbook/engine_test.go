package orderbook_test

import (
	"errors"
	"testing"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
	"github.com/1995parham-teaching/redpanda101/internal/domain/orderbook"
)

func submit(t *testing.T, e *orderbook.Engine, o model.Order) []model.Trade {
	t.Helper()

	trades, _, err := e.Submit(o)
	if err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}

	return trades
}

func TestEngine_ValidatesOrders(t *testing.T) {
	t.Parallel()

	e := orderbook.NewEngine()

	cases := map[string]struct {
		order model.Order
		want  error
	}{
		"bad side":      {order("a", "hold", 100, 1), orderbook.ErrInvalidSide},
		"zero price":    {order("a", model.Buy, 0, 1), orderbook.ErrNonPositivePrice},
		"zero quantity": {order("a", model.Buy, 100, 0), orderbook.ErrNonPositiveQty},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, _, err := e.Submit(tc.order); !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestEngine_SymbolsAreIsolated(t *testing.T) {
	t.Parallel()

	e := orderbook.NewEngine()

	// Same prices, different markets: must NOT cross.
	//nolint:exhaustruct
	m4 := model.Order{ID: "x", SrcCurrency: 3, DstCurrency: 4, Side: model.Sell, Price: 100, Quantity: 5}
	//nolint:exhaustruct
	mb := model.Order{ID: "y", SrcCurrency: 1, DstCurrency: 2, Side: model.Buy, Price: 100, Quantity: 5}

	if trades := submit(t, e, m4); len(trades) != 0 {
		t.Fatalf("first order should rest, got trades %+v", trades)
	}

	if trades := submit(t, e, mb); len(trades) != 0 {
		t.Fatalf("cross-symbol orders must not match, got %+v", trades)
	}

	if got := len(e.Symbols()); got != 2 {
		t.Fatalf("expected 2 symbols tracked, got %d", got)
	}
}

func TestEngine_DeterministicReplay(t *testing.T) {
	t.Parallel()

	orders := []model.Order{
		order("a", model.Sell, 101, 5),
		order("b", model.Sell, 100, 5),
		order("c", model.Buy, 100, 3),
		order("d", model.Buy, 102, 8),
		order("e", model.Sell, 99, 4),
	}

	run := func() []model.Trade {
		e := orderbook.NewEngine()

		all := make([]model.Trade, 0, len(orders))
		for _, o := range orders {
			all = append(all, submit(t, e, o)...)
		}

		return all
	}

	first := run()
	second := run()

	if len(first) != len(second) {
		t.Fatalf("replay produced different trade counts: %d vs %d", len(first), len(second))
	}

	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("replay diverged at trade %d: %+v vs %+v", i, first[i], second[i])
		}
	}

	if len(first) == 0 {
		t.Fatal("expected the scenario to generate trades")
	}
}
