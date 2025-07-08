package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakGroup_basic(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")
	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")

	runTestBasicGroup(t, groupName, attributeName, attributeValue)
}

func TestAccKeycloakGroup_basicGroupNameContainsBackSlash(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")
	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")

	runTestBasicGroup(t, groupName, attributeName, attributeValue)
}

func runTestBasicGroup(t *testing.T, groupName, attributeName, attributeValue string) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakGroupDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroup_basic(groupName, attributeName, attributeValue),
				Check:  testAccCheckKeycloakGroupExists("keycloak_group.group"),
			},
			{
				ResourceName:        "keycloak_group.group",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakGroup_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var group = &keycloak.Group{}

	groupName := acctest.RandomWithPrefix("tf-acc")
	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakGroupDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroup_basic(groupName, attributeName, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExists("keycloak_group.group"),
					testAccCheckKeycloakGroupFetch("keycloak_group.group", group),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteGroup(testCtx, group.RealmId, group.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakGroup_basic(groupName, attributeName, attributeValue),
				Check:  testAccCheckKeycloakGroupExists("keycloak_group.group"),
			},
		},
	})
}

func TestAccKeycloakGroup_updateGroupName(t *testing.T) {
	t.Parallel()

	groupNameBefore := acctest.RandomWithPrefix("tf-acc")
	groupNameAfter := acctest.RandomWithPrefix("tf-acc")
	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakGroupDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroup_basic(groupNameBefore, attributeName, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExists("keycloak_group.group"),
					resource.TestCheckResourceAttr("keycloak_group.group", "name", groupNameBefore),
				),
			},
			{
				Config: testKeycloakGroup_basic(groupNameAfter, attributeName, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExists("keycloak_group.group"),
					resource.TestCheckResourceAttr("keycloak_group.group", "name", groupNameAfter),
				),
			},
		},
	})
}

func TestAccKeycloakGroup_updateRealm(t *testing.T) {
	t.Parallel()

	group := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakGroupDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroup_updateRealmBefore(group),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExists("keycloak_group.group"),
					testAccCheckKeycloakGroupBelongsToRealm("keycloak_group.group", testAccRealm.Realm),
				),
			},
			{
				Config: testKeycloakGroup_updateRealmAfter(group),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExists("keycloak_group.group"),
					testAccCheckKeycloakGroupBelongsToRealm("keycloak_group.group", testAccRealmTwo.Realm),
				),
			},
		},
	})
}

func TestAccKeycloakGroup_nested(t *testing.T) {
	t.Parallel()

	parentGroupName := acctest.RandomWithPrefix("tf-acc")
	firstChildGroupName := acctest.RandomWithPrefix("tf-acc")
	secondChildGroupName := acctest.RandomWithPrefix("tf-acc")

	runTestNestedGroup(t, parentGroupName, firstChildGroupName, secondChildGroupName)
}

func TestAccKeycloakGroup_nestedGroupNameContainsBackSlash(t *testing.T) {
	t.Parallel()

	parentGroupName := acctest.RandomWithPrefix("tf-acc")
	firstChildGroupName := acctest.RandomWithPrefix("tf-acc")
	secondChildGroupName := acctest.RandomWithPrefix("tf-acc")

	runTestNestedGroup(t, parentGroupName, firstChildGroupName, secondChildGroupName)
}

func runTestNestedGroup(t *testing.T, parentGroupName, firstChildGroupName, secondChildGroupName string) {
	parentGroupResource := "keycloak_group.parent_group"
	firstChildGroupResource := "keycloak_group.first_child_group"
	secondChildGroupResource := "keycloak_group.second_child_group"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakGroupDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroup_nested(parentGroupName, firstChildGroupName, secondChildGroupName, firstChildGroupResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExists(parentGroupResource),
					testAccCheckKeycloakGroupExists(firstChildGroupResource),
					testAccCheckKeycloakGroupExists(secondChildGroupResource),

					resource.TestCheckNoResourceAttr(parentGroupResource, "parent_id"),
					resource.TestCheckResourceAttrPair(firstChildGroupResource, "parent_id", parentGroupResource, "id"),
					resource.TestCheckResourceAttrPair(secondChildGroupResource, "parent_id", firstChildGroupResource, "id"),
				),
			},
			{
				ResourceName:        parentGroupResource,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
			{
				ResourceName:        firstChildGroupResource,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
			{
				ResourceName:        secondChildGroupResource,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
			{
				Config: testKeycloakGroup_nested(parentGroupName, firstChildGroupName, secondChildGroupName, parentGroupResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExists(parentGroupResource),
					testAccCheckKeycloakGroupExists(firstChildGroupResource),
					testAccCheckKeycloakGroupExists(secondChildGroupResource),

					resource.TestCheckNoResourceAttr(parentGroupResource, "parent_id"),
					resource.TestCheckResourceAttrPair(firstChildGroupResource, "parent_id", parentGroupResource, "id"),
					resource.TestCheckResourceAttrPair(secondChildGroupResource, "parent_id", parentGroupResource, "id"),
				),
			},
			{
				ResourceName:        parentGroupResource,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
			{
				ResourceName:        firstChildGroupResource,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
			{
				ResourceName:        secondChildGroupResource,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakGroup_unsetOptionalAttributes(t *testing.T) {
	t.Parallel()

	attributeName := acctest.RandomWithPrefix("tf-acc")
	groupWithOptionalAttributes := &keycloak.Group{
		RealmId: "terraform-" + acctest.RandString(10),
		Name:    "terraform-group-" + acctest.RandString(10),
		Attributes: map[string][]string{
			attributeName: {
				acctest.RandString(230),
				acctest.RandString(12),
			},
		},
	}

	resourceName := "keycloak_group.group"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakGroupDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroup_fromInterface(groupWithOptionalAttributes),
				Check:  testAccCheckKeycloakGroupExists(resourceName),
			},
			{
				Config: testKeycloakGroup_basic(groupWithOptionalAttributes.Name, attributeName, strings.Join(groupWithOptionalAttributes.Attributes[attributeName], "")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", groupWithOptionalAttributes.Name),
				),
			},
		},
	})
}

func testAccCheckKeycloakGroupExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getGroupFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakGroupFetch(resourceName string, group *keycloak.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedGroup, err := getGroupFromState(s, resourceName)
		if err != nil {
			return err
		}

		group.Id = fetchedGroup.Id
		group.RealmId = fetchedGroup.RealmId

		return nil
	}
}

func testAccCheckKeycloakGroupBelongsToRealm(resourceName, realm string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		group, err := getGroupFromState(s, resourceName)
		if err != nil {
			return err
		}

		if group.RealmId != realm {
			return fmt.Errorf("expected group with id %s to have realm_id of %s, but got %s", group.Id, realm, group.RealmId)
		}

		return nil
	}
}

func testAccCheckKeycloakGroupDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_group" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			group, _ := keycloakClient.GetGroup(testCtx, realm, id)
			if group != nil {
				return fmt.Errorf("group with id %s still exists", id)
			}
		}

		return nil
	}
}

func getGroupFromState(s *terraform.State, resourceName string) (*keycloak.Group, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	group, err := keycloakClient.GetGroup(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting group with id %s: %s", id, err)
	}

	return group, nil
}

func testKeycloakGroup_basic(group string, attributeName string, attributeValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
	attributes = {
		"%s" = "%s"
	}
}
	`, testAccRealm.Realm, group, attributeName, attributeValue)
}

func testKeycloakGroup_updateRealmBefore(group string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_1" {
	realm = "%s"
}

data "keycloak_realm" "realm_2" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm_1.id
}
	`, testAccRealm.Realm, testAccRealmTwo.Realm, group)
}

func testKeycloakGroup_updateRealmAfter(group string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_1" {
	realm = "%s"
}

data "keycloak_realm" "realm_2" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm_2.id
}
	`, testAccRealm.Realm, testAccRealmTwo.Realm, group)
}

func testKeycloakGroup_nested(parentGroup, firstChildGroup, secondChildGroup, secondChildGroupParent string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "parent_group" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_group" "first_child_group" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	parent_id = keycloak_group.parent_group.id
}

resource "keycloak_group" "second_child_group" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	parent_id = %s.id
}
	`, testAccRealm.Realm, parentGroup, firstChildGroup, secondChildGroup, secondChildGroupParent)
}

func testKeycloakGroup_fromInterface(group *keycloak.Group) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}
	`, testAccRealm.Realm, group.Name)
}
