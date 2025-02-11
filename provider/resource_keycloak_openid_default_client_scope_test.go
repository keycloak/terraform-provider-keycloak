package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strings"
	"testing"
)

func TestAccKeycloakDataSourceOpenidDefaultClientScope_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakOpenidDefaultClientScope_basic(clientId),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScope("keycloak_openid_default_client_scope"),
			},
		},
	})
}

func testAccCheckKeycloakOpenidClientHasDefaultScope(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realm := rs.Primary.Attributes["realm_id"]
		clientScopeId := rs.Primary.Attributes["client_scope_id"]

		var client string
		if strings.HasPrefix(resourceName, "keycloak_openid_client") {
			client = rs.Primary.Attributes["client_id"]
		} else {
			client = rs.Primary.ID
		}

		keycloakDefaultClientScopes, err := keycloakClient.GetOpenidClientDefaultScopes(realm, client)

		if err != nil {
			return err
		}

		var found = false
		for _, keycloakDefaultScope := range keycloakDefaultClientScopes {
			if keycloakDefaultScope.Id == clientScopeId {
				found = true

				break
			}
		}

		if !found {
			return fmt.Errorf("default scope %s is not assigned to client", clientScopeId)
		}

		return nil
	}
}

func testAccKeycloakOpenidDefaultClientScope_basic(clientId string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
  realm   = "%s"
  enabled = true
}

resource "keycloak_openid_client_scope" "openid_client_scope" {
  realm_id               = keycloak_realm.realm.id
  name                   = "groups"
}

resource "keycloak_openid_default_client_scope" "openid_default_client_scope" {
	realm_id = keycloak_realm.realm.id
	client_scope_id = keycloak_openid_client_scope.client_scope.id
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = keycloak_realm.realm.id
	client_id = "%s"
}
`, testAccRealm.Realm, clientId)
}
