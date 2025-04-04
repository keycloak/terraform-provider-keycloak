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

func TestAccKeycloakGroupRoles_basic(t *testing.T) {
	t.Parallel()

	realmRoleName := acctest.RandomWithPrefix("tf-acc")
	openIdClientName := acctest.RandomWithPrefix("tf-acc")
	openIdRoleName := acctest.RandomWithPrefix("tf-acc")
	samlClientName := acctest.RandomWithPrefix("tf-acc")
	samlRoleName := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupRoles_basic(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, groupName),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
			{
				ResourceName:      "keycloak_group_roles.group_roles",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// check destroy
			{
				Config: testKeycloakGroupRoles_noGroupRoles(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, groupName),
				Check:  testAccCheckKeycloakGroupHasNoRoles("keycloak_group.group"),
			},
		},
	})
}

func TestAccKeycloakGroupRoles_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var group = &keycloak.Group{}

	realmRoleName := acctest.RandomWithPrefix("tf-acc")
	openIdClientName := acctest.RandomWithPrefix("tf-acc")
	openIdRoleName := acctest.RandomWithPrefix("tf-acc")
	samlClientName := acctest.RandomWithPrefix("tf-acc")
	samlRoleName := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupRoles_basic(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
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
				Config: testKeycloakGroupRoles_basic(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, groupName),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
		},
	})
}

func TestAccKeycloakGroupRoles_update(t *testing.T) {
	t.Parallel()

	realmRoleOneName := acctest.RandomWithPrefix("tf-acc")
	realmRoleTwoName := acctest.RandomWithPrefix("tf-acc")
	openIdClientName := acctest.RandomWithPrefix("tf-acc")
	openIdRoleOneName := acctest.RandomWithPrefix("tf-acc")
	openIdRoleTwoName := acctest.RandomWithPrefix("tf-acc")
	samlClientName := acctest.RandomWithPrefix("tf-acc")
	samlRoleOneName := acctest.RandomWithPrefix("tf-acc")
	samlRoleTwoName := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	allRoleIds := []string{
		"${keycloak_role.realm_role_one.id}",
		"${keycloak_role.realm_role_two.id}",
		"${keycloak_role.openid_client_role_one.id}",
		"${keycloak_role.openid_client_role_two.id}",
		"${keycloak_role.saml_client_role_one.id}",
		"${keycloak_role.saml_client_role_two.id}",
		"${data.keycloak_role.offline_access.id}",
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// initial setup, resource is defined but no roles are specified
			{
				Config: testKeycloakGroupRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, []string{}),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
			// add all roles
			{
				Config: testKeycloakGroupRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, allRoleIds),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
			// remove some
			{
				Config: testKeycloakGroupRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, []string{
					"${keycloak_role.realm_role_two.id}",
					"${keycloak_role.openid_client_role_one.id}",
					"${keycloak_role.openid_client_role_two.id}",
					"${data.keycloak_role.offline_access.id}",
				}),
				Check: testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
			// add some and remove some
			{
				Config: testKeycloakGroupRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, []string{
					"${keycloak_role.saml_client_role_one.id}",
					"${keycloak_role.saml_client_role_two.id}",
					"${keycloak_role.realm_role_one.id}",
				}),
				Check: testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
			// add some and remove some again
			{
				Config: testKeycloakGroupRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, []string{
					"${keycloak_role.saml_client_role_one.id}",
					"${keycloak_role.openid_client_role_two.id}",
					"${keycloak_role.realm_role_two.id}",
					"${data.keycloak_role.offline_access.id}",
				}),
				Check: testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
			// add all back
			{
				Config: testKeycloakGroupRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, allRoleIds),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
			// random scenario 1
			{
				Config: testKeycloakGroupRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, randomStringSliceSubset(allRoleIds)),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
			// random scenario 2
			{
				Config: testKeycloakGroupRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, randomStringSliceSubset(allRoleIds)),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
			// random scenario 3
			{
				Config: testKeycloakGroupRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, randomStringSliceSubset(allRoleIds)),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
			// remove all
			{
				Config: testKeycloakGroupRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, []string{}),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
		},
	})
}

func TestAccKeycloakGroupRoles_basicNonExhaustive(t *testing.T) {
	t.Parallel()

	realmRoleName := acctest.RandomWithPrefix("tf-acc")
	openIdClientName := acctest.RandomWithPrefix("tf-acc")
	openIdRoleName := acctest.RandomWithPrefix("tf-acc")
	samlClientName := acctest.RandomWithPrefix("tf-acc")
	samlRoleName := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	// no group roles
	// multiple non_exhaustive
	// no group roles -> nothing.
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupRoles_nonExhaustive(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, groupName),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles1", false),
			},
			// check destroy
			{
				Config: testKeycloakGroupRoles_noGroupRoles(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, groupName),
				Check:  testAccCheckKeycloakGroupHasNoRoles("keycloak_group.group"),
			},
		},
	})
}

func TestAccKeycloakGroupRoles_updateNonExhaustive(t *testing.T) {
	t.Parallel()

	realmRoleOneName := acctest.RandomWithPrefix("tf-acc")
	realmRoleTwoName := acctest.RandomWithPrefix("tf-acc")
	openIdClientName := acctest.RandomWithPrefix("tf-acc")
	openIdRoleOneName := acctest.RandomWithPrefix("tf-acc")
	openIdRoleTwoName := acctest.RandomWithPrefix("tf-acc")
	samlClientName := acctest.RandomWithPrefix("tf-acc")
	samlRoleOneName := acctest.RandomWithPrefix("tf-acc")
	samlRoleTwoName := acctest.RandomWithPrefix("tf-acc")
	groupName := acctest.RandomWithPrefix("tf-acc")

	allRoleIdSet1 := []string{
		"${keycloak_role.realm_role_one.id}",
		"${keycloak_role.openid_client_role_one.id}",
		"${keycloak_role.saml_client_role_one.id}",
	}

	allRoleIdSet2 := []string{
		"${keycloak_role.realm_role_two.id}",
		"${keycloak_role.openid_client_role_two.id}",
		"${keycloak_role.saml_client_role_two.id}",
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// initial setup, resource is defined but no roles are specified
			{
				Config: testKeycloakGroupRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, []string{}, []string{}),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles1", false),
			},
			// add all roles
			{
				Config: testKeycloakGroupRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, allRoleIdSet1, allRoleIdSet2),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles1", false),
			},
			// remove some
			{
				Config: testKeycloakGroupRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, []string{
					"${keycloak_role.openid_client_role_one.id}",
				}, allRoleIdSet2),
				Check: testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles1", false),
			},
			// add some and remove some
			{
				Config: testKeycloakGroupRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, []string{
					"${data.keycloak_role.offline_access.id}",
					"${keycloak_role.saml_client_role_one.id}",
				}, allRoleIdSet2),
				Check: testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles1", false),
			},
			// add some and remove some again
			{
				Config: testKeycloakGroupRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, []string{
					"${keycloak_role.realm_role_one.id}",
					"${keycloak_role.openid_client_role_one.id}",
				}, allRoleIdSet2),
				Check: testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles1", false),
			},
			// add all back
			{
				Config: testKeycloakGroupRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, allRoleIdSet1, allRoleIdSet2),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles1", false),
			},
			// random scenario 1
			{
				Config: testKeycloakGroupRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, randomStringSliceSubset(allRoleIdSet1), randomStringSliceSubset(allRoleIdSet2)),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles1", false),
			},
			// remove all
			{
				Config: testKeycloakGroupRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, []string{}, []string{}),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles1", false),
			},
		},
	})
}

func TestAccKeycloakGroupRoles_simultaneousRoleAndAssignmentUpdate(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakGroupRoles_simultaneousRoleAndAssignmentUpdate(groupName, 1),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
			{
				Config: testKeycloakGroupRoles_simultaneousRoleAndAssignmentUpdate(groupName, 2),
				Check:  testAccCheckKeycloakGroupHasRoles("keycloak_group_roles.group_roles", true),
			},
		},
	})
}

func testAccCheckKeycloakGroupHasRoles(resourceName string, exhaustive bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realm := rs.Primary.Attributes["realm_id"]
		groupId := rs.Primary.Attributes["group_id"]

		var roles []*keycloak.Role
		for k, v := range rs.Primary.Attributes {
			if match, _ := regexp.MatchString("role_ids\\.[^#]+", k); !match {
				continue
			}

			role, err := keycloakClient.GetRole(testCtx, realm, v)
			if err != nil {
				return err
			}

			roles = append(roles, role)
		}

		group, err := keycloakClient.GetGroup(testCtx, realm, groupId)
		if err != nil {
			return err
		}

		groupRoleMappings, err := keycloakClient.GetGroupRoleMappings(testCtx, realm, groupId)
		if err != nil {
			return err
		}

		groupRoles, err := flattenRoleMapping(groupRoleMappings)
		if err != nil {
			return err
		}

		if exhaustive {
			if len(groupRoles) != len(roles) {
				return fmt.Errorf("expected number of group roles to be %d, got %d", len(roles), len(groupRoles))
			}
		} else {
			if len(groupRoles) < len(roles) {
				return fmt.Errorf("expected number of group roles to be greater than %d, got %d", len(roles), len(groupRoles))
			}
		}

		for _, role := range roles {
			var expectedRoleString string
			if role.ClientRole {
				expectedRoleString = fmt.Sprintf("%s/%s", role.ClientId, role.Name)
			} else {
				expectedRoleString = role.Name
			}

			found := false

			for _, groupRole := range groupRoles {
				if groupRole == expectedRoleString {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("expected to find role %s assigned to group %s", expectedRoleString, group.Name)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakGroupHasNoRoles(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realm := rs.Primary.Attributes["realm_id"]
		id := rs.Primary.ID

		group, err := keycloakClient.GetGroup(testCtx, realm, id)
		if err != nil {
			return err
		}

		if len(group.RealmRoles) != 0 || len(group.ClientRoles) != 0 {
			return fmt.Errorf("expected group %s to have no roles", group.Name)
		}

		return nil
	}
}

func testKeycloakGroupRoles_basic(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, groupName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
}

resource "keycloak_saml_client" "saml_client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id
}

resource "keycloak_role" "realm_role" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_role" "openid_client_role" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.openid_client.id
}

resource "keycloak_role" "saml_client_role" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_saml_client.saml_client.id
}

data "keycloak_role" "offline_access" {
	realm_id  = data.keycloak_realm.realm.id
	name      = "offline_access"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s"
}

resource "keycloak_group_roles" "group_roles" {
	realm_id = data.keycloak_realm.realm.id
	group_id = keycloak_group.group.id

	role_ids = [
		keycloak_role.realm_role.id,
		keycloak_role.openid_client_role.id,
		keycloak_role.saml_client_role.id,
		data.keycloak_role.offline_access.id,
	]
}
	`, testAccRealm.Realm, openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, groupName)
}

func testKeycloakGroupRoles_nonExhaustive(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, groupName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
}

resource "keycloak_saml_client" "saml_client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id
}

resource "keycloak_role" "realm_role" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_role" "openid_client_role" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.openid_client.id
}

resource "keycloak_role" "saml_client_role" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_saml_client.saml_client.id
}

data "keycloak_role" "offline_access" {
	realm_id  = data.keycloak_realm.realm.id
	name      = "offline_access"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s"
}

resource "keycloak_group_roles" "group_roles1" {
	realm_id = data.keycloak_realm.realm.id
	group_id = keycloak_group.group.id
	exhaustive = false

	role_ids = [
		keycloak_role.realm_role.id,
		keycloak_role.openid_client_role.id,
	]
}
	`, testAccRealm.Realm, openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, groupName)
}

func testKeycloakGroupRoles_noGroupRoles(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, groupName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
}

resource "keycloak_saml_client" "saml_client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id
}

resource "keycloak_role" "realm_role" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_role" "openid_client_role" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.openid_client.id
}

resource "keycloak_role" "saml_client_role" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_saml_client.saml_client.id
}

data "keycloak_role" "offline_access" {
	realm_id  = data.keycloak_realm.realm.id
	name      = "offline_access"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s"
}
	`, testAccRealm.Realm, openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, groupName)
}

func testKeycloakGroupRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName string, roleIds []string) string {
	tfRoleIds := fmt.Sprintf("role_ids = %s", arrayOfStringsForTerraformResource(roleIds))

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
}

resource "keycloak_saml_client" "saml_client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id
}

resource "keycloak_role" "realm_role_one" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_role" "realm_role_two" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_role" "openid_client_role_one" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.openid_client.id
}

resource "keycloak_role" "openid_client_role_two" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.openid_client.id
}

resource "keycloak_role" "saml_client_role_one" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_saml_client.saml_client.id
}

resource "keycloak_role" "saml_client_role_two" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_saml_client.saml_client.id
}

data "keycloak_role" "offline_access" {
	realm_id  = data.keycloak_realm.realm.id
	name      = "offline_access"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s"
}

resource "keycloak_group_roles" "group_roles" {
	realm_id = data.keycloak_realm.realm.id
	group_id = keycloak_group.group.id

	%s
}
	`, testAccRealm.Realm, openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, tfRoleIds)
}

func testKeycloakGroupRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName string, roleIds1, roleIds2 []string) string {
	tfRoleIds1 := fmt.Sprintf("role_ids = %s", arrayOfStringsForTerraformResource(roleIds1))
	tfRoleIds2 := fmt.Sprintf("role_ids = %s", arrayOfStringsForTerraformResource(roleIds2))

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "openid_client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
}

resource "keycloak_saml_client" "saml_client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id
}

resource "keycloak_role" "realm_role_one" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_role" "realm_role_two" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_role" "openid_client_role_one" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.openid_client.id
}

resource "keycloak_role" "openid_client_role_two" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.openid_client.id
}

resource "keycloak_role" "saml_client_role_one" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_saml_client.saml_client.id
}

resource "keycloak_role" "saml_client_role_two" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_saml_client.saml_client.id
}

data "keycloak_role" "offline_access" {
	realm_id  = data.keycloak_realm.realm.id
	name      = "offline_access"
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name = "%s"
}

resource "keycloak_group_roles" "group_roles1" {
	realm_id   = data.keycloak_realm.realm.id
	group_id   = keycloak_group.group.id
	exhaustive = false

	%s
}

resource "keycloak_group_roles" "group_roles2" {
	realm_id   = data.keycloak_realm.realm.id
	group_id   = keycloak_group.group.id
	exhaustive = false

	%s
}
	`, testAccRealm.Realm, openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, groupName, tfRoleIds1, tfRoleIds2)
}

func testKeycloakGroupRoles_simultaneousRoleAndAssignmentUpdate(groupName string, id int) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_role" "realm_role_%[2]d" {
	name     = "role-%[2]d"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_group" "group" {
	realm_id = data.keycloak_realm.realm.id
	name     = "%s"
}

resource "keycloak_group_roles" "group_roles" {
	realm_id = data.keycloak_realm.realm.id
	group_id = keycloak_group.group.id

	role_ids = [
		keycloak_role.realm_role_%[2]d.id
	]
}
`, testAccRealm.Realm, id, groupName)
}
