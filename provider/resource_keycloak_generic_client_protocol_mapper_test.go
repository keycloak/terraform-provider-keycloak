package provider

import (
	"fmt"
	"testing"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeycloakGenericClientProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_generic_client_protocol_mapper.client_protocol_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakGenericClientProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericClientProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakGenericClientProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakGenericClientProtocolMapper_basicClientScope(t *testing.T) {
	t.Parallel()

	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_generic_client_protocol_mapper.client_protocol_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakGenericClientProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericClientProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakGenericClientProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakGenericClientProtocolMapper_import(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_generic_client_protocol_mapper.client_protocol_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakGenericClientProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericClientProtocolMapper_import(clientId, mapperName),
				Check:  testKeycloakGenericClientProtocolMapperExists(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getGenericProtocolMapperIdForClient(resourceName),
			},
		},
	})
}

func TestAccKeycloakGenericClientProtocolMapper_update(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_generic_client_protocol_mapper.client_protocol_mapper"

	oldAttributeName := acctest.RandomWithPrefix("tf-acc")
	oldAttributeValue := acctest.RandomWithPrefix("tf-acc")
	newAttributeName := acctest.RandomWithPrefix("tf-acc")
	newAttributeValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakGenericClientProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericClientProtocolMapper_update(clientId, mapperName, oldAttributeName, oldAttributeValue),
				Check:  testKeycloakGenericClientProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakGenericClientProtocolMapper_update(clientId, mapperName, newAttributeName, newAttributeValue),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakGenericClientProtocolMapperExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "config.attribute.name", newAttributeName),
					resource.TestCheckResourceAttr(resourceName, "config.attribute.value", newAttributeValue)),
			},
		},
	})
}

func testAccKeycloakGenericClientProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_generic_client_protocol_mapper" {
				continue
			}

			mapper, _ := getGenericClientProtocolMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("generic client protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func getGenericClientProtocolMapperUsingState(state *terraform.State, resourceName string) (*keycloak.GenericProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	mapperId := rs.Primary.ID
	realmId := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	return keycloakClient.GetGenericProtocolMapper(testCtx, realmId, clientId, clientScopeId, mapperId)
}

func testKeycloakGenericClientProtocolMapper_basic_client(clientId string, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"
}

resource "keycloak_generic_client_protocol_mapper" "client_protocol_mapper" {
	client_id       = keycloak_saml_client.saml_client.id
	name            = "%s"
	protocol        = "saml"
	protocol_mapper = "saml-hardcode-attribute-mapper"
	realm_id        = data.keycloak_realm.realm.id
	config = {
		"attribute.name"       = "name"
		"attribute.nameformat" = "Basic"
		"attribute.value"      = "value"
		"friendly.name"        = "%s"
	}
}`, testAccRealm.Realm, clientId, mapperName, mapperName)
}

func testKeycloakGenericClientProtocolMapper_basic_clientScope(clientScopeId string, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_generic_client_protocol_mapper" "client_protocol_mapper" {
	name            = "%s"
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = keycloak_openid_client_scope.client_scope.id
	protocol        = "openid-connect"
	protocol_mapper = "oidc-usermodel-property-mapper"
	config = {
		"user.attribute" = "foo"
		"claim.name"     = "bar"
	}
}`, testAccRealm.Realm, clientScopeId, mapperName)
}

func testKeycloakGenericClientProtocolMapper_import(clientId string, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
  realm_id  = data.keycloak_realm.realm.id
  client_id = "%s"
}

resource "keycloak_generic_client_protocol_mapper" "client_protocol_mapper" {
  client_id       = keycloak_saml_client.saml_client.id
  name            = "%s"
  protocol        = "saml"
  protocol_mapper = "saml-hardcode-attribute-mapper"
  realm_id        = data.keycloak_realm.realm.id
  config = {
    "attribute.name"       = "name"
    "attribute.nameformat" = "Basic"
    "attribute.value"      = "value"
    "friendly.name"        = "%s"
  }
}`, testAccRealm.Realm, clientId, mapperName, mapperName)
}

func testKeycloakGenericClientProtocolMapper_update(clientId string, mapperName string, attributeName string, attributeValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
  realm_id  = data.keycloak_realm.realm.id
  client_id = "%s"
}

resource "keycloak_generic_client_protocol_mapper" "client_protocol_mapper" {
  client_id       = keycloak_saml_client.saml_client.id
  name            = "%s"
  protocol        = "saml"
  protocol_mapper = "saml-hardcode-attribute-mapper"
  realm_id        = data.keycloak_realm.realm.id
  config = {
    "attribute.name"       = "%s"
    "attribute.nameformat" = "Basic"
    "attribute.value"      = "%s"
    "friendly.name"        = "%s"
  }
}`, testAccRealm.Realm, clientId, mapperName, attributeName, attributeValue, mapperName)
}

func testKeycloakGenericClientProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getGenericClientProtocolMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}
