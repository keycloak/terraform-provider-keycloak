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

func TestAccKeycloakOpenIdUserSessionNoteProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_user_session_note_protocol_mapper.user_session_note_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserSessionNoteProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserSessionNoteProtocolMapper_basicClientScope(t *testing.T) {
	t.Parallel()
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_user_session_note_protocol_mapper.user_session_note_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserSessionNoteProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserSessionNoteProtocolMapper_import(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	clientResourceName := "keycloak_openid_user_session_note_protocol_mapper.user_session_note_mapper_client"
	clientScopeResourceName := "keycloak_openid_user_session_note_protocol_mapper.user_session_note_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserSessionNoteProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_import(clientId, clientScopeId, mapperName),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakOpenIdUserSessionNoteProtocolMapperExists(clientResourceName),
					testKeycloakOpenIdUserSessionNoteProtocolMapperExists(clientScopeResourceName),
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

func TestAccKeycloakOpenIdUserSessionNoteProtocolMapper_updateClaim(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	claimName := acctest.RandomWithPrefix("tf-acc")
	updatedClaimName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_user_session_note_protocol_mapper.user_session_note_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserSessionNoteProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_claim(clientId, mapperName, claimName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_claim(clientId, mapperName, updatedClaimName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserSessionNoteProtocolMapper_updateNote(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	noteName := acctest.RandomWithPrefix("tf-acc")
	updatedNoteName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_user_session_note_protocol_mapper.user_session_note_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserSessionNoteProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_note(clientId, mapperName, noteName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_note(clientId, mapperName, updatedNoteName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserSessionNoteProtocolMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.OpenIdUserSessionNoteProtocolMapper{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_user_session_note_protocol_mapper.user_session_note_mapper_client"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserSessionNoteProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperFetch(resourceName, mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenIdUserSessionNoteProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
					if err != nil {
						t.Error(err)
					}
				},
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserSessionNoteProtocolMapper_validateClaimValueType(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	invalidClaimValueType := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserSessionNoteProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenIdUserSessionNoteProtocolMapper_validateClaimValueType(mapperName, invalidClaimValueType),
				ExpectError: regexp.MustCompile("expected claim_value_type to be one of .+ got " + invalidClaimValueType),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserSessionNoteProtocolMapper_updateClientIdForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	updatedClientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	claimName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_session_note_protocol_mapper.user_session_note_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserSessionNoteProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_claim(clientId, mapperName, claimName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_claim(updatedClientId, mapperName, claimName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserSessionNoteProtocolMapper_updateClientScopeForceNew(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	newClientScopeId := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_session_note_protocol_mapper.user_session_note_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserSessionNoteProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_basic_clientScope(newClientScopeId, mapperName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserSessionNoteProtocolMapper_updateRealmIdForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	claimName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_session_note_protocol_mapper.user_session_note_mapper"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccKeycloakOpenIdUserSessionNoteProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_claim(clientId, mapperName, claimName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserSessionNoteProtocolMapper_claim(clientId, mapperName, claimName),
				Check:  testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName),
			},
		},
	})
}

func testAccKeycloakOpenIdUserSessionNoteProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_openid_user_session_note_protocol_mapper" {
				continue
			}

			mapper, _ := getUserSessionNoteMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("openid user attribute protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakOpenIdUserSessionNoteProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getUserSessionNoteMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testKeycloakOpenIdUserSessionNoteProtocolMapperFetch(resourceName string, mapper *keycloak.OpenIdUserSessionNoteProtocolMapper) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedMapper, err := getUserSessionNoteMapperUsingState(state, resourceName)
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

func getUserSessionNoteMapperUsingState(state *terraform.State, resourceName string) (*keycloak.OpenIdUserSessionNoteProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	return keycloakClient.GetOpenIdUserSessionNoteProtocolMapper(testCtx, realm, clientId, clientScopeId, id)
}

func testKeycloakOpenIdUserSessionNoteProtocolMapper_basic_client(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}
resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"
	access_type = "BEARER-ONLY"
}
resource "keycloak_openid_user_session_note_protocol_mapper" "user_session_note_mapper_client" {
	name               = "%s"
	realm_id           = data.keycloak_realm.realm.id
	client_id          = "${keycloak_openid_client.openid_client.id}"
	claim_name         = "foo"
	claim_value_type   = "String"
	session_note       = "bar"
}`, testAccRealm.Realm, clientId, mapperName)
}

func testKeycloakOpenIdUserSessionNoteProtocolMapper_basic_clientScope(clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}
resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}
resource "keycloak_openid_user_session_note_protocol_mapper" "user_session_note_mapper_client_scope" {
	name               = "%s"
	realm_id           = data.keycloak_realm.realm.id
	client_scope_id    = "${keycloak_openid_client_scope.client_scope.id}"
	claim_name         = "foo"
	claim_value_type   = "String"
	session_note       = "bar"
}`, testAccRealm.Realm, clientScopeId, mapperName)
}

func testKeycloakOpenIdUserSessionNoteProtocolMapper_claim(clientId, mapperName, claimName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}
resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"
	access_type = "BEARER-ONLY"
}
resource "keycloak_openid_user_session_note_protocol_mapper" "user_session_note_mapper" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	client_id        = "${keycloak_openid_client.openid_client.id}"
	claim_name       = "%s"
	claim_value_type = "String"
}`, testAccRealm.Realm, clientId, mapperName, claimName)
}

func testKeycloakOpenIdUserSessionNoteProtocolMapper_note(clientId, mapperName, noteName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}
resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "%s"
	access_type = "BEARER-ONLY"
}
resource "keycloak_openid_user_session_note_protocol_mapper" "user_session_note_mapper" {
	name               = "%s"
	realm_id           = data.keycloak_realm.realm.id
	client_id          = "${keycloak_openid_client.openid_client.id}"
	claim_name         = "foo"
	claim_value_type   = "String"
	session_note       = "%s"
}`, testAccRealm.Realm, clientId, mapperName, noteName)
}

func testKeycloakOpenIdUserSessionNoteProtocolMapper_import(clientId, clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}
resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	access_type = "BEARER-ONLY"
}
resource "keycloak_openid_user_session_note_protocol_mapper" "user_session_note_mapper_client" {
	name               = "%s"
	realm_id           = data.keycloak_realm.realm.id
	client_id          = "${keycloak_openid_client.openid_client.id}"
	claim_name         = "foo"
	claim_value_type   = "String"
	session_note       = "bar"
}
resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}
resource "keycloak_openid_user_session_note_protocol_mapper" "user_session_note_mapper_client_scope" {
	name               = "%s"
	realm_id           = data.keycloak_realm.realm.id
	client_scope_id    = "${keycloak_openid_client_scope.client_scope.id}"
	claim_name         = "foo"
	claim_value_type   = "String"
	session_note       = "bar"
}`, testAccRealm.Realm, clientId, mapperName, clientScopeId, mapperName)
}

func testKeycloakOpenIdUserSessionNoteProtocolMapper_validateClaimValueType(mapperName, claimValueType string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}
resource "keycloak_openid_client" "openid_client" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "openid-client"
	access_type = "BEARER-ONLY"
}
resource "keycloak_openid_user_session_note_protocol_mapper" "user_session_note_mapper_validation" {
	name               = "%s"
	realm_id           = data.keycloak_realm.realm.id
	client_id          = "${keycloak_openid_client.openid_client.id}"
	claim_name         = "foo"
	claim_value_type   = "%s"
	session_note       = "bar"
}`, testAccRealm.Realm, mapperName, claimValueType)
}
