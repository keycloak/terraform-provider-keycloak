package provider

import (
	"context"
	"fmt"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkterraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestKeycloakClientScopeValidation(t *testing.T) {
	saml := []interface{}{map[string]interface{}{}}
	openidConnect := []interface{}{map[string]interface{}{}}
	for name, raw := range map[string]map[string]interface{}{
		"no protocol block":    {},
		"both protocol blocks": {"saml": saml, "openid_connect": openidConnect},
		"modeled extra_config": {"saml": saml, "extra_config": map[string]interface{}{"gui.order": "1"}},
		"empty extra_config":   {"saml": saml, "extra_config": map[string]interface{}{"custom": " \t"}},
	} {
		raw["realm_id"], raw["name"] = "realm", "custom"
		if diags := resourceKeycloakClientScope().Validate(sdkterraform.NewResourceConfigRaw(raw)); !diags.HasError() {
			t.Errorf("%s: expected validation error", name)
		}
	}
}

func TestKeycloakClientScopeProtocolBlocks(t *testing.T) {
	r := resourceKeycloakClientScope()
	data := schema.TestResourceDataRaw(t, r.Schema, nil)
	if err := setKeycloakClientScopeData(data, &keycloak.ClientScope{Id: "scope-id", RealmId: "realm", Name: "custom", Protocol: "saml"}); err != nil {
		t.Fatal(err)
	}
	samlConfig := map[string]interface{}{"realm_id": "realm", "name": "custom", "saml": []interface{}{map[string]interface{}{}}}
	diff, err := r.Diff(context.Background(), data.State(), sdkterraform.NewResourceConfigRaw(samlConfig), nil)
	if err != nil {
		t.Fatal(err)
	}
	if diff != nil && !diff.Empty() {
		t.Fatalf("expected empty saml diff, got %#v", diff.Attributes)
	}

	oidcConfig := map[string]interface{}{"realm_id": "realm", "name": "custom", "openid_connect": []interface{}{map[string]interface{}{}}}
	diff, err = r.Diff(context.Background(), data.State(), sdkterraform.NewResourceConfigRaw(oidcConfig), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !diff.RequiresNew() {
		t.Fatal("switching protocol blocks must replace the scope")
	}

	if err := setKeycloakClientScopeData(data, &keycloak.ClientScope{Id: "scope-id", RealmId: "realm", Name: "custom", Protocol: "openid-connect"}); err != nil {
		t.Fatal(err)
	}
	oidcConfig["openid_connect"] = []interface{}{map[string]interface{}{"include_in_token_scope": false}}
	diff, err = r.Diff(context.Background(), data.State(), sdkterraform.NewResourceConfigRaw(oidcConfig), nil)
	if err != nil {
		t.Fatal(err)
	}
	if diff.Empty() || diff.RequiresNew() {
		t.Fatalf("expected in-place update for openid_connect fields, got %#v", diff)
	}

	if err := setKeycloakClientScopeData(data, &keycloak.ClientScope{Id: "scope-id", RealmId: "realm", Name: "custom", Protocol: "docker-v2"}); err == nil {
		t.Fatal("expected unsupported protocol error")
	}
}

func TestKeycloakClientScopeReadKeycloakDefaults(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceKeycloakClientScope().Schema, nil)
	err := setKeycloakClientScopeData(data, &keycloak.ClientScope{
		Id: "scope-id", RealmId: "realm", Name: "legacy", Protocol: "openid-connect",
		Attributes: map[string]string{"gui.order": "", "include.in.token.scope": "false"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]interface{}{
		"display_on_consent_screen":                            true,
		"gui_order":                                            0,
		"openid_connect.0.include_in_token_scope":              false,
		"openid_connect.0.include_in_openid_provider_metadata": true,
	} {
		if got := data.Get(key); got != want {
			t.Errorf("%s: got %v, want %v", key, got, want)
		}
	}
}

func TestAccKeycloakClientScope_openidConnect(t *testing.T) {
	testAccKeycloakClientScopeProtocol(t, keycloakClientScopeProtocolOpenidConnect)
}

func TestAccKeycloakClientScope_saml(t *testing.T) {
	testAccKeycloakClientScopeProtocol(t, keycloakClientScopeProtocolSaml)
}

func testAccKeycloakClientScopeProtocol(t *testing.T, protocol string) {
	t.Helper()
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_client_scope.scope"
	protocolBlock := map[string]string{
		keycloakClientScopeProtocolOpenidConnect: "openid_connect {\n\t\tinclude_in_token_scope = false\n\t}",
		keycloakClientScopeProtocolSaml:          "saml {}",
	}[protocol]
	config := func(description, consentScreenText string, guiOrder int, customAttribute string) string {
		extraConfig := "{}"
		if customAttribute != "" {
			extraConfig = fmt.Sprintf(`{"custom.attribute" = %q}`, customAttribute)
		}
		return fmt.Sprintf(`
resource "keycloak_client_scope" "scope" {
	realm_id            = %q
	name                = %q
	description         = %q
	consent_screen_text = %q
	gui_order           = %d
	extra_config        = %s

	%s
}
`, testAccRealm.Realm, name, description, consentScreenText, guiOrder, extraConfig, protocolBlock)
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakUnifiedClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: config("scope description", "consent", 3, "before"),
				Check: testAccCheckKeycloakUnifiedClientScope(resourceName, protocol, "scope description", map[string]string{
					"consent.screen.text": "consent", "display.on.consent.screen": "true", "gui.order": "3", "custom.attribute": "before",
				}),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					resourceState, ok := state.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					return fmt.Sprintf("realm/%s/%s", resourceState.Primary.Attributes["realm_id"], resourceState.Primary.ID), nil
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"extra_config"},
			},
			{
				Config: config("updated description", "", 0, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUnifiedClientScope(resourceName, protocol, "updated description", map[string]string{
						"consent.screen.text": "", "display.on.consent.screen": "false", "gui.order": "", "custom.attribute": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "extra_config.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "display_on_consent_screen", "false"),
				),
			},
		},
	})
}

// testAccCheckKeycloakUnifiedClientScope checks scope attributes; an empty wanted value means the key must be absent or empty.
func testAccCheckKeycloakUnifiedClientScope(resourceName, protocol, description string, attributes map[string]string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		scope, err := keycloakClient.GetClientScope(testCtx, resourceState.Primary.Attributes["realm_id"], resourceState.Primary.ID)
		if err != nil {
			return err
		}
		if scope.Protocol != protocol {
			return fmt.Errorf("scope protocol mismatch: got %q, want %q", scope.Protocol, protocol)
		}
		if scope.Description != description {
			return fmt.Errorf("scope description mismatch: got %q, want %q", scope.Description, description)
		}
		if protocol == keycloakClientScopeProtocolOpenidConnect {
			attributes = maps.Clone(attributes)
			attributes["include.in.token.scope"] = "false"
			attributes["include.in.openid.provider.metadata"] = "true"
		}
		for key, want := range attributes {
			if got := scope.Attributes[key]; got != want {
				return fmt.Errorf("attribute %q mismatch: got %q, want %q", key, got, want)
			}
		}
		return nil
	}
}

func testAccCheckKeycloakUnifiedClientScopeDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, resourceState := range state.RootModule().Resources {
			if resourceState.Type != "keycloak_client_scope" {
				continue
			}
			_, err := keycloakClient.GetClientScope(testCtx, resourceState.Primary.Attributes["realm_id"], resourceState.Primary.ID)
			if err == nil {
				return fmt.Errorf("client scope %s still exists", resourceState.Primary.ID)
			}
			if !keycloak.ErrorIs404(err) {
				return fmt.Errorf("checking client scope deletion: %w", err)
			}
		}
		return nil
	}
}
