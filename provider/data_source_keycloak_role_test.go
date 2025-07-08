package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccKeycloakDataSourceRole_basic(t *testing.T) {
	t.Parallel()

	client := acctest.RandomWithPrefix("tf-acc")
	realmRole := acctest.RandomWithPrefix("tf-acc")
	clientRole := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakRole_basic(client, realmRole, clientRole),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRoleExists("keycloak_role.realm_role"),
					testAccCheckKeycloakRoleExists("keycloak_role.client_role"),
					// realm role
					resource.TestCheckResourceAttrPair("keycloak_role.realm_role", "id", "data.keycloak_role.realm_role", "id"),
					resource.TestCheckResourceAttrPair("keycloak_role.realm_role", "realm_id", "data.keycloak_role.realm_role", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_role.realm_role", "name", "data.keycloak_role.realm_role", "name"),
					resource.TestCheckResourceAttrPair("keycloak_role.realm_role", "description", "data.keycloak_role.realm_role", "description"),
					testAccCheckDataKeycloakRole("data.keycloak_role.realm_role"),
					// client role
					resource.TestCheckResourceAttrPair("keycloak_role.client_role", "id", "data.keycloak_role.client_role", "id"),
					resource.TestCheckResourceAttrPair("keycloak_role.client_role", "realm_id", "data.keycloak_role.client_role", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_role.client_role", "client_id", "data.keycloak_role.client_role", "client_id"),
					resource.TestCheckResourceAttrPair("keycloak_role.client_role", "name", "data.keycloak_role.client_role", "name"),
					resource.TestCheckResourceAttrPair("keycloak_role.client_role", "description", "data.keycloak_role.client_role", "description"),
					testAccCheckDataKeycloakRole("data.keycloak_role.client_role"),
					// offline_access
					resource.TestCheckResourceAttrPair("data.keycloak_realm.realm", "realm", "data.keycloak_role.realm_offline_access", "realm_id"),
					resource.TestCheckResourceAttr("data.keycloak_role.realm_offline_access", "name", "offline_access"),
					testAccCheckDataKeycloakRole("data.keycloak_role.realm_offline_access"),
				),
			},
		},
	})
}

func testAccCheckDataKeycloakRole(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmId := rs.Primary.Attributes["realm_id"]
		name := rs.Primary.Attributes["name"]

		role, err := keycloakClient.GetRole(testCtx, realmId, id)
		if err != nil {
			return err
		}

		if role.Name != name {
			return fmt.Errorf("expected role with ID %s to have name %s, but got %s", id, name, role.Name)
		}

		return nil
	}
}

func testDataSourceKeycloakRole_basic(client, realmRole, clientRole string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
}

resource "keycloak_role" "realm_role" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_role" "client_role" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.client.id
}

data "keycloak_role" "realm_role" {
	realm_id = data.keycloak_realm.realm.id
	name     = keycloak_role.realm_role.name

	depends_on = [
		keycloak_role.realm_role
	]
}

data "keycloak_role" "client_role" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.client.id
	name      = keycloak_role.client_role.name

	depends_on = [
		keycloak_role.client_role
	]
}

data "keycloak_role" "realm_offline_access" {
	realm_id = data.keycloak_realm.realm.id
	name     = "offline_access"
}
	`, testAccRealm.Realm, client, realmRole, clientRole)
}
