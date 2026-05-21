package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKeycloakUserPassword_basic(t *testing.T) {
	username := acctest.RandomWithPrefix("tf-acc")
	password := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_user_password.user_password"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUserPassword_basic(username, password, clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserPasswordExists(resourceName),
					testAccCheckKeycloakUserInitialPasswordLogin(username, password, clientId),
					resource.TestCheckResourceAttr(resourceName, "temporary", "false"),
				),
			},
		},
	})
}

func TestAccKeycloakUserPassword_update(t *testing.T) {
	username := acctest.RandomWithPrefix("tf-acc")
	passwordOne := acctest.RandomWithPrefix("tf-acc")
	passwordTwo := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUserPassword_basic(username, passwordOne, clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserInitialPasswordLogin(username, passwordOne, clientId),
				),
			},
			{
				Config: testKeycloakUserPassword_basic(username, passwordTwo, clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserInitialPasswordLogin(username, passwordTwo, clientId),
				),
			},
		},
	})
}

func TestAccKeycloakUserPassword_temporary(t *testing.T) {
	username := acctest.RandomWithPrefix("tf-acc")
	password := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_user_password.user_password"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUserPassword_temporary(username, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserPasswordExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "temporary", "true"),
				),
			},
		},
	})
}

func TestAccKeycloakUserPassword_import(t *testing.T) {
	username := acctest.RandomWithPrefix("tf-acc")
	password := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_user_password.user_password"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUserPassword_basic(username, password, clientId),
				Check:  testAccCheckKeycloakUserPasswordExists(resourceName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value", "value_hash", "temporary"},
				ImportStateIdFunc:       testAccKeycloakUserPasswordImportId(resourceName),
			},
		},
	})
}

func testAccCheckKeycloakUserPasswordExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		realmId := rs.Primary.Attributes["realm_id"]
		userId := rs.Primary.Attributes["user_id"]

		if _, err := keycloakClient.GetUser(testCtx, realmId, userId); err != nil {
			return fmt.Errorf("error fetching user %s in realm %s: %s", userId, realmId, err)
		}

		return nil
	}
}

func testAccKeycloakUserPasswordImportId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		realmId := rs.Primary.Attributes["realm_id"]
		userId := rs.Primary.Attributes["user_id"]

		return fmt.Sprintf("%s/%s", realmId, userId), nil
	}
}

func testKeycloakUserPassword_basic(username, password, clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	realm_id                     = data.keycloak_realm.realm.id
	client_id                    = "%s"
	name                         = "test client"
	enabled                      = true
	access_type                  = "PUBLIC"
	direct_access_grants_enabled = true
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_user_password" "user_password" {
	realm_id  = data.keycloak_realm.realm.id
	user_id   = keycloak_user.user.id
	value     = "%s"
	temporary = false
}
`, testAccRealm.Realm, clientId, username, password)
}

func testKeycloakUserPassword_temporary(username, password string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_user_password" "user_password" {
	realm_id  = data.keycloak_realm.realm.id
	user_id   = keycloak_user.user.id
	value     = "%s"
	temporary = true
}
`, testAccRealm.Realm, username, password)
}
