package keycloak

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/keycloak/terraform-provider-keycloak/helper"
)

var requiredEnvironmentVariables = []string{
	"KEYCLOAK_CLIENT_ID",
	"KEYCLOAK_URL",
	"KEYCLOAK_REALM",
}

func init() {
	helper.UpdateEnvFromTestEnvIfPresent()
}

func checkRequiredEnvironmentVariables(t *testing.T) {
	for _, requiredEnvironmentVariable := range requiredEnvironmentVariables {
		if value := os.Getenv(requiredEnvironmentVariable); value == "" {
			t.Fatalf("%s must be set before running acceptance tests.", requiredEnvironmentVariable)
		}
	}

	if v := os.Getenv("KEYCLOAK_CLIENT_SECRET"); v == "" {
		if v := os.Getenv("KEYCLOAK_USER"); v == "" {
			t.Fatal("KEYCLOAK_USER must be set for acceptance tests")
		}
		if v := os.Getenv("KEYCLOAK_PASSWORD"); v == "" {
			t.Fatal("KEYCLOAK_PASSWORD must be set for acceptance tests")
		}
	}
}

func TestAccKeycloakClientConnect(t *testing.T) {

	ctx := context.Background()

	checkRequiredEnvironmentVariables(t)

	clientTimeout := checkClientTimeout(t)

	keycloakClient, err := NewKeycloakClient(ctx, os.Getenv("KEYCLOAK_URL"), "", os.Getenv("KEYCLOAK_ADMIN_URL"), os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), os.Getenv("KEYCLOAK_USER"), os.Getenv("KEYCLOAK_PASSWORD"), "", "", true, clientTimeout, os.Getenv("KEYCLOAK_TLS_CA_CERT"), true, os.Getenv("KEYCLOAK_TLS_CLIENT_CERT"), os.Getenv("KEYCLOAK_TLS_CLIENT_KEY"), "", false, map[string]string{
		"foo": "bar",
	})
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

func TestAccKeycloakClientConnectHttpsMtlsAuth(t *testing.T) {

	ctx := context.Background()

	checkRequiredEnvironmentVariables(t)

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
	keycloakClient, err := NewKeycloakClient(ctx, keycloakHttpUrl, "", os.Getenv("KEYCLOAK_ADMIN_URL"), os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), os.Getenv("KEYCLOAK_USER"), os.Getenv("KEYCLOAK_PASSWORD"), "", "", true, clientTimeout, "", true, "", "", "", false, map[string]string{})

	version, err := keycloakClient.Version(ctx)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// skip test if running 26.0 or lower
	if v, _ := keycloakClient.VersionIsLessThanOrEqualTo(ctx, Version_26); v {
		t.Skip("We only test Keycloak > 26.0")
	}

	keycloakUrl := os.Getenv("KEYCLOAK_URL")
	if !strings.HasPrefix(keycloakUrl, "https://") {
		// only run tests for https URL
		t.Skip("We only test mtls when Keycloak is used with an https:// url")
	}

	// then try again to connect with Keycloak but this time via https with mtls client auth
	mtlsKeycloakClient, err := NewKeycloakClient(ctx, keycloakUrl, "", os.Getenv("KEYCLOAK_ADMIN_URL"), os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), os.Getenv("KEYCLOAK_USER"), os.Getenv("KEYCLOAK_PASSWORD"), "", "", true, clientTimeout, os.Getenv("KEYCLOAK_TLS_CA_CERT"), true, os.Getenv("KEYCLOAK_TLS_CLIENT_CERT"), os.Getenv("KEYCLOAK_TLS_CLIENT_KEY"), "", false, map[string]string{})
	if err != nil {
		t.Fatalf("%s", err)
	}

	version, err = mtlsKeycloakClient.Version(ctx)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if version == nil {
		t.Fatalf("%s", "Server Version not found")
	}
}

// Some actions, such as creating a realm, require a refresh
// before a GET can be performed on that realm
//
// This test ensures that, after creating a realm and performing
// a GET, the access token and refresh token have changed
//
// Any action that returns a 403 or a 401 could be used for this test
// Creating a realm is just the only one I'm aware of
//
// This appears to have been fixed as of Keycloak 12.x
func TestAccKeycloakApiClientRefresh(t *testing.T) {
	ctx := context.Background()

	checkRequiredEnvironmentVariables(t)

	clientTimeout := checkClientTimeout(t)

	keycloakClient, err := NewKeycloakClient(ctx, os.Getenv("KEYCLOAK_URL"), "", os.Getenv("KEYCLOAK_ADMIN_URL"), os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), os.Getenv("KEYCLOAK_USER"), os.Getenv("KEYCLOAK_PASSWORD"), "", "", true, clientTimeout, os.Getenv("KEYCLOAK_TLS_CA_CERT"), false, os.Getenv("KEYCLOAK_TLS_CLIENT_CERT"), os.Getenv("KEYCLOAK_TLS_CLIENT_KEY"), "", false, map[string]string{
		"foo": "bar",
	})
	if err != nil {
		t.Fatalf("%s", err)
	}

	// skip test if running 12.x or greater
	if v, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(ctx, Version_12); v {
		t.Skip()
	}

	realmName := "terraform-" + acctest.RandString(10)
	realm := &Realm{
		Realm: realmName,
		Id:    realmName,
	}

	err = keycloakClient.NewRealm(ctx, realm)
	if err != nil {
		t.Fatalf("%s", err)
	}

	var oldAccessToken, oldRefreshToken, oldTokenType string

	// A following GET for this realm will result in a 403, so we should save the current access and refresh token
	if keycloakClient.clientCredentials.GrantType == "client_credentials" {
		oldAccessToken = keycloakClient.clientCredentials.AccessToken
		oldRefreshToken = keycloakClient.clientCredentials.RefreshToken
		oldTokenType = keycloakClient.clientCredentials.TokenType
	}

	_, err = keycloakClient.GetRealm(ctx, realmName) // This should not fail since it will automatically refresh and try again
	if err != nil {
		t.Fatalf("%s", err)
	}

	// Clean up - the realm doesn't need to exist in order for us to assert against the refreshed tokens
	err = keycloakClient.DeleteRealm(ctx, realmName)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if keycloakClient.clientCredentials.GrantType == "client_credentials" {
		newAccessToken := keycloakClient.clientCredentials.AccessToken
		newRefreshToken := keycloakClient.clientCredentials.RefreshToken
		newTokenType := keycloakClient.clientCredentials.TokenType

		if oldAccessToken == newAccessToken {
			t.Fatalf("expected access token to update after refresh")
		}

		if oldRefreshToken == newRefreshToken {
			t.Fatalf("expected refresh token to update after refresh")
		}

		if oldTokenType != newTokenType {
			t.Fatalf("expected token type to remain the same after refresh")
		}
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
