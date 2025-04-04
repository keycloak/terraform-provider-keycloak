package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakDefaultRoles_basic(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakDefaultRoles_basic(realmName),
				Check:  testAccCheckDefaultRolesExists("keycloak_default_roles.default_roles"),
			},
			{
				Config: testKeycloakDefaultRoles_destroy(realmName),
				Check:  testAccCheckKeycloakDefaultRolesDestroy(realmName),
			},
		},
	})
}

func TestAccKeycloakDefaultRoles_updateDefaultRoles(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")

	groupDefaultRolesOne := &keycloak.DefaultRoles{
		RealmId:      testAccRealmUserFederation.Realm,
		DefaultRoles: []string{"\"uma_authorization\""},
	}

	groupDefaultRolesTwo := &keycloak.DefaultRoles{
		RealmId:      testAccRealmUserFederation.Realm,
		DefaultRoles: []string{"\"uma_authorization\",", "\"offline_access\""},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakDefaultRoles_basicFromInterface(realmName, groupDefaultRolesOne),
				Check:  testAccCheckDefaultRolesExists("keycloak_default_roles.default_roles"),
			},
			{
				Config: testKeycloakDefaultRoles_basicFromInterface(realmName, groupDefaultRolesTwo),
				Check:  testAccCheckDefaultRolesExists("keycloak_default_roles.default_roles"),
			},
			{
				Config: testKeycloakDefaultRoles_destroy(realmName),
				Check:  testAccCheckKeycloakDefaultRolesDestroy(realmName),
			},
		},
	})
}

func testAccCheckDefaultRolesExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakDefaultRolesFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakDefaultRolesDestroy(realmId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := keycloakClient.GetRealm(testCtx, realmId)
		if err != nil {
			return err
		}

		composites, err := keycloakClient.GetDefaultRoles(testCtx, realmId, realm.DefaultRole.Id)
		if err != nil {
			return fmt.Errorf("error getting defaultRoles with id %s: %s", realm.DefaultRole.Id, err)
		}

		if len(composites) != 0 {
			return fmt.Errorf("realm %s still has %d default roles, expected zero", realmId, len(composites))
		}

		return nil
	}
}

func getKeycloakDefaultRolesFromState(s *terraform.State, resourceName string) (*keycloak.DefaultRoles, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	composites, err := keycloakClient.GetDefaultRoles(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting defaultRoles with id %s: %s", id, err)
	}

	var defaultRoleNamesList []string
	for _, composite := range composites {
		name, err := keycloakClient.GetQualifiedRoleName(testCtx, realm, composite)
		if err != nil {
			return nil, fmt.Errorf("error getting qualified name for role id %s: %s", composite.Id, err)
		}
		defaultRoleNamesList = append(defaultRoleNamesList, name)
	}

	defaultRoles := &keycloak.DefaultRoles{
		Id:           id,
		RealmId:      realm,
		DefaultRoles: defaultRoleNamesList,
	}

	return defaultRoles, nil
}

func testKeycloakDefaultRoles_basic(realmName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm   = "%s"
	enabled = true
}

resource "keycloak_default_roles" "default_roles" {
	realm_id      = keycloak_realm.realm.id
	default_roles = ["uma_authorization"]
}
	`, realmName)
}

func testKeycloakDefaultRoles_destroy(realmName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm   = "%s"
	enabled = true
}
	`, realmName)
}

func testKeycloakDefaultRoles_basicFromInterface(realmName string, defaultRoles *keycloak.DefaultRoles) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm   = "%s"
	enabled = true
}

resource "keycloak_default_roles" "default_roles" {
	realm_id  = keycloak_realm.realm.id
	default_roles = %s
}
	`, realmName, defaultRoles.DefaultRoles)
}
