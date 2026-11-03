package provider

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakRealmDefaultClientScopes_basic(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmDefaultScopes_basic(realmName, clientScope),
				Check: testAccCheckKeycloakRealmHasDefaultScopes(
					"keycloak_realm_default_client_scopes.default_scopes",
					[]string{"profile", "email", "web-origins", "roles", "role_list", clientScope}),
			},
			{
				ResourceName:      "keycloak_realm_default_client_scopes.default_scopes",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["keycloak_realm_default_client_scopes.default_scopes"]
					if !ok {
						return "", fmt.Errorf("resource not found: keycloak_realm_default_client_scopes.default_scopes")
					}
					return rs.Primary.Attributes["realm_id"], nil
				},
			},
		},
	})
}

func TestAccKeycloakRealmDefaultClientScopes_empty(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmDefaultScopes_empty(realmName),
				Check: testAccCheckKeycloakRealmHasDefaultScopes(
					"keycloak_realm_default_client_scopes.default_scopes",
					[]string{}),
			},
		},
	})
}

func getRealmDefaultClientScopesFromState(resourceName string, s *terraform.State) ([]*keycloak.OpenidClientScope, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]

	keycloakDefaultClientScopes, err := keycloakClient.GetRealmDefaultClientScopes(testCtx, realm)
	if err != nil {
		return nil, err
	}

	return keycloakDefaultClientScopes, nil
}

func testAccCheckKeycloakRealmHasDefaultScopes(resourceName string, expectedClientScopeNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keycloakDefaultClientScopes, err := getRealmDefaultClientScopesFromState(resourceName, s)
		if err != nil {
			return err
		}

		var assignedClientScopeNames []string
		for _, keycloakDefaultScope := range keycloakDefaultClientScopes {
			assignedClientScopeNames = append(assignedClientScopeNames, keycloakDefaultScope.Name)
		}

		sort.Strings(expectedClientScopeNames)
		sort.Strings(assignedClientScopeNames)

		if !slices.Equal(assignedClientScopeNames, expectedClientScopeNames) {
			return fmt.Errorf(
				"assigned and expected realm default client scopes do not match %v != %v",
				assignedClientScopeNames,
				expectedClientScopeNames,
			)
		}

		return nil
	}
}

func testKeycloakRealmDefaultScopes_basic(realmName string, clientScope string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = keycloak_realm.realm.id

	description = "test description"
}

resource "keycloak_realm_default_client_scopes" "default_scopes" {
	realm_id       = keycloak_realm.realm.id
	default_scopes = [
		"profile",
		"email",
		"roles",
		"role_list",
		"web-origins",
		keycloak_openid_client_scope.client_scope.name
	]
}
	`, realmName, clientScope)
}

func testKeycloakRealmDefaultScopes_empty(realmName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_default_client_scopes" "default_scopes" {
	realm_id       = keycloak_realm.realm.id
	default_scopes = []
}
	`, realmName)
}

func TestAccKeycloakRealmDefaultClientScopes_conflictWithOptionalScopesResource(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmScopes_defaultAndOptional(realmName, clientScope,
					append(slices.Clone(preAssignedRealmDefaultClientScopes), clientScope),
					append(slices.Clone(preAssignedRealmOptionalClientScopes), clientScope)),
				ExpectError: regexp.MustCompile(fmt.Sprintf("scopes \\[%s\\] are configured as both default and optional scopes", clientScope)),
			},
		},
	})
}

// Adding a scope to the optional scopes resource that an already applied default scopes resource reconciles must fail
// the plan, even though the default scopes resource itself does not change.
func TestAccKeycloakRealmDefaultClientScopes_conflictWithAppliedResourceFailsPlan(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	defaultWithScope := append(slices.Clone(preAssignedRealmDefaultClientScopes), clientScope)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmScopes_defaultAndOptional(realmName, clientScope, defaultWithScope, preAssignedRealmOptionalClientScopes),
			},
			{
				Config: testKeycloakRealmScopes_defaultAndOptional(realmName, clientScope, defaultWithScope,
					append(slices.Clone(preAssignedRealmOptionalClientScopes), clientScope)),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(fmt.Sprintf("scopes \\[%s\\] are configured as both default and optional scopes", clientScope)),
			},
			{
				Config: testKeycloakRealmScopes_defaultAndOptional(realmName, clientScope, defaultWithScope, preAssignedRealmOptionalClientScopes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmHasDefaultScopes("keycloak_realm_default_client_scopes.default_scopes", slices.Clone(defaultWithScope)),
					testAccCheckKeycloakRealmHasOptionalScopes("keycloak_realm_optional_client_scopes.optional_scopes", slices.Clone(preAssignedRealmOptionalClientScopes)),
				),
			},
		},
	})
}

// A scope that is moved to optional outside of Terraform after the default scopes resource was applied is moved back
// to default by the next apply, and the optional scopes resource does not take it over.
func TestAccKeycloakRealmDefaultClientScopes_overridesOptionalAssignmentAfterApply(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	defaultWithScope := append(slices.Clone(preAssignedRealmDefaultClientScopes), clientScope)
	config := testKeycloakRealmScopes_defaultAndOptional(realmName, clientScope, defaultWithScope, preAssignedRealmOptionalClientScopes)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				PreConfig: func() {
					err := keycloakClient.MarkClientScopesAsRealmOptional(testCtx, realmName, []string{clientScope})
					if err != nil {
						t.Fatal(err)
					}

					optionalScopes, err := keycloakClient.GetRealmOptionalClientScopes(testCtx, realmName)
					if err != nil {
						t.Fatal(err)
					}
					if !slices.ContainsFunc(optionalScopes, func(s *keycloak.OpenidClientScope) bool { return s.Name == clientScope }) {
						t.Fatalf("expected %s to be moved to optional outside of terraform", clientScope)
					}
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmHasDefaultScopes("keycloak_realm_default_client_scopes.default_scopes", slices.Clone(defaultWithScope)),
					testAccCheckKeycloakRealmHasOptionalScopes("keycloak_realm_optional_client_scopes.optional_scopes", slices.Clone(preAssignedRealmOptionalClientScopes)),
				),
			},
		},
	})
}

// The default and optional scopes resources are not ordered against each other, so moving a scope between them has to
// work whichever resource reconciles first.
func TestAccKeycloakRealmDefaultClientScopes_moveBetweenResourcesWithoutDependency(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	defaultWithScope := append(slices.Clone(preAssignedRealmDefaultClientScopes), clientScope)
	optionalWithScope := append(slices.Clone(preAssignedRealmOptionalClientScopes), clientScope)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmScopes_defaultAndOptional(realmName, clientScope, preAssignedRealmDefaultClientScopes, optionalWithScope),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmHasDefaultScopes("keycloak_realm_default_client_scopes.default_scopes", slices.Clone(preAssignedRealmDefaultClientScopes)),
					testAccCheckKeycloakRealmHasOptionalScopes("keycloak_realm_optional_client_scopes.optional_scopes", slices.Clone(optionalWithScope)),
				),
			},
			// optional to default
			{
				Config: testKeycloakRealmScopes_defaultAndOptional(realmName, clientScope, defaultWithScope, preAssignedRealmOptionalClientScopes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmHasDefaultScopes("keycloak_realm_default_client_scopes.default_scopes", slices.Clone(defaultWithScope)),
					testAccCheckKeycloakRealmHasOptionalScopes("keycloak_realm_optional_client_scopes.optional_scopes", slices.Clone(preAssignedRealmOptionalClientScopes)),
				),
			},
			// and back to optional
			{
				Config: testKeycloakRealmScopes_defaultAndOptional(realmName, clientScope, preAssignedRealmDefaultClientScopes, optionalWithScope),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmHasDefaultScopes("keycloak_realm_default_client_scopes.default_scopes", slices.Clone(preAssignedRealmDefaultClientScopes)),
					testAccCheckKeycloakRealmHasOptionalScopes("keycloak_realm_optional_client_scopes.optional_scopes", slices.Clone(optionalWithScope)),
				),
			},
		},
	})
}

// Keycloak removes a scope assignment whatever its type, so unmarking a stale realm optional scope must not remove the
// scope once it has moved to default.
//
// This tests the API client rather than the resources: it calls the mark and unmark functions directly (from Check,
// after applying a config that only creates the realm and scope) so that the default mark always runs before the
// stale optional unmark. Terraform does not guarantee this order when a scope moves between the resources, so
// moveBetweenResourcesWithoutDependency cannot reliably cover it.
func TestAccKeycloakRealmDefaultClientScopes_apiClientStaleOptionalUnmarkKeepsDefaultScope(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmScopes_realmAndClientScope(realmName, clientScope),
				Check: func(s *terraform.State) error {
					if err := keycloakClient.MarkClientScopesAsRealmDefault(testCtx, realmName, []string{clientScope}); err != nil {
						return err
					}

					// stale optional unmark, as if the optional resource had read the scope before it moved
					if err := keycloakClient.UnmarkClientScopesAsRealmOptional(testCtx, realmName, []string{clientScope}); err != nil {
						return err
					}

					defaultScopes, err := keycloakClient.GetRealmDefaultClientScopes(testCtx, realmName)
					if err != nil {
						return err
					}

					for _, defaultScope := range defaultScopes {
						if defaultScope.Name == clientScope {
							return nil
						}
					}

					return fmt.Errorf("expected client scope %s to still be a realm default scope", clientScope)
				},
			},
		},
	})
}

func testKeycloakRealmScopes_realmAndClientScope(realmName, clientScope string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = keycloak_realm.realm.id

	description = "test description"
}
	`, realmName, clientScope)
}

var preAssignedRealmDefaultClientScopes = []string{"profile", "email", "roles", "role_list", "web-origins"}
var preAssignedRealmOptionalClientScopes = []string{"address", "phone", "offline_access"}

func testKeycloakRealmScopes_defaultAndOptional(realmName, clientScope string, defaultScopes, optionalScopes []string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = keycloak_realm.realm.id

	description = "test description"
}

resource "keycloak_realm_default_client_scopes" "default_scopes" {
	realm_id       = keycloak_realm.realm.id
	default_scopes = %s

	depends_on = [keycloak_openid_client_scope.client_scope]
}

resource "keycloak_realm_optional_client_scopes" "optional_scopes" {
	realm_id        = keycloak_realm.realm.id
	optional_scopes = %s

	depends_on = [keycloak_openid_client_scope.client_scope]
}
	`, realmName, clientScope, arrayOfStringsForTerraformResource(defaultScopes), arrayOfStringsForTerraformResource(optionalScopes))
}

// Keycloak returns a 409 if you attempt to mark a client scope as a realm default scope while it is marked as a realm
// optional scope, but the provider will unmark the existing scope and re-mark it as a realm default scope
func TestAccKeycloakRealmDefaultClientScopes_duplicateScopeAssignment(t *testing.T) {
	t.Parallel()
	realmName := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmOptionalScopes_basic(realmName, clientScope),
				Check: testAccCheckKeycloakRealmHasOptionalScopes(
					"keycloak_realm_optional_client_scopes.optional_scopes",
					[]string{"address", "phone", "offline_access", clientScope},
				),
			},
			// move the custom scope from optional to default while it is still marked as optional
			{
				Config: testKeycloakRealmDefaultScopes_duplicateScopeAssignment(realmName, clientScope),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmHasDefaultScopes(
						"keycloak_realm_default_client_scopes.default_scopes",
						[]string{"profile", "email", "roles", "role_list", "web-origins", clientScope},
					),
					testAccCheckKeycloakRealmHasOptionalScopes(
						"keycloak_realm_optional_client_scopes.optional_scopes",
						[]string{"address", "phone", "offline_access"},
					),
				),
			},
		},
	})
}

// the optional scopes resource depends on the default scopes resource so that the custom scope is still marked as a
// realm optional scope at the point it is marked as a realm default scope
func testKeycloakRealmDefaultScopes_duplicateScopeAssignment(realmName, clientScope string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = keycloak_realm.realm.id

	description = "test description"
}

resource "keycloak_realm_optional_client_scopes" "optional_scopes" {
	realm_id        = keycloak_realm.realm.id
	optional_scopes = [
		"address",
		"phone",
		"offline_access"
	]

	depends_on = [keycloak_realm_default_client_scopes.default_scopes]
}

resource "keycloak_realm_default_client_scopes" "default_scopes" {
	realm_id       = keycloak_realm.realm.id
	default_scopes = [
		"profile",
		"email",
		"roles",
		"role_list",
		"web-origins",
		keycloak_openid_client_scope.client_scope.name
	]
}
	`, realmName, clientScope)
}
