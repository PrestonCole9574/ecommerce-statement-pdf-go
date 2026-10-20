package statement

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Order struct {
	ID              string     `json:"id"`
	CheckedOutAt    time.Time  `json:"checked_out_at"`
	FulfilledAt     *time.Time `json:"fulfilled_at,omitempty"`
	ReceiptNumber   string     `json:"receipt_number,omitempty"`
	TotalCents      int64      `json:"total_cents"`
	CustomerUpdates []Update   `json:"customer_updates,omitempty"`
}

type Update struct {
	At      time.Time `json:"at"`
	Message string    `json:"message"`
}

type Request struct {
	CustomerName string    `json:"customer_name"`
	Currency     string    `json:"currency"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	Orders       []Order   `json:"orders"`
}

type Summary struct {
	CompletedOrders int
	TotalCents      int64
}

func Summarize(req Request) Summary {
	var result Summary
	for _, order := range req.Orders {
		if order.FulfilledAt == nil || order.ReceiptNumber == "" {
			continue
		}
		result.CompletedOrders++
		result.TotalCents += order.TotalCents
	}
	return result
}

func Markdown(req Request, usageJSON []byte) string {
	summary := Summarize(req)
	orders := append([]Order(nil), req.Orders...)
	sort.Slice(orders, func(i, j int) bool { return orders[i].CheckedOutAt.Before(orders[j].CheckedOutAt) })

	var b strings.Builder
	fmt.Fprintf(&b, "# Account statement\n\nCustomer: %s  \nPeriod: %s to %s  \nCurrency: %s\n\n", req.CustomerName, req.PeriodStart.Format("2006-01-02"), req.PeriodEnd.Format("2006-01-02"), req.Currency)
	fmt.Fprintf(&b, "## Settled orders\n\n| Checkout | Order | Receipt | Fulfilled | Amount |\n|---|---|---|---|---:|\n")
	for _, order := range orders {
		if order.FulfilledAt == nil || order.ReceiptNumber == "" {
			continue
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %.2f %s |\n", order.CheckedOutAt.Format("2006-01-02"), order.ID, order.ReceiptNumber, order.FulfilledAt.Format("2006-01-02"), float64(order.TotalCents)/100, req.Currency)
		for _, update := range order.CustomerUpdates {
			fmt.Fprintf(&b, "| | update | | %s | %s |\n", update.At.Format("2006-01-02"), update.Message)
		}
	}
	fmt.Fprintf(&b, "\nCompleted orders: %d  \nStatement total: %.2f %s\n\n", summary.CompletedOrders, float64(summary.TotalCents)/100, req.Currency)
	fmt.Fprintf(&b, "## Metered account usage\n\n```json\n%s\n```\n", usageJSON)
	return b.String()
}
