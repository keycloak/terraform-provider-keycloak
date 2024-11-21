package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"os"
	"testing"
)

var testAccProviderFactories map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider
var keycloakClient *keycloak.KeycloakClient
var testAccRealm *keycloak.Realm
var testAccRealmTwo *keycloak.Realm
var testAccRealmUserFederation *keycloak.Realm
var testCtx context.Context

var requiredEnvironmentVariables = []string{
	"KEYCLOAK_CLIENT_ID",
	"KEYCLOAK_CLIENT_SECRET",
	"KEYCLOAK_REALM",
	"KEYCLOAK_URL",
}

func init() {
	testCtx = context.Background()
	userAgent := fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io) Terraform Plugin SDK/%s", schema.Provider{}.TerraformVersion, meta.SDKVersionString())
	keycloakClient, _ = keycloak.NewKeycloakClient(testCtx, os.Getenv("KEYCLOAK_URL"), "", os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), "", "", true, 5, "", false, userAgent, false, map[string]string{
		"foo": "bar",
	})
	testAccProvider = KeycloakProvider(keycloakClient)
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"keycloak": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func TestMain(m *testing.M) {
	testAccRealm = createTestRealm(testCtx)
	testAccRealmTwo = createTestRealm(testCtx)
	testAccRealmUserFederation = createTestRealm(testCtx)

	code := m.Run()

	err := keycloakClient.DeleteRealm(testCtx, testAccRealm.Realm)
	if err != nil {
		os.Exit(1)
	}

	err = keycloakClient.DeleteRealm(testCtx, testAccRealmTwo.Realm)
	if err != nil {
		os.Exit(1)
	}

	err = keycloakClient.DeleteRealm(testCtx, testAccRealmUserFederation.Realm)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(code)
}

func createTestRealm(testCtx context.Context) *keycloak.Realm {
	name := acctest.RandomWithPrefix("tf-acc")
	r := &keycloak.Realm{
		Id:      name,
		Realm:   name,
		Enabled: true,
	}

	err := keycloakClient.NewRealm(testCtx, r)
	if err != nil {
		os.Exit(1)
	}

	return r
}

func TestProvider(t *testing.T) {
	t.Parallel()

	if err := testAccProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	for _, requiredEnvironmentVariable := range requiredEnvironmentVariables {
		if value := os.Getenv(requiredEnvironmentVariable); value == "" {
			t.Fatalf("%s must be set before running acceptance tests.", requiredEnvironmentVariable)
		}
	}
}
