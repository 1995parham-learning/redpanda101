package orderbook_test

import (
	"testing"

	"github.com/1995parham-teaching/redpanda101/internal/domain/model"
	"github.com/1995parham-teaching/redpanda101/internal/domain/orderbook"
)

// order is a tiny helper to build limit orders for the tests.
func order(id string, side model.Side, price, qty uint64) model.Order {
	// nolint: exhaustruct
	return model.Order{ID: id, Side: side, Price: price, Quantity: qty}
}

func TestMatch_RestsWhenBookEmpty(t *testing.T) {
	t.Parallel()

	b := orderbook.New("1-2")

	trades, resting := b.Match(order("a", model.Buy, 100, 5))

	if len(trades) != 0 {
		t.Fatalf("expected no trades, got %d", len(trades))
	}

	if resting != 5 {
		t.Fatalf("expected full quantity to rest, got %d", resting)
	}

	if bid, ok := b.BestBid(); !ok || bid != 100 {
		t.Fatalf("expected best bid 100, got %d (ok=%v)", bid, ok)
	}
}

func TestMatch_FullFillAtMakerPrice(t *testing.T) {
	t.Parallel()

	b := orderbook.New("1-2")
	// Resting ask at 100; incoming buy crosses at 105 but fills at the maker's 100.
	b.Match(order("maker", model.Sell, 100, 5))

	trades, resting := b.Match(order("taker", model.Buy, 105, 5))

	if resting != 0 {
		t.Fatalf("expected taker fully filled, %d left resting", resting)
	}

	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}

	tr := trades[0]
	if tr.Price != 100 {
		t.Errorf("expected execution at maker price 100, got %d", tr.Price)
	}

	if tr.Quantity != 5 {
		t.Errorf("expected quantity 5, got %d", tr.Quantity)
	}

	if tr.BuyOrderID != "taker" || tr.SellOrderID != "maker" {
		t.Errorf("trade order ids wrong: buy=%s sell=%s", tr.BuyOrderID, tr.SellOrderID)
	}

	if tr.TakerSide != model.Buy {
		t.Errorf("expected taker side buy, got %s", tr.TakerSide)
	}

	if _, ok := b.BestAsk(); ok {
		t.Error("expected ask side to be empty after full fill")
	}
}

func TestMatch_PartialFillTakerRests(t *testing.T) {
	t.Parallel()

	b := orderbook.New("1-2")
	b.Match(order("maker", model.Sell, 100, 3))

	trades, resting := b.Match(order("taker", model.Buy, 100, 5))

	if len(trades) != 1 || trades[0].Quantity != 3 {
		t.Fatalf("expected one trade of qty 3, got %+v", trades)
	}

	if resting != 2 {
		t.Fatalf("expected 2 to rest, got %d", resting)
	}

	if bid, ok := b.BestBid(); !ok || bid != 100 {
		t.Fatalf("expected leftover buy resting at 100, got %d (ok=%v)", bid, ok)
	}

	if _, ok := b.BestAsk(); ok {
		t.Error("expected ask fully consumed")
	}
}

func TestMatch_PartialFillMakerStays(t *testing.T) {
	t.Parallel()

	b := orderbook.New("1-2")
	b.Match(order("maker", model.Sell, 100, 10))

	_, resting := b.Match(order("taker", model.Buy, 100, 4))

	if resting != 0 {
		t.Fatalf("expected taker filled, %d resting", resting)
	}

	snap := b.Snapshot(5)
	if len(snap.Asks) != 1 || snap.Asks[0].Quantity != 6 {
		t.Fatalf("expected 6 left on the ask, got %+v", snap.Asks)
	}
}

func TestMatch_NoCrossWhenPricesDoNotOverlap(t *testing.T) {
	t.Parallel()

	b := orderbook.New("1-2")
	b.Match(order("maker", model.Sell, 101, 5))

	trades, resting := b.Match(order("taker", model.Buy, 100, 5))

	if len(trades) != 0 {
		t.Fatalf("expected no trades when buy < ask, got %d", len(trades))
	}

	if resting != 5 {
		t.Fatalf("expected buy to rest, got %d", resting)
	}

	// Now both sides have resting orders with a spread.
	bid, okBid := b.BestBid()
	ask, okAsk := b.BestAsk()

	if !okBid || !okAsk || bid != 100 || ask != 101 {
		t.Fatalf("expected spread 100/101, got %d/%d (%v/%v)", bid, ask, okBid, okAsk)
	}
}

func TestMatch_PriceThenTimePriority(t *testing.T) {
	t.Parallel()

	b := orderbook.New("1-2")
	// Two asks at 100 (FIFO: m1 then m2) and a worse ask at 101.
	b.Match(order("m1", model.Sell, 100, 5))
	b.Match(order("m2", model.Sell, 100, 5))
	b.Match(order("m3", model.Sell, 101, 5))

	// Buy 12 at 101 sweeps: m1 (5) and m2 (5) at 100, then 2 from m3 at 101.
	trades, resting := b.Match(order("taker", model.Buy, 101, 12))

	if resting != 0 {
		t.Fatalf("expected taker filled, %d resting", resting)
	}

	want := []struct {
		seller string
		price  uint64
		qty    uint64
	}{
		{"m1", 100, 5},
		{"m2", 100, 5},
		{"m3", 101, 2},
	}

	if len(trades) != len(want) {
		t.Fatalf("expected %d trades, got %d: %+v", len(want), len(trades), trades)
	}

	for i, w := range want {
		if trades[i].SellOrderID != w.seller || trades[i].Price != w.price || trades[i].Quantity != w.qty {
			t.Errorf("trade %d = %+v, want seller=%s price=%d qty=%d", i, trades[i], w.seller, w.price, w.qty)
		}
	}

	// m3 has 3 left at 101.
	if ask, ok := b.BestAsk(); !ok || ask != 101 {
		t.Fatalf("expected remaining ask at 101, got %d (ok=%v)", ask, ok)
	}
}

func TestMatch_SellTakerHitsBestBidFirst(t *testing.T) {
	t.Parallel()

	b := orderbook.New("1-2")
	// Bids: 100 and 99. A sell should hit the higher (better) bid first.
	b.Match(order("b1", model.Buy, 99, 5))
	b.Match(order("b2", model.Buy, 100, 5))

	trades, resting := b.Match(order("taker", model.Sell, 99, 6))

	if resting != 0 {
		t.Fatalf("expected sell filled, %d resting", resting)
	}

	if len(trades) != 2 {
		t.Fatalf("expected 2 trades, got %d: %+v", len(trades), trades)
	}

	// First fill against b2 at 100 (best bid), then b1 at 99.
	if trades[0].BuyOrderID != "b2" || trades[0].Price != 100 || trades[0].Quantity != 5 {
		t.Errorf("first fill wrong: %+v", trades[0])
	}

	if trades[1].BuyOrderID != "b1" || trades[1].Price != 99 || trades[1].Quantity != 1 {
		t.Errorf("second fill wrong: %+v", trades[1])
	}

	if trades[0].TakerSide != model.Sell {
		t.Errorf("expected taker side sell, got %s", trades[0].TakerSide)
	}
}

func TestSnapshot_AggregatesPriceLevels(t *testing.T) {
	t.Parallel()

	b := orderbook.New("1-2")
	b.Match(order("b1", model.Buy, 100, 5))
	b.Match(order("b2", model.Buy, 100, 3)) // same level as b1
	b.Match(order("b3", model.Buy, 99, 7))

	snap := b.Snapshot(10)

	if len(snap.Bids) != 2 {
		t.Fatalf("expected 2 bid levels, got %+v", snap.Bids)
	}

	if snap.Bids[0].Price != 100 || snap.Bids[0].Quantity != 8 || snap.Bids[0].Orders != 2 {
		t.Errorf("top bid level wrong: %+v", snap.Bids[0])
	}

	if snap.Bids[1].Price != 99 || snap.Bids[1].Quantity != 7 {
		t.Errorf("second bid level wrong: %+v", snap.Bids[1])
	}
}
