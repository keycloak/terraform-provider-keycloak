package provider

import (
	"fmt"
	"slices"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakRealmOptionalClientScopes_basic(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(
		t,
		resource.TestCase{
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
			PreCheck:                 func() { testAccPreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testKeycloakRealmOptionalScopes_basic(realmName, clientScope),
					Check: testAccCheckKeycloakRealmHasOptionalScopes(
						"keycloak_realm_optional_client_scopes.optional_scopes",
						[]string{"address", "phone", "offline_access", clientScope},
					),
				},
				{
					ResourceName:      "keycloak_realm_optional_client_scopes.optional_scopes",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: func(s *terraform.State) (string, error) {
						rs, ok := s.RootModule().Resources["keycloak_realm_optional_client_scopes.optional_scopes"]
						if !ok {
							return "", fmt.Errorf("resource not found: keycloak_realm_optional_client_scopes.optional_scopes")
						}
						return rs.Primary.Attributes["realm_id"], nil
					},
				},
			},
		},
	)
}

func TestAccKeycloakRealmOptionalClientScopes_empty(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(
		t,
		resource.TestCase{
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
			PreCheck:                 func() { testAccPreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testKeycloakRealmOptionalScopes_empty(realmName),
					Check: testAccCheckKeycloakRealmHasOptionalScopes(
						"keycloak_realm_optional_client_scopes.optional_scopes",
						[]string{},
					),
				},
			},
		},
	)
}

func getRealmOptionalClientScopesFromState(resourceName string, s *terraform.State) ([]*keycloak.OpenidClientScope, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]

	keycloakDefaultClientScopes, err := keycloakClient.GetRealmOptionalClientScopes(testCtx, realm)
	if err != nil {
		return nil, err
	}

	return keycloakDefaultClientScopes, nil
}

func testAccCheckKeycloakRealmHasOptionalScopes(resourceName string, expectedClientScopeNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keycloakOptionalClientScopes, err := getRealmOptionalClientScopesFromState(resourceName, s)
		if err != nil {
			return err
		}

		var assignedClientScopeNames []string
		for _, keycloakDefaultScope := range keycloakOptionalClientScopes {
			assignedClientScopeNames = append(assignedClientScopeNames, keycloakDefaultScope.Name)
		}

		sort.Strings(expectedClientScopeNames)
		sort.Strings(assignedClientScopeNames)

		if !slices.Equal(assignedClientScopeNames, expectedClientScopeNames) {
			return fmt.Errorf(
				"assigned and expected realm optional client scopes do not match %v != %v",
				assignedClientScopeNames,
				expectedClientScopeNames,
			)
		}

		return nil
	}
}

func testKeycloakRealmOptionalScopes_basic(realmName, clientScope string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = keycloak_realm.realm.id

	description = "test description"
}

resource "keycloak_realm_optional_client_scopes" "optional_scopes" {
	realm_id       = keycloak_realm.realm.id
	optional_scopes = [
		"address",
		"phone",
		"offline_access",
		keycloak_openid_client_scope.client_scope.name
	]
}
	`, realmName, clientScope)
}

func testKeycloakRealmOptionalScopes_empty(realmName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_optional_client_scopes" "optional_scopes" {
	realm_id       = keycloak_realm.realm.id
	optional_scopes = []
}
	`, realmName)
}

// Keycloak throws a 500 if you attempt to mark a client scope as a realm optional scope while it is marked as a realm
// default scope, but the provider will unmark the existing scope and re-mark it as a realm optional scope
func TestAccKeycloakRealmOptionalClientScopes_duplicateScopeAssignment(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmDefaultScopes_basic(realmName, clientScope),
				Check: testAccCheckKeycloakRealmHasDefaultScopes(
					"keycloak_realm_default_client_scopes.default_scopes",
					[]string{"profile", "email", "roles", "role_list", "web-origins", clientScope},
				),
			},
			// move the custom scope from default to optional while it is still marked as default
			{
				Config: testKeycloakRealmOptionalScopes_duplicateScopeAssignment(realmName, clientScope),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmHasOptionalScopes(
						"keycloak_realm_optional_client_scopes.optional_scopes",
						[]string{"address", "phone", "offline_access", clientScope},
					),
					testAccCheckKeycloakRealmHasDefaultScopes(
						"keycloak_realm_default_client_scopes.default_scopes",
						[]string{"profile", "email", "roles", "role_list", "web-origins"},
					),
				),
			},
		},
	})
}

// the default scopes resource depends on the optional scopes resource so that the custom scope is still marked as a
// realm default scope at the point it is marked as a realm optional scope
func testKeycloakRealmOptionalScopes_duplicateScopeAssignment(realmName, clientScope string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = keycloak_realm.realm.id

	description = "test description"
}

resource "keycloak_realm_default_client_scopes" "default_scopes" {
	realm_id       = keycloak_realm.realm.id
	default_scopes = [
		"profile",
		"email",
		"roles",
		"role_list",
		"web-origins"
	]

	depends_on = [keycloak_realm_optional_client_scopes.optional_scopes]
}

resource "keycloak_realm_optional_client_scopes" "optional_scopes" {
	realm_id        = keycloak_realm.realm.id
	optional_scopes = [
		"address",
		"phone",
		"offline_access",
		keycloak_openid_client_scope.client_scope.name
	]
}
	`, realmName, clientScope)
}
