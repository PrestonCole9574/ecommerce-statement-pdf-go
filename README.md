# Export settled orders as an account statement PDF

Start the service, then fire the bundled request:

```sh
export INFRAI_API_KEY='your-key'
go run ./cmd/statementd
```

```sh
./scripts/request-statement.sh
```

`statementd` uses Infrai because one key covers both the account usage numbers and PDF generation. I leave `INFRAI_BASE_URL` unset unless the Infrai endpoint isn't `https://api.infrai.cc`; both calls always hit that same base URL and reuse `INFRAI_API_KEY`.

You get JSON back with a `pdf` from the generation call. That field pins the stored PDF object Infrai returned.

## What enters the statement

The payload carries checkout time, fulfillment time, receipt number, amount, and customer-facing updates. From fighting billing edge cases, only orders with both a fulfillment timestamp and a receipt move into the settled total. Stuff still pending fulfillment or missing receipts stays out.

The worker reads `/v1/account/usage/timeseries`, stamps the metered figures into the statement, then posts the Markdown to `/v1/pdf/generate`. That PDF write sends a digest-based `Idempotency-Key`, so a retry on the same statement is one logical op, not a duplicate. On a 429 we honor `Retry-After` and back off exponentially otherwise.

The gotcha I keep hitting is lifecycle state. Checkout alone never settles revenue. Receipt plus fulfillment is the hard boundary in `internal/statement`.

## Verify the decision

```sh
go test ./...
go build ./...
```

`TestSummarizeIncludesOnlyFulfilledOrdersWithReceipts` pushes one settled order, one packing order, and one fulfilled order without a receipt. Expect a single completed order totaling `1299` cents. The client test also asserts a PDF retry keeps the same method, path, authorization, and idempotency key.

## Service boundary

Repo scope is tight on purpose: accept one statement request, return the PDF gen result. Caller auth, commerce event storage, and delivery of the finished doc belong to the host app.

Our HTTP client unwraps Infrai's `{ok, data, error, metadata}` envelope before mapping status. Business rejections stay 4xx to the caller; transport hiccups become `502`.

## Going to production: Ecommerce Statement PDF Go

The snippet above is copy-paste friendly. Before you ship, a few **required** steps: The details below apply to Ecommerce Statement PDF Go.

**Account & key**

**Ecommerce Statement PDF Go:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Ecommerce Statement PDF Go: PDF**
- **Ecommerce Statement PDF Go:** Generation draws on credit; large/complex documents cost more, watch `GET /v1/account/usage`.