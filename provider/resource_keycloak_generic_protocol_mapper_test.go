package provider

import (
	"fmt"
	"testing"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeycloakGenericProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_generic_protocol_mapper.client_protocol_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakGenericProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakGenericProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakGenericProtocolMapper_basicClientScope(t *testing.T) {
	t.Parallel()

	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_generic_protocol_mapper.client_protocol_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakGenericProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakGenericProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakGenericProtocolMapper_import(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_generic_protocol_mapper.client_protocol_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakGenericProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericProtocolMapper_import(clientId, mapperName),
				Check:  testKeycloakGenericProtocolMapperExists(resourceName),
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

func TestAccKeycloakGenericProtocolMapper_update(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_generic_protocol_mapper.client_protocol_mapper"

	oldAttributeName := acctest.RandomWithPrefix("tf-acc")
	oldAttributeValue := acctest.RandomWithPrefix("tf-acc")
	newAttributeName := acctest.RandomWithPrefix("tf-acc")
	newAttributeValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakGenericProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGenericProtocolMapper_update(clientId, mapperName, oldAttributeName, oldAttributeValue),
				Check:  testKeycloakGenericProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakGenericProtocolMapper_update(clientId, mapperName, newAttributeName, newAttributeValue),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakGenericProtocolMapperExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "config.attribute.name", newAttributeName),
					resource.TestCheckResourceAttr(resourceName, "config.attribute.value", newAttributeValue)),
			},
		},
	})
}

func testAccKeycloakGenericProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_generic_protocol_mapper" {
				continue
			}

			mapper, _ := getGenericProtocolMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("generic client protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func getGenericProtocolMapperUsingState(state *terraform.State, resourceName string) (*keycloak.GenericProtocolMapper, error) {
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

func testKeycloakGenericProtocolMapper_basic_client(clientId string, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"
}

resource "keycloak_generic_protocol_mapper" "client_protocol_mapper" {
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

func testKeycloakGenericProtocolMapper_basic_clientScope(clientScopeId string, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_generic_protocol_mapper" "client_protocol_mapper" {
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

func testKeycloakGenericProtocolMapper_import(clientId string, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
  realm_id  = data.keycloak_realm.realm.id
  client_id = "%s"
}

resource "keycloak_generic_protocol_mapper" "client_protocol_mapper" {
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

func testKeycloakGenericProtocolMapper_update(clientId string, mapperName string, attributeName string, attributeValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
  realm_id  = data.keycloak_realm.realm.id
  client_id = "%s"
}

resource "keycloak_generic_protocol_mapper" "client_protocol_mapper" {
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

func testKeycloakGenericProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getGenericProtocolMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}
