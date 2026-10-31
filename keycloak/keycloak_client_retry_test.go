package keycloak

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestNewHttpClient_retriesRecoverableResponses(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) < 3 {
			// Retry-After: 0 keeps the backoff out of the test run time.
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, "ok")
	}))
	defer server.Close()

	httpClient, err := newHttpClient(false, 15, "", "", "")
	if err != nil {
		t.Fatalf("newHttpClient: %v", err)
	}

	resp, err := httpClient.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 after retries, got %d", resp.StatusCode)
	}
	if got := attempts.Load(); got != 3 {
		t.Errorf("expected 3 attempts, got %d", got)
	}
}

func TestNewHttpClient_returnsLastResponseWhenRetriesRunOut(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"error":"too many requests"}`)
	}))
	defer server.Close()

	httpClient, err := newHttpClient(false, 15, "", "", "")
	if err != nil {
		t.Fatalf("newHttpClient: %v", err)
	}

	resp, err := httpClient.Get(server.URL)
	if err != nil {
		t.Fatalf("expected the last response instead of an error, got: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	if string(body) != `{"error":"too many requests"}` {
		t.Errorf("unexpected body %q", body)
	}
	// One initial attempt plus RetryMax (5) retries.
	if got := attempts.Load(); got != 6 {
		t.Errorf("expected 6 attempts, got %d", got)
	}
}

func TestRetryPolicy_connectionErrorWithoutResponse(t *testing.T) {
	retry, err := RetryPolicy(context.Background(), nil, errors.New("connection refused"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !retry {
		t.Error("expected a connection error to be retried")
	}
}
