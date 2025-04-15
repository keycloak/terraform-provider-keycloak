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

func TestAccKeycloakUserRoles_basic(t *testing.T) {
	t.Parallel()

	realmRoleName := acctest.RandomWithPrefix("tf-acc")
	openIdClientName := acctest.RandomWithPrefix("tf-acc")
	openIdRoleName := acctest.RandomWithPrefix("tf-acc")
	samlClientName := acctest.RandomWithPrefix("tf-acc")
	samlRoleName := acctest.RandomWithPrefix("tf-acc")
	userName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUserRoles_basic(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, userName),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles", true),
			},
			{
				ResourceName:      "keycloak_user_roles.user_roles",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// check destroy
			{
				Config: testKeycloakUserRoles_noUserRoles(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, userName),
				Check:  testAccCheckKeycloakUserHasNoRoles("keycloak_user.user"),
			},
		},
	})
}

func TestAccKeycloakUserRoles_update(t *testing.T) {
	t.Parallel()

	realmRoleOneName := acctest.RandomWithPrefix("tf-acc")
	realmRoleTwoName := acctest.RandomWithPrefix("tf-acc")
	openIdClientName := acctest.RandomWithPrefix("tf-acc")
	openIdRoleOneName := acctest.RandomWithPrefix("tf-acc")
	openIdRoleTwoName := acctest.RandomWithPrefix("tf-acc")
	samlClientName := acctest.RandomWithPrefix("tf-acc")
	samlRoleOneName := acctest.RandomWithPrefix("tf-acc")
	samlRoleTwoName := acctest.RandomWithPrefix("tf-acc")
	userName := acctest.RandomWithPrefix("tf-acc")

	allRoleIds := []string{
		"${keycloak_role.realm_role_one.id}",
		"${keycloak_role.realm_role_two.id}",
		"${keycloak_role.openid_client_role_one.id}",
		"${keycloak_role.openid_client_role_two.id}",
		"${keycloak_role.saml_client_role_one.id}",
		"${keycloak_role.saml_client_role_two.id}",
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// initial setup, resource is defined but no roles are specified
			{
				Config: testKeycloakUserRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, []string{}),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles", true),
			},
			// add all roles
			{
				Config: testKeycloakUserRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, allRoleIds),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles", true),
			},
			// remove some
			{
				Config: testKeycloakUserRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, []string{
					"${keycloak_role.realm_role_two.id}",
					"${keycloak_role.openid_client_role_one.id}",
					"${keycloak_role.openid_client_role_two.id}",
				}),
				Check: testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles", true),
			},
			// add some and remove some
			{
				Config: testKeycloakUserRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, []string{
					"${keycloak_role.saml_client_role_one.id}",
					"${keycloak_role.saml_client_role_two.id}",
					"${keycloak_role.realm_role_one.id}",
				}),
				Check: testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles", true),
			},
			// add some and remove some again
			{
				Config: testKeycloakUserRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, []string{
					"${keycloak_role.saml_client_role_one.id}",
					"${keycloak_role.openid_client_role_two.id}",
					"${keycloak_role.realm_role_two.id}",
				}),
				Check: testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles", true),
			},
			// add all back
			{
				Config: testKeycloakUserRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, allRoleIds),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles", true),
			},
			// random scenario 1
			{
				Config: testKeycloakUserRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, randomStringSliceSubset(allRoleIds)),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles", true),
			},
			// random scenario 2
			{
				Config: testKeycloakUserRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, randomStringSliceSubset(allRoleIds)),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles", true),
			},
			// random scenario 3
			{
				Config: testKeycloakUserRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, randomStringSliceSubset(allRoleIds)),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles", true),
			},
			// remove all
			{
				Config: testKeycloakUserRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, []string{}),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles", true),
			},
		},
	})
}

func TestAccKeycloakUserRoles_updateNonExhaustive(t *testing.T) {
	t.Parallel()

	realmRoleOneName := acctest.RandomWithPrefix("tf-acc")
	realmRoleTwoName := acctest.RandomWithPrefix("tf-acc")
	openIdClientName := acctest.RandomWithPrefix("tf-acc")
	openIdRoleOneName := acctest.RandomWithPrefix("tf-acc")
	openIdRoleTwoName := acctest.RandomWithPrefix("tf-acc")
	samlClientName := acctest.RandomWithPrefix("tf-acc")
	samlRoleOneName := acctest.RandomWithPrefix("tf-acc")
	samlRoleTwoName := acctest.RandomWithPrefix("tf-acc")
	userName := acctest.RandomWithPrefix("tf-acc")

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
				Config: testKeycloakUserRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, []string{}, []string{}),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles1", false),
			},
			// add all roles
			{
				Config: testKeycloakUserRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, allRoleIdSet1, allRoleIdSet2),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles1", false),
			},
			// remove some
			{
				Config: testKeycloakUserRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, []string{
					"${keycloak_role.openid_client_role_one.id}",
				}, allRoleIdSet2),
				Check: testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles1", false),
			},
			// add some and remove some
			{
				Config: testKeycloakUserRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, []string{
					"${keycloak_role.saml_client_role_one.id}",
				}, allRoleIdSet2),
				Check: testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles1", false),
			},
			// add some and remove some again
			{
				Config: testKeycloakUserRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, []string{
					"${keycloak_role.realm_role_one.id}",
					"${keycloak_role.openid_client_role_one.id}",
				}, allRoleIdSet2),
				Check: testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles1", false),
			},
			// add all back
			{
				Config: testKeycloakUserRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, allRoleIdSet1, allRoleIdSet2),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles1", false),
			},
			// random scenario 1
			{
				Config: testKeycloakUserRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, randomStringSliceSubset(allRoleIdSet1), randomStringSliceSubset(allRoleIdSet2)),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles1", false),
			},
			// remove all
			{
				Config: testKeycloakUserRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, []string{}, []string{}),
				Check:  testAccCheckKeycloakUserHasRoles("keycloak_user_roles.user_roles1", false),
			},
		},
	})
}

func flattenRoleMapping(roleMapping *keycloak.RoleMapping) ([]string, error) {
	var roles []string

	for _, realmRole := range roleMapping.RealmMappings {
		roles = append(roles, realmRole.Name)
	}

	for _, clientRoleMapping := range roleMapping.ClientMappings {
		for _, clientRole := range clientRoleMapping.Mappings {
			roles = append(roles, fmt.Sprintf("%s/%s", clientRoleMapping.Id, clientRole.Name))
		}
	}

	return roles, nil
}

func testAccCheckKeycloakUserHasRoles(resourceName string, exhaustive bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realm := rs.Primary.Attributes["realm_id"]
		userId := rs.Primary.Attributes["user_id"]

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

		user, err := keycloakClient.GetUser(testCtx, realm, userId)
		if err != nil {
			return err
		}

		userRoleMappings, err := keycloakClient.GetUserRoleMappings(testCtx, realm, userId)
		if err != nil {
			return err
		}

		userRoles, err := flattenRoleMapping(userRoleMappings)
		if err != nil {
			return err
		}

		if exhaustive {
			if len(userRoles) != len(roles) {
				return fmt.Errorf("expected number of user roles to be %d, got %d", len(roles), len(userRoles))
			}
		} else {
			if len(userRoles) < len(roles) {
				return fmt.Errorf("expected number of user roles to be greater than %d, got %d", len(roles), len(userRoles))
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

			for _, userRole := range userRoles {
				if userRole == expectedRoleString {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("expected to find role %s assigned to user %s", expectedRoleString, user.Username)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakUserHasNoRoles(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realm := rs.Primary.Attributes["realm_id"]
		id := rs.Primary.ID

		user, err := keycloakClient.GetUser(testCtx, realm, id)
		if err != nil {
			return err
		}

		userRoleMapping, err := keycloakClient.GetUserRoleMappings(testCtx, realm, id)
		if err != nil {
			return err
		}

		if len(userRoleMapping.RealmMappings) != 0 || len(userRoleMapping.ClientMappings) != 0 {
			return fmt.Errorf("expected user %s to have no roles", user.Username)
		}

		return nil
	}
}

func testKeycloakUserRoles_basic(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, userName string) string {
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

data "keycloak_openid_client" "account" {
	realm_id = data.keycloak_realm.realm.id
	client_id = "account"
}

data "keycloak_role" "view_consent" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = data.keycloak_openid_client.account.id
	name 	  = "view-consent"
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_user_roles" "user_roles" {
	realm_id = data.keycloak_realm.realm.id
	user_id = keycloak_user.user.id

	role_ids = [
		keycloak_role.realm_role.id,
		keycloak_role.openid_client_role.id,
		keycloak_role.saml_client_role.id,

		data.keycloak_role.view_consent.id,
	]
}
	`, testAccRealm.Realm, openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, userName)
}

func testKeycloakUserRoles_noUserRoles(openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, userName string) string {
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

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}
	`, testAccRealm.Realm, openIdClientName, samlClientName, realmRoleName, openIdRoleName, samlRoleName, userName)
}

func testKeycloakUserRoles_update(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName string, roleIds []string) string {
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

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_user_roles" "user_roles" {
	realm_id = data.keycloak_realm.realm.id
	user_id = keycloak_user.user.id

	%s
}
	`, testAccRealm.Realm, openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, tfRoleIds)
}

func testKeycloakUserRoles_updateNonExhaustive(openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName string, roleIds1, roleIds2 []string) string {
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

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_user_roles" "user_roles1" {
	realm_id   = data.keycloak_realm.realm.id
	user_id    = keycloak_user.user.id
	exhaustive = false

	%s
}

resource "keycloak_user_roles" "user_roles2" {
	realm_id   = data.keycloak_realm.realm.id
	user_id    = keycloak_user.user.id
	exhaustive = false

	%s
}
	`, testAccRealm.Realm, openIdClientName, samlClientName, realmRoleOneName, realmRoleTwoName, openIdRoleOneName, openIdRoleTwoName, samlRoleOneName, samlRoleTwoName, userName, tfRoleIds1, tfRoleIds2)
}
