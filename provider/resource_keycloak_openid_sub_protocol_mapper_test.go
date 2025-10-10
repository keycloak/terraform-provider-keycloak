package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenIdSubProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_25)

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_sub_protocol_mapper.sub_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdSubProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdSubProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdSubProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdSubProtocolMapper_basicClientScope(t *testing.T) {
	t.Parallel()
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_25)

	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_sub_protocol_mapper.sub_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdSubProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdSubProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdSubProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdSubProtocolMapper_import(t *testing.T) {
	t.Parallel()
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_25)

	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	clientResourceName := "keycloak_openid_sub_protocol_mapper.sub_mapper_client"
	clientScopeResourceName := "keycloak_openid_sub_protocol_mapper.sub_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdSubProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdSubProtocolMapper_import(clientId, clientScopeId, mapperName),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakOpenIdSubProtocolMapperExists(clientResourceName),
					testKeycloakOpenIdSubProtocolMapperExists(clientScopeResourceName),
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

func TestAccKeycloakOpenIdSubProtocolMapper_update(t *testing.T) {
	t.Parallel()
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_25)

	resourceName := "keycloak_openid_sub_protocol_mapper.sub_mapper"

	mapperOne := &keycloak.OpenIdSubProtocolMapper{
		Name:                    acctest.RandString(10),
		ClientId:                "terraform-client-" + acctest.RandString(10),
		AddToAccessToken:        randomBool(),
		AddToTokenIntrospection: randomBool(),
	}

	mapperTwo := &keycloak.OpenIdSubProtocolMapper{
		Name:                    mapperOne.Name,
		ClientId:                mapperOne.ClientId,
		AddToAccessToken:        randomBool(),
		AddToTokenIntrospection: randomBool(),
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdSubProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdSubProtocolMapper_fromInterface(mapperOne),
				Check:  testKeycloakOpenIdSubProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdSubProtocolMapper_fromInterface(mapperTwo),
				Check:  testKeycloakOpenIdSubProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdSubProtocolMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_25)

	var mapper = &keycloak.OpenIdSubProtocolMapper{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_sub_protocol_mapper.sub_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdSubProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdSubProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdSubProtocolMapperFetch(resourceName, mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenIdSubProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
					if err != nil {
						t.Error(err)
					}
				},
				Config: testKeycloakOpenIdSubProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdSubProtocolMapperExists(resourceName),
			},
		},
	})
}

func testAccKeycloakOpenIdSubProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_openid_sub_protocol_mapper" {
				continue
			}

			mapper, _ := getSubMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("openid sub protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakOpenIdSubProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getSubMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}
		return nil
	}
}

func testKeycloakOpenIdSubProtocolMapperFetch(resourceName string, mapper *keycloak.OpenIdSubProtocolMapper) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedMapper, err := getSubMapperUsingState(state, resourceName)
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

func getSubMapperUsingState(state *terraform.State, resourceName string) (*keycloak.OpenIdSubProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	return keycloakClient.GetOpenIdSubProtocolMapper(testCtx, realm, clientId, clientScopeId, id)
}

func testKeycloakOpenIdSubProtocolMapper_basic_client(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
    realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
    realm_id    = data.keycloak_realm.realm.id
    client_id   = "%s"

    access_type = "BEARER-ONLY"
}

resource "keycloak_openid_sub_protocol_mapper" "sub_mapper_client" {
    name       = "%s"
    realm_id   = data.keycloak_realm.realm.id
    client_id  = "${keycloak_openid_client.openid_client.id}"
}`, testAccRealm.Realm, clientId, mapperName)
}

func testKeycloakOpenIdSubProtocolMapper_basic_clientScope(clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
    realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
    name     = "%s"
    realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_sub_protocol_mapper" "sub_mapper_client_scope" {
    name            = "%s"
    realm_id        = data.keycloak_realm.realm.id
    client_scope_id = "${keycloak_openid_client_scope.client_scope.id}"
}`, testAccRealm.Realm, clientScopeId, mapperName)
}

func testKeycloakOpenIdSubProtocolMapper_import(clientId, clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
    realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
    realm_id    = data.keycloak_realm.realm.id
    client_id   = "%s"

    access_type = "BEARER-ONLY"
}

resource "keycloak_openid_sub_protocol_mapper" "sub_mapper_client" {
    name       = "%s"
    realm_id   = data.keycloak_realm.realm.id
    client_id  = "${keycloak_openid_client.openid_client.id}"
}

resource "keycloak_openid_client_scope" "client_scope" {
    name     = "%s"
    realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_sub_protocol_mapper" "sub_mapper_client_scope" {
    name            = "%s"
    realm_id        = data.keycloak_realm.realm.id
    client_scope_id = "${keycloak_openid_client_scope.client_scope.id}"
}`, testAccRealm.Realm, clientId, mapperName, clientScopeId, mapperName)
}

func testKeycloakOpenIdSubProtocolMapper_fromInterface(mapper *keycloak.OpenIdSubProtocolMapper) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
    realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
    realm_id    = data.keycloak_realm.realm.id
    client_id   = "%s"

    access_type = "BEARER-ONLY"
}

resource "keycloak_openid_sub_protocol_mapper" "sub_mapper" {
    name                      = "%s"
    realm_id                  = data.keycloak_realm.realm.id
    client_id                 = "${keycloak_openid_client.openid_client.id}"

    add_to_access_token        = %t
    add_to_token_introspection = %t
}`, testAccRealm.Realm, mapper.ClientId, mapper.Name, mapper.AddToAccessToken, mapper.AddToTokenIntrospection)
}
