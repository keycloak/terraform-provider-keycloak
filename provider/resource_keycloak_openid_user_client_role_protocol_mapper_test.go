package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenIdUserClientRoleProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserClientRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserClientRoleProtocolMapper_basicClientScope(t *testing.T) {
	t.Parallel()
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserClientRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserClientRoleProtocolMapper_import(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	clientResourceName := "keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper_client"
	clientScopeResourceName := "keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserClientRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_import(clientId, clientScopeId, mapperName),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakOpenIdUserClientRoleProtocolMapperExists(clientResourceName),
					testKeycloakOpenIdUserClientRoleProtocolMapperExists(clientScopeResourceName),
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

func TestAccKeycloakOpenIdUserClientRoleProtocolMapper_update(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	claimName := acctest.RandomWithPrefix("tf-acc")
	updatedClaimName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserClientRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_claim(clientId, mapperName, claimName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_claim(clientId, mapperName, updatedClaimName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserClientRoleProtocolMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.OpenIdUserClientRoleProtocolMapper{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserClientRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperFetch(resourceName, mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenIdUserClientRoleProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
					if err != nil {
						t.Error(err)
					}
				},
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserClientRoleProtocolMapper_validateClaimValueType(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	invalidClaimValueType := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserClientRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenIdUserClientRoleProtocolMapper_validateClaimValueType(mapperName, invalidClaimValueType),
				ExpectError: regexp.MustCompile("expected claim_value_type to be one of .+ got " + invalidClaimValueType),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserClientRoleProtocolMapper_updateClientIdForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	updatedClientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	claimName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserClientRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_claim(clientId, mapperName, claimName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_claim(updatedClientId, mapperName, claimName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserClientRoleProtocolMapper_updateClientScopeForceNew(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	newClientScopeId := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserClientRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_basic_clientScope(newClientScopeId, mapperName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserClientRoleProtocolMapper_updateRealmIdForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	claimName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserClientRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_claim(clientId, mapperName, claimName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_claim(clientId, mapperName, claimName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserClientRoleProtocolMapper_clientAssignment(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	assignedClientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper_validation"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserClientRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_clientAssignment(clientId, assignedClientId, mapperName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_clientAssignment(clientId, assignedClientId, mapperName),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserClientRoleProtocolMapper_clientAssignmentRolePrefix(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	assignedClientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")
	rolePrefix := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_client_role_protocol_mapper.user_client_role_mapper_validation"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserClientRoleProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_clientAssignmentRolePrefix(clientId, assignedClientId, mapperName, rolePrefix),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserClientRoleProtocolMapper_clientAssignmentRolePrefix(clientId, assignedClientId, mapperName, rolePrefix),
				Check:  testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName),
			},
		},
	})
}

func testAccKeycloakOpenIdUserClientRoleProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_openid_user_client_role_protocol_mapper" {
				continue
			}

			mapper, _ := getUserClientRoleMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("openid user attribute protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakOpenIdUserClientRoleProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getUserClientRoleMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testKeycloakOpenIdUserClientRoleProtocolMapperFetch(resourceName string, mapper *keycloak.OpenIdUserClientRoleProtocolMapper) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedMapper, err := getUserClientRoleMapperUsingState(state, resourceName)
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

func getUserClientRoleMapperUsingState(state *terraform.State, resourceName string) (*keycloak.OpenIdUserClientRoleProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	return keycloakClient.GetOpenIdUserClientRoleProtocolMapper(testCtx, realm, clientId, clientScopeId, id)
}

func testKeycloakOpenIdUserClientRoleProtocolMapper_basic_client(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}
resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"
	access_type = "BEARER-ONLY"
}
resource "keycloak_openid_user_client_role_protocol_mapper" "user_client_role_mapper_client" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_id        = "${keycloak_openid_client.openid_client.id}"
	claim_name       = "foo"
	claim_value_type = "String"
}`, testAccRealm.Realm, clientId, mapperName)
}

func testKeycloakOpenIdUserClientRoleProtocolMapper_basic_clientScope(clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}
resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}
resource "keycloak_openid_user_client_role_protocol_mapper" "user_client_role_mapper_client_scope" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_scope_id  = "${keycloak_openid_client_scope.client_scope.id}"
	claim_name       = "foo"
	claim_value_type = "String"
}`, testAccRealm.Realm, clientScopeId, mapperName)
}

func testKeycloakOpenIdUserClientRoleProtocolMapper_claim(clientId, mapperName, claimName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}
resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"
	access_type = "BEARER-ONLY"
}
resource "keycloak_openid_user_client_role_protocol_mapper" "user_client_role_mapper" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_id        = "${keycloak_openid_client.openid_client.id}"
	claim_name       = "%s"
	claim_value_type = "String"
}`, testAccRealm.Realm, clientId, mapperName, claimName)
}

func testKeycloakOpenIdUserClientRoleProtocolMapper_import(clientId, clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}
resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	access_type = "BEARER-ONLY"
}
resource "keycloak_openid_user_client_role_protocol_mapper" "user_client_role_mapper_client" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_id        = "${keycloak_openid_client.openid_client.id}"
	claim_name       = "foo"
	claim_value_type = "String"
}
resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}
resource "keycloak_openid_user_client_role_protocol_mapper" "user_client_role_mapper_client_scope" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_scope_id  = "${keycloak_openid_client_scope.client_scope.id}"
	claim_name       = "foo"
	claim_value_type = "String"
}`, testAccRealm.Realm, clientId, mapperName, clientScopeId, mapperName)
}

func testKeycloakOpenIdUserClientRoleProtocolMapper_validateClaimValueType(mapperName, claimValueType string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}
resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "openid-client"
	access_type = "BEARER-ONLY"
}
resource "keycloak_openid_user_client_role_protocol_mapper" "user_client_role_mapper_validation" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_id        = "${keycloak_openid_client.openid_client.id}"
	claim_name      = "foo"
	claim_value_type = "%s"
}`, testAccRealm.Realm, mapperName, claimValueType)
}

func testKeycloakOpenIdUserClientRoleProtocolMapper_clientAssignment(clientId, assignedClientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}
resource "keycloak_openid_client" "openid_client_assigned" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_user_client_role_protocol_mapper" "user_client_role_mapper_validation" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_id        = "${keycloak_openid_client.openid_client.id}"

	claim_name      = "foo"
	claim_value_type = "String"
	client_id_for_role_mappings = "${keycloak_openid_client.openid_client_assigned.id}"
}`, testAccRealm.Realm, clientId, assignedClientId, mapperName)
}

func testKeycloakOpenIdUserClientRoleProtocolMapper_clientAssignmentRolePrefix(clientId, assignedClientId, mapperName, rolePrefix string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}
resource "keycloak_openid_client" "openid_client_assigned" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_user_client_role_protocol_mapper" "user_client_role_mapper_validation" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_id        = "${keycloak_openid_client.openid_client.id}"

	claim_name      = "foo"
	claim_value_type = "String"
	client_id_for_role_mappings = "${keycloak_openid_client.openid_client_assigned.id}"
	client_role_prefix= "%s"
}`, testAccRealm.Realm, clientId, assignedClientId, mapperName, rolePrefix)
}
