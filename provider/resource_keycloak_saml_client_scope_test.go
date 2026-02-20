package provider

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakSamlClientScope_basic(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
			},
			{
				ResourceName:        "keycloak_saml_client_scope.client_scope",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakSamlClientScope_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var clientScope = &keycloak.SamlClientScope{}

	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClientScope_basic(clientScopeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
					testAccCheckKeycloakSamlClientScopeFetch("keycloak_saml_client_scope.client_scope", clientScope),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteSamlClientScope(testCtx, clientScope.RealmId, clientScope.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakSamlClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
			},
		},
	})
}

func TestAccKeycloakSamlClientScope_updateRealm(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClientScope_updateRealmBefore(clientScopeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
					testAccCheckKeycloakSamlClientScopeBelongsToRealm("keycloak_saml_client_scope.client_scope", testAccRealm.Realm),
				),
			},
			{
				Config: testKeycloakSamlClientScope_updateRealmAfter(clientScopeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
					testAccCheckKeycloakSamlClientScopeBelongsToRealm("keycloak_saml_client_scope.client_scope", testAccRealmTwo.Realm),
				),
			},
		},
	})
}

func TestAccKeycloakSamlClientScope_consentScreenText(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
			},
			{
				Config: testKeycloakSamlClientScope_withConsentText(clientScopeName, acctest.RandString(10)),
				Check:  testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
			},
			{
				Config: testKeycloakSamlClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
			},
		},
	})
}

func TestAccKeycloakSamlClientScope_guiOrder(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc")
	guiOrder := acctest.RandIntRange(0, 1000)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
			},
			{
				Config: testKeycloakSamlClientScope_withGuiOrder(clientScopeName, guiOrder),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
					testAccCheckKeycloakSamlClientScopeExistsWithCorrectGuiOrder("keycloak_saml_client_scope.client_scope", guiOrder),
				),
			},
			{
				Config: testKeycloakSamlClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
			},
		},
	})
}

func TestAccKeycloakSamlClientScope_extraConfig(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakSamlClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClientScope_withExtraConfigMap(clientScopeName, map[string]string{
					"key1": "value1",
					"key2": "value2",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
					testAccCheckKeycloakSamlClientScopeExtraConfigHasValue("keycloak_saml_client_scope.client_scope", "key1", "value1"),
					testAccCheckKeycloakSamlClientScopeExtraConfigHasValue("keycloak_saml_client_scope.client_scope", "key2", "value2"),
				),
			},
			{
				Config: testKeycloakSamlClientScope_withExtraConfigMap(clientScopeName, map[string]string{
					"key2": "value2",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol("keycloak_saml_client_scope.client_scope"),
					testAccCheckKeycloakSamlClientScopeExtraConfigHasValue("keycloak_saml_client_scope.client_scope", "key2", "value2"),
					testAccCheckKeycloakSamlClientScopeExtraConfigMissing("keycloak_saml_client_scope.client_scope", "key1"),
				),
			},
		},
	})
}

func testAccCheckKeycloakSamlClientScopeExistsWithCorrectProtocol(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clientScope, err := getSamlClientScopeFromState(s, resourceName)
		if err != nil {
			return err
		}

		if clientScope.Protocol != "saml" {
			return fmt.Errorf("expected saml client scope to have saml protocol, but got %s", clientScope.Protocol)
		}

		return nil
	}
}

func testAccCheckKeycloakSamlClientScopeExistsWithCorrectGuiOrder(resourceName string, guiOrder int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clientScope, err := getSamlClientScopeFromState(s, resourceName)
		if err != nil {
			return err
		}

		if clientScope.Attributes.GuiOrder != strconv.Itoa(guiOrder) {
			return fmt.Errorf("expected saml client guiOrder to have %d, but got %s", guiOrder, clientScope.Attributes.GuiOrder)
		}

		return nil
	}
}

func testAccCheckKeycloakSamlClientScopeFetch(resourceName string, clientScope *keycloak.SamlClientScope) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedClientScope, err := getSamlClientScopeFromState(s, resourceName)
		if err != nil {
			return err
		}

		clientScope.Id = fetchedClientScope.Id
		clientScope.RealmId = fetchedClientScope.RealmId

		return nil
	}
}

func testAccCheckKeycloakSamlClientScopeBelongsToRealm(resourceName, realm string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clientScope, err := getSamlClientScopeFromState(s, resourceName)
		if err != nil {
			return err
		}

		if clientScope.RealmId != realm {
			return fmt.Errorf("expected saml client scope %s to have realm_id of %s, but got %s", clientScope.Id, realm, clientScope.RealmId)
		}

		return nil
	}
}

func testAccCheckKeycloakSamlClientScopeDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_saml_client_scope" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			clientScope, _ := keycloakClient.GetSamlClientScope(testCtx, realm, id)
			if clientScope != nil {
				return fmt.Errorf("saml client scope %s still exists", id)
			}
		}

		return nil
	}
}

func getSamlClientScopeFromState(s *terraform.State, resourceName string) (*keycloak.SamlClientScope, error) {
	keycloakClientScope := testAccProvider.Meta().(*keycloak.KeycloakClient)

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	clientScope, err := keycloakClientScope.GetSamlClientScope(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting saml client scope %s: %s", id, err)
	}

	return clientScope, nil
}

func testKeycloakSamlClientScope_basic(clientScopeName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id

	description = "test description"
}
	`, testAccRealm.Realm, clientScopeName)
}

func testKeycloakSamlClientScope_withConsentText(clientScopeName, consentText string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client_scope" "client_scope" {
	name                = "%s"
	realm_id            = data.keycloak_realm.realm.id

	description         = "test description"

	consent_screen_text = "%s"
}
	`, testAccRealm.Realm, clientScopeName, consentText)
}

func testKeycloakSamlClientScope_withGuiOrder(clientScopeName string, guiOrder int) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client_scope" "client_scope" {
	name                = "%s"
	realm_id            = data.keycloak_realm.realm.id

	description         = "test description"

	gui_order           = %d
}
	`, testAccRealm.Realm, clientScopeName, guiOrder)
}

func testKeycloakSamlClientScope_updateRealmBefore(clientScopeName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_1" {
	realm = "%s"
}

data "keycloak_realm" "realm_2" {
	realm = "%s"
}

resource "keycloak_saml_client_scope" "client_scope" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm_1.id
}
	`, testAccRealm.Realm, testAccRealmTwo.Realm, clientScopeName)
}

func testKeycloakSamlClientScope_updateRealmAfter(clientScopeName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_1" {
	realm = "%s"
}

data "keycloak_realm" "realm_2" {
	realm = "%s"
}

resource "keycloak_saml_client_scope" "client_scope" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm_2.id
}
	`, testAccRealm.Realm, testAccRealmTwo.Realm, clientScopeName)
}

// check that a particular extra config key is missing
func testAccCheckKeycloakSamlClientScopeExtraConfigMissing(resourceName string, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clientScope, err := getSamlClientScopeFromState(s, resourceName)
		if err != nil {
			return err
		}

		if val, ok := clientScope.Attributes.ExtraConfig[key]; ok {
			// an empty string value is treated as removed (see setExtraConfigData)
			if val == "" {
				return nil
			}

			return fmt.Errorf("expected saml client scope to not have attribute %v, but got %v", key, val)
		}

		return nil
	}
}

func testAccCheckKeycloakSamlClientScopeExtraConfigHasValue(resourceName, key, expectedValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clientScope, err := getSamlClientScopeFromState(s, resourceName)
		if err != nil {
			return err
		}

		value, ok := clientScope.Attributes.ExtraConfig[key]
		if !ok {
			return fmt.Errorf("expected extra_config to contain key %q, but it was not found", key)
		}

		if value != expectedValue {
			return fmt.Errorf("expected extra_config key %q to have value %q, but got %q", key, expectedValue, value)
		}

		return nil
	}
}

func testKeycloakSamlClientScope_withExtraConfigMap(clientScopeName string, extraConfig map[string]string) string {
	var sb strings.Builder
	sb.WriteString("{\n")
	for k, v := range extraConfig {
		sb.WriteString(fmt.Sprintf("\t\t\"%s\" = \"%s\"\n", k, v))
	}
	sb.WriteString("\t}")

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id
	description = "test description"

	extra_config = %s
}
	`, testAccRealm.Realm, clientScopeName, sb.String())
}
