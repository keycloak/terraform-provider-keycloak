package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakGenericClientRoleMapper_basic(t *testing.T) {
	t.Parallel()

	parentClientName := acctest.RandomWithPrefix("tf-acc")
	parentRoleName := acctest.RandomWithPrefix("tf-acc")
	childClientName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericClientRoleMapper_basic(parentClientName, parentRoleName, childClientName),
				Check:  testAccCheckKeycloakGenericClientRoleMapperExists("keycloak_generic_client_role_mapper.child-client-with-parent-client-role"),
			},
		},
	})
}

func TestAccKeycloakGenericClientRoleMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var role = &keycloak.Role{}
	var childClient = &keycloak.GenericClient{}

	parentClientName := acctest.RandomWithPrefix("tf-acc")
	parentRoleName := acctest.RandomWithPrefix("tf-acc")
	childClientName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericClientRoleMapper_basic(parentClientName, parentRoleName, childClientName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGenericClientRoleMapperExists("keycloak_generic_client_role_mapper.child-client-with-parent-client-role"),
					testAccCheckKeycloakRoleFetch("keycloak_role.parent-role", role),
					testAccCheckKeycloakGenericClientFetch("keycloak_openid_client.child-client", childClient),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRoleScopeMapping(testCtx, childClient.RealmId, childClient.Id, "", role)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakGenericClientRoleMapper_basic(parentClientName, parentRoleName, childClientName),
				Check:  testAccCheckKeycloakGenericClientRoleMapperExists("keycloak_generic_client_role_mapper.child-client-with-parent-client-role"),
			},
		},
	})
}

func TestAccKeycloakGenericClientRoleMapper_import(t *testing.T) {
	t.Parallel()

	parentClientName := acctest.RandomWithPrefix("tf-acc")
	parentRoleName := acctest.RandomWithPrefix("tf-acc")
	childClientName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_generic_client_role_mapper.child-client-with-parent-client-role"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericClientRoleMapper_basic(parentClientName, parentRoleName, childClientName),
				Check:  testAccCheckKeycloakGenericClientRoleMapperExists(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getGenericClientRoleMapperId(resourceName),
			},
		},
	})
}

func TestAccKeycloakGenericClientRoleMapper_basicClientScope(t *testing.T) {
	t.Parallel()

	clientName := acctest.RandomWithPrefix("tf-acc")
	roleName := acctest.RandomWithPrefix("tf-acc")
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericClientRoleMapper_basicClientScope(clientName, roleName, clientScopeName),
				Check:  testAccCheckKeycloakGenericClientRoleMapperExists("keycloak_generic_client_role_mapper.clientscope-with-client-role"),
			},
		},
	})
}

func TestAccKeycloakGenericClientRoleMapper_importClientScope(t *testing.T) {
	t.Parallel()

	clientName := acctest.RandomWithPrefix("tf-acc")
	roleName := acctest.RandomWithPrefix("tf-acc")
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_generic_client_role_mapper.clientscope-with-client-role"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericClientRoleMapper_basicClientScope(clientName, roleName, clientScopeName),
				Check:  testAccCheckKeycloakGenericClientRoleMapperExists(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getGenericClientRoleMapperId(resourceName),
			},
		},
	})
}

func TestAccKeycloakGenericClientRoleMapper_basicClientScopeRealmRole(t *testing.T) {
	t.Parallel()

	roleName := acctest.RandomWithPrefix("tf-acc")
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericClientRoleMapper_basicClientScopeRealmRole(roleName, clientScopeName),
				Check:  testAccCheckKeycloakGenericClientRoleMapperExists("keycloak_generic_client_role_mapper.clientscope-with-realm-role"),
			},
		},
	})
}

func testAccCheckKeycloakGenericClientRoleMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		return nil
	}
}

func testAccCheckKeycloakGenericClientFetch(resourceName string, client *keycloak.GenericClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedClient, err := getGenericClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		client.Id = fetchedClient.Id
		client.ClientId = fetchedClient.ClientId
		client.RealmId = fetchedClient.RealmId

		return nil
	}
}

func getGenericClientFromState(s *terraform.State, resourceName string) (*keycloak.GenericClient, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	client, err := keycloakClient.GetGenericClient(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting generic client %s: %s", id, err)
	}

	return client, nil
}

func getGenericClientRoleMapperId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		return rs.Primary.ID, nil
	}
}

func testKeycloakGenericClientRoleMapper_basic(parentClientName, parentRoleName, childClientName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "parent-client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	access_type = "PUBLIC"
}

resource "keycloak_role" "parent-role" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.parent-client.id
	name      = "%s"
}

resource "keycloak_openid_client" "child-client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	access_type = "PUBLIC"
}

resource "keycloak_generic_client_role_mapper" "child-client-with-parent-client-role" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.child-client.id
	role_id   = keycloak_role.parent-role.id
}
	`, testAccRealm.Realm, parentClientName, parentRoleName, childClientName)
}

func testKeycloakGenericClientRoleMapper_basicClientScope(clientName, roleName, clientScopeName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	access_type = "PUBLIC"
}

resource "keycloak_role" "role" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.client.id
	name      = "%s"
}

resource "keycloak_openid_client_scope" "clientscope" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_generic_client_role_mapper" "clientscope-with-client-role" {
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = keycloak_openid_client_scope.clientscope.id
	role_id         = keycloak_role.role.id
}
	`, testAccRealm.Realm, clientName, roleName, clientScopeName)
}

func testKeycloakGenericClientRoleMapper_basicClientScopeRealmRole(roleName, clientScopeName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_role" "role" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_openid_client_scope" "clientscope" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_generic_client_role_mapper" "clientscope-with-realm-role" {
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = keycloak_openid_client_scope.clientscope.id
	role_id         = keycloak_role.role.id
}
	`, testAccRealm.Realm, roleName, clientScopeName)
}
