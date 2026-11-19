package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeycloakRealm_tokenSettingsCanBeReset(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	config := func(idle, maximum int, enabled bool) string {
		return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
 realm = %q
 enabled = true
 sso_session_idle_timeout_remember_me = "%ds"
 sso_session_max_lifespan_remember_me = "%ds"
 offline_session_max_lifespan_enabled = %t
}
`, realmName, idle, maximum, enabled)
	}
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{Config: config(123, 456, true)},
			{
				Config: config(0, 0, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("keycloak_realm.realm", "sso_session_idle_timeout_remember_me", "0s"),
					resource.TestCheckResourceAttr("keycloak_realm.realm", "sso_session_max_lifespan_remember_me", "0s"),
					resource.TestCheckResourceAttr("keycloak_realm.realm", "offline_session_max_lifespan_enabled", "false"),
				),
			},
		},
	})
}

func TestAccKeycloakRealm_organizationsCanBeDisabled(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	config := func(enabled bool) string {
		return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
 realm = %q
 enabled = true
 organizations_enabled = %t
}
`, realmName, enabled)
	}
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{Config: config(true)},
			{Config: config(false), Check: resource.TestCheckResourceAttr("keycloak_realm.realm", "organizations_enabled", "false")},
		},
	})
}

func TestAccKeycloakRealm_adminPermissionsCanBeDisabled(t *testing.T) {
	skipIfFGAPv2NotEnabled(testCtx, t, keycloakClient)
	realmName := acctest.RandomWithPrefix("tf-acc")
	config := func(enabled bool) string {
		return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
 realm = %q
 enabled = true
 admin_permissions_enabled = %t
}
`, realmName, enabled)
	}
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{Config: config(true)},
			{Config: config(false), Check: resource.TestCheckResourceAttr("keycloak_realm.realm", "admin_permissions_enabled", "false")},
		},
	})
}

func TestAccKeycloakRealm_maxTemporaryLockoutsCanBeReset(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	config := func(maximum int) string {
		return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
 realm = %q
 security_defenses {
  brute_force_detection {
   permanent_lockout = true
   max_temporary_lockouts = %d
  }
 }
}
`, realmName, maximum)
	}
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{Config: config(3), Check: resource.TestCheckResourceAttr("keycloak_realm.realm", "security_defenses.0.brute_force_detection.0.max_temporary_lockouts", "3")},
			{Config: config(0), Check: resource.TestCheckResourceAttr("keycloak_realm.realm", "security_defenses.0.brute_force_detection.0.max_temporary_lockouts", "0")},
			{Config: config(0), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}
