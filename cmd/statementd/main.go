package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/ecommerce-statement/internal/infrai"
	"example.com/ecommerce-statement/internal/statement"
)

type server struct{ client *infrai.Client }

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	baseURL := os.Getenv("INFRAI_BASE_URL")
	if baseURL == "" {
		baseURL = infrai.DefaultBaseURL
	}

	s := &server{client: infrai.New(baseURL, key)}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /statements", s.createStatement)
	addr := ":8080"
	log.Printf("statementd listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *server) createStatement(w http.ResponseWriter, r *http.Request) {
	var input statement.Request
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid statement request"})
		return
	}
	if input.CustomerName == "" || input.Currency == "" || !input.PeriodEnd.After(input.PeriodStart) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "customer_name, currency, and an ordered period are required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	usage, err := s.client.UsageTimeseries(ctx)
	if err != nil {
		s.writeUpstreamError(w, err)
		return
	}
	markdown := statement.Markdown(input, usage)
	digest := sha256.Sum256([]byte(markdown))
	pdf, err := s.client.GeneratePDF(ctx, markdown, "statement-"+hex.EncodeToString(digest[:]))
	if err != nil {
		s.writeUpstreamError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]json.RawMessage{"pdf": pdf})
}

func (s *server) writeUpstreamError(w http.ResponseWriter, err error) {
	var apiErr *infrai.APIError
	if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
		writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Code, "message": apiErr.Message})
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": "statement provider unavailable"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}
