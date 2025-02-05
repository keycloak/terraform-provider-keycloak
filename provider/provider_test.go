package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"log"
	"os"
	"testing"
	"time"
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
	var err error
	// Load environment variables from a json file if it exists
	// This is useful for running tests locally

	if _, err := os.Stat("../test_env.json"); err == nil {
		println("Using test_env.json to load environment variables...")
		file, err := os.Open("../test_env.json")
		if err != nil {
			log.Fatalf("Unable to open env.json: %s", err)
		}
		defer file.Close()

		var envVars map[string]string
		if err := json.NewDecoder(file).Decode(&envVars); err != nil {
			log.Fatalf("Unable to decode env.json: %s", err)
		}

		for key, value := range envVars {
			if err := os.Setenv(key, value); err != nil {
				log.Fatalf("Unable to set environment variable %s: %s", key, err)
			}
		}
	}

	keycloakClient, err = keycloak.NewKeycloakClient(testCtx, os.Getenv("KEYCLOAK_URL"), "", os.Getenv("KEYCLOAK_ADMIN_URL"), os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), "", "", true, 5, "", false, userAgent, false, map[string]string{
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
	for _, requiredEnvironmentVariable := range requiredEnvironmentVariables {
		if value := os.Getenv(requiredEnvironmentVariable); value == "" {
			t.Fatalf("%s must be set before running acceptance tests.", requiredEnvironmentVariable)
		}
	}
}
