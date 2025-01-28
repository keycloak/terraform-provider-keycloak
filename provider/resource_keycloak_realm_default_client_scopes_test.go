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

func TestAccKeycloakRealmDefaultClientScopes_basic(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmDefaultScopes_basic(realmName, clientScope),
				Check: testAccCheckKeycloakRealmHasDefaultScopes(
					"keycloak_realm_default_client_scopes.default_scopes",
					[]string{"profile", "email", "web-origins", "roles", "role_list", clientScope}),
			},
		},
	})
}

func TestAccKeycloakRealmDefaultClientScopes_empty(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmDefaultScopes_empty(realmName),
				Check: testAccCheckKeycloakRealmHasDefaultScopes(
					"keycloak_realm_default_client_scopes.default_scopes",
					[]string{}),
			},
		},
	})
}

func getRealmDefaultClientScopesFromState(resourceName string, s *terraform.State) ([]*keycloak.OpenidClientScope, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]

	keycloakDefaultClientScopes, err := keycloakClient.GetRealmDefaultClientScopes(testCtx, realm)
	if err != nil {
		return nil, err
	}

	return keycloakDefaultClientScopes, nil
}

func testAccCheckKeycloakRealmHasDefaultScopes(resourceName string, expectedClientScopeNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keycloakDefaultClientScopes, err := getRealmDefaultClientScopesFromState(resourceName, s)
		if err != nil {
			return err
		}

		var assignedClientScopeNames []string
		for _, keycloakDefaultScope := range keycloakDefaultClientScopes {
			assignedClientScopeNames = append(assignedClientScopeNames, keycloakDefaultScope.Name)
		}

		sort.Strings(expectedClientScopeNames)
		sort.Strings(assignedClientScopeNames)

		if !slices.Equal(assignedClientScopeNames, expectedClientScopeNames) {
			return fmt.Errorf(
				"assigned and expected realm default client scopes do not match %v != %v",
				assignedClientScopeNames,
				expectedClientScopeNames,
			)
		}

		return nil
	}
}

func testKeycloakRealmDefaultScopes_basic(realmName string, clientScope string) string {
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
		"web-origins",
		keycloak_openid_client_scope.client_scope.name
	]
}
	`, realmName, clientScope)
}

func testKeycloakRealmDefaultScopes_empty(realmName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_default_client_scopes" "default_scopes" {
	realm_id       = keycloak_realm.realm.id
	default_scopes = []
}
	`, realmName)
}
