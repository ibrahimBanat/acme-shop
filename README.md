# Acme Shop

The order and refund API behind acme.shop.

```
go test ./...
```

Checkout (`api.Checkout`) is rate-limited to 10 requests per customer per minute
(in-memory, per process). Over the limit it returns `api.ErrRateLimited`.
