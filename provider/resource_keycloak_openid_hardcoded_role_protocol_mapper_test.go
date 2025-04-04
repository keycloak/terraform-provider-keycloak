package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"testing"
)

func TestAccKeycloakOpenIdHardcodedRoleProtocolMapper_basicRealmRole_client(t *testing.T) {
	t.Parallel()
	role := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_hardcoded_role_protocol_mapper.hardcoded_role_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedRoleProtocolMapper_basicRealmRole_client(role, clientId, mapperName),
				Check:  testKeycloakOpenIdHardcodedRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdHardcodedRoleProtocolMapper_basicClientRole_client(t *testing.T) {
	t.Parallel()
	clientIdForRole := acctest.RandomWithPrefix("tf-acc")
	role := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_hardcoded_role_protocol_mapper.hardcoded_role_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedRoleProtocolMapper_basicClientRole_client(clientIdForRole, role, clientId, mapperName),
				Check:  testKeycloakOpenIdHardcodedRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdHardcodedRoleProtocolMapper_basicRealmRole_clientScope(t *testing.T) {
	t.Parallel()
	role := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_hardcoded_role_protocol_mapper.hardcoded_role_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedRoleProtocolMapper_basicRealmRole_clientScope(role, clientScopeId, mapperName),
				Check:  testKeycloakOpenIdHardcodedRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdHardcodedRoleProtocolMapper_import(t *testing.T) {
	t.Parallel()
	role := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	clientResourceName := "keycloak_openid_hardcoded_role_protocol_mapper.hardcoded_role_mapper_client"
	clientScopeResourceName := "keycloak_openid_hardcoded_role_protocol_mapper.hardcoded_role_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdFullNameProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedRoleProtocolMapper_import(role, clientId, clientScopeId, mapperName),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakOpenIdHardcodedRoleProtocolMapperExists(clientResourceName),
					testKeycloakOpenIdHardcodedRoleProtocolMapperExists(clientScopeResourceName),
				),
			},
			{
				ResourceName:      clientResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getGenericProtocolMapperIdForClient(clientResourceName),
			},
			{
				ResourceName:      clientScopeResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getGenericProtocolMapperIdForClientScope(clientScopeResourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdHardcodedRoleProtocolMapper_update(t *testing.T) {
	t.Parallel()
	roleOne := acctest.RandomWithPrefix("tf-acc")
	roleTwo := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_hardcoded_role_protocol_mapper.hardcoded_role_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedRoleProtocolMapper_basicRealmRole_clientUpdateBefore(roleOne, roleTwo, clientId, mapperName),
				Check:  testKeycloakOpenIdHardcodedRoleProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdHardcodedRoleProtocolMapper_basicRealmRole_clientUpdateAfter(roleOne, roleTwo, clientId, mapperName),
				Check:  testKeycloakOpenIdHardcodedRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdHardcodedRoleProtocolMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.OpenIdHardcodedRoleProtocolMapper{}

	role := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_hardcoded_role_protocol_mapper.hardcoded_role_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedRoleProtocolMapper_basicRealmRole_client(role, clientId, mapperName),
				Check:  testKeycloakOpenIdHardcodedRoleProtocolMapperFetch(resourceName, mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenIdHardcodedRoleProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
					if err != nil {
						t.Error(err)
					}
				},
				Config: testKeycloakOpenIdHardcodedRoleProtocolMapper_basicRealmRole_client(role, clientId, mapperName),
				Check:  testKeycloakOpenIdHardcodedRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func testAccKeycloakOpenIdHardcodedRoleProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_openid_hardcoded_role_protocol_mapper" {
				continue
			}

			mapper, _ := getHardcodedRoleMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("openid user attribute protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakOpenIdHardcodedRoleProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getHardcodedRoleMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testKeycloakOpenIdHardcodedRoleProtocolMapperFetch(resourceName string, mapper *keycloak.OpenIdHardcodedRoleProtocolMapper) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedMapper, err := getHardcodedRoleMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		mapper.Id = fetchedMapper.Id
		mapper.ClientId = fetchedMapper.ClientId
		mapper.ClientScopeId = fetchedMapper.ClientScopeId
		mapper.RealmId = fetchedMapper.RealmId

		return nil
	}
}

func getHardcodedRoleMapperUsingState(state *terraform.State, resourceName string) (*keycloak.OpenIdHardcodedRoleProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	return keycloakClient.GetOpenIdHardcodedRoleProtocolMapper(testCtx, realm, clientId, clientScopeId, id)
}

func testKeycloakOpenIdHardcodedRoleProtocolMapper_basicRealmRole_client(role, clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_role" "role" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_hardcoded_role_protocol_mapper" "hardcoded_role_mapper_client" {
	name           = "%s"
	realm_id       = data.keycloak_realm.realm.id
	client_id      = "${keycloak_openid_client.openid_client.id}"
	role_id        = "${keycloak_role.role.id}"
}`, testAccRealm.Realm, role, clientId, mapperName)
}

func testKeycloakOpenIdHardcodedRoleProtocolMapper_basicRealmRole_clientScope(role, clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_role" "role" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_hardcoded_role_protocol_mapper" "hardcoded_role_mapper_client_scope" {
	name            = "%s"
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = "${keycloak_openid_client_scope.client_scope.id}"
	role_id         = "${keycloak_role.role.id}"
}`, testAccRealm.Realm, role, clientScopeId, mapperName)
}

func testKeycloakOpenIdHardcodedRoleProtocolMapper_import(role, clientId, clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_role" "role" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_hardcoded_role_protocol_mapper" "hardcoded_role_mapper_client" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_id        = "${keycloak_openid_client.openid_client.id}"

	role_id          = "${keycloak_role.role.id}"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_hardcoded_role_protocol_mapper" "hardcoded_role_mapper_client_scope" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_scope_id  = "${keycloak_openid_client_scope.client_scope.id}"

	role_id          = "${keycloak_role.role.id}"
}`, testAccRealm.Realm, role, clientId, mapperName, clientScopeId, mapperName)
}

func testKeycloakOpenIdHardcodedRoleProtocolMapper_basicRealmRole_clientUpdateBefore(roleOne, roleTwo, clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_role" "role_one" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_role" "role_two" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_hardcoded_role_protocol_mapper" "hardcoded_role_mapper_client" {
	name           = "%s"
	realm_id       = data.keycloak_realm.realm.id
	client_id      = "${keycloak_openid_client.openid_client.id}"
	role_id        = "${keycloak_role.role_one.id}"
}`, testAccRealm.Realm, roleOne, roleTwo, clientId, mapperName)
}

func testKeycloakOpenIdHardcodedRoleProtocolMapper_basicRealmRole_clientUpdateAfter(roleOne, roleTwo, clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_role" "role_one" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_role" "role_two" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_hardcoded_role_protocol_mapper" "hardcoded_role_mapper_client" {
	name           = "%s"
	realm_id       = data.keycloak_realm.realm.id
	client_id      = "${keycloak_openid_client.openid_client.id}"
	role_id        = "${keycloak_role.role_two.id}"
}`, testAccRealm.Realm, roleOne, roleTwo, clientId, mapperName)
}

func testKeycloakOpenIdHardcodedRoleProtocolMapper_basicClientRole_client(clientIdForRole, role, clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client_for_role" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_role" "role" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = "${keycloak_openid_client.openid_client_for_role.id}"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_hardcoded_role_protocol_mapper" "hardcoded_role_mapper_client" {
	name           = "%s"
	realm_id       = data.keycloak_realm.realm.id
	client_id      = "${keycloak_openid_client.openid_client.id}"
	role_id        = "${keycloak_role.role.id}"
}`, testAccRealm.Realm, clientIdForRole, role, clientId, mapperName)
}
