package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"strings"
	"testing"
)

func TestAccKeycloakRole_basicRealm(t *testing.T) {
	t.Parallel()
	roleName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRole_basicRealm(roleName),
				Check:  testAccCheckKeycloakRoleExists("keycloak_role.role"),
			},
			{
				ResourceName:        "keycloak_role.role",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakRole_basicRealmUrlRoleName(t *testing.T) {
	t.Parallel()
	roleName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRole_basicRealm(roleName),
				Check:  testAccCheckKeycloakRoleExists("keycloak_role.role"),
			},
			{
				ResourceName:        "keycloak_role.role",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakRole_basicClient(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	roleName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRole_basicClient(clientId, roleName),
				Check:  testAccCheckKeycloakRoleExists("keycloak_role.role"),
			},
			{
				ResourceName:        "keycloak_role.role",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakRole_basicSamlClient(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	roleName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRole_basicSamlClient(clientId, roleName),
				Check:  testAccCheckKeycloakRoleExists("keycloak_role.role"),
			},
			{
				ResourceName:        "keycloak_role.role",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakRole_basicRealmUpdate(t *testing.T) {
	t.Parallel()
	roleName := acctest.RandomWithPrefix("tf-acc")
	descriptionOne := acctest.RandomWithPrefix("tf-acc")
	descriptionTwo := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRole_basicRealmWithDescription(roleName, descriptionOne),
				Check:  testAccCheckKeycloakRoleExists("keycloak_role.role"),
			},
			{
				Config: testKeycloakRole_basicRealmWithDescription(roleName, descriptionTwo),
				Check:  testAccCheckKeycloakRoleExists("keycloak_role.role"),
			},
			{
				Config: testKeycloakRole_basicRealm(roleName),
				Check:  testAccCheckKeycloakRoleExists("keycloak_role.role"),
			},
		},
	})
}

func TestAccKeycloakRole_basicClientUpdate(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	roleName := acctest.RandomWithPrefix("tf-acc")
	descriptionOne := acctest.RandomWithPrefix("tf-acc")
	descriptionTwo := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRole_basicClientWithDescription(clientId, roleName, descriptionOne),
				Check:  testAccCheckKeycloakRoleExists("keycloak_role.role"),
			},
			{
				Config: testKeycloakRole_basicClientWithDescription(clientId, roleName, descriptionTwo),
				Check:  testAccCheckKeycloakRoleExists("keycloak_role.role"),
			},
			{
				Config: testKeycloakRole_basicClient(clientId, roleName),
				Check:  testAccCheckKeycloakRoleExists("keycloak_role.role"),
			},
		},
	})
}

func TestAccKeycloakRole_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var role = &keycloak.Role{}

	roleName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRole_basicRealm(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRoleExists("keycloak_role.role"),
					testAccCheckKeycloakRoleFetch("keycloak_role.role", role),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRole(testCtx, role.RealmId, role.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRole_basicRealm(roleName),
				Check:  testAccCheckKeycloakRoleExists("keycloak_role.role"),
			},
		},
	})
}

func TestAccKeycloakRole_composites(t *testing.T) {
	t.Parallel()
	clientOne := acctest.RandomWithPrefix("tf-acc")
	clientTwo := acctest.RandomWithPrefix("tf-acc")
	roleOne := acctest.RandomWithPrefix("tf-acc")
	roleTwo := acctest.RandomWithPrefix("tf-acc")
	roleThree := acctest.RandomWithPrefix("tf-acc")
	roleFour := acctest.RandomWithPrefix("tf-acc")
	roleWithComposites := acctest.RandomWithPrefix("tf-acc")
	roleWithCompositesResourceName := "keycloak_role.role_with_composites"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		Steps: []resource.TestStep{
			// initial setup - no composites attached
			{
				Config: testKeycloakRole_composites(clientOne, clientTwo, roleOne, roleTwo, roleThree, roleFour, roleWithComposites, []string{}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRoleExists("keycloak_role.role_1"),
					testAccCheckKeycloakRoleExists("keycloak_role.role_2"),
					testAccCheckKeycloakRoleExists("keycloak_role.role_3"),
					testAccCheckKeycloakRoleExists("keycloak_role.role_with_composites"),
					testAccCheckKeycloakRoleHasComposites(roleWithCompositesResourceName, []string{}),
				),
			},
			// add all composites
			{
				Config: testKeycloakRole_composites(clientOne, clientTwo, roleOne, roleTwo, roleThree, roleFour, roleWithComposites, []string{
					"${keycloak_role.role_1.id}",
					"${keycloak_role.role_2.id}",
					"${keycloak_role.role_3.id}",
					"${keycloak_role.role_4.id}",
				}),
				Check: testAccCheckKeycloakRoleHasComposites(roleWithCompositesResourceName, []string{
					roleOne,
					roleTwo,
					roleThree,
					roleFour,
				}),
			},
			// remove two composites
			{
				Config: testKeycloakRole_composites(clientOne, clientTwo, roleOne, roleTwo, roleThree, roleFour, roleWithComposites, []string{
					"${keycloak_role.role_1.id}",
					"${keycloak_role.role_2.id}",
				}),
				Check: testAccCheckKeycloakRoleHasComposites(roleWithCompositesResourceName, []string{
					roleOne,
					roleTwo,
				}),
			},
			// add them back and remove the others
			{
				Config: testKeycloakRole_composites(clientOne, clientTwo, roleOne, roleTwo, roleThree, roleFour, roleWithComposites, []string{
					"${keycloak_role.role_3.id}",
					"${keycloak_role.role_4.id}",
				}),
				Check: testAccCheckKeycloakRoleHasComposites(roleWithCompositesResourceName, []string{
					roleThree,
					roleFour,
				}),
			},
			// remove them all
			{
				Config: testKeycloakRole_composites(clientOne, clientTwo, roleOne, roleTwo, roleThree, roleFour, roleWithComposites, []string{}),
				Check:  testAccCheckKeycloakRoleHasComposites(roleWithCompositesResourceName, []string{}),
			},
		},
	})
}

func TestAccKeycloakRole_basicWithAttributes(t *testing.T) {
	t.Parallel()
	roleName := acctest.RandomWithPrefix("tf-acc")
	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRole_basicWithAttributes(roleName, attributeName, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRoleExists("keycloak_role.role"),
					testAccCheckKeycloakRoleHasAttribute("keycloak_role.role", attributeName, attributeValue),
				),
			},
			{
				ResourceName:        "keycloak_role.role",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func testAccCheckKeycloakRoleExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getRoleFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakRoleDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_role" || strings.HasPrefix(name, "data") {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			role, _ := keycloakClient.GetRole(testCtx, realm, id)
			if role != nil {
				return fmt.Errorf("%s with id %s still exists", name, id)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakRoleFetch(resourceName string, role *keycloak.Role) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		fetchedRole, err := getRoleFromState(state, resourceName)
		if err != nil {
			return err
		}

		role.Id = fetchedRole.Id
		role.Name = fetchedRole.Name
		role.RealmId = fetchedRole.RealmId
		role.ClientId = fetchedRole.ClientId

		return nil
	}
}

func testAccCheckKeycloakRoleHasAttribute(resourceName, attributeName, attributeValue string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		role, err := getRoleFromState(state, resourceName)
		if err != nil {
			return err
		}

		if len(role.Attributes) != 1 || role.Attributes[attributeName][0] != attributeValue {
			return fmt.Errorf("expected role %s to have attribute %s with value %s", role.Name, attributeName, attributeValue)
		}

		return nil
	}
}

func testAccCheckKeycloakRoleHasComposites(resourceName string, compositeRoleNames []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		role, err := getRoleFromState(state, resourceName)
		if err != nil {
			return err
		}

		if len(compositeRoleNames) != 0 && !role.Composite {
			return fmt.Errorf("expected role %s to have composites, but has none", role.Name)
		}

		if len(compositeRoleNames) == 0 && role.Composite {
			return fmt.Errorf("expected role %s to have no composites, but has some", role.Name)
		}

		composites, err := keycloakClient.GetRoleComposites(testCtx, role)
		if err != nil {
			return err
		}

		for _, compositeRoleName := range compositeRoleNames {
			var found bool

			for _, composite := range composites {
				if composite.Name == compositeRoleName {
					found = true
				}
			}

			if !found {
				return fmt.Errorf("expected role %s to have composite %s", role.Name, compositeRoleName)
			}
		}

		for _, composite := range composites {
			var found bool

			for _, compositeRoleName := range compositeRoleNames {
				if composite.Name == compositeRoleName {
					found = true
				}
			}

			if !found {
				return fmt.Errorf("role %s had unexpected composite %s", role.Name, composite.Name)
			}
		}

		return nil
	}
}

func getRoleFromState(s *terraform.State, resourceName string) (*keycloak.Role, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	role, err := keycloakClient.GetRole(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting role with id %s: %s", id, err)
	}

	return role, nil
}

func testKeycloakRole_basicRealm(role string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_role" "role" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}
	`, testAccRealm.Realm, role)
}

func testKeycloakRole_basicRealmWithDescription(role, description string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_role" "role" {
	name        = "%s"
	description = "%s"
	realm_id    = data.keycloak_realm.realm.id
}
	`, testAccRealm.Realm, role, description)
}

func testKeycloakRole_basicClient(clientId, role string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
}

resource "keycloak_role" "role" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.client.id
}
	`, testAccRealm.Realm, clientId, role)
}

func testKeycloakRole_basicSamlClient(clientId, role string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "client" {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id
}

resource "keycloak_role" "role" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_saml_client.client.id
}
	`, testAccRealm.Realm, clientId, role)
}

func testKeycloakRole_basicClientWithDescription(clientId, role, description string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
}

resource "keycloak_role" "role" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id
	client_id   = keycloak_openid_client.client.id
	description = "%s"
}
	`, testAccRealm.Realm, clientId, role, description)
}

func testKeycloakRole_composites(clientOne, clientTwo, roleOne, roleTwo, roleThree, roleFour, roleWithComposites string, composites []string) string {
	var tfComposites string
	if len(composites) != 0 {
		tfComposites = fmt.Sprintf("composite_roles = %s", arrayOfStringsForTerraformResource(composites))
	}

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client_one" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
}

resource "keycloak_openid_client" "client_two" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
}

resource "keycloak_role" "role_1" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_role" "role_2" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.client_one.id
}

resource "keycloak_role" "role_3" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
}

resource "keycloak_role" "role_4" {
	name      = "%s"
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.client_two.id
}

resource "keycloak_role" "role_with_composites" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id

	%s
}
	`, testAccRealm.Realm, clientOne, clientTwo, roleOne, roleTwo, roleThree, roleFour, roleWithComposites, tfComposites)
}

func testKeycloakRole_basicWithAttributes(role, attributeName, attributeValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_role" "role" {
	name     = "%s"
	realm_id = data.keycloak_realm.realm.id
	attributes = {
		"%s" = "%s"
	}
}
	`, testAccRealm.Realm, role, attributeName, attributeValue)
}
