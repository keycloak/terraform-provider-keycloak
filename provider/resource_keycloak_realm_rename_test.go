package provider

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// Some Keycloak versions only show a renamed realm in full to tokens issued after the rename. Every
// terraform command starts with a new token, but the acceptance tests share one client, so the rename
// steps get a fresh token before the test framework refreshes the state.
func testAccRefreshClientToken(*terraform.State) error {
	return keycloakClient.Refresh(testCtx)
}

// Walks through the migration path: a group that identifies its realm by name is switched to
// realm_uuid, after which the realm is renamed. The group must be neither replaced nor drift.
func TestAccKeycloakRealm_renameWithRealmUuidChildren(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	renamedRealmName := acctest.RandomWithPrefix("tf-acc-renamed")

	var groupId, realmUuid string
	recordIds := func(s *terraform.State) error {
		groupId = s.RootModule().Resources["keycloak_group.group"].Primary.ID
		realmUuid = s.RootModule().Resources["keycloak_realm.realm"].Primary.Attributes["internal_id"]
		return nil
	}
	checkGroupAndRealmKept := func(expectedRealmName string) resource.TestCheckFunc {
		return resource.ComposeTestCheckFunc(
			func(s *terraform.State) error {
				if id := s.RootModule().Resources["keycloak_group.group"].Primary.ID; id != groupId {
					return fmt.Errorf("group was replaced: id %s -> %s", groupId, id)
				}
				if id := s.RootModule().Resources["keycloak_realm.realm"].Primary.Attributes["internal_id"]; id != realmUuid {
					return fmt.Errorf("realm internal id changed: %s -> %s", realmUuid, id)
				}
				return nil
			},
			resource.TestCheckResourceAttr("keycloak_realm.realm", "id", expectedRealmName),
			resource.TestCheckResourceAttrPair("keycloak_group.group", "realm_uuid", "keycloak_realm.realm", "internal_id"),
		)
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				// realm_uuid is filled in on create even when only realm_id is configured
				Config: testKeycloakRealmRename_byName(realmName),
				Check: resource.ComposeTestCheckFunc(
					recordIds,
					resource.TestCheckResourceAttrPair("keycloak_group.group", "realm_uuid", "keycloak_realm.realm", "internal_id"),
				),
			},
			{
				Config: testKeycloakRealmRename_byUuid(realmName),
				Check:  checkGroupAndRealmKept(realmName),
			},
			{
				Config: testKeycloakRealmRename_byUuid(renamedRealmName),
				Check:  resource.ComposeTestCheckFunc(testAccRefreshClientToken, checkGroupAndRealmKept(renamedRealmName)),
			},
			{
				// realm_id of the untouched group is only refreshed on the next read
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					checkGroupAndRealmKept(renamedRealmName),
					resource.TestCheckResourceAttr("keycloak_group.group", "realm_id", renamedRealmName),
				),
			},
			{
				// an out of band lookup must still find the group under the new name
				Config:       testKeycloakRealmRename_byUuid(renamedRealmName),
				ResourceName: "keycloak_group.group",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s", renamedRealmName, groupId), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func testKeycloakRealmRename_byName(realm string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	realm_id = keycloak_realm.realm.realm
	name     = "group"
}
`, realm)
}

func testKeycloakRealmRename_byUuid(realm string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_group" "group" {
	realm_uuid = keycloak_realm.realm.internal_id
	name       = "group"
}
`, realm)
}

// Renaming a realm must leave resources of every flavour untouched: plain ones, realm components
// (whose parentId is the realm's internal id), realm roles (containerId) and resources with
// an id that contains the realm name.
func TestAccKeycloakRealm_renameWithMixedChildren(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	renamedRealmName := acctest.RandomWithPrefix("tf-acc-renamed")

	// resources whose id does not contain the realm name, so it must not change on a rename
	stableResources := []string{
		"keycloak_group.group",
		"keycloak_role.role",
		"keycloak_openid_client.client",
		"keycloak_realm_keystore_aes_generated.aes",
		"keycloak_authentication_flow.flow",
		"keycloak_authentication_execution.execution",
		"keycloak_authentication_execution_config.config",
		"keycloak_oidc_identity_provider.oidc",
		"keycloak_hardcoded_role_identity_provider_mapper.mapper",
		"keycloak_organization.organization",
	}

	ids := map[string]string{}
	recordIds := func(s *terraform.State) error {
		for _, name := range stableResources {
			ids[name] = s.RootModule().Resources[name].Primary.ID
		}
		return nil
	}
	checkIds := func(s *terraform.State) error {
		for _, name := range stableResources {
			if id := s.RootModule().Resources[name].Primary.ID; id != ids[name] {
				return fmt.Errorf("%s was replaced: id %s -> %s", name, ids[name], id)
			}
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmRename_mixedChildren(realmName),
				Check:  recordIds,
			},
			{
				// data sources are read while planning, before the realm has been renamed
				Config: testKeycloakRealmRename_mixedChildren(renamedRealmName) + `
data "keycloak_group" "group" {
	realm_uuid = keycloak_realm.realm.internal_id
	name       = "group"
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccRefreshClientToken,
					checkIds,
					resource.TestCheckResourceAttr("keycloak_realm.realm", "id", renamedRealmName),
					resource.TestCheckResourceAttrPair("data.keycloak_group.group", "id", "keycloak_group.group", "id"),
				),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					checkIds,
					resource.TestCheckResourceAttr("keycloak_group.group", "realm_id", renamedRealmName),
					resource.TestCheckResourceAttr("keycloak_role.role", "realm_id", renamedRealmName),
					resource.TestCheckResourceAttr("keycloak_realm_keystore_aes_generated.aes", "realm_id", renamedRealmName),
					resource.TestCheckResourceAttr("keycloak_default_groups.defaults", "realm_id", renamedRealmName),
					// resources that are identified by their realm: the id is the realm name
					resource.TestCheckResourceAttr("keycloak_authentication_bindings.bindings", "id", renamedRealmName),
					resource.TestCheckResourceAttr("keycloak_authentication_bindings.bindings", "realm_id", renamedRealmName),
					resource.TestCheckResourceAttr("keycloak_realm_events.events", "id", renamedRealmName),
					// an id that contains the realm name follows the rename
					resource.TestCheckResourceAttr("keycloak_default_groups.defaults", "id", renamedRealmName+"/default-groups"),
					// these resources call the field "realm" instead of "realm_id"
					resource.TestCheckResourceAttr("keycloak_oidc_identity_provider.oidc", "realm", renamedRealmName),
					resource.TestCheckResourceAttr("keycloak_organization.organization", "realm", renamedRealmName),
				),
			},
			{
				// a follow-up change must still work on the renamed realm
				Config: testKeycloakRealmRename_mixedChildren(renamedRealmName) + `
data "keycloak_group" "group" {
	realm_uuid = keycloak_realm.realm.internal_id
	name       = "group"
}

resource "keycloak_group" "another" {
	realm_uuid = keycloak_realm.realm.internal_id
	name       = "another"
}
`,
				Check: checkIds,
			},
		},
	})
}

func testKeycloakRealmRename_mixedChildren(realm string) string {
	return testKeycloakRealmRename_mixedChildrenTemplate(realm, "realm_uuid = keycloak_realm.realm.internal_id", "realm_uuid = keycloak_realm.realm.internal_id")
}

// the same resources as written before realm_uuid existed
func testKeycloakRealmRename_mixedChildrenLegacy(realm string) string {
	return testKeycloakRealmRename_mixedChildrenTemplate(realm, "realm_id = keycloak_realm.realm.id", "realm = keycloak_realm.realm.id")
}

func testKeycloakRealmRename_mixedChildrenTemplate(realm, ref, realmRef string) string {
	template := fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm                 = "%s"
	organizations_enabled = true
}

resource "keycloak_group" "group" {
	{{REF}}
	name       = "group"
}

resource "keycloak_role" "role" {
	{{REF}}
	name       = "role"
}

resource "keycloak_openid_client" "client" {
	{{REF}}
	client_id   = "client"
	access_type = "CONFIDENTIAL"
}

resource "keycloak_realm_keystore_aes_generated" "aes" {
	{{REF}}
	name       = "aes"
	priority   = 100
}

resource "keycloak_default_groups" "defaults" {
	{{REF}}
	group_ids  = [keycloak_group.group.id]
}

resource "keycloak_authentication_bindings" "bindings" {
	{{REF}}
	browser_flow = "browser"
}

resource "keycloak_realm_events" "events" {
	{{REF}}
	events_enabled   = true
	events_listeners = ["jboss-logging"]
}

resource "keycloak_authentication_flow" "flow" {
	{{REF}}
	alias      = "flow"
}

resource "keycloak_authentication_execution" "execution" {
	{{REF}}
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	authenticator     = "identity-provider-redirector"
	requirement       = "REQUIRED"
}

resource "keycloak_oidc_identity_provider" "oidc" {
	{{REALM_REF}}
	alias             = "oidc"
	authorization_url = "https://example.com/auth"
	token_url         = "https://example.com/token"
	client_id         = "example_id"
	client_secret     = "example_token"
}

resource "keycloak_hardcoded_role_identity_provider_mapper" "mapper" {
	{{REALM_REF}}
	name                    = "mapper"
	identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
	role                    = keycloak_role.role.name
}

resource "keycloak_organization" "organization" {
	{{REALM_REF}}
	name       = "organization"

	domain {
		name     = "example.com"
		verified = true
	}
}

resource "keycloak_authentication_execution_config" "config" {
	{{REF}}
	execution_id = keycloak_authentication_execution.execution.id
	alias        = "config"
	config = {
		defaultProvider = "x"
	}
}
`, realm)

	return strings.NewReplacer("{{REF}}", ref, "{{REALM_REF}}", realmRef).Replace(template)
}

// Upgrading from a provider version without realm_uuid: nothing may change in the plan, switching
// to realm_uuid must not replace anything, and the realm can be renamed afterwards.
func TestAccKeycloakRealm_renameAfterUpgrade(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	renamedRealmName := acctest.RandomWithPrefix("tf-acc-renamed")

	var ids map[string]string
	recordIds := func(s *terraform.State) error {
		ids = map[string]string{}
		for name, rs := range s.RootModule().Resources {
			ids[name] = rs.Primary.ID
		}
		return nil
	}
	checkIds := func(s *terraform.State) error {
		for name, id := range ids {
			rs, ok := s.RootModule().Resources[name]
			if !ok {
				return fmt.Errorf("%s disappeared", name)
			}
			// ids that contain the realm name are not stable by design
			if rs.Primary.ID != id && !strings.Contains(id, realmName) {
				return fmt.Errorf("%s was replaced: id %s -> %s", name, id, rs.Primary.ID)
			}
		}
		return nil
	}
	expectEmptyPlan := resource.ConfigPlanChecks{
		PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"keycloak": {Source: "keycloak/keycloak", VersionConstraint: "5.10.0"},
				},
				Config: testKeycloakRealmRename_mixedChildrenLegacy(realmName),
				// the realm was created by another provider process, after the shared client got its token
				Check: resource.ComposeTestCheckFunc(recordIds, testAccRefreshClientToken),
			},
			{
				// the new provider reads the old state and must not want to change anything
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				Config:                   testKeycloakRealmRename_mixedChildrenLegacy(realmName),
				ConfigPlanChecks:         expectEmptyPlan,
				Check:                    checkIds,
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				Config:                   testKeycloakRealmRename_mixedChildren(realmName),
				ConfigPlanChecks:         expectEmptyPlan,
				Check:                    checkIds,
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				Config:                   testKeycloakRealmRename_mixedChildren(renamedRealmName),
				Check:                    resource.ComposeTestCheckFunc(testAccRefreshClientToken, checkIds),
			},
		},
	})
}

// The fine-grained admin permission resources have an id that starts with the realm name. They must
// keep working, and be deletable, after the realm was renamed.
func TestAccKeycloakRealm_renameWithAdminPermissions(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_2)
	skipIfFGAPv2NotEnabled(testCtx, t, keycloakClient)
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	renamedRealmName := acctest.RandomWithPrefix("tf-acc-renamed")

	permissions := []string{
		"keycloak_group_admin_permissions.group",
		"keycloak_users_admin_permissions.users",
		"keycloak_openid_client_admin_permissions.client",
		"keycloak_role_admin_permissions.role",
	}
	checkPermissions := func(s *terraform.State) error {
		for _, name := range permissions {
			rs, ok := s.RootModule().Resources[name]
			if !ok {
				return fmt.Errorf("%s was removed from state", name)
			}
			if rs.Primary.Attributes["permission_id"] == "" {
				return fmt.Errorf("%s lost its permission id", name)
			}
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmRename_adminPermissions(realmName),
				Check:  checkPermissions,
			},
			{
				Config: testKeycloakRealmRename_adminPermissions(renamedRealmName),
				Check:  resource.ComposeTestCheckFunc(testAccRefreshClientToken, checkPermissions),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					checkPermissions,
					resource.TestCheckResourceAttr("keycloak_group_admin_permissions.group", "realm_id", renamedRealmName),
					resource.TestCheckResourceAttr("keycloak_role_admin_permissions.role", "realm_id", renamedRealmName),
					resource.TestCheckResourceAttrWith("keycloak_group_admin_permissions.group", "id", func(id string) error {
						if !strings.HasPrefix(id, renamedRealmName+"/") {
							return fmt.Errorf("id %q does not start with the new realm name %q", id, renamedRealmName)
						}
						return nil
					}),
				),
			},
		},
	})
}

func testKeycloakRealmRename_adminPermissions(realm string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm                   = "%s"
	admin_permissions_enabled = true
}

resource "keycloak_group_admin_permissions" "group" {
	realm_uuid = keycloak_realm.realm.internal_id
	name       = "group-permission"
	scopes     = ["view"]
}

resource "keycloak_users_admin_permissions" "users" {
	realm_uuid = keycloak_realm.realm.internal_id
	name       = "users-permission"
	scopes     = ["view"]
}

resource "keycloak_openid_client_admin_permissions" "client" {
	realm_uuid = keycloak_realm.realm.internal_id
	name       = "client-permission"
	scopes     = ["view"]
}

resource "keycloak_role" "role" {
	realm_uuid = keycloak_realm.realm.internal_id
	name       = "role"
}

resource "keycloak_role_admin_permissions" "role" {
	realm_uuid = keycloak_realm.realm.internal_id
	name       = "role-permission"
	role_ids   = [keycloak_role.role.id]
	scopes     = ["map-role"]
}
`, realm)
}

// A renamed realm keeps its internal id, which is the old name for realms created by this provider, so a
// new realm with the old name is refused. The error has to say why.
func TestAccKeycloakRealm_reuseRenamedRealmName(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	renamedRealmName := acctest.RandomWithPrefix("tf-acc-renamed")

	realm := func(resourceName, name string) string {
		return fmt.Sprintf(`
resource "keycloak_realm" "%s" {
	realm = "%s"
}
`, resourceName, name)
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: realm("first", realmName),
			},
			{
				Config: realm("first", renamedRealmName),
				Check:  testAccRefreshClientToken,
			},
			{
				Config:      realm("first", renamedRealmName) + realm("second", realmName),
				ExpectError: regexp.MustCompile(`internal id is still`),
			},
			{
				// with an internal id of its own the old name can be used again
				Config: realm("first", renamedRealmName) + fmt.Sprintf(`
resource "keycloak_realm" "second" {
	realm       = "%s"
	internal_id = "%s-second"
}
`, realmName, realmName),
			},
		},
	})
}

// Keycloak 26.0 does not show a realm to a token that was issued before the realm was created. A realm uuid
// that cannot be found is retried with a new token.
func TestAccKeycloakRealm_uuidOfRealmCreatedOutsideTerraform(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	realmUuid := acctest.RandomWithPrefix("uuid")

	// another client, so the client used by the provider keeps the token it got before the realm existed
	otherClient, err := keycloak.NewKeycloakClient(testCtx, os.Getenv("KEYCLOAK_URL"), "", os.Getenv("KEYCLOAK_ADMIN_URL"), os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"), os.Getenv("KEYCLOAK_USER"), os.Getenv("KEYCLOAK_PASSWORD"), "", "", "", "", "", true, 120, os.Getenv("KEYCLOAK_TLS_CA_CERT"), false, os.Getenv("KEYCLOAK_TLS_CLIENT_CERT"), os.Getenv("KEYCLOAK_TLS_CLIENT_KEY"), "", false, map[string]string{}, os.Getenv("KEYCLOAK_VERSION"))
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := otherClient.NewRealm(testCtx, &keycloak.Realm{Id: realmUuid, Realm: realmName, Enabled: true}); err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = otherClient.DeleteRealm(testCtx, realmName) })
				},
				Config: fmt.Sprintf(`
resource "keycloak_group" "group" {
	realm_uuid = "%s"
	name       = "group"
}
`, realmUuid),
				Check: resource.TestCheckResourceAttr("keycloak_group.group", "realm_id", realmName),
			},
		},
	})
}
