#!/bin/sh
set -eu

curl --fail-with-body --request POST http://localhost:8080/statements \
  --header 'Content-Type: application/json' \
  --data @- <<'JSON'
{
  "customer_name": "Ada Lovelace",
  "currency": "USD",
  "period_start": "2026-09-01T00:00:00Z",
  "period_end": "2026-10-01T00:00:00Z",
  "orders": [
    {
      "id": "ord-1042",
      "checked_out_at": "2026-09-02T09:15:00Z",
      "fulfilled_at": "2026-09-03T14:10:00Z",
      "receipt_number": "rcpt-8831",
      "total_cents": 1299,
      "customer_updates": [
        {"at": "2026-09-02T11:00:00Z", "message": "Order packed"},
        {"at": "2026-09-03T14:10:00Z", "message": "Parcel delivered"}
      ]
    }
  ]
}
JSON
