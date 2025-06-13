package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccKeycloakDataSourceGroupMembers(t *testing.T) {
	t.Parallel()
	username := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakGroupMembers(groupName, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.keycloak_group_members.group_members", "users.0", username),
				),
			},
		},
	})
}

func testDataSourceKeycloakGroupMembers(groupName, username string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_user" "user" {
	username    = "%s"
	realm_id 	= data.keycloak_realm.realm.id
	enabled    	= true

	first_name 	= "Bob"
	last_name  	= "Bobson"
}

resource "keycloak_group_memberships" "group_members" {
	realm_id = data.keycloak_realm.realm.id
	group_id = keycloak_group.group.id

	members = [keycloak_user.user.username]
}
	`, testAccRealm.Realm, groupName, username)
}
