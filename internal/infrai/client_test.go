package infrai

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestGeneratePDFRetries429WithSameIdempotencyKey(t *testing.T) {
	requests := 0
	c := New("https://example.invalid", "test-key")
	c.sleep = func(context.Context, time.Duration) error { return nil }
	c.http.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		if req.Method != http.MethodPost || req.URL.Path != "/v1/pdf/generate" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		if req.Header.Get("Authorization") != "Bearer test-key" || req.Header.Get("Idempotency-Key") != "statement-abc" {
			t.Fatal("request headers are incomplete")
		}
		status, body := http.StatusOK, `{"ok":true,"data":{"job_id":"job-1"},"error":null,"metadata":{}}`
		if requests == 1 {
			status, body = http.StatusTooManyRequests, `{"ok":false,"data":null,"error":{"code":"RATE_LIMITED","message":"retry later"},"metadata":{}}`
		}
		return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})

	data, err := c.GeneratePDF(context.Background(), "# Statement", "statement-abc")
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 || !strings.Contains(string(data), "job-1") {
		t.Fatalf("requests=%d data=%s", requests, data)
	}
}
