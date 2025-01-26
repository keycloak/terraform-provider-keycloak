package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"slices"
	"sort"
	"testing"
)

func TestAccKeycloakRealmOptionalClientScopes_basic(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(
		t,
		resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testKeycloakRealmOptionalScopes_basic(realmName, clientScope),
					Check: testAccCheckKeycloakRealmHasOptionalScopes(
						"keycloak_realm_optional_client_scopes.optional_scopes",
						[]string{"address", "phone", "offline_access", clientScope},
					),
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
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
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
