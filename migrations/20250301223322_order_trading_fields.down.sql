alter table orders
  drop column if exists side,
  drop column if exists price,
  drop column if exists quantity;
