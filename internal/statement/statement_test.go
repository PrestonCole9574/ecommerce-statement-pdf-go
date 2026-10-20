package statement

import (
	"strings"
	"testing"
	"time"
)

func TestSummarizeIncludesOnlyFulfilledOrdersWithReceipts(t *testing.T) {
	fulfilled := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		in   Request
		want Summary
	}{
		{
			name: "settled and pending orders",
			in: Request{Orders: []Order{
				{ID: "ord-settled", FulfilledAt: &fulfilled, ReceiptNumber: "rcpt-19", TotalCents: 1299},
				{ID: "ord-packing", ReceiptNumber: "rcpt-20", TotalCents: 800},
				{ID: "ord-no-receipt", FulfilledAt: &fulfilled, TotalCents: 500},
			}},
			want: Summary{CompletedOrders: 1, TotalCents: 1299},
		},
		{name: "empty period", in: Request{}, want: Summary{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Summarize(tt.in); got != tt.want {
				t.Fatalf("Summarize() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestMarkdownCarriesUsageAndCustomerUpdate(t *testing.T) {
	fulfilled := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	req := Request{CustomerName: "Ada", Currency: "USD", Orders: []Order{{
		ID: "ord-1", FulfilledAt: &fulfilled, ReceiptNumber: "rcpt-1", TotalCents: 2100,
		CustomerUpdates: []Update{{At: fulfilled, Message: "Parcel delivered"}},
	}}}
	got := Markdown(req, []byte(`{"series":[{"value":7}]}`))
	for _, want := range []string{"21.00 USD", "Parcel delivered", `"value":7`} {
		if !strings.Contains(got, want) {
			t.Fatalf("markdown missing %q", want)
		}
	}
}
