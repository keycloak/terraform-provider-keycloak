package keycloak

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestNewHttpClient_keepsCookiesAcrossRedirects(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc", Path: "/"})
		http.Redirect(w, r, "/target", http.StatusFound)
	})
	mux.HandleFunc("/target", func(w http.ResponseWriter, r *http.Request) {
		cookies := r.Header.Values("Cookie")
		if len(cookies) != 1 || cookies[0] != "session=abc" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, strings.Join(cookies, "; "))
			return
		}
		_, _ = io.WriteString(w, "ok")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	httpClient, err := newHttpClient(false, 15, "", "", "")
	if err != nil {
		t.Fatalf("newHttpClient: %v", err)
	}

	// The cookie set by the redirect response must reach the redirect
	// target, and later requests must send it exactly once.
	for _, path := range []string{"/start", "/target"} {
		resp, err := httpClient.Get(server.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s: expected status 200, got %d (Cookie headers %q)", path, resp.StatusCode, body)
		}
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

func TestNewHttpClient_retryAfterCookieRotationSendsOnlyNewCookie(t *testing.T) {
	attempts := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "old", Path: "/"})
		_, _ = io.WriteString(w, "ok")
	})
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		attempts++
		cookies := r.Header.Values("Cookie")
		if attempts == 1 {
			// Rotate the session cookie on a retryable response.
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "new", Path: "/"})
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		if len(cookies) != 1 || cookies[0] != "extra=1; session=new" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, strings.Join(cookies, " | "))
			return
		}
		_, _ = io.WriteString(w, "ok")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	httpClient, err := newHttpClient(false, 15, "", "", "")
	if err != nil {
		t.Fatalf("newHttpClient: %v", err)
	}

	resp, err := httpClient.Get(server.URL + "/login")
	if err != nil {
		t.Fatalf("GET /login: %v", err)
	}
	resp.Body.Close()

	// A cookie set by the caller must be kept on every attempt, while the
	// jar cookies must be rebuilt, so the retry sends the rotated value only.
	req, err := http.NewRequest(http.MethodGet, server.URL+"/api", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "extra", Value: "1"})
	resp, err = httpClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 after retry, got %d (Cookie headers %q)", resp.StatusCode, body)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
	if got := req.Header.Values("Cookie"); len(got) != 1 || got[0] != "extra=1" {
		t.Errorf("caller's request was modified, Cookie headers %q", got)
	}
}

func TestNewHttpClient_redirectReplacesCallerCookie(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		// Rotate one caller cookie and delete another, then redirect.
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "new", Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "gone", Value: "", Path: "/", MaxAge: -1})
		http.Redirect(w, r, "/target", http.StatusFound)
	})
	mux.HandleFunc("/target", func(w http.ResponseWriter, r *http.Request) {
		cookies := r.Header.Values("Cookie")
		if len(cookies) != 1 || cookies[0] != "extra=1; session=new" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, strings.Join(cookies, " | "))
			return
		}
		_, _ = io.WriteString(w, "ok")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	httpClient, err := newHttpClient(false, 15, "", "", "")
	if err != nil {
		t.Fatalf("newHttpClient: %v", err)
	}

	// Cookies the caller set (e.g. through additional_headers) must be
	// replaced or removed when a redirect response sets the same name.
	req, err := http.NewRequest(http.MethodGet, server.URL+"/start", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Cookie", "session=old; gone=1; extra=1")
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("GET /start: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d (Cookie headers %q)", resp.StatusCode, body)
	}
}
