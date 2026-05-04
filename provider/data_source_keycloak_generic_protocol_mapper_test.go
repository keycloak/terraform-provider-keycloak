package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKeycloakDataSourceGenericProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccKeycloakGenericProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakGenericProtocolMapper_basicClient(clientId, mapperName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("keycloak_generic_protocol_mapper.client_protocol_mapper", "id", "data.keycloak_generic_protocol_mapper.mapper", "id"),
					testAccCheckDataSourceKeycloakGenericProtocolMapper("data.keycloak_generic_protocol_mapper.mapper"),
				),
			},
		},
	})
}

func TestAccKeycloakDataSourceGenericProtocolMapper_basicClientScope(t *testing.T) {
	t.Parallel()

	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccKeycloakGenericProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakGenericProtocolMapper_basicClientScope(clientScopeId, mapperName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("keycloak_generic_protocol_mapper.client_protocol_mapper", "id", "data.keycloak_generic_protocol_mapper.mapper", "id"),
					testAccCheckDataSourceKeycloakGenericProtocolMapper("data.keycloak_generic_protocol_mapper.mapper"),
				),
			},
		},
	})
}

func TestAccKeycloakDataSourceGenericProtocolMapper_notFound(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccKeycloakGenericProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testDataSourceKeycloakGenericProtocolMapper_notFound(clientId, mapperName),
				ExpectError: regexp.MustCompile(`protocol mapper with name .* not found`),
			},
		},
	})
}

func testAccCheckDataSourceKeycloakGenericProtocolMapper(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmId := rs.Primary.Attributes["realm_id"]
		clientId := rs.Primary.Attributes["client_id"]
		clientScopeId := rs.Primary.Attributes["client_scope_id"]

		mapper, err := keycloakClient.GetGenericProtocolMapper(testCtx, realmId, clientId, clientScopeId, id)
		if err != nil {
			return err
		}

		if mapper.Id != id {
			return fmt.Errorf("expected mapper with ID %s but got %s", id, mapper.Id)
		}

		return nil
	}
}

func testDataSourceKeycloakGenericProtocolMapper_basicClient(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"
}

resource "keycloak_generic_protocol_mapper" "client_protocol_mapper" {
	realm_id        = data.keycloak_realm.realm.id
	client_id       = keycloak_saml_client.saml_client.id
	name            = "%s"
	protocol        = "saml"
	protocol_mapper = "saml-hardcode-attribute-mapper"
	config = {
		"attribute.name"       = "name"
		"attribute.nameformat" = "Basic"
		"attribute.value"      = "value"
		"friendly.name"        = "%s"
	}
}

data "keycloak_generic_protocol_mapper" "mapper" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_saml_client.saml_client.id
	name      = keycloak_generic_protocol_mapper.client_protocol_mapper.name

	depends_on = [keycloak_generic_protocol_mapper.client_protocol_mapper]
}
`, testAccRealm.Realm, clientId, mapperName, mapperName)
}

func testDataSourceKeycloakGenericProtocolMapper_basicClientScope(clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_generic_protocol_mapper" "client_protocol_mapper" {
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = keycloak_openid_client_scope.client_scope.id
	name            = "%s"
	protocol        = "openid-connect"
	protocol_mapper = "oidc-usermodel-property-mapper"
	config = {
		"user.attribute" = "foo"
		"claim.name"     = "bar"
	}
}

data "keycloak_generic_protocol_mapper" "mapper" {
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = keycloak_openid_client_scope.client_scope.id
	name            = keycloak_generic_protocol_mapper.client_protocol_mapper.name

	depends_on = [keycloak_generic_protocol_mapper.client_protocol_mapper]
}
`, testAccRealm.Realm, clientScopeId, mapperName)
}

func testDataSourceKeycloakGenericProtocolMapper_notFound(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"
}

resource "keycloak_generic_protocol_mapper" "client_protocol_mapper" {
	realm_id        = data.keycloak_realm.realm.id
	client_id       = keycloak_saml_client.saml_client.id
	name            = "%s"
	protocol        = "saml"
	protocol_mapper = "saml-hardcode-attribute-mapper"
	config = {
		"attribute.name"       = "name"
		"attribute.nameformat" = "Basic"
		"attribute.value"      = "value"
		"friendly.name"        = "%s"
	}
}

data "keycloak_generic_protocol_mapper" "mapper" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_saml_client.saml_client.id
	name      = "nonexistent-mapper-name"

	depends_on = [keycloak_generic_protocol_mapper.client_protocol_mapper]
}
`, testAccRealm.Realm, clientId, mapperName, mapperName)
}
