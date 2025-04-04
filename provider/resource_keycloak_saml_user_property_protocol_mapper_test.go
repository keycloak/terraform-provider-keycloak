package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"regexp"
	"testing"
)

// Tests for attaching SAML mappers to SAML client scopes are omitted
// because the keycloak_saml_client_scope resource does not exist yet.

func TestAccKeycloakSamlUserPropertyProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_saml_user_property_protocol_mapper.saml_user_property_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlUserPropertyProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakSamlUserPropertyProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakSamlUserPropertyProtocolMapper_import(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	clientResourceName := "keycloak_saml_user_property_protocol_mapper.saml_user_property_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlUserPropertyProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakSamlUserPropertyProtocolMapperExists(clientResourceName),
			},
			{
				ResourceName:      clientResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getGenericProtocolMapperIdForClient(clientResourceName),
			},
		},
	})
}

func TestAccKeycloakSamlUserPropertyProtocolMapper_update(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	userAttribute := acctest.RandomWithPrefix("tf-acc")
	updatedUserAttribute := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_saml_user_property_protocol_mapper.saml_user_property_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlUserPropertyProtocolMapper_userProperty(clientId, mapperName, userAttribute),
				Check:  testKeycloakSamlUserPropertyProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakSamlUserPropertyProtocolMapper_userProperty(clientId, mapperName, updatedUserAttribute),
				Check:  testKeycloakSamlUserPropertyProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakSamlUserPropertyProtocolMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.SamlUserPropertyProtocolMapper{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_saml_user_property_protocol_mapper.saml_user_property_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlUserPropertyProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakSamlUserPropertyProtocolMapperFetch(resourceName, mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteSamlUserPropertyProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
					if err != nil {
						t.Error(err)
					}
				},
				Config: testKeycloakSamlUserPropertyProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakSamlUserPropertyProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakSamlUserPropertyProtocolMapper_validateClaimValueType(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")
	invalidSamlNameFormat := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakSamlUserPropertyProtocolMapper_samlAttributeNameFormat(clientId, mapperName, invalidSamlNameFormat),
				ExpectError: regexp.MustCompile("expected saml_attribute_name_format to be one of .+ got " + invalidSamlNameFormat),
			},
		},
	})
}

func TestAccKeycloakSamlUserPropertyProtocolMapper_updateClientIdForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	updatedClientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	userAttribute := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_saml_user_property_protocol_mapper.saml_user_property_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlUserPropertyProtocolMapper_userProperty(clientId, mapperName, userAttribute),
				Check:  testKeycloakSamlUserPropertyProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakSamlUserPropertyProtocolMapper_userProperty(updatedClientId, mapperName, userAttribute),
				Check:  testKeycloakSamlUserPropertyProtocolMapperExists(resourceName),
			},
		},
	})
}

func testAccKeycloakSamlUserPropertyProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_saml_user_property_protocol_mapper" {
				continue
			}

			mapper, _ := getSamlUserPropertyMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("saml user property protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakSamlUserPropertyProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getSamlUserPropertyMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testKeycloakSamlUserPropertyProtocolMapperFetch(resourceName string, mapper *keycloak.SamlUserPropertyProtocolMapper) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedMapper, err := getSamlUserPropertyMapperUsingState(state, resourceName)
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

func getSamlUserPropertyMapperUsingState(state *terraform.State, resourceName string) (*keycloak.SamlUserPropertyProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	return keycloakClient.GetSamlUserPropertyProtocolMapper(testCtx, realm, clientId, clientScopeId, id)
}

func testKeycloakSamlUserPropertyProtocolMapper_basic_client(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"
}

resource "keycloak_saml_user_property_protocol_mapper" "saml_user_property_mapper" {
	name                       = "%s"
	realm_id                   = data.keycloak_realm.realm.id
	client_id                  = keycloak_saml_client.saml_client.id

	user_property              = "email"
	saml_attribute_name        = "email"
	saml_attribute_name_format = "Unspecified"
}`, testAccRealm.Realm, clientId, mapperName)
}

func testKeycloakSamlUserPropertyProtocolMapper_userProperty(clientId, mapperName, userProperty string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"
}

resource "keycloak_saml_user_property_protocol_mapper" "saml_user_property_mapper" {
	name                       = "%s"
	realm_id                   = data.keycloak_realm.realm.id
	client_id                  = keycloak_saml_client.saml_client.id

	user_property              = "%s"
	saml_attribute_name        = "test"
	saml_attribute_name_format = "Unspecified"
}`, testAccRealm.Realm, clientId, mapperName, userProperty)
}

func testKeycloakSamlUserPropertyProtocolMapper_samlAttributeNameFormat(clientName, mapperName, samlAttributeNameFormat string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"
}

resource "keycloak_saml_user_property_protocol_mapper" "saml_user_property_mapper" {
	name                       = "%s"
	realm_id                   = data.keycloak_realm.realm.id
	client_id                  = keycloak_saml_client.saml_client.id

	user_property              = "email"
	saml_attribute_name        = "email"
	saml_attribute_name_format = "%s"
}`, testAccRealm.Realm, clientName, mapperName, samlAttributeNameFormat)
}
