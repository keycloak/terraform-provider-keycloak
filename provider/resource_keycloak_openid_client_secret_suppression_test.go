package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkterraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakOpenidClient_suppressSecretDuringAccessTypeDrift(t *testing.T) {
	clientID := acctest.RandomWithPrefix("tf-acc")
	config := fmt.Sprintf(`resource "keycloak_openid_client" "client" {
 realm_id = "%s"
 client_id = "%s"
 access_type = "PUBLIC"
 standard_flow_enabled = false
 implicit_flow_enabled = false
 direct_access_grants_enabled = false
 service_accounts_enabled = false
 suppress_client_secret = true
}`, testAccRealm.Realm, clientID)
	check := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("keycloak_openid_client.client", "access_type", "PUBLIC"),
		resource.TestCheckResourceAttr("keycloak_openid_client.client", "suppress_client_secret", "true"),
		resource.TestCheckResourceAttr("keycloak_openid_client.client", "client_secret", ""),
	)
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{Config: config, Check: check},
			{RefreshState: true, ExpectNonEmptyPlan: true, PreConfig: func() {
				client, err := keycloakClient.GetOpenidClientByClientId(testCtx, testAccRealm.Realm, clientID)
				if err != nil {
					t.Fatal(err)
				}
				client.PublicClient = false
				if err := keycloakClient.UpdateOpenidClient(testCtx, client); err != nil {
					t.Fatal(err)
				}
				if _, err := keycloakClient.RegenerateOpenIdClientSecret(testCtx, client); err != nil {
					t.Fatal(err)
				}
			}, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("keycloak_openid_client.client", "access_type", "CONFIDENTIAL"),
				resource.TestCheckResourceAttr("keycloak_openid_client.client", "client_secret", ""),
			)},
			{Config: config, Check: check},
		},
	})
}

func TestKeycloakOpenidClient_suppressionDefaultFromLegacyState(t *testing.T) {
	data, err := schema.InternalMap(resourceKeycloakOpenidClient().Schema).Data(&sdkterraform.InstanceState{
		ID:         "legacy-client",
		Attributes: map[string]string{"client_id": "legacy", "realm_id": "example", "access_type": "PUBLIC"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &keycloak.OpenidClient{Id: "legacy-client", ClientId: "legacy", RealmId: "example", PublicClient: true}
	if err := setOpenidClientData(context.Background(), nil, data, client); err != nil {
		t.Fatal(err)
	}
	value, exists := data.State().Attributes["suppress_client_secret"]
	if !exists || value != "false" {
		t.Fatalf("refresh must persist false for older state, got %q (present=%t)", value, exists)
	}
}

func TestKeycloakOpenidClient_dataSourceWithoutSuppressionOption(t *testing.T) {
	data, err := schema.InternalMap(dataSourceKeycloakOpenidClient().Schema).Data(&sdkterraform.InstanceState{
		ID:         "existing-client",
		Attributes: map[string]string{"client_id": "existing", "realm_id": "example", "access_type": "PUBLIC"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &keycloak.OpenidClient{Id: "existing-client", ClientId: "existing", RealmId: "example", PublicClient: true}
	if err := setOpenidClientData(context.Background(), nil, data, client); err != nil {
		t.Fatal(err)
	}
	if _, exists := data.State().Attributes["suppress_client_secret"]; exists {
		t.Fatal("data source must not acquire the resource-only suppression field")
	}
}

func TestAccKeycloakOpenidClient_suppressComputedSecretAndRotateWriteOnly(t *testing.T) {
	t.Parallel()
	clientID := acctest.RandomWithPrefix("tf-acc")
	firstSecret := acctest.RandomWithPrefix("tf-acc")
	secondSecret := acctest.RandomWithPrefix("tf-acc")
	config := func(suppress bool, arguments string) string {
		return fmt.Sprintf(`resource "keycloak_openid_client" "client" {
 realm_id = "%s"
 client_id = "%s"
 access_type = "CONFIDENTIAL"
 suppress_client_secret = %t
 %s
}`, testAccRealm.Realm, clientID, suppress, arguments)
	}
	originalSecret := ""
	checkRemoteSecret := func(expected *string) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			client, err := keycloakClient.GetOpenidClientByClientId(testCtx, testAccRealm.Realm, clientID)
			if err != nil {
				return err
			}
			if client.ClientSecret == "" {
				return fmt.Errorf("expected a nonempty remote client secret")
			}
			if *expected == "" {
				*expected = client.ClientSecret
			} else if client.ClientSecret != *expected {
				return fmt.Errorf("remote client secret did not match the expected credential")
			}
			return nil
		}
	}
	suppressed := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("keycloak_openid_client.client", "suppress_client_secret", "true"),
		resource.TestCheckResourceAttr("keycloak_openid_client.client", "client_secret", ""),
		resource.TestCheckNoResourceAttr("keycloak_openid_client.client", "client_secret_wo"),
	)
	writeOnly := func(secret, version string) string {
		return config(true, fmt.Sprintf("client_secret_wo = %q\nclient_secret_wo_version = %q", secret, version))
	}
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{Config: config(false, ""), Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet("keycloak_openid_client.client", "client_secret"),
				checkRemoteSecret(&originalSecret),
			)},
			// A previously computed secret is allowed and is removed from state, without rotation.
			{Config: config(true, ""), Check: resource.ComposeTestCheckFunc(suppressed, checkRemoteSecret(&originalSecret))},
			{Config: config(true, fmt.Sprintf("client_secret = %q", firstSecret)),
				ExpectError: regexp.MustCompile("suppress_client_secret cannot be combined with client_secret")},
			{Config: writeOnly(firstSecret, "v1"), Check: resource.ComposeTestCheckFunc(
				suppressed, checkRemoteSecret(&firstSecret),
				resource.TestCheckResourceAttr("keycloak_openid_client.client", "client_secret_wo_version", "v1"),
			)},
			// Changing only the write-only value must not rotate the remote credential.
			{Config: writeOnly(secondSecret, "v1"), Check: resource.ComposeTestCheckFunc(suppressed, checkRemoteSecret(&firstSecret))},
			{Config: writeOnly(secondSecret, "v2"), Check: resource.ComposeTestCheckFunc(
				suppressed, checkRemoteSecret(&secondSecret),
				resource.TestCheckResourceAttr("keycloak_openid_client.client", "client_secret_wo_version", "v2"),
			)},
		},
	})
}
