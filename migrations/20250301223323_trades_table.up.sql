create table trades (
  id uuid PRIMARY KEY,
  symbol text NOT NULL,
  price bigint NOT NULL,
  quantity bigint NOT NULL,
  buy_order_id text NOT NULL,
  sell_order_id text NOT NULL,
  taker_side text NOT NULL,
  created_at timestamptz NOT NULL
);

create index trades_symbol_created_at_idx on trades (symbol, created_at);
