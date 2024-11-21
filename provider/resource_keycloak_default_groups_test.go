package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakDefaultGroups_basic(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakDefaultGroups_basic(realmName, groupName),
				Check:  testAccCheckGroupsAreDefault("keycloak_default_groups.group_default", []string{groupName}),
			},
			{
				// we need a separate test for destroy instead of using CheckDestroy because this resource is implicitly
				// destroyed at the end of each test via destroying users or groups they're tied to
				Config: testKeycloakDefaultGroups_noDefaultGroups(realmName, groupName),
				Check:  testAccNoDefaultGroups("keycloak_group.group"),
			},
		},
	})
}

func TestAccKeycloakDefaultGroups_import(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakDefaultGroups_basic(realmName, groupName),
				Check:  testAccCheckGroupsAreDefault("keycloak_default_groups.group_default", []string{groupName}),
			},
			{
				ResourceName:      "keycloak_default_groups.group_default",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     realmName,
			},
		},
	})
}

func TestAccKeycloakDefaultGroups_updateInPlace(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")

	allGroupsForTest := []string{
		"terraform-group-" + acctest.RandString(10),
		"terraform-group-" + acctest.RandString(10),
		"terraform-group-" + acctest.RandString(10),
	}
	indexOfRandomGroupToRemove := acctest.RandIntRange(0, len(allGroupsForTest)-1)
	randomGroupToRemove := allGroupsForTest[indexOfRandomGroupToRemove]

	var subsetOfGroups []string
	for index, group := range allGroupsForTest {
		if index != indexOfRandomGroupToRemove {
			subsetOfGroups = append(subsetOfGroups, group)
		}
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// init
			{
				Config: testKeycloakDefaultGroups_multipleGroups(realmName, allGroupsForTest, allGroupsForTest),
				Check:  testAccCheckGroupsAreDefault("keycloak_default_groups.group_default", allGroupsForTest),
			},
			// remove
			{
				Config: testKeycloakDefaultGroups_multipleGroups(realmName, allGroupsForTest, subsetOfGroups),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupsAreDefault("keycloak_default_groups.group_default", subsetOfGroups),
					testAccCheckGroupsArentDefault("keycloak_default_groups.group_default", []string{randomGroupToRemove}),
				),
			},
			// add
			{
				Config: testKeycloakDefaultGroups_multipleGroups(realmName, allGroupsForTest, allGroupsForTest),
				Check:  testAccCheckGroupsAreDefault("keycloak_default_groups.group_default", allGroupsForTest),
			},
		},
	})
}

func testAccCheckGroupsAreDefault(resourceName string, groupNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		stateGroups, err := testAccGetGroupsFromDefaultGroup(resourceName, s)
		if err != nil {
			return err
		}

		for _, groupName := range groupNames {
			found := false

			for _, stateGroup := range stateGroups {
				if stateGroup.Name == groupName {
					found = true
				}
			}

			if !found {
				return fmt.Errorf("unable to find group %s", groupName)
			}
		}

		return nil
	}
}

func testAccCheckGroupsArentDefault(resourceName string, groupNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		stateGroups, err := testAccGetGroupsFromDefaultGroup(resourceName, s)
		if err != nil {
			return err
		}

		for _, stateGroup := range stateGroups {
			for _, groupName := range groupNames {
				if groupName == stateGroup.Name {
					return fmt.Errorf("didnt expect to find group %s", groupName)
				}
			}
		}

		return nil
	}
}

func testAccNoDefaultGroups(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		stateGroups, err := testAccGetGroupsFromDefaultGroup(resourceName, s)
		if err != nil {
			return err
		}

		if len(stateGroups) != 0 {
			return fmt.Errorf("expected 0 stateGroups, got %d", len(stateGroups))
		}

		return nil
	}
}

func testAccGetGroupsFromDefaultGroup(resourceName string, s *terraform.State) ([]keycloak.Group, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realmId := rs.Primary.Attributes["realm_id"]

	return keycloakClient.GetDefaultGroups(testCtx, realmId)
}

func testKeycloakDefaultGroups_basic(realmName, groupName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = keycloak_realm.realm.id
}

resource "keycloak_default_groups" "group_default" {
	realm_id = keycloak_realm.realm.id
	group_ids = [
		keycloak_group.group.id
	]
}
	`, realmName, groupName)
}

func testKeycloakDefaultGroups_noDefaultGroups(realmName, groupName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = keycloak_realm.realm.id
}`, realmName, groupName)
}

// this tf config provides a good way to test groups that exist within keycloak but aren't defaults
func testKeycloakDefaultGroups_multipleGroups(realm string, groups []string, defaultGroups []string) string {
	out := fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}`, realm)

	for _, group := range groups {
		out += fmt.Sprintf(`
resource "keycloak_group" "%s" {
	name     = "%s"
	realm_id = keycloak_realm.realm.id
}`, group, group)
	}

	var defaultGroupResources []string
	for _, defaultGroup := range defaultGroups {
		defaultGroupResources = append(defaultGroupResources, fmt.Sprintf(`${keycloak_group.%s.id}`, defaultGroup))
	}

	out += fmt.Sprintf(`
resource "keycloak_default_groups" "group_default" {
	realm_id = keycloak_realm.realm.id
	group_ids = %s
}`, arrayOfStringsForTerraformResource(defaultGroupResources))

	return out
}
