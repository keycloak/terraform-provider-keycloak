package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeycloakOpenidDefaultClientScope_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenidDefaultClientScopeConfigDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakOpenidDefaultClientScopeConfig(clientId, clientScopeId),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScope(),
			},
			{
				ResourceName:      "keycloak_openid_default_client_scope.openid_default_client_scope",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getKeycloakOpenidDefaultClientScopeImportId("keycloak_openid_default_client_scope.openid_default_client_scope"),
			},
		},
	})
}

func testAccKeycloakOpenidDefaultClientScopeConfigDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_default_client_scope" {
				continue
			}

			id := rs.Primary.ID
			realmId := rs.Primary.Attributes["realm_id"]
			clientScopeId := rs.Primary.Attributes["client_scope_id"]

			clientScope, _ := keycloakClient.GetOpenidRealmDefaultClientScope(testCtx, realmId, clientScopeId)
			if clientScope != nil {
				return fmt.Errorf("default client scope mapping with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakOpenidDefaultClientScopeImportId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found", resourceName)
		}

		id := rs.Primary.ID

		return id, nil
	}
}

func testAccCheckKeycloakOpenidClientHasDefaultScope() resource.TestCheckFunc {
	return func(s *terraform.State) error {

		resourceName := "keycloak_openid_default_client_scope.openid_default_client_scope"
		rsClientScope, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		resourceName = "keycloak_openid_client.openid_client"
		rsClient, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realm := rsClientScope.Primary.Attributes["realm_id"]
		clientScopeId := rsClientScope.Primary.Attributes["client_scope_id"]

		clientId := rsClient.Primary.ID

		keycloakDefaultClientScopes, err := keycloakClient.GetOpenidClientDefaultScopes(testCtx, realm, clientId)

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

func testAccKeycloakOpenidDefaultClientScopeConfig(clientId string, clientScopeId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "openid_client_scope" {
  realm_id = data.keycloak_realm.realm.id
  name     = "%s"
}

resource "keycloak_openid_default_client_scope" "openid_default_client_scope" {
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = keycloak_openid_client_scope.openid_client_scope.id
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	access_type = "CONFIDENTIAL"
	depends_on  = [keycloak_openid_default_client_scope.openid_default_client_scope]
}
`, testAccRealm.Realm, clientScopeId, clientId)
}
