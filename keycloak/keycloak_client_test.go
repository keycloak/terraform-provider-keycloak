package keycloak

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/keycloak/terraform-provider-keycloak/helper"
)

func init() {
	helper.UpdateEnvFromTestEnvIfPresent()
}

func TestAccKeycloakClientConnect(t *testing.T) {

	ctx := context.Background()

	helper.CheckRequiredEnvironmentVariables(t)

	clientTimeout := checkClientTimeout(t)

	keycloakClient, err := NewKeycloakClient(ctx, os.Getenv("KEYCLOAK_URL"), "", os.Getenv("KEYCLOAK_ADMIN_URL"), os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), os.Getenv("KEYCLOAK_USER"), os.Getenv("KEYCLOAK_PASSWORD"), os.Getenv("KEYCLOAK_ACCESS_TOKEN"), "", "", os.Getenv("KEYCLOAK_JWT_TOKEN"), os.Getenv("KEYCLOAK_JWT_TOKEN_FILE"), true, clientTimeout, os.Getenv("KEYCLOAK_TLS_CA_CERT"), true, os.Getenv("KEYCLOAK_TLS_CLIENT_CERT"), os.Getenv("KEYCLOAK_TLS_CLIENT_KEY"), "", false, map[string]string{
		"foo": "bar",
	})

	keycloakClientChecks(t, err, keycloakClient, ctx)
}

func TestAccKeycloakClientConnectAccessTokenAuth(t *testing.T) {

	ctx := context.Background()

	helper.CheckRequiredEnvironmentVariables(t)

	if os.Getenv("KEYCLOAK_ACCESS_TOKEN") == "" {
		t.Skip("Skipping: KEYCLOAK_ACCESS_TOKEN must be present to test auth with provided access token")
	}

	clientTimeout := checkClientTimeout(t)

	keycloakClient, err := NewKeycloakClient(ctx, os.Getenv("KEYCLOAK_URL"), "", os.Getenv("KEYCLOAK_ADMIN_URL"), os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), os.Getenv("KEYCLOAK_USER"), os.Getenv("KEYCLOAK_PASSWORD"), os.Getenv("KEYCLOAK_ACCESS_TOKEN"), "", "", os.Getenv("KEYCLOAK_JWT_TOKEN"), os.Getenv("KEYCLOAK_JWT_TOKEN_FILE"), true, clientTimeout, os.Getenv("KEYCLOAK_TLS_CA_CERT"), true, os.Getenv("KEYCLOAK_TLS_CLIENT_CERT"), os.Getenv("KEYCLOAK_TLS_CLIENT_KEY"), "", false, map[string]string{
		"foo": "bar",
	})
	keycloakClientChecks(t, err, keycloakClient, ctx)
}

func TestAccKeycloakClientConnectHttpsMtlsAuth(t *testing.T) {

	ctx := context.Background()

	helper.CheckRequiredEnvironmentVariables(t)

	clientTimeout := checkClientTimeout(t)

	if os.Getenv("KEYCLOAK_TLS_CLIENT_CERT") == "" || os.Getenv("KEYCLOAK_TLS_CLIENT_KEY") == "" {
		t.Skip("Skipping: KEYCLOAK_TLS_CLIENT_CERT and KEYCLOAK_TLS_CLIENT_KEY must both be set to test mTLS")
	}

	// use the keycloak client with plain http to read Keycloak version
	keycloakHttpUrl := os.Getenv("KEYCLOAK_URL_HTTP")
	if keycloakHttpUrl == "" {
		if keycloakHttpUrl = os.Getenv("KEYCLOAK_URL"); strings.HasPrefix(keycloakHttpUrl, "https") {
			t.Fatalf("KEYCLOAK_URL_HTTP must also be set to https when using https")
		}
	}
	keycloakClient, err := NewKeycloakClient(ctx, keycloakHttpUrl, "", os.Getenv("KEYCLOAK_ADMIN_URL"), os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), os.Getenv("KEYCLOAK_USER"), os.Getenv("KEYCLOAK_PASSWORD"), os.Getenv("KEYCLOAK_ACCESS_TOKEN"), "", "", os.Getenv("KEYCLOAK_JWT_TOKEN"), os.Getenv("KEYCLOAK_JWT_TOKEN_FILE"), true, clientTimeout, "", true, "", "", "", false, map[string]string{})

	_, err = keycloakClient.Version(ctx)
	if err != nil {
		t.Fatalf("%s", err)
	}

	keycloakUrl := os.Getenv("KEYCLOAK_URL")
	if !strings.HasPrefix(keycloakUrl, "https://") {
		// only run tests for https URL
		t.Skip("We only test mtls when Keycloak is used with an https:// url")
	}

	// then try again to connect with Keycloak but this time via https with mtls client auth
	mtlsKeycloakClient, err := NewKeycloakClient(ctx, keycloakUrl, "", os.Getenv("KEYCLOAK_ADMIN_URL"), os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), os.Getenv("KEYCLOAK_USER"), os.Getenv("KEYCLOAK_PASSWORD"), os.Getenv("KEYCLOAK_ACCESS_TOKEN"), "", "", os.Getenv("KEYCLOAK_JWT_TOKEN"), os.Getenv("KEYCLOAK_JWT_TOKEN_FILE"), true, clientTimeout, os.Getenv("KEYCLOAK_TLS_CA_CERT"), true, os.Getenv("KEYCLOAK_TLS_CLIENT_CERT"), os.Getenv("KEYCLOAK_TLS_CLIENT_KEY"), "", false, map[string]string{})
	keycloakClientChecks(t, err, mtlsKeycloakClient, ctx)
}

func TestNewKeycloakClient_clientIdValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name          string
		clientId      string
		clientSecret  string
		username      string
		password      string
		accessToken   string
		jwtSigningKey string
		jwtToken      string
		jwtTokenFile  string
		initialLogin  bool
		expectError   string
	}{
		{
			name:        "password grant without client_id",
			username:    "admin",
			password:    "password",
			expectError: "client_id is required for password grant",
		},
		{
			name:         "client secret without client_id",
			clientSecret: "secret",
			expectError:  "client_id is required for client secret authentication",
		},
		{
			name:          "jwt_signing_key without client_id",
			jwtSigningKey: "some-key",
			expectError:   "client_id is required when using jwt_signing_key because it is used for the JWT iss/sub claims",
		},
		{
			name:         "no credentials with initial_login",
			initialLogin: true,
			expectError:  "must specify client id",
		},
		{
			name:     "jwt_token without client_id succeeds",
			jwtToken: "some-jwt-token",
		},
		{
			name:         "jwt_token_file without client_id succeeds",
			jwtTokenFile: "/tmp/nonexistent",
		},
		{
			name:        "access_token without client_id succeeds",
			accessToken: "some-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewKeycloakClient(
				ctx,
				"http://localhost:8080", // url
				"",                      // basePath
				"",                      // adminUrl
				tt.clientId,
				tt.clientSecret,
				"master",    // realm
				tt.username, // username
				tt.password, // password
				tt.accessToken,
				"RS256", // jwtSigningAlg
				tt.jwtSigningKey,
				tt.jwtToken,
				tt.jwtTokenFile,
				tt.initialLogin,
				5,     // clientTimeout
				"",    // caCert
				true,  // tlsInsecureSkipVerify
				"",    // tlsClientCert
				"",    // tlsClientPrivateKey
				"",    // userAgent
				false, // redHatSSO
				map[string]string{},
			)

			if tt.expectError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q but got nil", tt.expectError)
				}
				if !strings.Contains(err.Error(), tt.expectError) {
					t.Errorf("expected error containing %q, got: %v", tt.expectError, err)
				}
			} else {
				// For success cases, we expect either no error or a connection error
				// (since there's no real Keycloak). The point is that validation passed.
				if err != nil && strings.Contains(err.Error(), "client_id is required") {
					t.Errorf("unexpected client_id validation error: %v", err)
				}
			}
		})
	}
}

func TestApplyAdditionalHeaders_HostHeader(t *testing.T) {
	keycloakClient := &KeycloakClient{
		additionalHeaders: map[string]string{
			"Host":         "custom.host.example.com",
			"X-Custom-Key": "custom-value",
		},
	}

	request, err := http.NewRequest(http.MethodGet, "http://localhost/test", nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %s", err)
	}

	keycloakClient.applyAdditionalHeaders(request)

	if request.Host != "custom.host.example.com" {
		t.Fatalf("expected request.Host to be 'custom.host.example.com', got '%s'", request.Host)
	}

	if request.Header.Get("Host") != "" {
		t.Fatal("expected Host to not be set in request.Header (Go uses request.Host instead)")
	}

	if request.Header.Get("X-Custom-Key") != "custom-value" {
		t.Fatalf("expected X-Custom-Key header to be 'custom-value', got '%s'", request.Header.Get("X-Custom-Key"))
	}
}

func TestApplyAdditionalHeaders_HostHeaderCaseInsensitive(t *testing.T) {
	variants := []string{"host", "HOST", "Host", "hOsT"}

	for _, hostKey := range variants {
		t.Run(hostKey, func(t *testing.T) {
			keycloakClient := &KeycloakClient{
				additionalHeaders: map[string]string{
					hostKey: "custom.host.example.com",
				},
			}

			request, err := http.NewRequest(http.MethodGet, "http://localhost/test", nil)
			if err != nil {
				t.Fatalf("unexpected error creating request: %s", err)
			}

			keycloakClient.applyAdditionalHeaders(request)

			if request.Host != "custom.host.example.com" {
				t.Fatalf("expected request.Host to be 'custom.host.example.com' for key '%s', got '%s'", hostKey, request.Host)
			}
		})
	}
}

func checkClientTimeout(t *testing.T) int {
	// Convert KEYCLOAK_CLIENT_TIMEOUT to int
	clientTimeout, err := strconv.Atoi(os.Getenv("KEYCLOAK_CLIENT_TIMEOUT"))
	if err != nil {
		t.Fatal("KEYCLOAK_CLIENT_TIMEOUT must be an integer")
	}
	return clientTimeout
}

func keycloakClientChecks(t *testing.T, err error, keycloakClient *KeycloakClient, ctx context.Context) {
	if err != nil {
		t.Fatalf("%s", err)
	}

	version, err := keycloakClient.Version(ctx)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if version == nil {
		t.Fatalf("%s", "Server Version not found")
	}
}
