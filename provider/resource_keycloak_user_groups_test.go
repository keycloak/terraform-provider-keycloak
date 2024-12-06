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

func TestAccKeycloakUserGroups_basic(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")
	userName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUserGroups_basic(groupName, userName),
				Check:  testAccCheckKeycloakUserHasGroups("keycloak_user_groups.user_groups"),
			},
			{
				ResourceName:      "keycloak_user_groups.user_groups",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// check destroy
			{
				Config: testKeycloakUserGroups_noUserGroups(groupName, userName),
				Check:  testAccCheckKeycloakUserHasNoGroups("keycloak_user.user"),
			},
		},
	})
}

func TestAccKeycloakUserGroups_basicNonExhaustive(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")
	userName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUserGroups_nonExhaustive(groupName, userName),
				Check:  testAccCheckKeycloakUserHasGroups("keycloak_user_groups.user_groups"),
			},
			// check destroy
			{
				Config: testKeycloakUserGroups_noUserGroups(groupName, userName),
				Check:  testAccCheckKeycloakUserHasNoGroups("keycloak_user.user"),
			},
		},
	})
}

func TestAccKeycloakUserGroups_update(t *testing.T) {
	t.Parallel()

	userName := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	allGroupIds := []string{
		"${keycloak_group.group1.id}",
		"${keycloak_group.group2.id}",
		"${keycloak_group.group3.id}",
		"${keycloak_group.group4.id}",
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// initial setup, resource is defined but no roles are specified
			{
				Config: testKeycloakUserGroups_update(userName, groupName, []string{}),
				Check:  testAccCheckKeycloakUserHasGroups("keycloak_user_groups.user_groups"),
			},
			// add all roles
			{
				Config: testKeycloakUserGroups_update(userName, groupName, allGroupIds),
				Check:  testAccCheckKeycloakUserHasGroups("keycloak_user_groups.user_groups"),
			},
			// remove some
			{
				Config: testKeycloakUserGroups_update(userName, groupName, []string{
					"${keycloak_group.group3.id}",
					"${keycloak_group.group4.id}",
				}),
				Check: testAccCheckKeycloakUserHasGroups("keycloak_user_groups.user_groups"),
			},
			// add some and remove some
			{
				Config: testKeycloakUserGroups_update(userName, groupName, []string{
					"${keycloak_group.group1.id}",
					"${keycloak_group.group4.id}",
				}),
				Check: testAccCheckKeycloakUserHasGroups("keycloak_user_groups.user_groups"),
			},
			// add some and remove some again
			{
				Config: testKeycloakUserGroups_update(userName, groupName, []string{
					"${keycloak_group.group1.id}",
					"${keycloak_group.group2.id}",
					"${keycloak_group.group3.id}",
				}),
				Check: testAccCheckKeycloakUserHasGroups("keycloak_user_groups.user_groups"),
			},
			// add all back
			{
				Config: testKeycloakUserGroups_update(userName, groupName, allGroupIds),
				Check:  testAccCheckKeycloakUserHasGroups("keycloak_user_groups.user_groups"),
			},
			// random scenario 1
			{
				Config: testKeycloakUserGroups_update(userName, groupName, randomStringSliceSubset(allGroupIds)),
				Check:  testAccCheckKeycloakUserHasGroups("keycloak_user_groups.user_groups"),
			},
			// random scenario 2
			{
				Config: testKeycloakUserGroups_update(userName, groupName, randomStringSliceSubset(allGroupIds)),
				Check:  testAccCheckKeycloakUserHasGroups("keycloak_user_groups.user_groups"),
			},
			// random scenario 3
			{
				Config: testKeycloakUserGroups_update(userName, groupName, randomStringSliceSubset(allGroupIds)),
				Check:  testAccCheckKeycloakUserHasGroups("keycloak_user_groups.user_groups"),
			},
			// remove all
			{
				Config: testKeycloakUserGroups_update(userName, groupName, []string{}),
				Check:  testAccCheckKeycloakUserHasGroups("keycloak_user_groups.user_groups"),
			},
		},
	})
}

func TestAccKeycloakUserGroups_updateNonExhaustive(t *testing.T) {
	t.Parallel()

	userName := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	allGroupIdsSet1 := []string{
		"${keycloak_group.group1.id}",
		"${keycloak_group.group2.id}",
		"${keycloak_group.group3.id}",
	}

	allGroupIdsSet2 := []string{
		"${keycloak_group.group4.id}",
		"${keycloak_group.group5.id}",
		"${keycloak_group.group6.id}",
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// initial setup, resource is defined but no roles are specified
			{
				Config: testKeycloakUserGroups_updateNonExhaustive(userName, groupName, []string{}, []string{}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_1"),
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_2")),
			},
			// add all roles
			{
				Config: testKeycloakUserGroups_updateNonExhaustive(userName, groupName, allGroupIdsSet1, allGroupIdsSet2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_1"),
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_2"))},
			// remove some
			{
				Config: testKeycloakUserGroups_updateNonExhaustive(userName, groupName, []string{
					"${keycloak_group.group3.id}",
				}, allGroupIdsSet2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_1"),
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_2"))},
			// add some and remove some
			{
				Config: testKeycloakUserGroups_updateNonExhaustive(userName, groupName, []string{
					"${keycloak_group.group1.id}",
					"${keycloak_group.group2.id}",
				}, allGroupIdsSet2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_1"),
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_2"))},
			// add some and remove some again
			{
				Config: testKeycloakUserGroups_updateNonExhaustive(userName, groupName, []string{
					"${keycloak_group.group1.id}",
					"${keycloak_group.group3.id}",
				}, allGroupIdsSet2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_1"),
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_2"))},
			// add all back
			{
				Config: testKeycloakUserGroups_updateNonExhaustive(userName, groupName, allGroupIdsSet1, allGroupIdsSet2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_1"),
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_2"))},
			// random scenario 1
			{
				Config: testKeycloakUserGroups_updateNonExhaustive(userName, groupName, randomStringSliceSubset(allGroupIdsSet1), randomStringSliceSubset(allGroupIdsSet2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_1"),
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_2"))},
			// remove all
			{
				Config: testKeycloakUserGroups_updateNonExhaustive(userName, groupName, []string{}, []string{}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_1"),
					testAccCheckKeycloakUserHasNonExhaustiveGroups("keycloak_user_groups.user_groups_2")),
			},
		},
	})
}

func testAccCheckKeycloakUserHasGroups(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realm := rs.Primary.Attributes["realm_id"]
		userId := rs.Primary.Attributes["user_id"]

		var expectedGroups []*keycloak.Group
		for k, v := range rs.Primary.Attributes {
			if match, _ := regexp.MatchString("group_ids\\.[^#]+", k); !match {
				continue
			}

			group, err := keycloakClient.GetGroup(testCtx, realm, v)
			if err != nil {
				return err
			}

			expectedGroups = append(expectedGroups, group)
		}

		userGroups, err := keycloakClient.GetUserGroups(testCtx, realm, userId)
		if err != nil {
			return err
		}

		if len(userGroups) != len(expectedGroups) {
			return fmt.Errorf("expected number of user groups to be %d, got %d", len(expectedGroups), len(userGroups))
		}

		for _, expectedGroup := range expectedGroups {

			found := false

			for _, group := range userGroups {
				if group.Id == expectedGroup.Id {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("expected to find group %s assigned to user %s", expectedGroup.Id, userId)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakUserHasNonExhaustiveGroups(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realm := rs.Primary.Attributes["realm_id"]
		userId := rs.Primary.Attributes["user_id"]

		var expectedGroups []*keycloak.Group
		for k, v := range rs.Primary.Attributes {
			if match, _ := regexp.MatchString("group_ids\\.[^#]+", k); !match {
				continue
			}

			group, err := keycloakClient.GetGroup(testCtx, realm, v)
			if err != nil {
				return err
			}

			expectedGroups = append(expectedGroups, group)
		}

		userGroups, err := keycloakClient.GetUserGroups(testCtx, realm, userId)
		if err != nil {
			return err
		}

		if len(userGroups) < len(expectedGroups) {
			return fmt.Errorf("expected number of user groups to be greater or equals to %d, got %d", len(expectedGroups), len(userGroups))
		}

		for _, expectedGroup := range expectedGroups {

			found := false

			for _, group := range userGroups {
				if group.Id == expectedGroup.Id {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("expected to find group %s assigned to user %s", expectedGroup.Id, userId)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakUserHasNoGroups(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realm := rs.Primary.Attributes["realm_id"]
		id := rs.Primary.ID

		userGroups, err := keycloakClient.GetUserGroups(testCtx, realm, id)
		if err != nil {
			return err
		}

		if len(userGroups) != 0 {
			return fmt.Errorf("expected user %s to have no groups", id)
		}

		return nil
	}
}

func testKeycloakUserGroups_basic(groupName, userName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s"
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_user_groups" "user_groups" {
	realm_id = data.keycloak_realm.realm.id
  	user_id  = keycloak_user.user.id
  	group_ids = [
    	keycloak_group.group.id
  	]
}
	`, testAccRealm.Realm, groupName, userName)
}

func testKeycloakUserGroups_noUserGroups(groupName, userName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s"
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}
	`, testAccRealm.Realm, groupName, userName)
}

func testKeycloakUserGroups_nonExhaustive(groupName, userName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s"
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_user_groups" "user_groups" {
	realm_id = data.keycloak_realm.realm.id
  	user_id  = keycloak_user.user.id
  	group_ids = [
    	keycloak_group.group.id
  	]
	exhaustive = false
}
	`, testAccRealm.Realm, groupName, userName)
}

func testKeycloakUserGroups_update(userName, groupName string, groupIds []string) string {
	tfGroupIds := fmt.Sprintf("group_ids = %s", arrayOfStringsForTerraformResource(groupIds))

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group1" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s1"
}

resource "keycloak_group" "group2" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s2"
}

resource "keycloak_group" "group3" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s3"
}

resource "keycloak_group" "group4" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s4"
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_user_groups" "user_groups" {
	realm_id = data.keycloak_realm.realm.id
  	user_id  = keycloak_user.user.id
  	%s
}
	`, testAccRealm.Realm, userName, groupName, groupName, groupName, groupName, tfGroupIds)
}

func testKeycloakUserGroups_updateNonExhaustive(userName, groupName string, groupIds1, groupIds2 []string) string {
	tfGroupIds1 := fmt.Sprintf("group_ids = %s", arrayOfStringsForTerraformResource(groupIds1))
	tfGroupIds2 := fmt.Sprintf("group_ids = %s", arrayOfStringsForTerraformResource(groupIds2))

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group1" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s1"
}

resource "keycloak_group" "group2" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s2"
}

resource "keycloak_group" "group3" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s3"
}

resource "keycloak_group" "group4" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s4"
}

resource "keycloak_group" "group5" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s5"
}

resource "keycloak_group" "group6" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s6"
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_user_groups" "user_groups_1" {
	realm_id = data.keycloak_realm.realm.id
  	user_id  = keycloak_user.user.id

	exhaustive = false
  	%s
}

resource "keycloak_user_groups" "user_groups_2" {
	realm_id = data.keycloak_realm.realm.id
  	user_id  = keycloak_user.user.id

	exhaustive = false
  	%s
}
	`, testAccRealm.Realm, userName, groupName, groupName, groupName, groupName, groupName, groupName, tfGroupIds1, tfGroupIds2)
}
