package provider

import (
	"fmt"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakClientScope_basic(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
			},
			{
				ResourceName:        "keycloak_openid_client_scope.client_scope",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakClientScope_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var clientScope = &keycloak.OpenidClientScope{}

	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientScope_basic(clientScopeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
					testAccCheckKeycloakClientScopeFetch("keycloak_openid_client_scope.client_scope", clientScope),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenidClientScope(testCtx, clientScope.RealmId, clientScope.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
			},
		},
	})
}

func TestAccKeycloakClientScope_updateRealm(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientScope_updateRealmBefore(clientScopeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
					testAccCheckKeycloakClientScopeBelongsToRealm("keycloak_openid_client_scope.client_scope", testAccRealm.Realm),
				),
			},
			{
				Config: testKeycloakClientScope_updateRealmAfter(clientScopeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
					testAccCheckKeycloakClientScopeBelongsToRealm("keycloak_openid_client_scope.client_scope", testAccRealmTwo.Realm),
				),
			},
		},
	})
}

func TestAccKeycloakClientScope_consentScreenText(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
			},
			{
				Config: testKeycloakClientScope_withConsentText(clientScopeName, acctest.RandString(10)),
				Check:  testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
			},
			{
				Config: testKeycloakClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
			},
		},
	})
}

func TestAccKeycloakClientScope_includeInTokenScope(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc")
	includeInTokenScope := false

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
			},
			{
				Config: testKeycloakClientScope_withIncludeInTokenScope(clientScopeName, includeInTokenScope),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
					testAccCheckKeycloakClientScopeExistsWithCorrectIncludeInTokenScope("keycloak_openid_client_scope.client_scope", includeInTokenScope),
				),
			},
			{
				Config: testKeycloakClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
			},
		},
	})
}

func TestAccKeycloakClientScope_guiOrder(t *testing.T) {
	t.Parallel()
	clientScopeName := acctest.RandomWithPrefix("tf-acc")
	guiOrder := acctest.RandIntRange(0, 1000)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakClientScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
			},
			{
				Config: testKeycloakClientScope_withGuiOrder(clientScopeName, guiOrder),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
					testAccCheckKeycloakClientScopeExistsWithCorrectGuiOrder("keycloak_openid_client_scope.client_scope", guiOrder),
				),
			},
			{
				Config: testKeycloakClientScope_basic(clientScopeName),
				Check:  testAccCheckKeycloakClientScopeExistsWithCorrectProtocol("keycloak_openid_client_scope.client_scope"),
			},
		},
	})
}

func testAccCheckKeycloakClientScopeExistsWithCorrectProtocol(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clientScope, err := getClientScopeFromState(s, resourceName)
		if err != nil {
			return err
		}

		if clientScope.Protocol != "openid-connect" {
			return fmt.Errorf("expected openid client scope to have openid-connect protocol, but got %s", clientScope.Protocol)
		}

		return nil
	}
}

func testAccCheckKeycloakClientScopeExistsWithCorrectIncludeInTokenScope(resourceName string, includeInTokenScope bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clientScope, err := getClientScopeFromState(s, resourceName)
		if err != nil {
			return err
		}

		if clientScope.Attributes.IncludeInTokenScope != types.KeycloakBoolQuoted(includeInTokenScope) {
			return fmt.Errorf("expected saml client includeInTokenScope to have %t, but got %t", includeInTokenScope, clientScope.Attributes.IncludeInTokenScope)
		}

		return nil
	}
}

func testAccCheckKeycloakClientScopeExistsWithCorrectGuiOrder(resourceName string, guiOrder int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clientScope, err := getClientScopeFromState(s, resourceName)
		if err != nil {
			return err
		}

		if clientScope.Attributes.GuiOrder != strconv.Itoa(guiOrder) {
			return fmt.Errorf("expected saml client guiOrder to have %d, but got %s", guiOrder, clientScope.Attributes.GuiOrder)
		}

		return nil
	}
}

func testAccCheckKeycloakClientScopeFetch(resourceName string, clientScope *keycloak.OpenidClientScope) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedClientScope, err := getClientScopeFromState(s, resourceName)
		if err != nil {
			return err
		}

		clientScope.Id = fetchedClientScope.Id
		clientScope.RealmId = fetchedClientScope.RealmId

		return nil
	}
}

func testAccCheckKeycloakClientScopeBelongsToRealm(resourceName, realm string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clientScope, err := getClientScopeFromState(s, resourceName)
		if err != nil {
			return err
		}

		if clientScope.RealmId != realm {
			return fmt.Errorf("expected openid client scope %s to have realm_id of %s, but got %s", clientScope.Id, realm, clientScope.RealmId)
		}

		return nil
	}
}

func testAccCheckKeycloakClientScopeDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client_scope" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			clientScope, _ := keycloakClient.GetOpenidClientScope(testCtx, realm, id)
			if clientScope != nil {
				return fmt.Errorf("openid client scope %s still exists", id)
			}
		}

		return nil
	}
}

func getClientScopeFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClientScope, error) {
	keycloakClientScope := testAccProvider.Meta().(*keycloak.KeycloakClient)

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	clientScope, err := keycloakClientScope.GetOpenidClientScope(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting openid client scope %s: %s", id, err)
	}

	return clientScope, nil
}

func testKeycloakClientScope_basic(clientScopeName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id

	description = "test description"
}
	`, testAccRealm.Realm, clientScopeName)
}

func testKeycloakClientScope_withConsentText(clientScopeName, consentText string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name                = "%s"
	realm_id            = data.keycloak_realm.realm.id

	description         = "test description"

	consent_screen_text = "%s"
}
	`, testAccRealm.Realm, clientScopeName, consentText)
}

func testKeycloakClientScope_withIncludeInTokenScope(clientScopeName string, includeInTokenScope bool) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name                = "%s"
	realm_id            = data.keycloak_realm.realm.id

	description         = "test description"

	include_in_token_scope = %t
}
	`, testAccRealm.Realm, clientScopeName, includeInTokenScope)
}

func testKeycloakClientScope_withGuiOrder(clientScopeName string, guiOrder int) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name                = "%s"
	realm_id            = data.keycloak_realm.realm.id

	description         = "test description"

	gui_order           = %d
}
	`, testAccRealm.Realm, clientScopeName, guiOrder)
}

func testKeycloakClientScope_updateRealmBefore(clientScopeName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_1" {
	realm = "%s"
}

data "keycloak_realm" "realm_2" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm_1.id
}
	`, testAccRealm.Realm, testAccRealmTwo.Realm, clientScopeName)
}

func testKeycloakClientScope_updateRealmAfter(clientScopeName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_1" {
	realm = "%s"
}

data "keycloak_realm" "realm_2" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm_2.id
}
	`, testAccRealm.Realm, testAccRealmTwo.Realm, clientScopeName)
}
