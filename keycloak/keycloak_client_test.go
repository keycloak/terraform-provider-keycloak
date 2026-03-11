package keycloak

import (
	"context"
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
