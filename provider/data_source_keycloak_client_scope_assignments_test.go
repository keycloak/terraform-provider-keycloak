package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakDataSourceClientScopeAssignments_realm(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")

	dataSourceName := "data.keycloak_client_scope_assignments.assignments"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakClientScopeAssignments_realm(realmName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "realm_id", "keycloak_realm.realm", "id"),
					resource.TestCheckNoResourceAttr(dataSourceName, "client_id"),
					resource.TestCheckResourceAttr(dataSourceName, "default_scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "default_scopes.*", "profile"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "default_scopes.*", "email"),
					resource.TestCheckResourceAttr(dataSourceName, "optional_scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "optional_scopes.*", "address"),
				),
			},
		},
	})
}

func TestAccKeycloakDataSourceClientScopeAssignments_client(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	dataSourceName := "data.keycloak_client_scope_assignments.assignments"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakClientScopeAssignments_client(clientId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id", "keycloak_openid_client.client", "id"),
					resource.TestCheckResourceAttr(dataSourceName, "default_scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "default_scopes.*", "profile"),
					resource.TestCheckResourceAttr(dataSourceName, "optional_scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "optional_scopes.*", "email"),
				),
			},
		},
	})
}

func testDataSourceKeycloakClientScopeAssignments_realm(realmName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_client_scope_assignments" "assignments" {
	realm_id        = keycloak_realm.realm.id
	default_scopes  = ["profile", "email"]
	optional_scopes = ["address"]
}

data "keycloak_client_scope_assignments" "assignments" {
	realm_id = keycloak_realm.realm.id

	depends_on = [keycloak_client_scope_assignments.assignments]
}`, realmName)
}

func testDataSourceKeycloakClientScopeAssignments_client(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s"
	access_type = "CONFIDENTIAL"
}

resource "keycloak_client_scope_assignments" "assignments" {
	realm_id        = data.keycloak_realm.realm.id
	client_id       = keycloak_openid_client.client.id
	default_scopes  = ["profile"]
	optional_scopes = ["email"]
}

data "keycloak_client_scope_assignments" "assignments" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.client.id

	depends_on = [keycloak_client_scope_assignments.assignments]
}`, testAccRealm.Realm, clientId)
}
