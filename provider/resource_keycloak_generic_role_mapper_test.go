package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakGenericRoleMapper_basic(t *testing.T) {
	t.Parallel()

	parentClientName := acctest.RandomWithPrefix("tf-acc")
	parentRoleName := acctest.RandomWithPrefix("tf-acc")
	childClientName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericRoleMapper_basic(parentClientName, parentRoleName, childClientName),
				Check:  testAccCheckKeycloakGenericRoleMapperExists("keycloak_generic_role_mapper.child-client-with-parent-client-role"),
			},
		},
	})
}

func TestAccKeycloakGenericRoleMapper_createAfterManualDestroy(t *testing.T) {
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
				Config: testKeycloakGenericRoleMapper_basic(parentClientName, parentRoleName, childClientName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGenericRoleMapperExists("keycloak_generic_role_mapper.child-client-with-parent-client-role"),
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
				Config: testKeycloakGenericRoleMapper_basic(parentClientName, parentRoleName, childClientName),
				Check:  testAccCheckKeycloakGenericRoleMapperExists("keycloak_generic_role_mapper.child-client-with-parent-client-role"),
			},
		},
	})
}

func TestAccKeycloakGenericRoleMapper_import(t *testing.T) {
	t.Parallel()

	parentClientName := acctest.RandomWithPrefix("tf-acc")
	parentRoleName := acctest.RandomWithPrefix("tf-acc")
	childClientName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_generic_role_mapper.child-client-with-parent-client-role"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericRoleMapper_basic(parentClientName, parentRoleName, childClientName),
				Check:  testAccCheckKeycloakGenericRoleMapperExists(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getGenericRoleMapperId(resourceName),
			},
		},
	})
}

func TestAccKeycloakGenericRoleMapper_basicClientScope(t *testing.T) {
	t.Parallel()

	clientName := acctest.RandomWithPrefix("tf-acc")
	roleName := acctest.RandomWithPrefix("tf-acc")
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericRoleMapper_basicClientScope(clientName, roleName, clientScopeName),
				Check:  testAccCheckKeycloakGenericRoleMapperExists("keycloak_generic_role_mapper.clientscope-with-client-role"),
			},
		},
	})
}

func TestAccKeycloakGenericRoleMapper_importClientScope(t *testing.T) {
	t.Parallel()

	clientName := acctest.RandomWithPrefix("tf-acc")
	roleName := acctest.RandomWithPrefix("tf-acc")
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_generic_role_mapper.clientscope-with-client-role"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericRoleMapper_basicClientScope(clientName, roleName, clientScopeName),
				Check:  testAccCheckKeycloakGenericRoleMapperExists(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getGenericRoleMapperId(resourceName),
			},
		},
	})
}

func TestAccKeycloakGenericRoleMapper_basicClientScopeRealmRole(t *testing.T) {
	t.Parallel()

	roleName := acctest.RandomWithPrefix("tf-acc")
	clientScopeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericRoleMapper_basicClientScopeRealmRole(roleName, clientScopeName),
				Check:  testAccCheckKeycloakGenericRoleMapperExists("keycloak_generic_role_mapper.clientscope-with-realm-role"),
			},
		},
	})
}

func TestAccKeycloakGenericRoleMapper_deleteIndividualMappers(t *testing.T) {
	t.Parallel()

	var someRole = &keycloak.Role{}
	var someOtherRole = &keycloak.Role{}
	var client = &keycloak.GenericClient{}

	clientName := acctest.RandomWithPrefix("tf-acc")
	someRoleName := acctest.RandomWithPrefix("tf-acc")
	someOtherRoleName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakGenericRoleMapperDestroy("keycloak_generic_role_mapper.client-with-some-role"),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericRoleMapper_basicClientDedicatedAllRealmRoles(clientName, someRoleName, someOtherRoleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGenericClientRoleMapperExists("keycloak_generic_role_mapper.client-with-some-role"),
					testAccCheckKeycloakGenericClientRoleMapperExists("keycloak_generic_role_mapper.client-with-some-other-role"),
					testAccCheckKeycloakRoleFetch("keycloak_role.some-role", someRole),
					testAccCheckKeycloakRoleFetch("keycloak_role.some-other-role", someOtherRole),
					testAccCheckKeycloakGenericClientFetch("keycloak_openid_client.client", client),
				),
			},
			{
				Config: testKeycloakGenericRoleMapper_basicClientDedicatedPartialRealmRoles(clientName, someRoleName, someOtherRoleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGenericClientRoleMapperExists("keycloak_generic_role_mapper.client-with-some-other-role"),
				),
			},
		},
	})
}

func testAccCheckKeycloakGenericRoleMapperExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		return nil
	}
}

func getGenericRoleMapperId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		return rs.Primary.ID, nil
	}
}

func testAccCheckKeycloakGenericRoleMapperDestroy(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if ok {
			return fmt.Errorf("resource should not exist: %s", resourceName)
		}
		return nil
	}
}

func testKeycloakGenericRoleMapper_basic(parentClientName, parentRoleName, childClientName string) string {
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

resource "keycloak_generic_role_mapper" "child-client-with-parent-client-role" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.child-client.id
	role_id   = keycloak_role.parent-role.id
}
	`, testAccRealm.Realm, parentClientName, parentRoleName, childClientName)
}

func testKeycloakGenericRoleMapper_basicClientScope(clientName, roleName, clientScopeName string) string {
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

resource "keycloak_generic_role_mapper" "clientscope-with-client-role" {
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = keycloak_openid_client_scope.clientscope.id
	role_id         = keycloak_role.role.id
}
	`, testAccRealm.Realm, clientName, roleName, clientScopeName)
}

func testKeycloakGenericRoleMapper_basicClientScopeRealmRole(roleName, clientScopeName string) string {
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

resource "keycloak_generic_role_mapper" "clientscope-with-realm-role" {
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = keycloak_openid_client_scope.clientscope.id
	role_id         = keycloak_role.role.id
}
	`, testAccRealm.Realm, roleName, clientScopeName)
}

func testKeycloakGenericRoleMapper_basicClientDedicatedAllRealmRoles(clientName, someRoleName, someOtherRoleName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	access_type = "PUBLIC"
}

resource "keycloak_role" "some-role" {
	realm_id  = data.keycloak_realm.realm.id
	name      = "%s"
}

resource "keycloak_role" "some-other-role" {
	realm_id  = data.keycloak_realm.realm.id
	name      = "%s"
}

resource "keycloak_generic_role_mapper" "client-with-some-role" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.client.id
	role_id   = keycloak_role.some-role.id
}

resource "keycloak_generic_role_mapper" "client-with-some-other-role" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.client.id
	role_id   = keycloak_role.some-other-role.id
}
	`, testAccRealm.Realm, clientName, someRoleName, someOtherRoleName)
}

func testKeycloakGenericRoleMapper_basicClientDedicatedPartialRealmRoles(clientName, someRoleName, someOtherRoleName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	access_type = "PUBLIC"
}

resource "keycloak_role" "some-role" {
	realm_id  = data.keycloak_realm.realm.id
	name      = "%s"
}

resource "keycloak_role" "some-other-role" {
	realm_id  = data.keycloak_realm.realm.id
	name      = "%s"
}

resource "keycloak_generic_role_mapper" "client-with-some-other-role" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.client.id
	role_id   = keycloak_role.some-other-role.id
}
	`, testAccRealm.Realm, clientName, someRoleName, someOtherRoleName)
}
