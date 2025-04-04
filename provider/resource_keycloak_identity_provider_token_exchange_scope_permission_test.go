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

func TestAccKeycloakIdpTokenExchangeScopePermission_basic(t *testing.T) {
	t.Parallel()

	providerAlias := acctest.RandomWithPrefix("tf-acc")
	providerClientId := acctest.RandomWithPrefix("tf-acc")
	webappClientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakIdpTokenExchangeScopePermissionDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakIdpTokenExchangeScopePermission_basic(providerAlias, providerClientId, webappClientId),
				Check:  testAccCheckKeycloakIdpTokenExchangeScopePermissionExists("keycloak_identity_provider_token_exchange_scope_permission.my_permission"),
			},
		},
	})
}

func TestAccKeycloakIdpTokenExchangeScopePermission_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var idpPermissions = &keycloak.IdentityProviderPermissions{}

	providerAlias := acctest.RandomWithPrefix("tf-acc")
	providerClientId := acctest.RandomWithPrefix("tf-acc")
	webappClientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakIdpTokenExchangeScopePermissionDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakIdpTokenExchangeScopePermission_basic(providerAlias, providerClientId, webappClientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakIdpTokenExchangeScopePermissionExists("keycloak_identity_provider_token_exchange_scope_permission.my_permission"),
					testAccCheckKeycloakIdpPermissionFetch("keycloak_identity_provider_token_exchange_scope_permission.my_permission", idpPermissions),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DisableIdentityProviderPermissions(testCtx, idpPermissions.RealmId, idpPermissions.ProviderAlias)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakIdpTokenExchangeScopePermission_basic(providerAlias, providerClientId, webappClientId),
				Check:  testAccCheckKeycloakIdpTokenExchangeScopePermissionExists("keycloak_identity_provider_token_exchange_scope_permission.my_permission"),
			},
		},
	})
}

func TestAccKeycloakIdpTokenExchangeScopePermission_import(t *testing.T) {
	t.Parallel()

	providerAlias := acctest.RandomWithPrefix("tf-acc")
	providerClientId := acctest.RandomWithPrefix("tf-acc")
	webappClientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakIdpTokenExchangeScopePermissionDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakIdpTokenExchangeScopePermission_basic(providerAlias, providerClientId, webappClientId),
				Check:  testAccCheckKeycloakIdpTokenExchangeScopePermissionExists("keycloak_identity_provider_token_exchange_scope_permission.my_permission"),
			},
			{
				ResourceName:            "keycloak_identity_provider_token_exchange_scope_permission.my_permission",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_type"},
			},
		},
	})
}

func TestAccKeycloakIdpTokenExchangeScopePermission_updatePolicyMultipleClients(t *testing.T) {
	t.Parallel()

	providerAlias := acctest.RandomWithPrefix("tf-acc")
	providerClientId := acctest.RandomWithPrefix("tf-acc")
	webappClientId := acctest.RandomWithPrefix("tf-acc")
	webappClientId2 := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakIdpTokenExchangeScopePermissionDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakIdpTokenExchangeScopePermission_basic(providerAlias, providerClientId, webappClientId),
				Check:  testAccCheckKeycloakIdpTokenExchangeScopePermissionClientPolicyHasClient("keycloak_identity_provider_token_exchange_scope_permission.my_permission", webappClientId),
			},
			{
				Config: testKeycloakIdpTokenExchangeScopePermission_multipleClients(providerAlias, providerClientId, webappClientId, webappClientId2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakIdpTokenExchangeScopePermissionClientPolicyHasClient("keycloak_identity_provider_token_exchange_scope_permission.my_permission", webappClientId),
					testAccCheckKeycloakIdpTokenExchangeScopePermissionClientPolicyHasClient("keycloak_identity_provider_token_exchange_scope_permission.my_permission", webappClientId2),
				),
			},
			{
				Config: testKeycloakIdpTokenExchangeScopePermission_basic(providerAlias, providerClientId, webappClientId),
				Check:  testAccCheckKeycloakIdpTokenExchangeScopePermissionClientPolicyHasClient("keycloak_identity_provider_token_exchange_scope_permission.my_permission", webappClientId),
			},
		},
	})
}

func TestAccKeycloakIdpTokenExchangeScopePermission_rolePolicy(t *testing.T) {
	t.Parallel()

	providerAlias := acctest.RandomWithPrefix("tf-acc")
	providerClientId := acctest.RandomWithPrefix("tf-acc")
	webappClientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakIdpTokenExchangeScopePermissionDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakIdpTokenExchangeScopePermission_rolePolicy(providerAlias, providerClientId, webappClientId),
				ExpectError: regexp.MustCompile(".*expected policy_type to be one of.*"),
			},
		},
	})
}

func testAccCheckKeycloakIdpTokenExchangeScopePermissionDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_identity_provider_token_exchange_scope_permission" {
				continue
			}

			realmId := rs.Primary.Attributes["realm_id"]
			providerAlias := rs.Primary.Attributes["provider_alias"]
			policyId := rs.Primary.Attributes["policy_id"]
			authorizationResourceServerId := rs.Primary.Attributes["authorization_resource_server_id"]
			authorizationIdpResourceId := rs.Primary.Attributes["authorization_idp_resource_id"]
			authorizationTokenExchangeScopePermissionId := rs.Primary.Attributes["authorization_token_exchange_scope_permission_id"]

			permissions, _ := keycloakClient.GetIdentityProviderPermissions(testCtx, realmId, providerAlias)
			if permissions != nil {
				return fmt.Errorf("idp permissions for realm id %s and provider alias %s still exists", realmId, providerAlias)
			}

			tokenExchangePermission, _ := keycloakClient.GetOpenidClientAuthorizationPermission(testCtx, realmId, authorizationResourceServerId, authorizationTokenExchangeScopePermissionId)
			if tokenExchangePermission != nil {
				return fmt.Errorf("tokenExchangePermission for realm id %s, resource server id %s and permission id %s still exists", realmId, authorizationResourceServerId, authorizationTokenExchangeScopePermissionId)
			}

			idpResource, _ := keycloakClient.GetOpenidClientAuthorizationResource(testCtx, realmId, authorizationResourceServerId, authorizationIdpResourceId)
			if idpResource != nil {
				return fmt.Errorf("idp resource for realm id%s, resource server id %s and resource id %s still exists", realmId, authorizationResourceServerId, authorizationIdpResourceId)
			}

			policy, _ := keycloakClient.GetOpenidClientAuthorizationClientPolicy(testCtx, realmId, authorizationResourceServerId, policyId)
			if policy != nil {
				return fmt.Errorf("client policy for realm id %s, resource server id %s and policy id %s still exists", realmId, authorizationResourceServerId, policyId)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakIdpTokenExchangeScopePermissionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		permissions, err := getIdpPermissionsFromState(s, resourceName)
		if err != nil {
			return err
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		policyId := rs.Primary.Attributes["policy_id"]
		authorizationResourceServerId := rs.Primary.Attributes["authorization_resource_server_id"]
		authorizationIdpResourceId := rs.Primary.Attributes["authorization_idp_resource_id"]
		authorizationTokenExchangeScopePermissionId := rs.Primary.Attributes["authorization_token_exchange_scope_permission_id"]

		var realmManagementId string
		clients, _ := keycloakClient.GetOpenidClients(testCtx, permissions.RealmId, false)
		for _, client := range clients {
			if client.ClientId == "realm-management" {
				realmManagementId = client.Id
				break
			}
		}

		if authorizationResourceServerId != realmManagementId {
			return fmt.Errorf("computed authorizationResourceServerId %s was not equal to %s (the id of the realm-management client)", authorizationResourceServerId, realmManagementId)
		}

		tokenExchangeScopedPermissionId, err := permissions.GetTokenExchangeScopedPermissionId()
		if err != nil {
			return err
		}

		if authorizationTokenExchangeScopePermissionId != tokenExchangeScopedPermissionId {
			return fmt.Errorf("computed authorizationTokenExchangeScopePermissionId %s was not equal to %s scope permission id set on the idp permission", authorizationTokenExchangeScopePermissionId, tokenExchangeScopedPermissionId)
		}

		tokenExchangeScopedPermission, err := keycloakClient.GetOpenidClientAuthorizationPermission(testCtx, permissions.RealmId, realmManagementId, tokenExchangeScopedPermissionId)
		if err != nil {
			return err
		}

		if tokenExchangeScopedPermission == nil {
			return fmt.Errorf("token exchange scope permission represented to idp permission could not be found")
		}

		if len(tokenExchangeScopedPermission.Policies) != 1 {
			return fmt.Errorf("token exchange scope permission has not exact 1 policy, it has %d", len(tokenExchangeScopedPermission.Policies))
		}

		policy, err := keycloakClient.GetOpenidClientAuthorizationClientPolicy(testCtx, permissions.RealmId, realmManagementId, tokenExchangeScopedPermission.Policies[0])
		if err != nil {
			return err
		}

		if policyId != policy.Id {
			return fmt.Errorf("computed policyId %s was not equal to %s policyId found on the token exchange scope based permission", policyId, policy.Id)
		}

		idpResource, err := keycloakClient.GetOpenidClientAuthorizationResource(testCtx, permissions.RealmId, realmManagementId, permissions.Resource)
		if err != nil {
			return err
		}

		if tokenExchangeScopedPermission.Resources[0] != idpResource.Id {
			return fmt.Errorf("fetched permission resources %s do not correspond with the idp resource provided id %s", tokenExchangeScopedPermission.Resources[0], idpResource.Id)
		}

		if authorizationIdpResourceId != idpResource.Id {
			return fmt.Errorf("computed authorizationIdpResourceId %s was not equal to %s sidp resource id found on the token exchange scope based permission", authorizationIdpResourceId, idpResource.Id)
		}

		return nil
	}
}

func testAccCheckKeycloakIdpTokenExchangeScopePermissionClientPolicyHasClient(resourceName, clientId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		realmId := rs.Primary.Attributes["realm_id"]
		authorizationResourceServerId := rs.Primary.Attributes["authorization_resource_server_id"]
		policyId := rs.Primary.Attributes["policy_id"]

		policy, err := keycloakClient.GetOpenidClientAuthorizationClientPolicy(testCtx, realmId, authorizationResourceServerId, policyId)
		if err != nil {
			return err
		}

		client, err := keycloakClient.GetOpenidClientByClientId(testCtx, realmId, clientId)
		if err != nil {
			return err
		}

		clientNotFound := true
		for _, idOfClientString := range policy.Clients {
			if idOfClientString == client.Id {
				clientNotFound = false
				break
			}
		}
		if clientNotFound {
			return fmt.Errorf("client with clientId %s was not linked to policy", clientId)
		}

		return nil
	}
}

func testAccCheckKeycloakIdpPermissionFetch(resourceName string, idpPermissions *keycloak.IdentityProviderPermissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedPermissions, err := getIdpPermissionsFromState(s, resourceName)
		if err != nil {
			return err
		}

		idpPermissions.RealmId = fetchedPermissions.RealmId
		idpPermissions.Enabled = fetchedPermissions.Enabled
		idpPermissions.ProviderAlias = fetchedPermissions.ProviderAlias
		idpPermissions.ScopePermissions = fetchedPermissions.ScopePermissions
		idpPermissions.Resource = fetchedPermissions.Resource

		return nil
	}
}

func getIdpPermissionsFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProviderPermissions, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realmId := rs.Primary.Attributes["realm_id"]
	providerAlias := rs.Primary.Attributes["provider_alias"]

	permissions, err := keycloakClient.GetIdentityProviderPermissions(testCtx, realmId, providerAlias)
	if err != nil {
		return nil, fmt.Errorf("error getting idp permissions with realm id %s and provider alias %s: %s", realmId, providerAlias, err)

	}
	return permissions, nil
}

func testKeycloakIdpTokenExchangeScopePermission_basic(providerAlias, providerClientId, webappClientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_identity_provider" "my_idp" {
	realm              = data.keycloak_realm.realm.id
	alias              = "%s"
	authorization_url  = "http://localhost:8080/auth/realms/something/protocol/openid-connect/auth"
	token_url          = "http://localhost:8080/auth/realms/something/protocol/openid-connect/token"
	client_id          = "%s"
	client_secret      = "secret"
}

resource "keycloak_openid_client" "webapp_client" {
	realm_id              = data.keycloak_realm.realm.id
	name                  = "webapp_client"
	client_id             = "%s"
	client_secret         = "secret"
	access_type           = "CONFIDENTIAL"
	standard_flow_enabled = true
	valid_redirect_uris = [
		"http://localhost:8080/*",
	]
}

resource "keycloak_identity_provider_token_exchange_scope_permission" "my_permission" {
	realm_id       = data.keycloak_realm.realm.id
	provider_alias = keycloak_oidc_identity_provider.my_idp.alias
	policy_type    = "client"
	clients        = [
		keycloak_openid_client.webapp_client.id
	]
}
	`, testAccRealm.Realm, providerAlias, providerClientId, webappClientId)
}

func testKeycloakIdpTokenExchangeScopePermission_multipleClients(providerAlias, providerClientId, webappClientId, webappClientId2 string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_identity_provider" "my_idp" {
	realm              = data.keycloak_realm.realm.id
	alias              = "%s"
	authorization_url  = "http://localhost:8080/auth/realms/something/protocol/openid-connect/auth"
	token_url          = "http://localhost:8080/auth/realms/something/protocol/openid-connect/token"
	client_id          = "%s"
	client_secret      = "secret"
}

resource "keycloak_openid_client" "webapp_client" {
	realm_id              = data.keycloak_realm.realm.id
	name                  = "webapp_client"
	client_id             = "%s"
	client_secret         = "secret"
	access_type           = "CONFIDENTIAL"
	standard_flow_enabled = true
	valid_redirect_uris = [
		"http://localhost:8080/*",
	]
}

resource "keycloak_openid_client" "webapp_client2" {
	realm_id              = data.keycloak_realm.realm.id
	name                  = "webapp_client"
	client_id             = "%s"
	client_secret         = "secret"
	access_type           = "CONFIDENTIAL"
	standard_flow_enabled = true
	valid_redirect_uris = [
		"http://localhost:8080/*",
	]
}

resource "keycloak_identity_provider_token_exchange_scope_permission" "my_permission" {
	realm_id       = data.keycloak_realm.realm.id
	provider_alias = keycloak_oidc_identity_provider.my_idp.alias
	policy_type    = "client"
	clients        = [
		keycloak_openid_client.webapp_client.id,
		keycloak_openid_client.webapp_client2.id,
	]
}
	`, testAccRealm.Realm, providerAlias, providerClientId, webappClientId, webappClientId2)
}

func testKeycloakIdpTokenExchangeScopePermission_rolePolicy(providerAlias, providerClientId, webappClientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_identity_provider" "my_idp" {
	realm              = data.keycloak_realm.realm.id
	alias              = "%s"
	authorization_url  = "http://localhost:8080/auth/realms/something/protocol/openid-connect/auth"
	token_url          = "http://localhost:8080/auth/realms/something/protocol/openid-connect/token"
	client_id          = "%s"
	client_secret      = "secret"
}

resource "keycloak_openid_client" "webapp_client" {
	realm_id              = data.keycloak_realm.realm.id
	name                  = "webapp_client"
	client_id             = "%s"
	client_secret         = "secret"
	access_type           = "CONFIDENTIAL"
	standard_flow_enabled = true
	valid_redirect_uris = [
		"http://localhost:8080/*",
	]
}

resource "keycloak_identity_provider_token_exchange_scope_permission" "my_permission" {
	realm_id       = data.keycloak_realm.realm.id
	provider_alias = keycloak_oidc_identity_provider.my_idp.alias
	policy_type    = "role"
	clients        = [
		keycloak_openid_client.webapp_client.id
	]
}
	`, testAccRealm.Realm, providerAlias, providerClientId, webappClientId)
}
