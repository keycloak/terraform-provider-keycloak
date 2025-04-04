package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccKeycloakDataSourceUserRoles(t *testing.T) {
	username := acctest.RandomWithPrefix("tf-acc")
	email := acctest.RandomWithPrefix("tf-acc") + "@fakedomain.com"
	realmRoleName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakUserRoles(username, email, realmRoleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.keycloak_user_realm_roles.user_realm_roles", "role_names.0", realmRoleName),
				),
			},
		},
	})
}

func testDataSourceKeycloakUserRoles(username, email, realmRoleName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_user" "user" {
	username    = "%s"
	realm_id 	= data.keycloak_realm.realm.id
	enabled    	= true

	email      	= "%s"
	first_name 	= "Bob"
	last_name  	= "Bobson"
}

resource "keycloak_role" "realm_role" {
	realm_id    = data.keycloak_realm.realm.id
	name        = "%s"
}

resource "keycloak_user_roles" "user_roles" {
	realm_id 	= data.keycloak_realm.realm.id
	user_id     = keycloak_user.user.id

	role_ids = [
		keycloak_role.realm_role.id,
	]
}

data "keycloak_user_realm_roles" "user_realm_roles" {
	realm_id 	= data.keycloak_realm.realm.id
	user_id     = keycloak_user.user.id

	depends_on = [
		keycloak_user_roles.user_roles
	]
}
	`, testAccRealm.Realm, username, email, realmRoleName)
}
