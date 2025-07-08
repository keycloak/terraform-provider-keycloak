package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenIdAudienceResolveProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_audience_resolve_protocol_mapper.audience_resolve_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceResolveProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceResolveProtocolMapper_basic_client(clientId),
				Check:  testKeycloakOpenIdAudienceResolveProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdAudienceResolveProtocolMapper_basicClientScope(t *testing.T) {
	t.Parallel()
	clientScopeId := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_audience_resolve_protocol_mapper.audience_resolve_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceResolveProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceResolveProtocolMapper_basic_clientScope(clientScopeId),
				Check:  testKeycloakOpenIdAudienceResolveProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdAudienceResolveProtocolMapper_import(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")

	clientResourceName := "keycloak_openid_audience_resolve_protocol_mapper.audience_resolve_mapper_client"
	clientScopeResourceName := "keycloak_openid_audience_resolve_protocol_mapper.audience_resolve_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceResolveProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceResolveProtocolMapper_import(clientId, clientScopeId),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakOpenIdAudienceResolveProtocolMapperExists(clientResourceName),
					testKeycloakOpenIdAudienceResolveProtocolMapperExists(clientScopeResourceName),
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

func TestAccKeycloakOpenIdAudienceResolveProtocolMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.OpenIdAudienceResolveProtocolMapper{}

	clientId := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_audience_resolve_protocol_mapper.audience_resolve_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceResolveProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceResolveProtocolMapper_basic_client(clientId),
				Check:  testKeycloakOpenIdAudienceResolveProtocolMapperFetch(resourceName, mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenIdAudienceResolveProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
					if err != nil {
						t.Error(err)
					}
				},
				Config: testKeycloakOpenIdAudienceResolveProtocolMapper_basic_client(clientId),
				Check:  testKeycloakOpenIdAudienceResolveProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdAudienceResolveProtocolMapper_updateClientScopeForceNew(t *testing.T) {
	t.Parallel()
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	newClientScopeId := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_audience_resolve_protocol_mapper.audience_resolve_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceResolveProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceResolveProtocolMapper_basic_clientScope(clientScopeId),
				Check:  testKeycloakOpenIdAudienceResolveProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdAudienceResolveProtocolMapper_basic_clientScope(newClientScopeId),
				Check:  testKeycloakOpenIdAudienceResolveProtocolMapperExists(resourceName),
			},
		},
	})
}

func testAccKeycloakOpenIdAudienceResolveProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_openid_audience_resolve_protocol_mapper" {
				continue
			}

			mapper, _ := getAudienceResolveMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("openid audience protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakOpenIdAudienceResolveProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getAudienceResolveMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testKeycloakOpenIdAudienceResolveProtocolMapperFetch(resourceName string, mapper *keycloak.OpenIdAudienceResolveProtocolMapper) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedMapper, err := getAudienceResolveMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		mapper.Id = fetchedMapper.Id
		mapper.Name = fetchedMapper.Name
		mapper.ClientId = fetchedMapper.ClientId
		mapper.ClientScopeId = fetchedMapper.ClientScopeId
		mapper.RealmId = fetchedMapper.RealmId

		return nil
	}
}

func getAudienceResolveMapperUsingState(state *terraform.State, resourceName string) (*keycloak.OpenIdAudienceResolveProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	return keycloakClient.GetOpenIdAudienceResolveProtocolMapper(testCtx, realm, clientId, clientScopeId, id)
}

func testKeycloakOpenIdAudienceResolveProtocolMapper_basic_client(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_audience_resolve_protocol_mapper" "audience_resolve_mapper_client" {
	name                     = "a-custom-name"
	realm_id                 = data.keycloak_realm.realm.id
	client_id                = "${keycloak_openid_client.openid_client.id}"
}`, testAccRealm.Realm, clientId)
}

func testKeycloakOpenIdAudienceResolveProtocolMapper_basic_clientScope(clientScopeId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_audience_resolve_protocol_mapper" "audience_resolve_mapper_client_scope" {
	realm_id                 = data.keycloak_realm.realm.id
	client_scope_id          = "${keycloak_openid_client_scope.client_scope.id}"
}`, testAccRealm.Realm, clientScopeId)
}

func testKeycloakOpenIdAudienceResolveProtocolMapper_import(clientId, clientScopeId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_audience_resolve_protocol_mapper" "audience_resolve_mapper_client" {
	realm_id                 = data.keycloak_realm.realm.id
	client_id                = "${keycloak_openid_client.openid_client.id}"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_audience_resolve_protocol_mapper" "audience_resolve_mapper_client_scope" {
	realm_id                 = data.keycloak_realm.realm.id
	client_scope_id          = "${keycloak_openid_client_scope.client_scope.id}"

}`, testAccRealm.Realm, clientId, clientScopeId)
}
