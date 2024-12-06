package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"regexp"
	"strings"
	"testing"
)

func TestAccKeycloakGroupMemberships_basic(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")
	username := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupMemberships_basic(groupName, username),
				Check:  testAccCheckUserBelongsToGroup("keycloak_group_memberships.group_members", username),
			},
			{
				// we need a separate test for destroy instead of using CheckDestroy because this resource is implicitly
				// destroyed at the end of each test via destroying users or groups they're tied to
				Config: testKeycloakGroupMemberships_noGroupMemberships(groupName, username),
				Check:  testAccCheckUsersDontBelongToGroup("keycloak_group.group", []string{username}),
			},
		},
	})
}

func TestAccKeycloakGroupMemberships_basicUserWithBackslash(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")
	username := acctest.RandString(5) + `\\` + acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupMemberships_basic(groupName, username),
				Check:  testAccCheckUserBelongsToGroup("keycloak_group_memberships.group_members", strings.ReplaceAll(username, `\\`, `\`)),
			},
		},
	})
}

func TestAccKeycloakGroupMemberships_moreThan100members(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupMemberships_moreThan100members(groupName),
			},
		},
	})
}

func TestAccKeycloakGroupMemberships_updateGroupForceNew(t *testing.T) {
	t.Parallel()

	groupOne := acctest.RandomWithPrefix("tf-acc")
	groupTwo := acctest.RandomWithPrefix("tf-acc")

	username := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupMemberships_updateGroupForceNew(groupOne, groupTwo, username, "group_one"),
				Check:  testAccCheckUserBelongsToGroup("keycloak_group_memberships.group_members", username),
			},
			{
				Config: testKeycloakGroupMemberships_updateGroupForceNew(groupOne, groupTwo, username, "group_two"),
				Check:  testAccCheckUserBelongsToGroup("keycloak_group_memberships.group_members", username),
			},
		},
	})
}

func TestAccKeycloakGroupMemberships_updateInPlace(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")

	allUsersForTest := []string{
		"terraform-user-" + acctest.RandString(10),
		"terraform-user-" + acctest.RandString(10),
		"terraform-user-" + acctest.RandString(10),
	}
	indexOfRandomUserToRemove := acctest.RandIntRange(0, len(allUsersForTest)-1)
	randomUserToRemove := allUsersForTest[indexOfRandomUserToRemove]

	var subsetOfUsers []string
	for index, user := range allUsersForTest {
		if index != indexOfRandomUserToRemove {
			subsetOfUsers = append(subsetOfUsers, user)
		}
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// init
			{
				Config: testKeycloakGroupMemberships_multipleUsers(groupName, allUsersForTest, allUsersForTest),
				Check:  testAccCheckUsersBelongToGroup("keycloak_group_memberships.group_members", allUsersForTest),
			},
			// remove
			{
				Config: testKeycloakGroupMemberships_multipleUsers(groupName, allUsersForTest, subsetOfUsers),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsersBelongToGroup("keycloak_group_memberships.group_members", subsetOfUsers),
					testAccCheckUsersDontBelongToGroup("keycloak_group_memberships.group_members", []string{randomUserToRemove}),
				),
			},
			// add
			{
				Config: testKeycloakGroupMemberships_multipleUsers(groupName, allUsersForTest, allUsersForTest),
				Check:  testAccCheckUsersBelongToGroup("keycloak_group_memberships.group_members", allUsersForTest),
			},
		},
	})
}

func TestAccKeycloakGroupMemberships_userDoesNotExist(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")
	username := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakGroupMemberships_userDoesNotExist(groupName, username),
				ExpectError: regexp.MustCompile("user with username .+ does not exist"),
			},
		},
	})
}

// if a user is removed from a group controlled by this resource, terraform should add them again
func TestAccKeycloakGroupMemberships_authoritativeAdd(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")

	usersInGroup := []string{
		"terraform-user-" + acctest.RandString(10),
		"terraform-user-" + acctest.RandString(10),
		"terraform-user-" + acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupMemberships_multipleUsers(groupName, usersInGroup, usersInGroup),
				Check:  testAccCheckUsersBelongToGroup("keycloak_group_memberships.group_members", usersInGroup),
			},
			{
				PreConfig: func() {
					groupsWithName, err := keycloakClient.ListGroupsWithName(testCtx, testAccRealm.Realm, groupName)
					if err != nil {
						t.Fatal(err)
					}

					userToManuallyRemove := usersInGroup[acctest.RandIntRange(0, len(usersInGroup)-1)]

					err = keycloakClient.RemoveUsersFromGroup(testCtx, testAccRealm.Realm, groupsWithName[0].Id, []interface{}{userToManuallyRemove})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakGroupMemberships_multipleUsers(groupName, usersInGroup, usersInGroup),
				Check:  testAccCheckUsersBelongToGroup("keycloak_group_memberships.group_members", usersInGroup),
			},
		},
	})
}

// if a user is added to a group controlled by this resource, terraform should remove them
func TestAccKeycloakGroupMemberships_authoritativeRemove(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")

	allUsersForTest := []string{
		"terraform-user-" + acctest.RandString(10),
		"terraform-user-" + acctest.RandString(10),
		"terraform-user-" + acctest.RandString(10),
		"terraform-user-" + acctest.RandString(10),
	}

	var usersInGroup []string
	indexOfUserToManuallyAdd := acctest.RandIntRange(0, len(allUsersForTest)-1)
	userToManuallyAdd := allUsersForTest[indexOfUserToManuallyAdd]
	for index, user := range allUsersForTest {
		if index != indexOfUserToManuallyAdd {
			usersInGroup = append(usersInGroup, user)
		}
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupMemberships_multipleUsers(groupName, allUsersForTest, usersInGroup),
				Check:  testAccCheckUsersBelongToGroup("keycloak_group_memberships.group_members", usersInGroup),
			},
			{
				PreConfig: func() {
					groupsWithName, err := keycloakClient.ListGroupsWithName(testCtx, testAccRealm.Realm, groupName)
					if err != nil {
						t.Fatal(err)
					}

					err = keycloakClient.AddUsersToGroup(testCtx, testAccRealm.Realm, groupsWithName[0].Id, []interface{}{userToManuallyAdd})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakGroupMemberships_multipleUsers(groupName, allUsersForTest, usersInGroup),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsersBelongToGroup("keycloak_group_memberships.group_members", usersInGroup),
					testAccCheckUsersDontBelongToGroup("keycloak_group_memberships.group_members", []string{userToManuallyAdd}),
				),
			},
		},
	})
}

// this resource doesn't support import because it can be created even if the desired state already exists in keycloak
func TestAccKeycloakGroupMemberships_noImportNeeded(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")
	username := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupMemberships_noGroupMemberships(groupName, username),
				Check:  testAccCheckUsersDontBelongToGroup("keycloak_group.group", []string{username}),
			},
			{
				PreConfig: func() {
					groupsWithName, err := keycloakClient.ListGroupsWithName(testCtx, testAccRealm.Realm, groupName)
					if err != nil {
						t.Fatal(err)
					}

					err = keycloakClient.AddUsersToGroup(testCtx, testAccRealm.Realm, groupsWithName[0].Id, []interface{}{username})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakGroupMemberships_basic(groupName, username),
				Check:  testAccCheckUserBelongsToGroup("keycloak_group.group", username),
			},
		},
	})
}

func TestAccKeycloakGroupMemberships_validateLowercaseUsernames(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")
	randomString := acctest.RandomWithPrefix("tf-acc")
	username := "terraform-user-" + randomString
	usernameWithUppercaseCharacters := "terraform-user-" + strings.ToUpper(randomString)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakGroupMemberships_hardcodedUsername(groupName, username, usernameWithUppercaseCharacters),
				ExpectError: regexp.MustCompile("expected all usernames within group membership to be lowercase"),
			},
		},
	})
}

func TestAccKeycloakGroupMemberships_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")
	username := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_group_memberships.group_members"

	var groupId *string

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupMemberships_basic(groupName, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserBelongsToGroup(resourceName, username), func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceName)
						}

						stateGroupId := rs.Primary.Attributes["group_id"]
						groupId = &stateGroupId

						return nil
					},
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteGroup(testCtx, testAccRealm.Realm, *groupId)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakGroupMemberships_basic(groupName, username),
				Check:  testAccCheckUserBelongsToGroup(resourceName, username),
			},
		},
	})
}

func testAccGetUsersInGroupFromGroupMembershipsState(resourceName string, s *terraform.State) ([]*keycloak.User, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realmId := rs.Primary.Attributes["realm_id"]

	var groupId string
	if strings.HasPrefix(resourceName, "keycloak_group_membership") {
		groupId = rs.Primary.Attributes["group_id"]
	} else {
		groupId = rs.Primary.ID
	}

	return keycloakClient.GetGroupMembers(testCtx, realmId, groupId)
}

func testAccCheckUserBelongsToGroup(resourceName, user string) resource.TestCheckFunc {
	return testAccCheckUsersBelongToGroup(resourceName, []string{user})
}

func testAccCheckUsersBelongToGroup(resourceName string, users []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		usersInGroup, err := testAccGetUsersInGroupFromGroupMembershipsState(resourceName, s)
		if err != nil {
			return err
		}

		for _, user := range users {
			userFound := false

			for _, userInGroup := range usersInGroup {
				if user == userInGroup.Username {
					userFound = true

					break
				}
			}

			if !userFound {
				return fmt.Errorf("unable to find user %s in group", user)
			}
		}

		return nil
	}
}

func testAccCheckUsersDontBelongToGroup(resourceName string, users []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		usersInGroup, err := testAccGetUsersInGroupFromGroupMembershipsState(resourceName, s)
		if err != nil {
			return err
		}

		for _, user := range users {
			for _, userInGroup := range usersInGroup {
				if user == userInGroup.Username {
					return fmt.Errorf("expected user %s to not belong to group", user)
				}
			}
		}

		return nil
	}
}

func testKeycloakGroupMemberships_basic(group, username string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_group_memberships" "group_members" {
	realm_id = data.keycloak_realm.realm.id
	group_id = keycloak_group.group.id

	members = [
		keycloak_user.user.username
	]
}
	`, testAccRealm.Realm, group, username)
}

func testKeycloakGroupMemberships_moreThan100members(group string) string {
	username := acctest.RandomWithPrefix("tf-acc")
	count := 110

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_user" "users" {
	count = %d

	realm_id = data.keycloak_realm.realm.id
	username = "%s-${count.index}"
}

resource "keycloak_group_memberships" "group_members" {
	realm_id = data.keycloak_realm.realm.id
	group_id = keycloak_group.group.id

	members = keycloak_user.users.*.username
}

        `, testAccRealm.Realm, group, count, username)
}

func testKeycloakGroupMemberships_updateGroupForceNew(groupOne, groupTwo, username, currentGroup string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group_one" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_group" "group_two" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_group_memberships" "group_members" {
	realm_id = data.keycloak_realm.realm.id
	group_id = keycloak_group.%s.id

	members = [
		keycloak_user.user.username
	]
}
	`, testAccRealm.Realm, groupOne, groupTwo, username, currentGroup)
}

// this tf config provides a good way to test users that exist within keycloak but are not necessarily part of a group
func testKeycloakGroupMemberships_multipleUsers(group string, definedUsers, usersInGroup []string) string {
	var userResources strings.Builder
	for _, username := range definedUsers {
		userResources.WriteString(fmt.Sprintf(`
resource "keycloak_user" "user_%s" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}
		`, username, username))
	}

	var usersInGroupInterpolated []string
	for _, userInGroup := range usersInGroup {
		usersInGroupInterpolated = append(usersInGroupInterpolated, fmt.Sprintf("${keycloak_user.user_%s.username}", userInGroup))
	}

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

%s

resource "keycloak_group_memberships" "group_members" {
	realm_id = data.keycloak_realm.realm.id
	group_id = keycloak_group.group.id

	members = %s
}
	`, testAccRealm.Realm, group, userResources.String(), arrayOfStringsForTerraformResource(usersInGroupInterpolated))
}

func testKeycloakGroupMemberships_userDoesNotExist(group, username string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_group_memberships" "group_members" {
	realm_id = data.keycloak_realm.realm.id
	group_id = keycloak_group.group.id

	members = [
		"%s"
	]
}
	`, testAccRealm.Realm, group, username)
}

func testKeycloakGroupMemberships_noGroupMemberships(group, username string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}
	`, testAccRealm.Realm, group, username)
}

func testKeycloakGroupMemberships_hardcodedUsername(group, username, hardcodedUsername string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_group_memberships" "group_members" {
	realm_id = data.keycloak_realm.realm.id
	group_id = keycloak_group.group.id

	members = [
		"%s"
	]
}
	`, testAccRealm.Realm, group, username, hardcodedUsername)
}
