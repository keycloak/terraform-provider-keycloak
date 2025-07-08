package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenidClientAuthorizationPermission_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	resourceName := acctest.RandomWithPrefix("tf-acc")
	permissionName := acctest.RandomWithPrefix("tf-acc")
	scopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientAuthorizationPermissionDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientAuthorizationPermission_basic(clientId, resourceName, permissionName, scopeName),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationPermissionExists("keycloak_openid_client_authorization_permission.test"),
			},
		},
	})
}

func TestAccKeycloakOpenidClientAuthorizationPermission_resourceType(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	resourceName := acctest.RandomWithPrefix("tf-acc")
	resourceType := acctest.RandomWithPrefix("tf-acc")
	permissionName := acctest.RandomWithPrefix("tf-acc")
	scopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientAuthorizationPermissionDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientAuthorizationPermission_resourceType(clientId, resourceName, resourceType, permissionName, scopeName),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationPermissionExists("keycloak_openid_client_authorization_permission.test"),
			},
		},
	})
}

func TestAccKeycloakOpenidClientAuthorizationPermission_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var authorizationPermission = &keycloak.OpenidClientAuthorizationPermission{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	resourceName := acctest.RandomWithPrefix("tf-acc")
	permissionName := acctest.RandomWithPrefix("tf-acc")
	scopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientAuthorizationPermissionDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientAuthorizationPermission_basic(clientId, resourceName, permissionName, scopeName),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationPermissionFetch("keycloak_openid_client_authorization_permission.test", authorizationPermission),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenidClientAuthorizationPermission(testCtx, authorizationPermission.RealmId, authorizationPermission.ResourceServerId, authorizationPermission.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOpenidClientAuthorizationPermission_basic(clientId, resourceName, permissionName, scopeName),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationPermissionExists("keycloak_openid_client_authorization_permission.test"),
			},
		},
	})
}

func TestAccKeycloakOpenidClientAuthorizationPermission_basicUpdateAll(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	scopeName := acctest.RandomWithPrefix("tf-acc")

	firstAuthorizationPermission := &keycloak.OpenidClientAuthorizationPermission{
		RealmId:     testAccRealm.Realm,
		Name:        acctest.RandString(10),
		Description: acctest.RandString(10),
	}

	secondAuthorizationPermission := &keycloak.OpenidClientAuthorizationPermission{
		RealmId:     testAccRealm.Realm,
		Name:        acctest.RandString(10),
		Description: acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientAuthorizationPermissionDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientAuthorizationPermission_basicFromInterface(clientId, firstAuthorizationPermission, acctest.RandString(10), scopeName),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationPermissionExists("keycloak_openid_client_authorization_permission.test"),
			},
			{
				Config: testKeycloakOpenidClientAuthorizationPermission_basicFromInterface(clientId, secondAuthorizationPermission, acctest.RandString(10), scopeName),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationPermissionExists("keycloak_openid_client_authorization_permission.test"),
			},
		},
	})
}

func testAccCheckKeycloakOpenidClientAuthorizationPermissionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakOpenidClientAuthorizationPermissionFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientAuthorizationPermissionFetch(resourceName string, mapper *keycloak.OpenidClientAuthorizationPermission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getKeycloakOpenidClientAuthorizationPermissionFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.ResourceServerId = fetchedMapper.ResourceServerId
		mapper.RealmId = fetchedMapper.RealmId
		mapper.Id = fetchedMapper.Id

		return nil
	}
}

func testAccCheckKeycloakOpenidClientAuthorizationPermissionDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client_authorization_permission" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			resourceServerId := rs.Primary.Attributes["resource_server_id"]
			id := rs.Primary.ID

			authorizationPermission, _ := keycloakClient.GetOpenidClientAuthorizationPermission(testCtx, realm, resourceServerId, id)
			if authorizationPermission != nil {
				return fmt.Errorf("test config with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakOpenidClientAuthorizationPermissionFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClientAuthorizationPermission, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	resourceServerId := rs.Primary.Attributes["resource_server_id"]
	id := rs.Primary.ID

	authorizationPermission, err := keycloakClient.GetOpenidClientAuthorizationPermission(testCtx, realm, resourceServerId, id)
	if err != nil {
		return nil, fmt.Errorf("error getting authorization permission config with id %s: %s", id, err)
	}

	return authorizationPermission, nil
}

func testKeycloakOpenidClientAuthorizationPermission_basic(clientId, resourceName, permissionName, scopeName string) string {
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

data keycloak_openid_client_authorization_policy default {
	realm_id           = data.keycloak_realm.realm.id
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	name               = "default"
}

resource keycloak_openid_client_authorization_resource test {
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	name               = "%s"
	realm_id           = data.keycloak_realm.realm.id

	uris = [
		"/endpoint/*"
	]
}

resource keycloak_openid_client_authorization_scope test {
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	name               = "%s"
	realm_id           = data.keycloak_realm.realm.id
}

resource keycloak_openid_client_authorization_permission test {
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	realm_id           = data.keycloak_realm.realm.id
	name               = "%s"
	policies           = ["${data.keycloak_openid_client_authorization_policy.default.id}"]
	resources          = ["${keycloak_openid_client_authorization_resource.test.id}"]

}
	`, testAccRealm.Realm, clientId, resourceName, scopeName, permissionName)
}

func testKeycloakOpenidClientAuthorizationPermission_resourceType(clientId, resourceName, resourceType, permissionName, scopeName string) string {
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

data keycloak_openid_client_authorization_policy default {
	realm_id           = data.keycloak_realm.realm.id
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	name               = "default"
}

resource keycloak_openid_client_authorization_resource test {
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	name               = "%s"
	realm_id           = data.keycloak_realm.realm.id

	type = "%s"

	uris = [
		"/endpoint/*"
	]
}

resource keycloak_openid_client_authorization_scope test {
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	name               = "%s"
	realm_id           = data.keycloak_realm.realm.id
}

resource keycloak_openid_client_authorization_permission test {
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	realm_id           = data.keycloak_realm.realm.id
	name               = "%s"
	policies           = ["${data.keycloak_openid_client_authorization_policy.default.id}"]
	resource_type      = "%s"

}
	`, testAccRealm.Realm, clientId, resourceName, resourceType, scopeName, permissionName, resourceType)
}

func testKeycloakOpenidClientAuthorizationPermission_basicFromInterface(clientId string, authorizationPermission *keycloak.OpenidClientAuthorizationPermission, resourceName, scopeName string) string {
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

data keycloak_openid_client_authorization_policy default {
	realm_id           = data.keycloak_realm.realm.id
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	name               = "default"
}

resource keycloak_openid_client_authorization_resource resource {
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	name               = "%s"
	realm_id           = data.keycloak_realm.realm.id

	uris = [
		"/endpoint/*"
	]
}

resource keycloak_openid_client_authorization_scope test {
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	name               = "%s"
	realm_id           = data.keycloak_realm.realm.id
}

resource keycloak_openid_client_authorization_permission test {
	resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
	realm_id           = data.keycloak_realm.realm.id
	name               = "%s"
	policies           = ["${data.keycloak_openid_client_authorization_policy.default.id}"]
	resources          = ["${keycloak_openid_client_authorization_resource.resource.id}"]
	description        = "%s"
	scopes = ["${keycloak_openid_client_authorization_scope.test.id}"]
}
	`, testAccRealm.Realm, clientId, resourceName, scopeName, authorizationPermission.Name, authorizationPermission.Description)
}
