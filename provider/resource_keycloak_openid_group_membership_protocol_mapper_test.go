package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"testing"
)

func TestAccKeycloakOpenIdGroupMembershipProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_group_membership_protocol_mapper.group_membership_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdGroupMembershipProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdGroupMembershipProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdGroupMembershipProtocolMapper_basicClientScope(t *testing.T) {
	t.Parallel()
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_group_membership_protocol_mapper.group_membership_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdGroupMembershipProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdGroupMembershipProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdGroupMembershipProtocolMapper_import(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	clientResourceName := "keycloak_openid_group_membership_protocol_mapper.group_membership_mapper_client"
	clientScopeResourceName := "keycloak_openid_group_membership_protocol_mapper.group_membership_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdFullNameProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_import(clientId, clientScopeId, mapperName),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakOpenIdGroupMembershipProtocolMapperExists(clientResourceName),
					testKeycloakOpenIdGroupMembershipProtocolMapperExists(clientScopeResourceName),
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

func TestAccKeycloakOpenIdGroupMembershipProtocolMapper_update(t *testing.T) {
	t.Parallel()
	resourceName := "keycloak_openid_group_membership_protocol_mapper.group_membership_mapper"

	mapperOne := &keycloak.OpenIdGroupMembershipProtocolMapper{
		Name:             acctest.RandString(10),
		ClientId:         "terraform-client-" + acctest.RandString(10),
		ClaimName:        acctest.RandString(10),
		FullPath:         randomBool(),
		AddToIdToken:     randomBool(),
		AddToAccessToken: randomBool(),
		AddToUserinfo:    randomBool(),
	}

	mapperTwo := &keycloak.OpenIdGroupMembershipProtocolMapper{
		Name:             mapperOne.Name,
		ClientId:         mapperOne.ClientId,
		ClaimName:        acctest.RandString(10),
		FullPath:         randomBool(),
		AddToIdToken:     randomBool(),
		AddToAccessToken: randomBool(),
		AddToUserinfo:    randomBool(),
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdGroupMembershipProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_fromInterface(mapperOne),
				Check:  testKeycloakOpenIdGroupMembershipProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_fromInterface(mapperTwo),
				Check:  testKeycloakOpenIdGroupMembershipProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdGroupMembershipProtocolMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.OpenIdGroupMembershipProtocolMapper{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_group_membership_protocol_mapper.group_membership_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdGroupMembershipProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdGroupMembershipProtocolMapperFetch(resourceName, mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenIdUserAttributeProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
					if err != nil {
						t.Error(err)
					}
				},
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdGroupMembershipProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdGroupMembershipProtocolMapper_updateMapperNameForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperNameOne := acctest.RandomWithPrefix("tf-acc")
	mapperNameTwo := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_group_membership_protocol_mapper.group_membership_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdGroupMembershipProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_basic_client(clientId, mapperNameOne),
				Check:  testKeycloakOpenIdGroupMembershipProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_basic_client(clientId, mapperNameTwo),
				Check:  testKeycloakOpenIdGroupMembershipProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdGroupMembershipProtocolMapper_updateClientIdForceNew(t *testing.T) {
	t.Parallel()
	clientIdOne := acctest.RandomWithPrefix("tf-acc")
	clientIdTwo := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_group_membership_protocol_mapper.group_membership_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdGroupMembershipProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_updateClientForceNew(clientIdOne, clientIdTwo, "openid_client_one"),
				Check:  testKeycloakOpenIdGroupMembershipProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_updateClientForceNew(clientIdOne, clientIdTwo, "openid_client_two"),
				Check:  testKeycloakOpenIdGroupMembershipProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdGroupMembershipProtocolMapper_updateClientScopeForceNew(t *testing.T) {
	t.Parallel()
	clientScopeOne := acctest.RandomWithPrefix("tf-acc")
	clientScopeTwo := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_group_membership_protocol_mapper.group_membership_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdGroupMembershipProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_updateClientScopeForceNew(clientScopeOne, clientScopeTwo, "client_scope_one"),
				Check:  testKeycloakOpenIdGroupMembershipProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdGroupMembershipProtocolMapper_updateClientScopeForceNew(clientScopeOne, clientScopeTwo, "client_scope_two"),
				Check:  testKeycloakOpenIdGroupMembershipProtocolMapperExists(resourceName),
			},
		},
	})
}

func testAccKeycloakOpenIdGroupMembershipProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_openid_group_membership_protocol_mapper" {
				continue
			}

			mapper, _ := getGroupMembershipMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("openid user attribute protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakOpenIdGroupMembershipProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getGroupMembershipMapperUsingState(state, resourceName)

		if err != nil {
			return err
		}

		return nil
	}
}

func testKeycloakOpenIdGroupMembershipProtocolMapperFetch(resourceName string, mapper *keycloak.OpenIdGroupMembershipProtocolMapper) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedMapper, err := getGroupMembershipMapperUsingState(state, resourceName)
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

func getGroupMembershipMapperUsingState(state *terraform.State, resourceName string) (*keycloak.OpenIdGroupMembershipProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	return keycloakClient.GetOpenIdGroupMembershipProtocolMapper(testCtx, realm, clientId, clientScopeId, id)
}

func testKeycloakOpenIdGroupMembershipProtocolMapper_basic_client(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_group_membership_protocol_mapper" "group_membership_mapper_client" {
	name       = "%s"
	realm_id   = data.keycloak_realm.realm.id
	client_id  = "${keycloak_openid_client.openid_client.id}"
	claim_name = "bar"
}`, testAccRealm.Realm, clientId, mapperName)
}

func testKeycloakOpenIdGroupMembershipProtocolMapper_basic_clientScope(clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_group_membership_protocol_mapper" "group_membership_mapper_client_scope" {
	name            = "%s"
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = "${keycloak_openid_client_scope.client_scope.id}"
	claim_name      = "bar"
}`, testAccRealm.Realm, clientScopeId, mapperName)
}

func testKeycloakOpenIdGroupMembershipProtocolMapper_import(clientId, clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_group_membership_protocol_mapper" "group_membership_mapper_client" {
	name       = "%s"
	realm_id   = data.keycloak_realm.realm.id
	client_id  = "${keycloak_openid_client.openid_client.id}"
	claim_name = "bar"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_group_membership_protocol_mapper" "group_membership_mapper_client_scope" {
	name            = "%s"
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = "${keycloak_openid_client_scope.client_scope.id}"
	claim_name      = "bar"
}`, testAccRealm.Realm, clientId, mapperName, clientScopeId, mapperName)
}

func testKeycloakOpenIdGroupMembershipProtocolMapper_fromInterface(mapper *keycloak.OpenIdGroupMembershipProtocolMapper) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_group_membership_protocol_mapper" "group_membership_mapper" {
	name                = "%s"
	realm_id            = data.keycloak_realm.realm.id
	client_id           = "${keycloak_openid_client.openid_client.id}"

	claim_name          = "%s"
	full_path           = %t
	add_to_id_token     = %t
	add_to_access_token = %t
	add_to_userinfo     = %t
}`, testAccRealm.Realm, mapper.ClientId, mapper.Name, mapper.ClaimName, mapper.FullPath, mapper.AddToIdToken, mapper.AddToAccessToken, mapper.AddToUserinfo)
}

func testKeycloakOpenIdGroupMembershipProtocolMapper_updateClientForceNew(clientIdOne, clientIdTwo, currentClient string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client_one" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_client" "openid_client_two" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_group_membership_protocol_mapper" "group_membership_mapper_client" {
	name       = "group-mapper"
	realm_id   = data.keycloak_realm.realm.id
	client_id  = "${keycloak_openid_client.%s.id}"

	claim_name = "foo"
}`, testAccRealm.Realm, clientIdOne, clientIdTwo, currentClient)
}

func testKeycloakOpenIdGroupMembershipProtocolMapper_updateClientScopeForceNew(clientScopeIdOne, clientScopeIdTwo, currentClientScope string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope_one" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_client_scope" "client_scope_two" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_group_membership_protocol_mapper" "group_membership_mapper_client_scope" {
	name            = "group-mapper"
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = "${keycloak_openid_client_scope.%s.id}"

	claim_name      = "foo"
}`, testAccRealm.Realm, clientScopeIdOne, clientScopeIdTwo, currentClientScope)
}
