package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenIdUserPropertyProtocolMapper_basicClient(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_user_property_protocol_mapper.user_property_mapper_client"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccKeycloakOpenIdUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserPropertyProtocolMapper_basicClientScope(t *testing.T) {
	t.Parallel()
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_user_property_protocol_mapper.user_property_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccKeycloakOpenIdUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserPropertyProtocolMapper_import(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	clientResourceName := "keycloak_openid_user_property_protocol_mapper.user_property_mapper_client"
	clientScopeResourceName := "keycloak_openid_user_property_protocol_mapper.user_property_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccKeycloakOpenIdUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_import(clientId, clientScopeId, mapperName),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakOpenIdUserPropertyProtocolMapperExists(clientResourceName),
					testKeycloakOpenIdUserPropertyProtocolMapperExists(clientScopeResourceName),
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

func TestAccKeycloakOpenIdUserPropertyProtocolMapper_importFlagClient_success(t *testing.T) {
	t.Parallel()

	mapperName := acctest.RandomWithPrefix("tf-acc")
	initialPropertyName := acctest.RandomWithPrefix("tf-acc")
	updatedPropertyName := acctest.RandomWithPrefix("tf-acc")
	updatedClaimName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_property_protocol_mapper.imported_user_property_mapper"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			createExternalOpenIdUserPropertyProtocolMapperForClient(t, "account", mapperName, initialPropertyName, "bar")
		},
		CheckDestroy: testAccKeycloakOpenIdUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_importFlagClient("account", mapperName, updatedPropertyName, updatedClaimName),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "user_property", updatedPropertyName),
					resource.TestCheckResourceAttr(resourceName, "claim_name", updatedClaimName),
				),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserPropertyProtocolMapper_importFlagClientScope_success(t *testing.T) {
	t.Parallel()

	mapperName := acctest.RandomWithPrefix("tf-acc")
	initialPropertyName := acctest.RandomWithPrefix("tf-acc")
	updatedPropertyName := acctest.RandomWithPrefix("tf-acc")
	updatedClaimName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_property_protocol_mapper.imported_user_property_mapper"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
			createExternalOpenIdUserPropertyProtocolMapperForClientScope(t, "profile", mapperName, initialPropertyName, "bar")
		},
		CheckDestroy: testAccKeycloakOpenIdUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_importFlagClientScope("profile", mapperName, updatedPropertyName, updatedClaimName),
				Check: resource.ComposeTestCheckFunc(
					testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "user_property", updatedPropertyName),
					resource.TestCheckResourceAttr(resourceName, "claim_name", updatedClaimName),
				),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserPropertyProtocolMapper_importFlag_notFound(t *testing.T) {
	t.Parallel()

	mapperName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenIdUserPropertyProtocolMapper_importFlagClient("account", mapperName, "foo", "bar"),
				ExpectError: regexp.MustCompile(fmt.Sprintf("protocol mapper with name %q not found for import", mapperName)),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserPropertyProtocolMapper_update(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	propertyName := acctest.RandomWithPrefix("tf-acc")
	updatedPropertyName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_property_protocol_mapper.user_property_mapper"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccKeycloakOpenIdUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_claim(clientId, mapperName, propertyName),
				Check:  testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_claim(clientId, mapperName, updatedPropertyName),
				Check:  testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserPropertyProtocolMapper_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var mapper = &keycloak.OpenIdUserPropertyProtocolMapper{}

	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_openid_user_property_protocol_mapper.user_property_mapper_client"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccKeycloakOpenIdUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdUserPropertyProtocolMapperFetch(resourceName, mapper),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenIdUserPropertyProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
					if err != nil {
						t.Error(err)
					}
				},
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_basic_client(clientId, mapperName),
				Check:  testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserPropertyProtocolMapper_validateClaimValueType(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	invalidClaimValueType := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccKeycloakOpenIdUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenIdUserPropertyProtocolMapper_claimValueType(mapperName, invalidClaimValueType),
				ExpectError: regexp.MustCompile("expected claim_value_type to be one of .+ got " + invalidClaimValueType),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserPropertyProtocolMapper_updateClientIdForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	updatedClientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	propertyName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_property_protocol_mapper.user_property_mapper"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccKeycloakOpenIdUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_claim(clientId, mapperName, propertyName),
				Check:  testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_claim(updatedClientId, mapperName, propertyName),
				Check:  testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserPropertyProtocolMapper_updateClientScopeForceNew(t *testing.T) {
	t.Parallel()
	mapperName := acctest.RandomWithPrefix("tf-acc")
	clientScopeId := acctest.RandomWithPrefix("tf-acc")
	newClientScopeId := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_property_protocol_mapper.user_property_mapper_client_scope"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccKeycloakOpenIdUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_basic_clientScope(clientScopeId, mapperName),
				Check:  testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_basic_clientScope(newClientScopeId, mapperName),
				Check:  testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenIdUserPropertyProtocolMapper_updateRealmIdForceNew(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	mapperName := acctest.RandomWithPrefix("tf-acc")

	propertyName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_user_property_protocol_mapper.user_property_mapper"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccKeycloakOpenIdUserPropertyProtocolMapperDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_claim(clientId, mapperName, propertyName),
				Check:  testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
			},
			{
				Config: testKeycloakOpenIdUserPropertyProtocolMapper_claim(clientId, mapperName, propertyName),
				Check:  testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName),
			},
		},
	})
}

func testAccKeycloakOpenIdUserPropertyProtocolMapperDestroy() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for resourceName, rs := range state.RootModule().Resources {
			if rs.Type != "keycloak_openid_user_property_protocol_mapper" {
				continue
			}

			mapper, _ := getUserPropertyMapperUsingState(state, resourceName)

			if mapper != nil {
				return fmt.Errorf("openid user property protocol mapper with id %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakOpenIdUserPropertyProtocolMapperExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := getUserPropertyMapperUsingState(state, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testKeycloakOpenIdUserPropertyProtocolMapperFetch(resourceName string, mapper *keycloak.OpenIdUserPropertyProtocolMapper) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedMapper, err := getUserPropertyMapperUsingState(state, resourceName)
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

func getUserPropertyMapperUsingState(state *terraform.State, resourceName string) (*keycloak.OpenIdUserPropertyProtocolMapper, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	clientId := rs.Primary.Attributes["client_id"]
	clientScopeId := rs.Primary.Attributes["client_scope_id"]

	return keycloakClient.GetOpenIdUserPropertyProtocolMapper(testCtx, realm, clientId, clientScopeId, id)
}

func createExternalOpenIdUserPropertyProtocolMapperForClient(t *testing.T, clientName, mapperName, userProperty, claimName string) *keycloak.OpenIdUserPropertyProtocolMapper {
	t.Helper()

	client, err := keycloakClient.GetOpenidClientByClientId(testCtx, testAccRealm.Realm, clientName)
	if err != nil {
		t.Fatal(err)
	}

	mapper := &keycloak.OpenIdUserPropertyProtocolMapper{
		Name:             mapperName,
		RealmId:          testAccRealm.Realm,
		ClientId:         client.Id,
		AddToIdToken:     true,
		AddToAccessToken: true,
		AddToUserInfo:    true,
		UserProperty:     userProperty,
		ClaimName:        claimName,
		ClaimValueType:   "String",
	}

	if err := keycloakClient.NewOpenIdUserPropertyProtocolMapper(testCtx, mapper); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = keycloakClient.DeleteOpenIdUserPropertyProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
	})

	return mapper
}

func createExternalOpenIdUserPropertyProtocolMapperForClientScope(t *testing.T, clientScopeName, mapperName, userProperty, claimName string) *keycloak.OpenIdUserPropertyProtocolMapper {
	t.Helper()

	clientScopes, err := keycloakClient.ListOpenidClientScopesWithFilter(testCtx, testAccRealm.Realm, keycloak.IncludeOpenidClientScopesMatchingNames([]string{clientScopeName}))
	if err != nil {
		t.Fatal(err)
	}

	if len(clientScopes) != 1 {
		t.Fatalf("expected client scope name %q to match 1 scope, matched %d", clientScopeName, len(clientScopes))
	}

	mapper := &keycloak.OpenIdUserPropertyProtocolMapper{
		Name:             mapperName,
		RealmId:          testAccRealm.Realm,
		ClientScopeId:    clientScopes[0].Id,
		AddToIdToken:     true,
		AddToAccessToken: true,
		AddToUserInfo:    true,
		UserProperty:     userProperty,
		ClaimName:        claimName,
		ClaimValueType:   "String",
	}

	if err := keycloakClient.NewOpenIdUserPropertyProtocolMapper(testCtx, mapper); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = keycloakClient.DeleteOpenIdUserPropertyProtocolMapper(testCtx, mapper.RealmId, mapper.ClientId, mapper.ClientScopeId, mapper.Id)
	})

	return mapper
}

func testKeycloakOpenIdUserPropertyProtocolMapper_basic_client(clientId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_user_property_protocol_mapper" "user_property_mapper_client" {
	name          = "%s"
	realm_id      = data.keycloak_realm.realm.id
	client_id     = "${keycloak_openid_client.openid_client.id}"
	user_property = "foo"
	claim_name    = "bar"
}`, testAccRealm.Realm, clientId, mapperName)
}

func testKeycloakOpenIdUserPropertyProtocolMapper_basic_clientScope(clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_user_property_protocol_mapper" "user_property_mapper_client_scope" {
	name            = "%s"
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = "${keycloak_openid_client_scope.client_scope.id}"
	user_property   = "foo"
	claim_name      = "bar"
}`, testAccRealm.Realm, clientScopeId, mapperName)
}

func testKeycloakOpenIdUserPropertyProtocolMapper_import(clientId, clientScopeId, mapperName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_user_property_protocol_mapper" "user_property_mapper_client" {
	name          = "%s"
	realm_id      = data.keycloak_realm.realm.id
	client_id     = "${keycloak_openid_client.openid_client.id}"
	user_property = "foo"
	claim_name    = "bar"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_openid_user_property_protocol_mapper" "user_property_mapper_client_scope" {
	name            = "%s"
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = "${keycloak_openid_client_scope.client_scope.id}"
	user_property   = "foo"
	claim_name      = "bar"
}`, testAccRealm.Realm, clientId, mapperName, clientScopeId, mapperName)
}

func testKeycloakOpenIdUserPropertyProtocolMapper_claim(clientId, mapperName, propertyName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"

	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_user_property_protocol_mapper" "user_property_mapper" {
	name          = "%s"
	realm_id      = data.keycloak_realm.realm.id
	client_id     = "${keycloak_openid_client.openid_client.id}"
	user_property = "%s"
	claim_name    = "bar"
}`, testAccRealm.Realm, clientId, mapperName, propertyName)
}

func testKeycloakOpenIdUserPropertyProtocolMapper_claimValueType(mapperName, claimValueType string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_user_property_protocol_mapper" "user_property_mapper_validation" {
	name             = "%s"
	realm_id         = data.keycloak_realm.realm.id
	user_property    = "foo"
	claim_name       = "bar"
	claim_value_type = "%s"
}`, testAccRealm.Realm, mapperName, claimValueType)
}

func testKeycloakOpenIdUserPropertyProtocolMapper_importFlagClient(clientName, mapperName, propertyName, claimName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "imported_client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	access_type = "PUBLIC"
	import      = true
}

resource "keycloak_openid_user_property_protocol_mapper" "imported_user_property_mapper" {
	name          = "%s"
	realm_id      = data.keycloak_realm.realm.id
	client_id     = keycloak_openid_client.imported_client.id
	import        = true
	user_property = "%s"
	claim_name    = "%s"
}`, testAccRealm.Realm, clientName, mapperName, propertyName, claimName)
}

func testKeycloakOpenIdUserPropertyProtocolMapper_importFlagClientScope(clientScopeName, mapperName, propertyName, claimName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_openid_client_scope" "imported_client_scope" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_openid_user_property_protocol_mapper" "imported_user_property_mapper" {
	name            = "%s"
	realm_id        = data.keycloak_realm.realm.id
	client_scope_id = data.keycloak_openid_client_scope.imported_client_scope.id
	import          = true
	user_property   = "%s"
	claim_name      = "%s"
}`, testAccRealm.Realm, clientScopeName, mapperName, propertyName, claimName)
}
