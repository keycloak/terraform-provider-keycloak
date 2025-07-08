package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeycloakDataSourceGroup_basic(t *testing.T) {
	t.Parallel()

	group := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakGroup_basic(group),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExists("keycloak_group.group"),
					// realm role
					resource.TestCheckResourceAttrPair("keycloak_group.group", "id", "data.keycloak_group.group", "id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "realm_id", "data.keycloak_group.group", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "name", "data.keycloak_group.group", "name"),
					testAccCheckDataKeycloakGroup("data.keycloak_group.group"),
				),
			},
		},
	})
}

func TestAccKeycloakDataSourceGroup_nested(t *testing.T) {
	t.Parallel()

	group := acctest.RandomWithPrefix("tf-acc")
	groupNested := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakGroup_nested(group, groupNested),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExists("keycloak_group.group"),
					testAccCheckKeycloakGroupExists("keycloak_group.group_nested"),
					// realm role
					resource.TestCheckResourceAttrPair("keycloak_group.group", "id", "data.keycloak_group.group", "id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "realm_id", "data.keycloak_group.group", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "name", "data.keycloak_group.group", "name"),
					resource.TestCheckResourceAttrPair("keycloak_group.group_nested", "id", "data.keycloak_group.group_nested", "id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group_nested", "realm_id", "data.keycloak_group.group_nested", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group_nested", "name", "data.keycloak_group.group_nested", "name"),
					testAccCheckDataKeycloakGroup("data.keycloak_group.group"),
					testAccCheckDataKeycloakGroup("data.keycloak_group.group_nested"),
				),
			},
		},
	})
}

func testAccCheckDataKeycloakGroup(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmId := rs.Primary.Attributes["realm_id"]
		name := rs.Primary.Attributes["name"]

		group, err := keycloakClient.GetGroup(testCtx, realmId, id)
		if err != nil {
			return err
		}

		if group.Name != name {
			return fmt.Errorf("expected group with ID %s to have name %s, but got %s", id, name, group.Name)
		}

		return nil
	}
}

func testDataSourceKeycloakGroup_basic(group string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

# we create another group with a similar name to make the data lookup more realistic
resource "keycloak_group" "similar_group" {
	name     = "%s_with_similar_name"
	realm_id = data.keycloak_realm.realm.id
}

data "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name     = keycloak_group.group.name

	depends_on = [
		keycloak_group.group,
		keycloak_group.similar_group,
	]
}
	`, testAccRealm.Realm, group, group)
}

func testDataSourceKeycloakGroup_nested(group, groupNested string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_group" "group_nested" {
	name     	= "%s"
	parent_id = keycloak_group.group.id
	realm_id 	= data.keycloak_realm.realm.id
}

data "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name     = keycloak_group.group.name

	depends_on = [
		keycloak_group.group
	]
}

data "keycloak_group" "group_nested" {
	realm_id = data.keycloak_realm.realm.id
	name     = keycloak_group.group_nested.name

	depends_on = [
		keycloak_group.group_nested
	]
}
	`, testAccRealm.Realm, group, groupNested)
}
