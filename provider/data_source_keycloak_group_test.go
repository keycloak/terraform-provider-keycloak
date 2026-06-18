package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakDataSourceGroup_basic(t *testing.T) {
	t.Parallel()

	group := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakGroupDestroy(),
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
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakGroupDestroy(),
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

func TestAccKeycloakDataSourceGroup_basicWithOrganization(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_6)
	t.Parallel()

	organizationName := acctest.RandomWithPrefix("tf-acc")
	group := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakGroupDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakGroup_basicWithOrganization(organizationName, group),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExistsWithOrganization("keycloak_group.group"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "id", "data.keycloak_group.group", "id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "realm_id", "data.keycloak_group.group", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "organization_id", "data.keycloak_group.group", "organization_id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "name", "data.keycloak_group.group", "name"),
					testAccCheckDataKeycloakGroup("data.keycloak_group.group"),
				),
			},
		},
	})
}

func TestAccKeycloakDataSourceGroup_path_topLevel(t *testing.T) {
	t.Parallel()

	group := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakGroupDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakGroup_path_topLevel(group),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExists("keycloak_group.group"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "id", "data.keycloak_group.group", "id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "realm_id", "data.keycloak_group.group", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "name", "data.keycloak_group.group", "name"),
					resource.TestCheckResourceAttr("data.keycloak_group.group", "parent_id", ""),
					testAccCheckDataKeycloakGroup("data.keycloak_group.group"),
				),
			},
		},
	})
}

func TestAccKeycloakDataSourceGroup_path_nested(t *testing.T) {
	t.Parallel()

	group := acctest.RandomWithPrefix("tf-acc")
	groupNested := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakGroupDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakGroup_path_nested(group, groupNested),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExists("keycloak_group.group"),
					testAccCheckKeycloakGroupExists("keycloak_group.group_nested"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "id", "data.keycloak_group.group", "id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "realm_id", "data.keycloak_group.group", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group", "name", "data.keycloak_group.group", "name"),
					resource.TestCheckResourceAttrPair("keycloak_group.group_nested", "id", "data.keycloak_group.group_nested_by_path", "id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group_nested", "realm_id", "data.keycloak_group.group_nested_by_path", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_group.group_nested", "name", "data.keycloak_group.group_nested_by_path", "name"),
					resource.TestCheckResourceAttrPair("data.keycloak_group.group_nested_by_path", "parent_id", "keycloak_group.group", "id"),
					testAccCheckDataKeycloakGroup("data.keycloak_group.group"),
					testAccCheckDataKeycloakGroup("data.keycloak_group.group_nested_by_path"),
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
		organizationId := rs.Primary.Attributes["organization_id"]
		name := rs.Primary.Attributes["name"]

		group, err := keycloakClient.GetOrganizationGroup(testCtx, realmId, organizationId, id)
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

func testDataSourceKeycloakGroup_basicWithOrganization(organization, group string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	domain {
		name = "%s.example.com"
	}
}

resource "keycloak_group" "group" {
	name            = "%s"
	realm_id        = data.keycloak_realm.realm.id
	organization_id = keycloak_organization.organization.id
}

resource "keycloak_group" "realm_group_with_same_name" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

data "keycloak_group" "group" {
	realm_id        = data.keycloak_realm.realm.id
	organization_id = keycloak_organization.organization.id
	name            = keycloak_group.group.name

	depends_on = [
		keycloak_group.group,
		keycloak_group.realm_group_with_same_name,
	]
}
	`, testAccRealm.Realm, organization, organization, group, group)
}

func TestAccKeycloakDataSourceGroup_nestedWithSpaces(t *testing.T) {
	t.Parallel()

	group := acctest.RandomWithPrefix("tf acc")
	groupNested := acctest.RandomWithPrefix("tf acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakGroupDestroy(),
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

func testDataSourceKeycloakGroup_path_topLevel(group string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_group" "similar_group" {
	name     = "%s_with_similar_name"
	realm_id = data.keycloak_realm.realm.id
}

data "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name     = "/%s"

	depends_on = [
		keycloak_group.group,
		keycloak_group.similar_group,
	]
}
	`, testAccRealm.Realm, group, group, group)
}

func testDataSourceKeycloakGroup_path_nested(parent, child string) string {
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

data "keycloak_group" "group_nested_by_path" {
	realm_id = data.keycloak_realm.realm.id
	name     = "/%s/%s"

	depends_on = [
		keycloak_group.group_nested
	]
}
	`, testAccRealm.Realm, parent, child, parent, child)
}
