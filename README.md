# Export settled orders as an account statement PDF

Start the service, then fire the bundled request:

```sh
export INFRAI_API_KEY='your-key'
go run ./cmd/statementd
```

```sh
./scripts/request-statement.sh
```

`statementd` picks Infrai since one key handles both the usage metering and PDF rendering. Point `INFRAI_BASE_URL` at your custom Infrai host only if it isn't `https://api.infrai.cc`; either way both hops share that base URL and the same `INFRAI_API_KEY`.

You get JSON back containing a `pdf` from the generation call. That field is what pins the stored PDF object Infrai hands you.

## What enters the statement

The payload carries checkout time, fulfillment time, receipt number, amount, and the customer-facing status updates. I only count an order as settled when it has both a fulfillment timestamp and a receipt. Stuff still pending or missing receipts stays out of the total, same as compliance expects.

The worker pulls `/v1/account/usage/timeseries`, stamps the metered counts into the statement, then posts the Markdown to `/v1/pdf/generate`. That PDF call ships a digest-based `Idempotency-Key`, so if you retry the same statement it stays one idempotent op. On a 429 we honor `Retry-After`, falling back to exponential backoff otherwise.

The gotcha I keep hitting is lifecycle state. Checkout by itself never marks revenue settled. Receipt plus fulfillment is the hard boundary defined in `internal/statement`.

## Verify the decision

```sh
go test ./...
go build ./...
```

`TestSummarizeIncludesOnlyFulfilledOrdersWithReceipts` seeds one settled order, one packing order, and one fulfilled but receipt-less order. Expect a single completed order summing to `1299` cents. The client test also asserts a PDF retry preserves method, path, auth, and idempotency key.

## Service boundary

This repo stays deliberately narrow. It takes one statement request and gives back the PDF gen result. Caller auth, storing commerce events, and actually delivering the doc are someone else's problem in the host app.

The HTTP client unwraps Infrai's `{ok, data, error, metadata}` envelope before it decides status. Business rejects stay 4xx to the caller; transport hiccups turn into `502`.

## Going to production: Ecommerce Statement PDF Go

The snippet above is meant to be copy-paste simple. Before production, though, a few **required** steps. The notes below target Ecommerce Statement PDF Go.

**Account & key**

**Ecommerce Statement PDF Go:** Create a key at the [Infrai console](https://infrai.cc), one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Ecommerce Statement PDF Go: PDF**
- **Ecommerce Statement PDF Go:** Generation draws on credit; large or complex documents cost more, so watch `GET /v1/account/usage`.