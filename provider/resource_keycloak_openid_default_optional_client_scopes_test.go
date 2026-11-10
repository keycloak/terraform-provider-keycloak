package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKeycloakOpenidDefaultOptionalClientScopes_basic(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	clientScopeName := acctest.RandomWithPrefix("tf-acc-scope")

	expectedScopes := []string{"address", "phone", clientScopeName}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidDefaultOptionalClientScopes_basic(realmName, clientScopeName, expectedScopes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidRealmHasDefaultOptionalClientScopes(realmName, expectedScopes),
				),
			},
			{
				ResourceName:      "keycloak_openid_default_optional_client_scopes.realm_optionals",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKeycloakOpenidDefaultOptionalClientScopes_updateInPlace(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	scopeA := acctest.RandomWithPrefix("tf-acc-scope-a")
	scopeB := acctest.RandomWithPrefix("tf-acc-scope-b")

	withA := []string{"address", scopeA}
	withB := []string{"address", scopeB}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidDefaultOptionalClientScopes_twoScopes(realmName, scopeA, scopeB, withA),
				Check:  testAccCheckKeycloakOpenidRealmHasDefaultOptionalClientScopes(realmName, withA),
			},
			{
				Config: testKeycloakOpenidDefaultOptionalClientScopes_twoScopes(realmName, scopeA, scopeB, withB),
				Check:  testAccCheckKeycloakOpenidRealmHasDefaultOptionalClientScopes(realmName, withB),
			},
		},
	})
}

func testAccCheckKeycloakOpenidRealmHasDefaultOptionalClientScopes(realmName string, expectedScopes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actual, err := keycloakClient.GetOpenidRealmDefaultOptionalClientScopes(testCtx, realmName)
		if err != nil {
			return err
		}

		actualNames := make(map[string]struct{}, len(actual))
		for _, scope := range actual {
			actualNames[scope.Name] = struct{}{}
		}

		if len(actualNames) != len(expectedScopes) {
			return fmt.Errorf("expected realm %q to have %d default-optional client scopes, but got %d", realmName, len(expectedScopes), len(actualNames))
		}

		for _, expected := range expectedScopes {
			if _, ok := actualNames[expected]; !ok {
				return fmt.Errorf("expected realm %q default-optional client scopes to contain %q", realmName, expected)
			}
		}

		return nil
	}
}

func testKeycloakOpenidDefaultOptionalClientScopes_basic(realm, clientScope string, scopes []string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	realm_id = keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_openid_default_optional_client_scopes" "realm_optionals" {
	realm_id        = keycloak_realm.realm.id
	optional_scopes = %s

	depends_on = [keycloak_openid_client_scope.client_scope]
}
	`, realm, clientScope, arrayOfStringsForTerraformResource(scopes))
}

func testKeycloakOpenidDefaultOptionalClientScopes_twoScopes(realm, scopeA, scopeB string, scopes []string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "scope_a" {
	realm_id = keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_openid_client_scope" "scope_b" {
	realm_id = keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_openid_default_optional_client_scopes" "realm_optionals" {
	realm_id        = keycloak_realm.realm.id
	optional_scopes = %s

	depends_on = [
		keycloak_openid_client_scope.scope_a,
		keycloak_openid_client_scope.scope_b,
	]
}
	`, realm, scopeA, scopeB, arrayOfStringsForTerraformResource(scopes))
}
