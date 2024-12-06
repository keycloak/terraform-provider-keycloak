package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"testing"
)

func TestAccKeycloakOpenidClientAuthorizationScope_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	scopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientAuthorizationScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientAuthorizationScope_basic(clientId, scopeName),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationScopeExists("keycloak_openid_client_authorization_scope.test"),
			},
		},
	})
}

func TestAccKeycloakOpenidClientAuthorizationScope_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var authorizationScope = &keycloak.OpenidClientAuthorizationScope{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	scopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientAuthorizationScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientAuthorizationScope_basic(clientId, scopeName),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationScopeFetch("keycloak_openid_client_authorization_scope.test", authorizationScope),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenidClientAuthorizationScope(testCtx, authorizationScope.RealmId, authorizationScope.ResourceServerId, authorizationScope.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOpenidClientAuthorizationScope_basic(clientId, scopeName),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationScopeExists("keycloak_openid_client_authorization_scope.test"),
			},
		},
	})
}

func TestAccKeycloakOpenidClientAuthorizationScope_basicUpdateAll(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	firstAuthorizationScope := &keycloak.OpenidClientAuthorizationScope{
		RealmId:     testAccRealm.Realm,
		Name:        acctest.RandString(10),
		DisplayName: acctest.RandString(10),
		IconUri:     acctest.RandString(10),
	}

	secondAuthorizationScope := &keycloak.OpenidClientAuthorizationScope{
		RealmId:     testAccRealm.Realm,
		Name:        acctest.RandString(10),
		DisplayName: acctest.RandString(10),
		IconUri:     acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientAuthorizationScopeDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientAuthorizationScope_basicFromInterface(clientId, firstAuthorizationScope),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationScopeExists("keycloak_openid_client_authorization_scope.test"),
			},
			{
				Config: testKeycloakOpenidClientAuthorizationScope_basicFromInterface(clientId, secondAuthorizationScope),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationScopeExists("keycloak_openid_client_authorization_scope.test"),
			},
		},
	})
}

func testAccCheckKeycloakOpenidClientAuthorizationScopeExists(scopeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakOpenidClientAuthorizationScopeFromState(s, scopeName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientAuthorizationScopeFetch(scopeName string, authorizationScope *keycloak.OpenidClientAuthorizationScope) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedAuthorizationScope, err := getKeycloakOpenidClientAuthorizationScopeFromState(s, scopeName)
		if err != nil {
			return err
		}

		authorizationScope.ResourceServerId = fetchedAuthorizationScope.ResourceServerId
		authorizationScope.RealmId = fetchedAuthorizationScope.RealmId
		authorizationScope.Id = fetchedAuthorizationScope.Id

		return nil
	}
}

func testAccCheckKeycloakOpenidClientAuthorizationScopeDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client_authorization_scope" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			resourceServerId := rs.Primary.Attributes["resource_server_id"]
			id := rs.Primary.ID

			authorizationScope, _ := keycloakClient.GetOpenidClientAuthorizationScope(testCtx, realm, resourceServerId, id)
			if authorizationScope != nil {
				return fmt.Errorf("test config with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakOpenidClientAuthorizationScopeFromState(s *terraform.State, scopeName string) (*keycloak.OpenidClientAuthorizationScope, error) {
	rs, ok := s.RootModule().Resources[scopeName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", scopeName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	resourceServerId := rs.Primary.Attributes["resource_server_id"]
	id := rs.Primary.ID

	authorizationScope, err := keycloakClient.GetOpenidClientAuthorizationScope(testCtx, realm, resourceServerId, id)
	if err != nil {
		return nil, fmt.Errorf("error getting authorization scope config with id %s: %s", id, err)
	}

	return authorizationScope, nil
}

func testKeycloakOpenidClientAuthorizationScope_basic(clientId, scopeName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource keycloak_openid_client test {
	client_id                = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	access_type              = "CONFIDENTIAL"
	service_accounts_enabled = true
	authorization {
		policy_enforcement_mode = "ENFORCING"
	}
}

resource keycloak_openid_client_authorization_scope test {
  resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
  name               = "%s"
  realm_id           = data.keycloak_realm.realm.id
}
	`, testAccRealm.Realm, clientId, scopeName)
}

func testKeycloakOpenidClientAuthorizationScope_basicFromInterface(clientId string, authorizationScope *keycloak.OpenidClientAuthorizationScope) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource keycloak_openid_client test {
	client_id                = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	access_type              = "CONFIDENTIAL"
	service_accounts_enabled = true
	authorization {
		policy_enforcement_mode = "ENFORCING"
	}
}

resource keycloak_openid_client_authorization_scope test {
  resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
  name                 = "%s"
  realm_id             = data.keycloak_realm.realm.id
  display_name         = "%s"
  icon_uri             = "%s"
}
	`, authorizationScope.RealmId, clientId, authorizationScope.Name, authorizationScope.DisplayName, authorizationScope.IconUri)
}
