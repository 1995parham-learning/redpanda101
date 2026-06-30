package constant

import "time"

const (
	// Topic carries incoming orders. It is keyed by symbol so every order for a
	// market lands on the same partition and is matched in arrival order.
	Topic = "orders"
	// TradesTopic carries the trades emitted by the matching engine.
	TradesTopic = "trades"
	PingTimeout = 5 * time.Second
)
