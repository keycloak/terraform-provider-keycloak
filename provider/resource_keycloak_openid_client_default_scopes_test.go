package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"regexp"
	"strings"
	"testing"
)

// All openid clients in Keycloak will automatically have these scopes listed as "default client scopes".
var preAssignedDefaultClientScopes = []string{"profile", "email", "web-origins", "roles"}

func TestAccKeycloakOpenidClientDefaultScopes_basic(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	clientScopes := append(preAssignedDefaultClientScopes, clientScope)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientDefaultScopes_basic(client, clientScope),
				Check:  testAccCheckKeycloakOpenidClientHasDefaultScopes("keycloak_openid_client_default_scopes.default_scopes", clientScopes),
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
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
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
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
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
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
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
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenidClientDefaultScopes_validationBearerOnlyClient(client, clientScope),
				ExpectError: regexp.MustCompile("validation error: client with id .+ uses access type BEARER-ONLY which does not use scopes"),
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
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
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
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
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
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
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

// Keycloak throws a 500 if you attempt to attach an optional scope that is already attached as an optional scope
func TestAccKeycloakOpenidClientDefaultScopes_validateDuplicateScopeAssignment(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	optionalClientScopes := append(getPreAssignedOptionalClientScopes(), clientScope)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// attach optional scopes, including the custom scope
			{
				Config: testKeycloakOpenidClientOptionalScopes_basic(client, clientScope),
				Check:  testAccCheckKeycloakOpenidClientHasOptionalScopes("keycloak_openid_client_optional_scopes.optional_scopes", optionalClientScopes),
			},
			// attach default scopes with the custom scope, expect an error since it is already in use
			{
				Config:      testKeycloakOpenidClientDefaultScopes_duplicateScopeAssignment(client, clientScope),
				ExpectError: regexp.MustCompile("validation error: scope .+ is already attached to client as an optional scope"),
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
	`, testKeycloakOpenidClientOptionalScopes_basic(client, clientScope))
}
