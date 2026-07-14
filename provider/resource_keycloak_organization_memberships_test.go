package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOrganizationMemberships_basic(t *testing.T) {
	t.Parallel()

	organizationName := acctest.RandomWithPrefix("tf-acc")
	username := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganizationMemberships_basic(organizationName, username),
				Check:  testAccCheckUserBelongsToOrganization("keycloak_organization_memberships.org_members", username),
			},
			{
				// we need a separate test for destroy instead of using CheckDestroy because this resource is implicitly
				// destroyed at the end of each test via destroying the users or organization they're tied to
				Config: testKeycloakOrganizationMemberships_noMembers(organizationName, username),
				Check:  testAccCheckOrganizationHasNoMembers("keycloak_organization.organization"),
			},
		},
	})
}

func TestAccKeycloakOrganizationMemberships_adoptPreExistingMembers(t *testing.T) {
	t.Parallel()

	organizationName := acctest.RandomWithPrefix("tf-acc")
	preExistingUser := acctest.RandomWithPrefix("tf-acc")
	managedUser := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: Pre-populate the organization with a member out of band
			{
				Config: testKeycloakOrganizationMemberships_noMembersAndPreExistingUser(organizationName, preExistingUser, managedUser),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						org, err := keycloakClient.GetOrganizationByName(testCtx, testAccRealm.Realm, organizationName)
						if err != nil {
							return err
						}
						user, err := keycloakClient.GetUserByUsername(testCtx, testAccRealm.Realm, preExistingUser)
						if err != nil {
							return err
						}
						return keycloakClient.AddUserToOrganization(testCtx, testAccRealm.Realm, org.Id, user.Id)
					},
				),
			},
			// Step 2: Declare memberships with `managedUser` only. This should adopt the organization,
			// adding `managedUser` and removing the pre-existing `preExistingUser`.
			{
				Config: testKeycloakOrganizationMemberships_basic(organizationName, managedUser),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserBelongsToOrganization("keycloak_organization_memberships.org_members", managedUser),
					testAccCheckUsersDontBelongToOrganization("keycloak_organization_memberships.org_members", []string{preExistingUser}),
				),
			},
		},
	})
}

func TestAccKeycloakOrganizationMemberships_updateInPlace(t *testing.T) {
	t.Parallel()

	organizationName := acctest.RandomWithPrefix("tf-acc")

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
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// init
			{
				Config: testKeycloakOrganizationMemberships_multipleUsers(organizationName, allUsersForTest, allUsersForTest),
				Check:  testAccCheckUsersBelongToOrganization("keycloak_organization_memberships.org_members", allUsersForTest),
			},
			// remove
			{
				Config: testKeycloakOrganizationMemberships_multipleUsers(organizationName, allUsersForTest, subsetOfUsers),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsersBelongToOrganization("keycloak_organization_memberships.org_members", subsetOfUsers),
					testAccCheckUsersDontBelongToOrganization("keycloak_organization_memberships.org_members", []string{randomUserToRemove}),
				),
			},
			// add
			{
				Config: testKeycloakOrganizationMemberships_multipleUsers(organizationName, allUsersForTest, allUsersForTest),
				Check:  testAccCheckUsersBelongToOrganization("keycloak_organization_memberships.org_members", allUsersForTest),
			},
		},
	})
}

func TestAccKeycloakOrganizationMemberships_userDoesNotExist(t *testing.T) {
	t.Parallel()

	organizationName := acctest.RandomWithPrefix("tf-acc")
	username := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOrganizationMemberships_userDoesNotExist(organizationName, username),
				ExpectError: regexp.MustCompile("user with username .+ does not exist"),
			},
		},
	})
}

// if a user is removed from an org controlled by this resource, terraform should add them again
func TestAccKeycloakOrganizationMemberships_authoritativeAdd(t *testing.T) {
	t.Parallel()

	organizationName := acctest.RandomWithPrefix("tf-acc")

	usersInOrg := []string{
		"terraform-user-" + acctest.RandString(10),
		"terraform-user-" + acctest.RandString(10),
		"terraform-user-" + acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganizationMemberships_multipleUsers(organizationName, usersInOrg, usersInOrg),
				Check:  testAccCheckUsersBelongToOrganization("keycloak_organization_memberships.org_members", usersInOrg),
			},
			{
				PreConfig: func() {
					org, err := keycloakClient.GetOrganizationByName(testCtx, testAccRealm.Realm, organizationName)
					if err != nil {
						t.Fatal(err)
					}

					userToManuallyRemove := usersInOrg[acctest.RandIntRange(0, len(usersInOrg)-1)]

					user, err := keycloakClient.GetUserByUsername(testCtx, testAccRealm.Realm, userToManuallyRemove)
					if err != nil {
						t.Fatal(err)
					}

					err = keycloakClient.RemoveUserFromOrganization(testCtx, testAccRealm.Realm, org.Id, user.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOrganizationMemberships_multipleUsers(organizationName, usersInOrg, usersInOrg),
				Check:  testAccCheckUsersBelongToOrganization("keycloak_organization_memberships.org_members", usersInOrg),
			},
		},
	})
}

// if a user is added to an org controlled by this resource, terraform should remove them
func TestAccKeycloakOrganizationMemberships_authoritativeRemove(t *testing.T) {
	t.Parallel()

	organizationName := acctest.RandomWithPrefix("tf-acc")

	allUsersForTest := []string{
		"terraform-user-" + acctest.RandString(10),
		"terraform-user-" + acctest.RandString(10),
		"terraform-user-" + acctest.RandString(10),
		"terraform-user-" + acctest.RandString(10),
	}

	var usersInOrg []string
	indexOfUserToManuallyAdd := acctest.RandIntRange(0, len(allUsersForTest)-1)
	userToManuallyAdd := allUsersForTest[indexOfUserToManuallyAdd]
	for index, user := range allUsersForTest {
		if index != indexOfUserToManuallyAdd {
			usersInOrg = append(usersInOrg, user)
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOrganizationMemberships_multipleUsers(organizationName, allUsersForTest, usersInOrg),
				Check:  testAccCheckUsersBelongToOrganization("keycloak_organization_memberships.org_members", usersInOrg),
			},
			{
				PreConfig: func() {
					org, err := keycloakClient.GetOrganizationByName(testCtx, testAccRealm.Realm, organizationName)
					if err != nil {
						t.Fatal(err)
					}

					user, err := keycloakClient.GetUserByUsername(testCtx, testAccRealm.Realm, userToManuallyAdd)
					if err != nil {
						t.Fatal(err)
					}

					err = keycloakClient.AddUserToOrganization(testCtx, testAccRealm.Realm, org.Id, user.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOrganizationMemberships_multipleUsers(organizationName, allUsersForTest, usersInOrg),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsersBelongToOrganization("keycloak_organization_memberships.org_members", usersInOrg),
					testAccCheckUsersDontBelongToOrganization("keycloak_organization_memberships.org_members", []string{userToManuallyAdd}),
				),
			},
		},
	})
}

// --- State helpers ---

func testAccGetUsersInOrganizationFromState(resourceName string, s *terraform.State) ([]*keycloak.User, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realmId := rs.Primary.Attributes["realm_id"]
	organizationId := rs.Primary.Attributes["organization_id"]

	return keycloakClient.GetOrganizationMembers(testCtx, realmId, organizationId)
}

// --- Check helpers ---

func testAccCheckUserBelongsToOrganization(resourceName, user string) resource.TestCheckFunc {
	return testAccCheckUsersBelongToOrganization(resourceName, []string{user})
}

func testAccCheckUsersBelongToOrganization(resourceName string, users []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		usersInOrg, err := testAccGetUsersInOrganizationFromState(resourceName, s)
		if err != nil {
			return err
		}

		for _, user := range users {
			userFound := false

			for _, userInOrg := range usersInOrg {
				if user == userInOrg.Username {
					userFound = true

					break
				}
			}

			if !userFound {
				return fmt.Errorf("unable to find user %s in organization", user)
			}
		}

		return nil
	}
}

func testAccCheckUsersDontBelongToOrganization(resourceName string, users []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		usersInOrg, err := testAccGetUsersInOrganizationFromState(resourceName, s)
		if err != nil {
			return err
		}

		for _, user := range users {
			for _, userInOrg := range usersInOrg {
				if user == userInOrg.Username {
					return fmt.Errorf("expected user %s to not belong to organization", user)
				}
			}
		}

		return nil
	}
}

func testAccCheckOrganizationHasNoMembers(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realmId := rs.Primary.Attributes["realm"]
		organizationId := rs.Primary.ID

		members, err := keycloakClient.GetOrganizationMembers(testCtx, realmId, organizationId)
		if err != nil {
			return err
		}

		if len(members) != 0 {
			return fmt.Errorf("expected organization %s to have no members, but it has %d", organizationId, len(members))
		}

		return nil
	}
}

// --- Config helpers ---

func testKeycloakOrganizationMemberships_basic(organizationName, username string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	domain {
		name = "example.com"
	}
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_organization_memberships" "org_members" {
	realm_id        = data.keycloak_realm.realm.id
	organization_id = keycloak_organization.organization.id

	members = [
		keycloak_user.user.username
	]
}
	`, testAccRealm.Realm, organizationName, username)
}

func testKeycloakOrganizationMemberships_noMembers(organizationName, username string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	domain {
		name = "example.com"
	}
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}
	`, testAccRealm.Realm, organizationName, username)
}

func testKeycloakOrganizationMemberships_multipleUsers(organizationName string, definedUsers, usersInOrg []string) string {
	var userResources strings.Builder
	for _, username := range definedUsers {
		userResources.WriteString(fmt.Sprintf(`
resource "keycloak_user" "user_%s" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}
		`, username, username))
	}

	var usersInOrgInterpolated []string
	for _, userInOrg := range usersInOrg {
		usersInOrgInterpolated = append(usersInOrgInterpolated, fmt.Sprintf("${keycloak_user.user_%s.username}", userInOrg))
	}

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	domain {
		name = "example.com"
	}
}

%s

resource "keycloak_organization_memberships" "org_members" {
	realm_id        = data.keycloak_realm.realm.id
	organization_id = keycloak_organization.organization.id

	members = %s
}
	`, testAccRealm.Realm, organizationName, userResources.String(), arrayOfStringsForTerraformResource(usersInOrgInterpolated))
}

func testKeycloakOrganizationMemberships_userDoesNotExist(organizationName, username string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	domain {
		name = "example.com"
	}
}

resource "keycloak_organization_memberships" "org_members" {
	realm_id        = data.keycloak_realm.realm.id
	organization_id = keycloak_organization.organization.id

	members = [
		"%s"
	]
}
	`, testAccRealm.Realm, organizationName, username)
}

func testKeycloakOrganizationMemberships_noMembersAndPreExistingUser(organizationName, preExistingUser, managedUser string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "organization" {
	name  = "%s"
	realm = data.keycloak_realm.realm.id

	domain {
		name = "example.com"
	}
}

resource "keycloak_user" "pre_existing_user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}
	`, testAccRealm.Realm, organizationName, preExistingUser, managedUser)
}
