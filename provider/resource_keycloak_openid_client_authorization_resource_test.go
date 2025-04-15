package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"testing"
)

func TestAccKeycloakOpenidClientAuthorizationResource_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	resourceName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientAuthorizationResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientAuthorizationResource_basic(clientId, resourceName),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationResourceExists("keycloak_openid_client_authorization_resource.test"),
			},
		},
	})
}

func TestAccKeycloakOpenidClientAuthorizationResource_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var authorizationResource = &keycloak.OpenidClientAuthorizationResource{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	resourceName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientAuthorizationResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientAuthorizationResource_basic(clientId, resourceName),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationResourceFetch("keycloak_openid_client_authorization_resource.test", authorizationResource),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenidClientAuthorizationResource(testCtx, authorizationResource.RealmId, authorizationResource.ResourceServerId, authorizationResource.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOpenidClientAuthorizationResource_basic(clientId, resourceName),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationResourceExists("keycloak_openid_client_authorization_resource.test"),
			},
		},
	})
}

func TestAccKeycloakOpenidClientAuthorizationResource_basicUpdateAll(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	ownerManagedAccess := randomBool()

	firstAuthorizationResource := &keycloak.OpenidClientAuthorizationResource{
		RealmId:            testAccRealm.Realm,
		Name:               acctest.RandString(10),
		DisplayName:        acctest.RandString(10),
		IconUri:            acctest.RandString(10),
		Type:               acctest.RandString(10),
		OwnerManagedAccess: ownerManagedAccess,
	}

	secondAuthorizationResource := &keycloak.OpenidClientAuthorizationResource{
		RealmId:            testAccRealm.Realm,
		Name:               acctest.RandString(10),
		DisplayName:        acctest.RandString(10),
		IconUri:            acctest.RandString(10),
		Type:               acctest.RandString(10),
		OwnerManagedAccess: !ownerManagedAccess,
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientAuthorizationResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientAuthorizationResource_basicFromInterface(clientId, firstAuthorizationResource),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationResourceExists("keycloak_openid_client_authorization_resource.test"),
			},
			{
				Config: testKeycloakOpenidClientAuthorizationResource_basicFromInterface(clientId, secondAuthorizationResource),
				Check:  testAccCheckKeycloakOpenidClientAuthorizationResourceExists("keycloak_openid_client_authorization_resource.test"),
			},
		},
	})
}

func testAccCheckKeycloakOpenidClientAuthorizationResourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakOpenidClientAuthorizationResourceFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientAuthorizationResourceFetch(resourceName string, mapper *keycloak.OpenidClientAuthorizationResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedMapper, err := getKeycloakOpenidClientAuthorizationResourceFromState(s, resourceName)
		if err != nil {
			return err
		}

		mapper.ResourceServerId = fetchedMapper.ResourceServerId
		mapper.RealmId = fetchedMapper.RealmId
		mapper.Id = fetchedMapper.Id

		return nil
	}
}

func testAccCheckKeycloakOpenidClientAuthorizationResourceDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client_authorization_resource" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			resourceServerId := rs.Primary.Attributes["resource_server_id"]
			id := rs.Primary.ID

			authorizationResource, _ := keycloakClient.GetOpenidClientAuthorizationResource(testCtx, realm, resourceServerId, id)
			if authorizationResource != nil {
				return fmt.Errorf("test config with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakOpenidClientAuthorizationResourceFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClientAuthorizationResource, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	resourceServerId := rs.Primary.Attributes["resource_server_id"]
	id := rs.Primary.ID

	authorizationResource, err := keycloakClient.GetOpenidClientAuthorizationResource(testCtx, realm, resourceServerId, id)
	if err != nil {
		return nil, fmt.Errorf("error getting authorization resource config with id %s: %s", id, err)
	}

	return authorizationResource, nil
}

func testKeycloakOpenidClientAuthorizationResource_basic(clientId, resourceName string) string {
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

resource keycloak_openid_client_authorization_resource test {
  resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
  name               = "%s"
  realm_id           = data.keycloak_realm.realm.id

  uris = [
    "/endpoint/*"
  ]
}
	`, testAccRealm.Realm, clientId, resourceName)
}

func testKeycloakOpenidClientAuthorizationResource_basicFromInterface(clientId string, authorizationResource *keycloak.OpenidClientAuthorizationResource) string {
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

resource keycloak_openid_client_authorization_resource test {
  resource_server_id = "${keycloak_openid_client.test.resource_server_id}"
  name                 = "%s"
  realm_id             = data.keycloak_realm.realm.id
  display_name         = "%s"
  icon_uri             = "%s"
  owner_managed_access = %t
  type                 = "%s"
  uris = [
    "/test/"
  ]
}
	`, testAccRealm.Realm, clientId, authorizationResource.Name, authorizationResource.DisplayName, authorizationResource.IconUri, authorizationResource.OwnerManagedAccess, authorizationResource.Type)
}
