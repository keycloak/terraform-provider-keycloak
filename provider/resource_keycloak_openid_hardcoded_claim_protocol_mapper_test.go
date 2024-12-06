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

func TestAccKeycloakOpenIdHardcodedClaimProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_hardcoded_claim_protocol_mapper.hardcoded_claim_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedClaimProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdHardcodedClaimProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdHardcodedClaimProtocolMapper_basicClientScope(t *testing.T) {
	t.Parallel()
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_hardcoded_claim_protocol_mapper.hardcoded_claim_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedClaimProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdHardcodedClaimProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdHardcodedClaimProtocolMapper_import(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	clientResourceName := "keycloak_openid_hardcoded_claim_protocol_mapper.hardcoded_claim_mapper_client"
	clientScopeResourceName := "keycloak_openid_hardcoded_claim_protocol_mapper.hardcoded_claim_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdFullNameProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_import(clientId, clientScopeId, mapperName),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakOpenIdHardcodedClaimProtocolMapperExists(clientResourceName),
					testKeycloakOpenIdHardcodedClaimProtocolMapperExists(clientScopeResourceName),
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

func TestAccKeycloakOpenIdHardcodedClaimProtocolMapper_update(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	claimName := acctest.RandomWithPrefix("tf-acc")
	updatedClaimName := acctest.RandomWithPrefix("tf-acc")
	claimValue := acctest.RandomWithPrefix("tf-acc")
	updatedClaimValue := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_hardcoded_claim_protocol_mapper.hardcoded_claim_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedClaimProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_claimNameAndValue(clientId, mapperName, claimName, claimValue),
				Check:  testKeycloakOpenIdHardcodedClaimProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_claimNameAndValue(clientId, mapperName, updatedClaimName, updatedClaimValue),
				Check:  testKeycloakOpenIdHardcodedClaimProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdHardcodedClaimProtocolMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.OpenIdHardcodedClaimProtocolMapper{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_hardcoded_claim_protocol_mapper.hardcoded_claim_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedClaimProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdHardcodedClaimProtocolMapperFetch(resourceName, mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenIdHardcodedClaimProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
					if err != nil {
						t.Error(err)
					}
				},
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdHardcodedClaimProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdHardcodedClaimProtocolMapper_validateClaimValueType(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	invalidClaimValueType := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedClaimProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenIdHardcodedClaimProtocolMapper_validateClaimValueType(mapperName, invalidClaimValueType),
				ExpectError: regexp.MustCompile("expected claim_value_type to be one of .+ got " + invalidClaimValueType),
			},
		},
	})
}

func TestAccKeycloakOpenIdHardcodedClaimProtocolMapper_updateClientIdForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	updatedClientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	claimName := acctest.RandomWithPrefix("tf-acc")
	claimValue := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_hardcoded_claim_protocol_mapper.hardcoded_claim_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedClaimProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_claimNameAndValue(clientId, mapperName, claimName, claimValue),
				Check:  testKeycloakOpenIdHardcodedClaimProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_claimNameAndValue(updatedClientId, mapperName, claimName, claimValue),
				Check:  testKeycloakOpenIdHardcodedClaimProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdHardcodedClaimProtocolMapper_updateClientScopeForceNew(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	newClientScopeId := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_hardcoded_claim_protocol_mapper.hardcoded_claim_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedClaimProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdHardcodedClaimProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_basic_clientScope(newClientScopeId, mapperName),
				Check:  testKeycloakOpenIdHardcodedClaimProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdHardcodedClaimProtocolMapper_updateRealmIdForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	claimName := acctest.RandomWithPrefix("tf-acc")
	claimValue := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_hardcoded_claim_protocol_mapper.hardcoded_claim_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdHardcodedClaimProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_claimNameAndValue(clientId, mapperName, claimName, claimValue),
				Check:  testKeycloakOpenIdHardcodedClaimProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdHardcodedClaimProtocolMapper_claimNameAndValue(clientId, mapperName, claimName, claimValue),
				Check:  testKeycloakOpenIdHardcodedClaimProtocolMapperExists(resourceName),
			},
		},
	})
}

func testAccKeycloakOpenIdHardcodedClaimProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_openid_hardcoded_claim_protocol_mapper" {
				continue
			}

			mapper, _ := getHardcodedClaimMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("openid user attribute protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakOpenIdHardcodedClaimProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getHardcodedClaimMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testKeycloakOpenIdHardcodedClaimProtocolMapperFetch(resourceName string, mapper *keycloak.OpenIdHardcodedClaimProtocolMapper) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedMapper, err := getHardcodedClaimMapperUsingState(state, resourceName)
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

func getHardcodedClaimMapperUsingState(state *terraform.State, resourceName string) (*keycloak.OpenIdHardcodedClaimProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	return keycloakClient.GetOpenIdHardcodedClaimProtocolMapper(testCtx, realm, clientId, clientScopeId, id)
}

func testKeycloakOpenIdHardcodedClaimProtocolMapper_basic_client(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_hardcoded_claim_protocol_mapper" "hardcoded_claim_mapper_client" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_id        = "${keycloak_openid_client.openid_client.id}"

	claim_name       = "foo"
	claim_value      = "bar"
	claim_value_type = "String"
}`, testAccRealm.Realm, clientId, mapperName)
}

func testKeycloakOpenIdHardcodedClaimProtocolMapper_basic_clientScope(clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_hardcoded_claim_protocol_mapper" "hardcoded_claim_mapper_client_scope" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_scope_id  = "${keycloak_openid_client_scope.client_scope.id}"

	claim_name       = "foo"
	claim_value      = "bar"
	claim_value_type = "String"
}`, testAccRealm.Realm, clientScopeId, mapperName)
}

func testKeycloakOpenIdHardcodedClaimProtocolMapper_import(clientId, clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_hardcoded_claim_protocol_mapper" "hardcoded_claim_mapper_client" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_id        = "${keycloak_openid_client.openid_client.id}"

	claim_name       = "foo"
	claim_value      = "bar"
	claim_value_type = "String"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_hardcoded_claim_protocol_mapper" "hardcoded_claim_mapper_client_scope" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_scope_id  = "${keycloak_openid_client_scope.client_scope.id}"

	claim_name       = "foo"
	claim_value      = "bar"
	claim_value_type = "String"
}`, testAccRealm.Realm, clientId, mapperName, clientScopeId, mapperName)
}

func testKeycloakOpenIdHardcodedClaimProtocolMapper_claimNameAndValue(clientId, mapperName, claimName, claimValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_hardcoded_claim_protocol_mapper" "hardcoded_claim_mapper" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "${keycloak_openid_client.openid_client.id}"

	claim_name  = "%s"
	claim_value = "%s"
}`, testAccRealm.Realm, clientId, mapperName, claimName, claimValue)
}

func testKeycloakOpenIdHardcodedClaimProtocolMapper_validateClaimValueType(mapperName, claimValueType string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "openid-client"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_hardcoded_claim_protocol_mapper" "hardcoded_claim_mapper_validation" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_id        = "${keycloak_openid_client.openid_client.id}"

	claim_value      = "foo"
	claim_name       = "bar"
	claim_value_type = "%s"
}`, testAccRealm.Realm, mapperName, claimValueType)
}
