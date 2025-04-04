package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"testing"
)

func TestAccKeycloakOpenIdFullNameProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_full_name_protocol_mapper.full_name_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdFullNameProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdFullNameProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdFullNameProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdFullNameProtocolMapper_basicClientScope(t *testing.T) {
	t.Parallel()
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_full_name_protocol_mapper.full_name_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdFullNameProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdFullNameProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdFullNameProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdFullNameProtocolMapper_import(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	clientResourceName := "keycloak_openid_full_name_protocol_mapper.full_name_mapper_client"
	clientScopeResourceName := "keycloak_openid_full_name_protocol_mapper.full_name_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdFullNameProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdFullNameProtocolMapper_import(clientId, clientScopeId, mapperName),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakOpenIdFullNameProtocolMapperExists(clientResourceName),
					testKeycloakOpenIdFullNameProtocolMapperExists(clientScopeResourceName),
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

func TestAccKeycloakOpenIdFullNameProtocolMapper_update(t *testing.T) {
	t.Parallel()
	resourceName := "keycloak_openid_full_name_protocol_mapper.full_name_mapper"

	mapperOne := &keycloak.OpenIdFullNameProtocolMapper{
		Name:             acctest.RandString(10),
		ClientId:         "terraform-client-" + acctest.RandString(10),
		AddToIdToken:     randomBool(),
		AddToAccessToken: randomBool(),
		AddToUserInfo:    randomBool(),
	}

	mapperTwo := &keycloak.OpenIdFullNameProtocolMapper{
		Name:             mapperOne.Name,
		ClientId:         mapperOne.ClientId,
		AddToIdToken:     randomBool(),
		AddToAccessToken: randomBool(),
		AddToUserInfo:    randomBool(),
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdFullNameProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdFullNameProtocolMapper_fromInterface(mapperOne),
				Check:  testKeycloakOpenIdFullNameProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdFullNameProtocolMapper_fromInterface(mapperTwo),
				Check:  testKeycloakOpenIdFullNameProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdFullNameProtocolMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.OpenIdFullNameProtocolMapper{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_full_name_protocol_mapper.full_name_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdFullNameProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdFullNameProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdFullNameProtocolMapperFetch(resourceName, mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenIdUserAttributeProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
					if err != nil {
						t.Error(err)
					}
				},
				Config: testKeycloakOpenIdFullNameProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdFullNameProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdFullNameProtocolMapper_updateMapperNameForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperNameOne := acctest.RandomWithPrefix("tf-acc")
	mapperNameTwo := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_full_name_protocol_mapper.full_name_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdFullNameProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdFullNameProtocolMapper_basic_client(clientId, mapperNameOne),
				Check:  testKeycloakOpenIdFullNameProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdFullNameProtocolMapper_basic_client(clientId, mapperNameTwo),
				Check:  testKeycloakOpenIdFullNameProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdFullNameProtocolMapper_updateClientIdForceNew(t *testing.T) {
	t.Parallel()
	clientIdOne := acctest.RandomWithPrefix("tf-acc")
	clientIdTwo := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_full_name_protocol_mapper.full_name_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdFullNameProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdFullNameProtocolMapper_updateClientForceNew(clientIdOne, clientIdTwo, "openid_client_one"),
				Check:  testKeycloakOpenIdFullNameProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdFullNameProtocolMapper_updateClientForceNew(clientIdOne, clientIdTwo, "openid_client_two"),
				Check:  testKeycloakOpenIdFullNameProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdFullNameProtocolMapper_updateClientScopeForceNew(t *testing.T) {
	t.Parallel()
	clientScopeOne := acctest.RandomWithPrefix("tf-acc")
	clientScopeTwo := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_full_name_protocol_mapper.full_name_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdFullNameProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdFullNameProtocolMapper_updateClientScopeForceNew(clientScopeOne, clientScopeTwo, "client_scope_one"),
				Check:  testKeycloakOpenIdFullNameProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdFullNameProtocolMapper_updateClientScopeForceNew(clientScopeOne, clientScopeTwo, "client_scope_two"),
				Check:  testKeycloakOpenIdFullNameProtocolMapperExists(resourceName),
			},
		},
	})
}

func testAccKeycloakOpenIdFullNameProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_openid_full_name_protocol_mapper" {
				continue
			}

			mapper, _ := getFullNameMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("openid user attribute protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakOpenIdFullNameProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getFullNameMapperUsingState(state, resourceName)

		if err != nil {
			return err
		}

		return nil
	}
}

func testKeycloakOpenIdFullNameProtocolMapperFetch(resourceName string, mapper *keycloak.OpenIdFullNameProtocolMapper) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedMapper, err := getFullNameMapperUsingState(state, resourceName)
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

func getFullNameMapperUsingState(state *terraform.State, resourceName string) (*keycloak.OpenIdFullNameProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	return keycloakClient.GetOpenIdFullNameProtocolMapper(testCtx, realm, clientId, clientScopeId, id)
}

func getGenericProtocolMapperIdForClient(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmId := rs.Primary.Attributes["realm_id"]
		clientId := rs.Primary.Attributes["client_id"]

		return fmt.Sprintf("%s/client/%s/%s", realmId, clientId, id), nil
	}
}

func getGenericProtocolMapperIdForClientScope(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmId := rs.Primary.Attributes["realm_id"]
		clientScopeId := rs.Primary.Attributes["client_scope_id"]

		return fmt.Sprintf("%s/client-scope/%s/%s", realmId, clientScopeId, id), nil
	}
}

func testKeycloakOpenIdFullNameProtocolMapper_basic_client(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_full_name_protocol_mapper" "full_name_mapper_client" {
	name       = "%s"
	realm_id   = data.keycloak_realm.realm.id
	client_id  = "${keycloak_openid_client.openid_client.id}"
}`, testAccRealm.Realm, clientId, mapperName)
}

func testKeycloakOpenIdFullNameProtocolMapper_basic_clientScope(clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_full_name_protocol_mapper" "full_name_mapper_client_scope" {
	name            = "%s"
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = "${keycloak_openid_client_scope.client_scope.id}"
}`, testAccRealm.Realm, clientScopeId, mapperName)
}

func testKeycloakOpenIdFullNameProtocolMapper_import(clientId, clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_full_name_protocol_mapper" "full_name_mapper_client" {
	name       = "%s"
	realm_id   = data.keycloak_realm.realm.id
	client_id  = "${keycloak_openid_client.openid_client.id}"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_full_name_protocol_mapper" "full_name_mapper_client_scope" {
	name            = "%s"
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = "${keycloak_openid_client_scope.client_scope.id}"
}`, testAccRealm.Realm, clientId, mapperName, clientScopeId, mapperName)
}

func testKeycloakOpenIdFullNameProtocolMapper_fromInterface(mapper *keycloak.OpenIdFullNameProtocolMapper) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_full_name_protocol_mapper" "full_name_mapper" {
	name                = "%s"
	realm_id            = data.keycloak_realm.realm.id
	client_id           = "${keycloak_openid_client.openid_client.id}"

	add_to_id_token     = %t
	add_to_access_token = %t
	add_to_userinfo     = %t
}`, testAccRealm.Realm, mapper.ClientId, mapper.Name, mapper.AddToIdToken, mapper.AddToAccessToken, mapper.AddToUserInfo)
}

func testKeycloakOpenIdFullNameProtocolMapper_updateClientForceNew(clientIdOne, clientIdTwo, currentClient string) string {
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

resource "keycloak_openid_full_name_protocol_mapper" "full_name_mapper_client" {
	name       = "group-mapper"
	realm_id   = data.keycloak_realm.realm.id
	client_id  = "${keycloak_openid_client.%s.id}"
}`, testAccRealm.Realm, clientIdOne, clientIdTwo, currentClient)
}

func testKeycloakOpenIdFullNameProtocolMapper_updateClientScopeForceNew(clientScopeIdOne, clientScopeIdTwo, currentClientScope string) string {
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

resource "keycloak_openid_full_name_protocol_mapper" "full_name_mapper_client_scope" {
	name            = "group-mapper"
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = "${keycloak_openid_client_scope.%s.id}"
}`, testAccRealm.Realm, clientScopeIdOne, clientScopeIdTwo, currentClientScope)
}
