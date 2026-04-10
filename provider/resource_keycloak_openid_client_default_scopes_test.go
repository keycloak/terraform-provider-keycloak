package provider

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// All openid clients in Keycloak will automatically have these scopes listed as "default client scopes". Note this list is not exhaustive and doesn't have to be for the sake of these tests.
var preAssignedDefaultClientScopes = []string{"profile", "email", "web-origins", "roles"}

func TestAccKeycloakOpenidClientDefaultScopes_basic(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	clientScopes := append(preAssignedDefaultClientScopes, clientScope)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientDefaultScopes_basic(client, clientScope),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", clientScopes),
			},
			{
				ResourceName:      "keycloak_openid_client_default_scopes.default_scopes",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// we need a separate test step for destroy instead of using CheckDestroy because this resource is implicitly
			// destroyed at the end of each test via destroying clients
			{
				Config: testKeycloakOpenidClientDefaultScopes_noDefaultScopes(client, clientScope),
				Check:  testAccCheckKeycloakOpenidClientHasNoDefaultScopes("keycloak_openid_client.client"),
			},
		},
	})
}

func TestAccKeycloakOpenidClientDefaultScopes_updateClientForceNew(t *testing.T) {
	t.Parallel()
	clientOne := acctest.RandomWithPrefix("tf-acc")
	clientTwo := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	clientScopes := append(preAssignedDefaultClientScopes, clientScope)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientDefaultScopes_basic(clientOne, clientScope),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", clientScopes),
			},
			{
				Config: testKeycloakOpenidClientDefaultScopes_basic(clientTwo, clientScope),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", clientScopes),
			},
		},
	})
}

func TestAccKeycloakOpenidClientDefaultScopes_updateInPlace(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	allClientScopes := append(preAssignedDefaultClientScopes, clientScope)

	clientScopeToRemove := allClientScopes[acctest.RandIntRange(0, 2)]
	var subsetOfClientScopes []string
	for _, cs := range allClientScopes {
		if cs != clientScopeToRemove {
			subsetOfClientScopes = append(subsetOfClientScopes, cs)
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// init
			{
				Config: testKeycloakOpenidClientDefaultScopes_listOfScopes(client, clientScope, allClientScopes),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", allClientScopes),
			},
			// remove
			{
				Config: testKeycloakOpenidClientDefaultScopes_listOfScopes(client, clientScope, subsetOfClientScopes),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", subsetOfClientScopes),
			},
			// add
			{
				Config: testKeycloakOpenidClientDefaultScopes_listOfScopes(client, clientScope, allClientScopes),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", allClientScopes),
			},
		},
	})
}

func TestAccKeycloakOpenidClientDefaultScopes_validateClientDoesNotExist(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenidClientDefaultScopes_validationNoClient(client, clientScope),
				ExpectError: regexp.MustCompile("validation error: client with id .+ does not exist"),
			},
		},
	})
}

func TestAccKeycloakOpenidClientDefaultScopes_validateClientAccessType(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenidClientDefaultScopes_validationBearerOnlyClient(client, clientScope),
				ExpectError: regexp.MustCompile("validation error: client with id .+ uses access type BEARER-ONLY which does not use scopes"),
			},
		},
	})
}

func TestAccKeycloakOpenidClientDefaultScopes_validateScopeDoesNotExist(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenidClientDefaultScopes_listOfScopes(client, acctest.RandomWithPrefix("tf-acc"), []string{"profile", "non_existent_scope_that_should_fail"}),
				ExpectError: regexp.MustCompile("scope .+ does not exist"),
			},
		},
	})
}

// if a default client scope is manually detached from a client with default scopes controlled by this resource, terraform should add it again
func TestAccKeycloakOpenidClientDefaultScopes_authoritativeAdd(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScopes := append(preAssignedDefaultClientScopes,
		"terraform-client-scope-"+acctest.RandString(10),
		"terraform-client-scope-"+acctest.RandString(10),
		"terraform-client-scope-"+acctest.RandString(10),
	)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientDefaultScopes_multipleClientScopes(client, clientScopes, clientScopes),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", clientScopes),
			},
			{
				PreConfig: func() {
					client, err := keycloakClient.GetOpenidClientByClientId(testCtx, testAccRealm.Realm, client)
					if err != nil {
						t.Fatal(err)
					}

					clientToManuallyDetach := clientScopes[acctest.RandIntRange(0, len(clientScopes)-1)]
					err = keycloakClient.DetachOpenidClientDefaultScopes(testCtx, testAccRealm.Realm, client.Id, []string{clientToManuallyDetach})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOpenidClientDefaultScopes_multipleClientScopes(client, clientScopes, clientScopes),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", clientScopes),
			},
		},
	})
}

// if a default client scope is manually attached to a client with default scopes controlled by this resource, terraform should detach it
func TestAccKeycloakOpenidClientDefaultScopes_authoritativeRemove(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")

	randomClientScopes := []string{
		"terraform-client-scope-" + acctest.RandString(10),
		"terraform-client-scope-" + acctest.RandString(10),
		"terraform-client-scope-" + acctest.RandString(10),
	}
	allClientScopes := append(preAssignedDefaultClientScopes, randomClientScopes...)

	clientToManuallyAttach := randomClientScopes[acctest.RandIntRange(0, len(randomClientScopes)-1)]
	var attachedClientScopes []string
	for _, clientScope := range allClientScopes {
		if clientScope != clientToManuallyAttach {
			attachedClientScopes = append(attachedClientScopes, clientScope)
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientDefaultScopes_multipleClientScopes(client, allClientScopes, attachedClientScopes),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", attachedClientScopes),
			},
			{
				PreConfig: func() {
					client, err := keycloakClient.GetOpenidClientByClientId(testCtx, testAccRealm.Realm, client)
					if err != nil {
						t.Fatal(err)
					}

					err = keycloakClient.AttachOpenidClientDefaultScopes(testCtx, testAccRealm.Realm, client.Id, []string{clientToManuallyAttach})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOpenidClientDefaultScopes_multipleClientScopes(client, allClientScopes, attachedClientScopes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", attachedClientScopes),
					testAccCheckKeycloakOpenidClientDefaultScopeIsNotAttached("keycloak_openid_client_default_scopes.default_scopes", clientToManuallyAttach),
				),
			},
		},
	})
}

// this resource doesn't support import because it can be created even if the desired state already exists in keycloak
func TestAccKeycloakOpenidClientDefaultScopes_noImportNeeded(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	clientScopes := append(preAssignedDefaultClientScopes, clientScope)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientDefaultScopes_noDefaultScopes(client, clientScope),
				Check:  testAccCheckKeycloakOpenidClientDefaultScopeIsNotAttached("keycloak_openid_client.client", clientScope),
			},
			{
				PreConfig: func() {
					openidClient, err := keycloakClient.GetOpenidClientByClientId(testCtx, testAccRealm.Realm, client)
					if err != nil {
						t.Fatal(err)
					}

					err = keycloakClient.AttachOpenidClientDefaultScopes(testCtx, testAccRealm.Realm, openidClient.Id, clientScopes)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOpenidClientDefaultScopes_basic(client, clientScope),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", clientScopes),
			},
		},
	})
}

// Keycloak silently ignores an attempt to attach a default scope that is already attached as an optional scope,
// but the provider will remove the existing scope and re-attach as a default scope
func TestAccKeycloakOpenidClientDefaultScopes_validateDuplicateScopeAssignment(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	expectedDefaultClientScopes := append(preAssignedDefaultClientScopes, clientScope)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientOptionalScopes_noOptionalScopes(client, clientScope),
			},
			// attach the custom scope as optional outside of terraform, so no resource owns it and cleans it up
			{
				PreConfig: func() {
					openidClient, err := keycloakClient.GetOpenidClientByClientId(testCtx, testAccRealm.Realm, client)
					if err != nil {
						t.Fatal(err)
					}

					err = keycloakClient.AttachOpenidClientOptionalScopes(testCtx, testAccRealm.Realm, openidClient.Id, []string{clientScope})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOpenidClientDefaultScopes_duplicateScopeAssignment(client, clientScope),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", expectedDefaultClientScopes),
					testAccCheckKeycloakOpenidClientOptionalScopeIsNotAttached("keycloak_openid_client.client", clientScope),
				),
			},
		},
	})
}

func TestAccKeycloakOpenidClientDefaultScopes_conflictWithOptionalScopesResource(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientScopes_defaultAndOptional(client, clientScope,
					append(preAssignedDefaultClientScopes, clientScope),
					append(getPreAssignedOptionalClientScopes(), clientScope)),
				ExpectError: regexp.MustCompile(fmt.Sprintf("scopes \\[%s\\] are configured as both default and optional scopes", clientScope)),
			},
		},
	})
}

// Adding a scope to the optional scopes resource that an already applied default scopes resource reconciles must fail
// the plan, even though the default scopes resource itself does not change.
func TestAccKeycloakOpenidClientDefaultScopes_conflictWithAppliedResourceFailsPlan(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	defaultWithScope := append(slices.Clone(preAssignedDefaultClientScopes), clientScope)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientScopes_defaultAndOptional(client, clientScope, defaultWithScope, getPreAssignedOptionalClientScopes()),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", defaultWithScope),
			},
			{
				Config: testKeycloakOpenidClientScopes_defaultAndOptional(client, clientScope, defaultWithScope,
					append(getPreAssignedOptionalClientScopes(), clientScope)),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(fmt.Sprintf("scopes \\[%s\\] are configured as both default and optional scopes", clientScope)),
			},
			{
				Config: testKeycloakOpenidClientScopes_defaultAndOptional(client, clientScope, defaultWithScope, getPreAssignedOptionalClientScopes()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", defaultWithScope),
					testAccCheckKeycloakOpenidClientOptionalScopeIsNotAttached("keycloak_openid_client.client", clientScope),
				),
			},
		},
	})
}

// A scope that is moved to optional outside of Terraform after the default scopes resource was applied is moved back
// to default by the next apply, and the optional scopes resource does not take it over.
func TestAccKeycloakOpenidClientDefaultScopes_overridesOptionalAssignmentAfterApply(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	defaultWithScope := append(slices.Clone(preAssignedDefaultClientScopes), clientScope)
	config := testKeycloakOpenidClientScopes_defaultAndOptional(client, clientScope, defaultWithScope, getPreAssignedOptionalClientScopes())

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				PreConfig: func() {
					openidClient, err := keycloakClient.GetOpenidClientByClientId(testCtx, testAccRealm.Realm, client)
					if err != nil {
						t.Fatal(err)
					}

					err = keycloakClient.AttachOpenidClientOptionalScopes(testCtx, testAccRealm.Realm, openidClient.Id, []string{clientScope})
					if err != nil {
						t.Fatal(err)
					}

					optionalScopes, err := keycloakClient.GetOpenidClientOptionalScopes(testCtx, testAccRealm.Realm, openidClient.Id)
					if err != nil {
						t.Fatal(err)
					}
					if !slices.ContainsFunc(optionalScopes, func(s *keycloak.OpenidClientScope) bool { return s.Name == clientScope }) {
						t.Fatalf("expected %s to be moved to optional outside of terraform", clientScope)
					}
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", defaultWithScope),
					testAccCheckKeycloakOpenidClientOptionalScopeIsNotAttached("keycloak_openid_client.client", clientScope),
				),
			},
		},
	})
}

// The default and optional scopes resources are not ordered against each other, so moving a scope between them has to
// work whichever resource reconciles first.
func TestAccKeycloakOpenidClientDefaultScopes_moveBetweenResourcesWithoutDependency(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	defaultWithScope := append(preAssignedDefaultClientScopes, clientScope)
	optionalWithScope := append(getPreAssignedOptionalClientScopes(), clientScope)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientScopes_defaultAndOptional(client, clientScope, preAssignedDefaultClientScopes, optionalWithScope),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasOptionalScopes("keycloak_openid_client_optional_scopes.optional_scopes", optionalWithScope),
					testAccCheckKeycloakOpenidClientDefaultScopeIsNotAttached("keycloak_openid_client.client", clientScope),
				),
			},
			// optional to default
			{
				Config: testKeycloakOpenidClientScopes_defaultAndOptional(client, clientScope, defaultWithScope, getPreAssignedOptionalClientScopes()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", defaultWithScope),
					testAccCheckKeycloakOpenidClientHasOptionalScopes("keycloak_openid_client_optional_scopes.optional_scopes", getPreAssignedOptionalClientScopes()),
					testAccCheckKeycloakOpenidClientOptionalScopeIsNotAttached("keycloak_openid_client.client", clientScope),
				),
			},
			// and back to optional
			{
				Config: testKeycloakOpenidClientScopes_defaultAndOptional(client, clientScope, preAssignedDefaultClientScopes, optionalWithScope),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", preAssignedDefaultClientScopes),
					testAccCheckKeycloakOpenidClientHasOptionalScopes("keycloak_openid_client_optional_scopes.optional_scopes", optionalWithScope),
					testAccCheckKeycloakOpenidClientDefaultScopeIsNotAttached("keycloak_openid_client.client", clientScope),
				),
			},
		},
	})
}

func getDefaultClientScopesFromState(resourceName string, s *terraform.State) ([]*keycloak.OpenidClientScope, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]

	var client string
	if strings.HasPrefix(resourceName, "keycloak_openid_client_default_scopes") {
		client = rs.Primary.Attributes["client_id"]
	} else {
		client = rs.Primary.ID
	}

	keycloakDefaultClientScopes, err := keycloakClient.GetOpenidClientDefaultScopes(testCtx, realm, client)
	if err != nil {
		return nil, err
	}

	return keycloakDefaultClientScopes, nil
}

func testAccCheckKeycloakOpenidClientHasDefaultScopes(resourceName string, tfDefaultClientScopes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keycloakDefaultClientScopes, err := getDefaultClientScopesFromState(resourceName, s)
		if err != nil {
			return err
		}

		for _, tfDefaultClientScope := range tfDefaultClientScopes {
			found := false

			for _, keycloakDefaultScope := range keycloakDefaultClientScopes {
				if keycloakDefaultScope.Name == tfDefaultClientScope {
					found = true

					break
				}
			}

			if !found {
				return fmt.Errorf("default scope %s is not assigned to client", tfDefaultClientScope)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientHasNoDefaultScopes(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keycloakDefaultClientScopes, err := getDefaultClientScopesFromState(resourceName, s)
		if err != nil {
			return err
		}

		if numberOfDefaultScopes := len(keycloakDefaultClientScopes); numberOfDefaultScopes != 0 {
			return fmt.Errorf("expected client to have no assigned default scopes, but it has %d", numberOfDefaultScopes)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientDefaultScopeIsNotAttached(resourceName, clientScope string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keycloakDefaultClientScopes, err := getDefaultClientScopesFromState(resourceName, s)
		if err != nil {
			return err
		}

		for _, keycloakDefaultClientScope := range keycloakDefaultClientScopes {
			if keycloakDefaultClientScope.Name == clientScope {
				return fmt.Errorf("expected client scope with name %s to not be attached to client", clientScope)
			}
		}

		return nil
	}
}

func testKeycloakOpenidClientDefaultScopes_basic(client, clientScope string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "PUBLIC"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id

	description = "test description"
}

resource "keycloak_openid_client_default_scopes" "default_scopes" {
	realm_id       = data.keycloak_realm.realm.id
	client_id      = "${keycloak_openid_client.client.id}"
	default_scopes = [
		"profile",
		"email",
		"roles",
		"web-origins",
		"${keycloak_openid_client_scope.client_scope.name}"
	]
}
	`, testAccRealm.Realm, client, clientScope)
}

func testKeycloakOpenidClientDefaultScopes_noDefaultScopes(client, clientScope string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "PUBLIC"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id

	description = "test description"
}
	`, testAccRealm.Realm, client, clientScope)
}

func testKeycloakOpenidClientDefaultScopes_listOfScopes(client, clientScope string, listOfDefaultScopes []string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "PUBLIC"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id

	description = "test description"
}

resource "keycloak_openid_client_default_scopes" "default_scopes" {
	realm_id       = data.keycloak_realm.realm.id
	client_id      = "${keycloak_openid_client.client.id}"
	default_scopes = %s

	depends_on = ["keycloak_openid_client_scope.client_scope"]
}
	`, testAccRealm.Realm, client, clientScope, arrayOfStringsForTerraformResource(listOfDefaultScopes))
}

func testKeycloakOpenidClientDefaultScopes_validationNoClient(client, clientScope string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id

	description = "test description"
}

resource "keycloak_openid_client_default_scopes" "default_scopes" {
	realm_id       = data.keycloak_realm.realm.id
	client_id      = "%s"
	default_scopes = [
		"profile",
		"email",
		"roles",
		"web-origins",
		"${keycloak_openid_client_scope.client_scope.name}"
	]
}
	`, testAccRealm.Realm, clientScope, client)
}

func testKeycloakOpenidClientDefaultScopes_validationBearerOnlyClient(client, clientScope string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "BEARER-ONLY"
}

resource "keycloak_openid_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id

	description = "test description"
}

resource "keycloak_openid_client_default_scopes" "default_scopes" {
	realm_id       = data.keycloak_realm.realm.id
	client_id      = "${keycloak_openid_client.client.id}"
	default_scopes = [
		"profile",
		"email",
		"roles",
		"web-origins",
		"${keycloak_openid_client_scope.client_scope.name}"
	]
}
	`, testAccRealm.Realm, client, clientScope)
}

func testKeycloakOpenidClientDefaultScopes_multipleClientScopes(client string, allClientScopes, attachedClientScopes []string) string {
	var clientScopeResources strings.Builder
	for _, clientScope := range allClientScopes {
		if strings.HasPrefix(clientScope, "terraform") {
			clientScopeResources.WriteString(fmt.Sprintf(`
resource "keycloak_openid_client_scope" "client_scope_%s" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id
}
		`, clientScope, clientScope))
		}
	}

	var attachedClientScopesInterpolated []string
	for _, attachedClientScope := range attachedClientScopes {
		if strings.HasPrefix(attachedClientScope, "terraform") {
			attachedClientScopesInterpolated = append(attachedClientScopesInterpolated, fmt.Sprintf("${keycloak_openid_client_scope.client_scope_%s.name}", attachedClientScope))
		} else {
			attachedClientScopesInterpolated = append(attachedClientScopesInterpolated, attachedClientScope)
		}
	}

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "PUBLIC"
}

%s

resource "keycloak_openid_client_default_scopes" "default_scopes" {
	realm_id       = data.keycloak_realm.realm.id
	client_id      = "${keycloak_openid_client.client.id}"
	default_scopes = %s
}
	`, testAccRealm.Realm, client, clientScopeResources.String(), arrayOfStringsForTerraformResource(attachedClientScopesInterpolated))
}

// Keycloak removes a scope assignment whatever its type, so detaching a stale optional scope must not remove the scope
// once it has moved to default.
//
// This tests the API client rather than the resources: it calls the attach and detach functions directly (from Check,
// after applying a config that only creates the client and scope) so that the default attach always runs before the
// stale optional detach. Terraform does not guarantee this order when a scope moves between the resources, so
// moveBetweenResourcesWithoutDependency cannot reliably cover it.
func TestAccKeycloakOpenidClientDefaultScopes_apiClientStaleOptionalDetachKeepsDefaultScope(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientOptionalScopes_noOptionalScopes(client, clientScope),
				Check: func(s *terraform.State) error {
					openidClient, err := keycloakClient.GetOpenidClientByClientId(testCtx, testAccRealm.Realm, client)
					if err != nil {
						return err
					}

					if err = keycloakClient.AttachOpenidClientDefaultScopes(testCtx, testAccRealm.Realm, openidClient.Id, []string{clientScope}); err != nil {
						return err
					}

					if err = keycloakClient.DetachOpenidClientOptionalScopes(testCtx, testAccRealm.Realm, openidClient.Id, []string{clientScope}); err != nil {
						return err
					}

					return testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client.client", append(slices.Clone(preAssignedDefaultClientScopes), clientScope))(s)
				},
			},
		},
	})
}

func testKeycloakOpenidClientScopes_defaultAndOptional(client, clientScope string, defaultScopes, optionalScopes []string) string {
	return fmt.Sprintf(`
%s

resource "keycloak_openid_client_default_scopes" "default_scopes" {
	realm_id       = data.keycloak_realm.realm.id
	client_id      = keycloak_openid_client.client.id
	default_scopes = %s

	depends_on = [keycloak_openid_client_scope.client_scope]
}

resource "keycloak_openid_client_optional_scopes" "optional_scopes" {
	realm_id        = data.keycloak_realm.realm.id
	client_id       = keycloak_openid_client.client.id
	optional_scopes = %s

	depends_on = [keycloak_openid_client_scope.client_scope]
}
	`, testKeycloakOpenidClientOptionalScopes_noOptionalScopes(client, clientScope), arrayOfStringsForTerraformResource(defaultScopes), arrayOfStringsForTerraformResource(optionalScopes))
}

func testKeycloakOpenidClientDefaultScopes_duplicateScopeAssignment(client, clientScope string) string {
	return fmt.Sprintf(`
%s

resource "keycloak_openid_client_default_scopes" "default_scopes" {
	realm_id       = data.keycloak_realm.realm.id
	client_id      = "${keycloak_openid_client.client.id}"
	default_scopes = [
		"profile",
		"email",
		"roles",
		"web-origins",
		"${keycloak_openid_client_scope.client_scope.name}"
	]
}
	`, testKeycloakOpenidClientOptionalScopes_noOptionalScopes(client, clientScope))
}
