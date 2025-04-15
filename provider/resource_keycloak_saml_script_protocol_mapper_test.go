package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"testing"
)

func TestAccKeycloakSamlScriptProtocolMapper_basicClient(t *testing.T) {
	skipIfVersionIsGreaterThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_19)

	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_saml_script_protocol_mapper.script_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlScriptProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlScriptProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakSamlScriptProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakSamlScriptProtocolMapper_basicClientScope(t *testing.T) {
	skipIfVersionIsGreaterThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_19)

	t.Parallel()

	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_saml_script_protocol_mapper.script_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlScriptProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlScriptProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakSamlScriptProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakSamlScriptProtocolMapper_import(t *testing.T) {
	skipIfVersionIsGreaterThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_19)

	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	clientResourceName := "keycloak_saml_script_protocol_mapper.script_mapper_client"
	clientScopeResourceName := "keycloak_saml_script_protocol_mapper.script_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlScriptProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlScriptProtocolMapper_import(clientId, clientScopeId, mapperName),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakSamlScriptProtocolMapperExists(clientResourceName),
					testKeycloakSamlScriptProtocolMapperExists(clientScopeResourceName),
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

func TestAccKeycloakSamlScriptProtocolMapper_update(t *testing.T) {
	skipIfVersionIsGreaterThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_19)

	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	attributeName := acctest.RandomWithPrefix("tf-acc")
	updatedAttributeName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_saml_script_protocol_mapper.script_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlScriptProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlScriptProtocolMapper_claim(clientId, mapperName, attributeName),
				Check:  testKeycloakSamlScriptProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakSamlScriptProtocolMapper_claim(clientId, mapperName, updatedAttributeName),
				Check:  testKeycloakSamlScriptProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakSamlScriptProtocolMapper_createAfterManualDestroy(t *testing.T) {
	skipIfVersionIsGreaterThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_19)

	t.Parallel()

	var mapper = &keycloak.SamlScriptProtocolMapper{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_saml_script_protocol_mapper.script_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlScriptProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlScriptProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakSamlScriptProtocolMapperFetch(resourceName, mapper),
			},
			{
				PreConfig: func() {
					keycloakClient := testAccProvider.Meta().(*keycloak.KeycloakClient)

					err := keycloakClient.DeleteSamlScriptProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
					if err != nil {
						t.Error(err)
					}
				},
				Config: testKeycloakSamlScriptProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakSamlScriptProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakSamlScriptProtocolMapper_updateClientIdForceNew(t *testing.T) {
	skipIfVersionIsGreaterThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_19)

	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	updatedClientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	attributeName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_saml_script_protocol_mapper.script_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlScriptProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlScriptProtocolMapper_claim(clientId, mapperName, attributeName),
				Check:  testKeycloakSamlScriptProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakSamlScriptProtocolMapper_claim(updatedClientId, mapperName, attributeName),
				Check:  testKeycloakSamlScriptProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakSamlScriptProtocolMapper_updateClientScopeForceNew(t *testing.T) {
	skipIfVersionIsGreaterThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_19)

	t.Parallel()

	mapperName := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	newClientScopeId := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_saml_script_protocol_mapper.script_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakSamlScriptProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlScriptProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakSamlScriptProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakSamlScriptProtocolMapper_basic_clientScope(newClientScopeId, mapperName),
				Check:  testKeycloakSamlScriptProtocolMapperExists(resourceName),
			},
		},
	})
}

func testAccKeycloakSamlScriptProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_saml_script_protocol_mapper" {
				continue
			}

			mapper, _ := getSamlScriptMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("saml script protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakSamlScriptProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getSamlScriptMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testKeycloakSamlScriptProtocolMapperFetch(resourceName string, mapper *keycloak.SamlScriptProtocolMapper) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedMapper, err := getSamlScriptMapperUsingState(state, resourceName)
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

func getSamlScriptMapperUsingState(state *terraform.State, resourceName string) (*keycloak.SamlScriptProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	keycloakClient := testAccProvider.Meta().(*keycloak.KeycloakClient)

	return keycloakClient.GetSamlScriptProtocolMapper(testCtx, realm, clientId, clientScopeId, id)
}

func testKeycloakSamlScriptProtocolMapper_basic_client(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
        realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
        realm_id  = data.keycloak_realm.realm.id
        client_id = "%s"
}

resource "keycloak_saml_script_protocol_mapper" "script_mapper_client" {
        name                       = "%s"
        realm_id                   = data.keycloak_realm.realm.id
        client_id                  = keycloak_saml_client.saml_client.id
        script                     = "exports = 'foo';"
        saml_attribute_name        = "bar"
        saml_attribute_name_format = "Unspecified"
}`, testAccRealm.Realm, clientId, mapperName)
}

func testKeycloakSamlScriptProtocolMapper_basic_clientScope(clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
        realm = "%s"
}

resource "keycloak_saml_client_scope" "client_scope" {
        name     = "%s"
        realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_saml_script_protocol_mapper" "script_mapper_client_scope" {
        name                       = "%s"
        realm_id                   = data.keycloak_realm.realm.id
        client_scope_id            = keycloak_saml_client_scope.client_scope.id
        script                     = "exports = 'foo';"
        saml_attribute_name        = "bar"
        saml_attribute_name_format = "Unspecified"
}`, testAccRealm.Realm, clientScopeId, mapperName)
}

func testKeycloakSamlScriptProtocolMapper_import(clientId, clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
        realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
        realm_id    = data.keycloak_realm.realm.id
        client_id   = "%s"
}

resource "keycloak_saml_script_protocol_mapper" "script_mapper_client" {
        name                       = "%s"
        realm_id                   = data.keycloak_realm.realm.id
        client_id                  = keycloak_saml_client.saml_client.id
        script                     = "exports = 'foo';"
        saml_attribute_name        = "bar"
        saml_attribute_name_format = "Unspecified"
}

resource "keycloak_saml_client_scope" "client_scope" {
        name     = "%s"
        realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_saml_script_protocol_mapper" "script_mapper_client_scope" {
        name                       = "%s"
        realm_id                   = data.keycloak_realm.realm.id
        client_scope_id            = keycloak_saml_client_scope.client_scope.id
        script                     = "exports = 'foo';"
        saml_attribute_name        = "bar"
        saml_attribute_name_format = "Unspecified"
}`, testAccRealm.Realm, clientId, mapperName, clientScopeId, mapperName)
}

func testKeycloakSamlScriptProtocolMapper_claim(clientId, mapperName, attributeName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
        realm = "%s"
}

resource "keycloak_saml_client" "saml_client" {
        realm_id  = data.keycloak_realm.realm.id
        client_id = "%s"
}

resource "keycloak_saml_script_protocol_mapper" "script_mapper" {
        name                       = "%s"
        realm_id                   = data.keycloak_realm.realm.id
        client_id                  = keycloak_saml_client.saml_client.id
        script                     = "exports = '%s';"
        saml_attribute_name        = "bar"
        saml_attribute_name_format = "Unspecified"
}`, testAccRealm.Realm, clientId, mapperName, attributeName)
}
