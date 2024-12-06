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

func TestAccKeycloakOpenIdAudienceProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_audience_protocol_mapper.audience_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdAudienceProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdAudienceProtocolMapper_basicClientScope(t *testing.T) {
	t.Parallel()
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_audience_protocol_mapper.audience_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdAudienceProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdAudienceProtocolMapper_import(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	clientResourceName := "keycloak_openid_audience_protocol_mapper.audience_mapper_client"
	clientScopeResourceName := "keycloak_openid_audience_protocol_mapper.audience_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceProtocolMapper_import(clientId, clientScopeId, mapperName),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakOpenIdAudienceProtocolMapperExists(clientResourceName),
					testKeycloakOpenIdAudienceProtocolMapperExists(clientScopeResourceName),
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

func TestAccKeycloakOpenIdAudienceProtocolMapper_update(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	customAudience := acctest.RandomWithPrefix("tf-acc")
	updatedCustomAudience := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_audience_protocol_mapper.audience_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceProtocolMapper_customAudience(clientId, mapperName, customAudience),
				Check:  testKeycloakOpenIdAudienceProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdAudienceProtocolMapper_customAudience(clientId, mapperName, updatedCustomAudience),
				Check:  testKeycloakOpenIdAudienceProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdAudienceProtocolMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.OpenIdAudienceProtocolMapper{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_audience_protocol_mapper.audience_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdAudienceProtocolMapperFetch(resourceName, mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenIdAudienceProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
					if err != nil {
						t.Error(err)
					}
				},
				Config: testKeycloakOpenIdAudienceProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdAudienceProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdAudienceProtocolMapper_updateClientIdForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	updatedClientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	customAudience := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_audience_protocol_mapper.audience_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceProtocolMapper_customAudience(clientId, mapperName, customAudience),
				Check:  testKeycloakOpenIdAudienceProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdAudienceProtocolMapper_customAudience(updatedClientId, mapperName, customAudience),
				Check:  testKeycloakOpenIdAudienceProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdAudienceProtocolMapper_updateClientScopeForceNew(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	newClientScopeId := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_audience_protocol_mapper.audience_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdAudienceProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdAudienceProtocolMapper_basic_clientScope(newClientScopeId, mapperName),
				Check:  testKeycloakOpenIdAudienceProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdAudienceProtocolMapper_updateRealmIdForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	customAudience := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_audience_protocol_mapper.audience_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdAudienceProtocolMapper_customAudience(clientId, mapperName, customAudience),
				Check:  testKeycloakOpenIdAudienceProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdAudienceProtocolMapper_customAudience(clientId, mapperName, customAudience),
				Check:  testKeycloakOpenIdAudienceProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdAudienceProtocolMapper_validateClientAudienceExists(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdAudienceProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenIdAudienceProtocolMapper_validateClientAudienceExists(clientId, mapperName),
				ExpectError: regexp.MustCompile("validation error: client .+ does not exist"),
			},
		},
	})
}

func testAccKeycloakOpenIdAudienceProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_openid_audience_protocol_mapper" {
				continue
			}

			mapper, _ := getAudienceMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("openid audience protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakOpenIdAudienceProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getAudienceMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testKeycloakOpenIdAudienceProtocolMapperFetch(resourceName string, mapper *keycloak.OpenIdAudienceProtocolMapper) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedMapper, err := getAudienceMapperUsingState(state, resourceName)
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

func getAudienceMapperUsingState(state *terraform.State, resourceName string) (*keycloak.OpenIdAudienceProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	return keycloakClient.GetOpenIdAudienceProtocolMapper(testCtx, realm, clientId, clientScopeId, id)
}

func testKeycloakOpenIdAudienceProtocolMapper_basic_client(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_audience_protocol_mapper" "audience_mapper_client" {
	name                     = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	client_id                = "${keycloak_openid_client.openid_client.id}"

	included_custom_audience = "foo"
}`, testAccRealm.Realm, clientId, mapperName)
}

func testKeycloakOpenIdAudienceProtocolMapper_basic_clientScope(clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_audience_protocol_mapper" "audience_mapper_client_scope" {
	name                     = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	client_scope_id          = "${keycloak_openid_client_scope.client_scope.id}"

	included_custom_audience = "foo"
}`, testAccRealm.Realm, clientScopeId, mapperName)
}

func testKeycloakOpenIdAudienceProtocolMapper_import(clientId, clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_audience_protocol_mapper" "audience_mapper_client" {
	name                     = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	client_id                = "${keycloak_openid_client.openid_client.id}"

	included_custom_audience = "foo"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_audience_protocol_mapper" "audience_mapper_client_scope" {
	name                     = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	client_scope_id          = "${keycloak_openid_client_scope.client_scope.id}"

	included_custom_audience = "foo"
}`, testAccRealm.Realm, clientId, mapperName, clientScopeId, mapperName)
}

func testKeycloakOpenIdAudienceProtocolMapper_customAudience(clientId, mapperName, customAudience string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "CONFIDENTIAL"

	standard_flow_enabled = true

	valid_redirect_uris = ["http://localhost:5555/callback"]
}

resource "keycloak_openid_audience_protocol_mapper" "audience_mapper" {
	name                     = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	client_id                = "${keycloak_openid_client.openid_client.id}"

	included_custom_audience = "%s"
}`, testAccRealm.Realm, clientId, mapperName, customAudience)
}

func testKeycloakOpenIdAudienceProtocolMapper_validateClientConflictsWithClientScope(clientId, clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_audience_protocol_mapper" "audience_mapper" {
	name                     = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	client_id                = "${keycloak_openid_client.openid_client.id}"
	client_scope_id          = "${keycloak_openid_client_scope.client_scope.id}"

	included_custom_audience = "foo"
}`, testAccRealm.Realm, clientId, clientScopeId, mapperName)
}

func testKeycloakOpenIdAudienceProtocolMapper_validateClientAudienceConflictsWithCustomAudience(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_audience_protocol_mapper" "audience_mapper" {
	name                     = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	client_id                = "${keycloak_openid_client.openid_client.id}"

	included_client_audience = "${keycloak_openid_client.openid_client.client_id}"
	included_custom_audience = "foo"
}`, testAccRealm.Realm, clientId, mapperName)
}

func testKeycloakOpenIdAudienceProtocolMapper_validateClientAudienceExists(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "openid-client"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_audience_protocol_mapper" "audience_mapper" {
	name                     = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	client_id                = "${keycloak_openid_client.openid_client.id}"

	included_client_audience = "%s"

	depends_on = [ "keycloak_openid_client.openid_client" ]
}`, testAccRealm.Realm, mapperName, clientId)
}
