package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakIdentityProviderPermissions_basic(t *testing.T) {
	skipIfFGAPv2Enabled(testCtx, t, keycloakClient)

	providerAlias := acctest.RandomWithPrefix("tf-acc")
	providerClientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakIdentityProviderPermissions_basic(providerAlias, providerClientId),
				Check:  testAccCheckKeycloakIdentityProviderPermissionsExists("keycloak_identity_provider_permissions.test"),
			},
			{
				ResourceName:      "keycloak_identity_provider_permissions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKeycloakIdentityProviderPermissions_withScopes(t *testing.T) {
	skipIfFGAPv2Enabled(testCtx, t, keycloakClient)

	providerAlias := acctest.RandomWithPrefix("tf-acc")
	providerClientId := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakIdentityProviderPermissions_withScopes(providerAlias, providerClientId, groupName),
				Check:  testAccCheckKeycloakIdentityProviderPermissionsExists("keycloak_identity_provider_permissions.test"),
			},
		},
	})
}

func testAccCheckKeycloakIdentityProviderPermissionsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		permissions, err := getIdentityProviderPermissionsFromState(s, resourceName)
		if err != nil {
			return err
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		authorizationResourceServerId := rs.Primary.Attributes["authorization_resource_server_id"]

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

		if !permissions.Enabled {
			return fmt.Errorf("expected identity provider permissions to be enabled")
		}

		return nil
	}
}

func getIdentityProviderPermissionsFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProviderPermissions, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realmId := rs.Primary.Attributes["realm_id"]
	providerAlias := rs.Primary.Attributes["provider_alias"]

	permissions, err := keycloakClient.GetIdentityProviderPermissions(testCtx, realmId, providerAlias)
	if err != nil {
		return nil, fmt.Errorf("error getting identity provider permissions for realm %s, alias %s: %s", realmId, providerAlias, err)
	}

	return permissions, nil
}

func testKeycloakIdentityProviderPermissions_basic(providerAlias, providerClientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_identity_provider" "idp" {
	realm             = data.keycloak_realm.realm.id
	alias             = "%s"
	authorization_url = "http://localhost:8080/auth/realms/something/protocol/openid-connect/auth"
	token_url         = "http://localhost:8080/auth/realms/something/protocol/openid-connect/token"
	client_id         = "%s"
	client_secret     = "secret"
}

resource "keycloak_identity_provider_permissions" "test" {
	realm_id       = data.keycloak_realm.realm.id
	provider_alias = keycloak_oidc_identity_provider.idp.alias
}
`, testAccRealm.Realm, providerAlias, providerClientId)
}

func testKeycloakIdentityProviderPermissions_withScopes(providerAlias, providerClientId, groupName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_openid_client" "realm_management" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "realm-management"
}

resource "keycloak_oidc_identity_provider" "idp" {
	realm             = data.keycloak_realm.realm.id
	alias             = "%s"
	authorization_url = "http://localhost:8080/auth/realms/something/protocol/openid-connect/auth"
	token_url         = "http://localhost:8080/auth/realms/something/protocol/openid-connect/token"
	client_id         = "%s"
	client_secret     = "secret"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

# In FGAPv2 realm-management authorization is already enabled by the feature flag,
# so policies can be created without an explicit enablement prerequisite.
resource "keycloak_openid_client_group_policy" "policy" {
	realm_id           = data.keycloak_realm.realm.id
	resource_server_id = data.keycloak_openid_client.realm_management.id
	name               = "idp_group_policy_test"
	groups {
		id              = keycloak_group.group.id
		path            = keycloak_group.group.path
		extend_children = false
	}
	logic             = "POSITIVE"
	decision_strategy = "UNANIMOUS"
}

resource "keycloak_identity_provider_permissions" "test" {
	realm_id       = data.keycloak_realm.realm.id
	provider_alias = keycloak_oidc_identity_provider.idp.alias
	manage_scope {
		policies          = [keycloak_openid_client_group_policy.policy.id]
		description       = "manage IDP"
		decision_strategy = "UNANIMOUS"
	}
}
`, testAccRealm.Realm, providerAlias, providerClientId, groupName)
}
