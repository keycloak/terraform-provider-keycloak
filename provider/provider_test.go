package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/keycloak/terraform-provider-keycloak/helper"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

var testAccProviderFactories map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider
var keycloakClient *keycloak.KeycloakClient
var testAccRealm *keycloak.Realm
var testAccRealmTwo *keycloak.Realm
var testAccRealmUserFederation *keycloak.Realm
var testAccRealmOrganization *keycloak.Realm
var testCtx context.Context

func init() {
	testCtx = context.Background()
	userAgent := fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io) Terraform Plugin SDK/%s", schema.Provider{}.TerraformVersion, meta.SDKVersionString())
	var err error
	// Load environment variables from a json file if it exists
	// This is useful for running tests locally

	helper.UpdateEnvFromTestEnvIfPresent()

	initialLogin := os.Getenv("KEYCLOAK_ACCESS_TOKEN") == ""
	keycloakClient, err = keycloak.NewKeycloakClient(testCtx, os.Getenv("KEYCLOAK_URL"), "", os.Getenv("KEYCLOAK_ADMIN_URL"), os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), "", "", os.Getenv("KEYCLOAK_ACCESS_TOKEN"), "", "", initialLogin, 120, os.Getenv("KEYCLOAK_TLS_CA_CERT"), false, os.Getenv("KEYCLOAK_TLS_CLIENT_CERT"), os.Getenv("KEYCLOAK_TLS_CLIENT_KEY"), userAgent, false, map[string]string{
		"foo": "bar",
	})
	if err != nil {
		panic(err)
	}
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

	// Clean up of tests is not fatal if it fails
	err := keycloakClient.DeleteRealm(testCtx, testAccRealm.Realm)
	if err != nil {
		log.Printf("Unable to delete realm %s: %s", testAccRealmUserFederation.Realm, err)
	}

	err = keycloakClient.DeleteRealm(testCtx, testAccRealmTwo.Realm)
	if err != nil {
		log.Printf("Unable to delete realm %s: %s", testAccRealmUserFederation.Realm, err)
	}

	err = keycloakClient.DeleteRealm(testCtx, testAccRealmUserFederation.Realm)
	if err != nil {
		log.Printf("Unable to delete realm %s: %s", testAccRealmUserFederation.Realm, err)
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

	var err error

	validVersion, err := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_26)
	if err != nil {
		log.Printf("Unable to check keycloak version: %s", err)
	}
	if validVersion {
		r.OrganizationsEnabled = true
	}

	for i := 0; i < 3; i++ { // on CI this sometimes fails and keycloak can't be reached
		err = keycloakClient.NewRealm(testCtx, r)
		if err != nil {
			log.Printf("Unable to create new realm: %s - retrying in 5s", err)
			time.Sleep(5 * time.Second) // 24.0.5 on CI seems to have issues creating a realm when locking the table
		} else {
			break
		}
	}
	if err != nil {
		log.Fatalf("Unable to create new realm: %s", err)
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
	helper.CheckRequiredEnvironmentVariables(t)
}

func testAccPreCheckDynamicScopes(t *testing.T) {
	testAccPreCheck(t)

	// Try to create a test dynamic scope to check if the feature is enabled
	testScope := &keycloak.OpenidClientScope{
		RealmId:            testAccRealm.Realm,
		Name:               "test-dynamic-feature-check",
		Dynamic:            true,
		DynamicScopeRegexp: "test:*",
	}

	err := keycloakClient.NewOpenidClientScope(testCtx, testScope)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "400 Bad Request" ||
			(errMsg != "" && (strings.Contains(errMsg, "Unexpected value") || strings.Contains(errMsg, "dynamic.scope.regexp"))) {
			t.Skip("Skipping test: Keycloak dynamic scopes feature is not enabled. Enable with KC_FEATURES=dynamic-scopes:v1")
		}
		// Other errors should fail the test
		t.Fatalf("Unexpected error checking dynamic scopes support: %v", err)
	}

	// Clean up the test scope
	_ = keycloakClient.DeleteOpenidClientScope(testCtx, testScope.RealmId, testScope.Id)
}
