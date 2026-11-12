package keycloak

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/go-version"
)

// Regression tests for https://github.com/keycloak/terraform-provider-keycloak/issues/1691:
// the DELETE is retried once, and only for server errors. The first failure must not be lost.
func deleteExecutionServer(t *testing.T, statuses ...int) (*KeycloakClient, *int32) {
	t.Helper()

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(atomic.AddInt32(&requests, 1))
		if r.Method != http.MethodDelete || !strings.HasSuffix(r.URL.Path, "/realms/realm/authentication/executions/abc") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		status := statuses[len(statuses)-1]
		if n <= len(statuses) {
			status = statuses[n-1]
		}
		w.WriteHeader(status)
		if status >= 400 {
			w.Write([]byte(`{"error":"status ` + http.StatusText(status) + `"}`))
		}
	}))
	t.Cleanup(server.Close)

	return &KeycloakClient{
		baseUrl: server.URL, httpClient: server.Client(), initialLogin: true, accessTokenProvided: true,
		version:           version.Must(version.NewVersion("26.8.1")),
		clientCredentials: &ClientCredentials{AccessToken: "test", TokenType: "Bearer"},
	}, &requests
}

func TestDeleteAuthenticationExecutionDoesNotRetryClientError(t *testing.T) {
	client, requests := deleteExecutionServer(t, http.StatusBadRequest)

	err := client.DeleteAuthenticationExecution(context.Background(), "realm", "abc")
	if err == nil {
		t.Fatal("a 400 must fail the delete")
	}
	if got := atomic.LoadInt32(requests); got != 1 {
		t.Fatalf("a 400 must not be retried, got %d requests", got)
	}

	apiErr, ok := errwrap.GetType(err, &ApiError{}).(*ApiError)
	if !ok || apiErr.Code != http.StatusBadRequest {
		t.Fatalf("want the original 400 ApiError, got %v", err)
	}
	if strings.Contains(err.Error(), "first attempt failed") {
		t.Fatalf("a single attempt must not be reported as a retry, got %v", err)
	}
}

func TestDeleteAuthenticationExecutionRetriesServerError(t *testing.T) {
	client, requests := deleteExecutionServer(t, http.StatusInternalServerError, http.StatusNoContent)

	if err := client.DeleteAuthenticationExecution(context.Background(), "realm", "abc"); err != nil {
		t.Fatalf("a 500 followed by success must succeed, got %v", err)
	}
	if got := atomic.LoadInt32(requests); got != 2 {
		t.Fatalf("want exactly 2 requests, got %d", got)
	}
}

func TestDeleteAuthenticationExecutionKeepsBothErrorsWhenRetryFails(t *testing.T) {
	client, requests := deleteExecutionServer(t, http.StatusInternalServerError, http.StatusBadGateway)

	err := client.DeleteAuthenticationExecution(context.Background(), "realm", "abc")
	if err == nil {
		t.Fatal("a 500 followed by a 502 must fail the delete")
	}
	if got := atomic.LoadInt32(requests); got != 2 {
		t.Fatalf("want exactly 2 requests, got %d", got)
	}

	for _, want := range []string{http.StatusText(http.StatusInternalServerError), http.StatusText(http.StatusBadGateway), "first attempt failed"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("want %q in the error, got %v", want, err)
		}
	}

	apiErr, ok := errwrap.GetType(err, &ApiError{}).(*ApiError)
	if !ok || apiErr.Code != http.StatusBadGateway {
		t.Fatalf("want the retry's 502 ApiError discoverable through errwrap, got %v", err)
	}
}
