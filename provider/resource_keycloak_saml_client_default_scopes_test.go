package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// All saml clients in Keycloak will automatically have these scopes listed as "default client scopes".
var preAssignedDefaultSamlClientScopes = []string{"role_list"}

func TestAccKeycloakSamlClientDefaultScopes_basic(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	clientScopes := append(preAssignedDefaultSamlClientScopes, clientScope)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClientDefaultScopes_basic(client, clientScope),
				Check:  testAccCheckKeycloakSamlClientHasDefaultScopes("keycloak_saml_client_default_scopes.default_scopes", clientScopes),
			},
			// we need a separate test step for destroy instead of using CheckDestroy because this resource is implicitly
			// destroyed at the end of each test via destroying clients
			{
				Config: testKeycloakSamlClientDefaultScopes_noDefaultScopes(client, clientScope),
				Check:  testAccCheckKeycloakSamlClientHasNoDefaultScopes("keycloak_saml_client.client"),
			},
		},
	})
}

func TestAccKeycloakSamlClientDefaultScopes_updateClientForceNew(t *testing.T) {
	t.Parallel()
	clientOne := acctest.RandomWithPrefix("tf-acc")
	clientTwo := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	clientScopes := append(preAssignedDefaultSamlClientScopes, clientScope)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClientDefaultScopes_basic(clientOne, clientScope),
				Check:  testAccCheckKeycloakSamlClientHasDefaultScopes("keycloak_saml_client_default_scopes.default_scopes", clientScopes),
			},
			{
				Config: testKeycloakSamlClientDefaultScopes_basic(clientTwo, clientScope),
				Check:  testAccCheckKeycloakSamlClientHasDefaultScopes("keycloak_saml_client_default_scopes.default_scopes", clientScopes),
			},
		},
	})
}

func TestAccKeycloakSamlClientDefaultScopes_updateInPlace(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	allClientScopes := append(preAssignedDefaultSamlClientScopes, clientScope)

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
				Config: testKeycloakSamlClientDefaultScopes_listOfScopes(client, clientScope, allClientScopes),
				Check:  testAccCheckKeycloakSamlClientHasDefaultScopes("keycloak_saml_client_default_scopes.default_scopes", allClientScopes),
			},
			// remove
			{
				Config: testKeycloakSamlClientDefaultScopes_listOfScopes(client, clientScope, subsetOfClientScopes),
				Check:  testAccCheckKeycloakSamlClientHasDefaultScopes("keycloak_saml_client_default_scopes.default_scopes", subsetOfClientScopes),
			},
			// add
			{
				Config: testKeycloakSamlClientDefaultScopes_listOfScopes(client, clientScope, allClientScopes),
				Check:  testAccCheckKeycloakSamlClientHasDefaultScopes("keycloak_saml_client_default_scopes.default_scopes", allClientScopes),
			},
		},
	})
}

func TestAccKeycloakSamlClientDefaultScopes_validateClientDoesNotExist(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakSamlClientDefaultScopes_validationNoClient(client, clientScope),
				ExpectError: regexp.MustCompile("validation error: client with id .+ does not exist"),
			},
		},
	})
}

// if a default client scope is manually detached from a client with default scopes controlled by this resource, terraform should add it again
func TestAccKeycloakSamlClientDefaultScopes_authoritativeAdd(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScopes := append(preAssignedDefaultSamlClientScopes,
		"terraform-client-scope-"+acctest.RandString(10),
		"terraform-client-scope-"+acctest.RandString(10),
		"terraform-client-scope-"+acctest.RandString(10),
	)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClientDefaultScopes_multipleClientScopes(client, clientScopes, clientScopes),
				Check:  testAccCheckKeycloakSamlClientHasDefaultScopes("keycloak_saml_client_default_scopes.default_scopes", clientScopes),
			},
			{
				PreConfig: func() {
					client, err := keycloakClient.GetSamlClientByClientId(testCtx, testAccRealm.Realm, client)
					if err != nil {
						t.Fatal(err)
					}

					clientToManuallyDetach := clientScopes[acctest.RandIntRange(0, len(clientScopes)-1)]
					err = keycloakClient.DetachSamlClientDefaultScopes(testCtx, testAccRealm.Realm, client.Id, []string{clientToManuallyDetach})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakSamlClientDefaultScopes_multipleClientScopes(client, clientScopes, clientScopes),
				Check:  testAccCheckKeycloakSamlClientHasDefaultScopes("keycloak_saml_client_default_scopes.default_scopes", clientScopes),
			},
		},
	})
}

// if a default client scope is manually attached to a client with default scopes controlled by this resource, terraform should detach it
func TestAccKeycloakSamlClientDefaultScopes_authoritativeRemove(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")

	randomClientScopes := []string{
		"terraform-client-scope-" + acctest.RandString(10),
		"terraform-client-scope-" + acctest.RandString(10),
		"terraform-client-scope-" + acctest.RandString(10),
	}
	allClientScopes := append(preAssignedDefaultSamlClientScopes, randomClientScopes...)

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
				Config: testKeycloakSamlClientDefaultScopes_multipleClientScopes(client, allClientScopes, attachedClientScopes),
				Check:  testAccCheckKeycloakSamlClientHasDefaultScopes("keycloak_saml_client_default_scopes.default_scopes", attachedClientScopes),
			},
			{
				PreConfig: func() {
					client, err := keycloakClient.GetSamlClientByClientId(testCtx, testAccRealm.Realm, client)
					if err != nil {
						t.Fatal(err)
					}

					err = keycloakClient.AttachSamlClientDefaultScopes(testCtx, testAccRealm.Realm, client.Id, []string{clientToManuallyAttach})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakSamlClientDefaultScopes_multipleClientScopes(client, allClientScopes, attachedClientScopes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakSamlClientHasDefaultScopes("keycloak_saml_client_default_scopes.default_scopes", attachedClientScopes),
					testAccCheckKeycloakSamlClientDefaultScopeIsNotAttached("keycloak_saml_client_default_scopes.default_scopes", clientToManuallyAttach),
				),
			},
		},
	})
}

// this resource doesn't support import because it can be created even if the desired state already exists in keycloak
func TestAccKeycloakSamlClientDefaultScopes_noImportNeeded(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	clientScopes := append(preAssignedDefaultSamlClientScopes, clientScope)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSamlClientDefaultScopes_noDefaultScopes(client, clientScope),
				Check:  testAccCheckKeycloakSamlClientDefaultScopeIsNotAttached("keycloak_saml_client.client", clientScope),
			},
			{
				PreConfig: func() {
					samlClient, err := keycloakClient.GetSamlClientByClientId(testCtx, testAccRealm.Realm, client)
					if err != nil {
						t.Fatal(err)
					}

					err = keycloakClient.AttachSamlClientDefaultScopes(testCtx, testAccRealm.Realm, samlClient.Id, clientScopes)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakSamlClientDefaultScopes_basic(client, clientScope),
				Check:  testAccCheckKeycloakSamlClientHasDefaultScopes("keycloak_saml_client_default_scopes.default_scopes", clientScopes),
			},
		},
	})
}

// by default, keycloak saml clients have the default scopes "role_list",
// attached. if you create this resource with only one scope, it
// won't remove these two scopes, because the creation of a new resource should not
// result in anything destructive. thus, a following plan will not be empty, as terraform
// will think it needs to remove these scopes, which is okay to do during an update
func TestAccKeycloakSamlClientDefaultScopes_profileAndEmailDefaultScopes(t *testing.T) {
	t.Parallel()
	client := acctest.RandomWithPrefix("tf-acc")
	clientScope := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:             testKeycloakSamlClientDefaultScopes_listOfScopes(client, clientScope, []string{clientScope}),
				Check:              testAccCheckKeycloakSamlClientHasDefaultScopes("keycloak_saml_client.client", append(preAssignedDefaultSamlClientScopes, clientScope)),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func getDefaultSamlClientScopesFromState(resourceName string, s *terraform.State) ([]*keycloak.SamlClientScope, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	var client string
	if strings.HasPrefix(resourceName, "keycloak_saml_client_default_scopes") {
		client = rs.Primary.Attributes["client_id"]
	} else {
		client = rs.Primary.ID
	}

	keycloakDefaultSamlClientScopes, err := keycloakClient.GetSamlClientDefaultScopes(testCtx, testAccRealm.Realm, client)
	if err != nil {
		return nil, err
	}

	return keycloakDefaultSamlClientScopes, nil
}

func testAccCheckKeycloakSamlClientHasDefaultScopes(resourceName string, tfDefaultClientScopes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keycloakDefaultClientScopes, err := getDefaultSamlClientScopesFromState(resourceName, s)
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

func testAccCheckKeycloakSamlClientHasNoDefaultScopes(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keycloakDefaultClientScopes, err := getDefaultSamlClientScopesFromState(resourceName, s)
		if err != nil {
			return err
		}

		if numberOfDefaultScopes := len(keycloakDefaultClientScopes); numberOfDefaultScopes != 0 {
			return fmt.Errorf("expected client to have no assigned default scopes, but it has %d", numberOfDefaultScopes)
		}

		return nil
	}
}

func testAccCheckKeycloakSamlClientDefaultScopeIsNotAttached(resourceName, clientScope string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keycloakDefaultClientScopes, err := getDefaultSamlClientScopesFromState(resourceName, s)
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

func testKeycloakSamlClientDefaultScopes_basic(client, clientScope string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id

	sign_documents          = false
	sign_assertions         = true
	include_authn_statement = true

	signing_certificate     = file("misc/saml-cert.pem")
	signing_private_key     = file("misc/saml-key.pem")
}

resource "keycloak_saml_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id

	description = "test description"
}

resource "keycloak_saml_client_default_scopes" "default_scopes" {
	realm_id       = data.keycloak_realm.realm.id
	client_id      = keycloak_saml_client.client.id
	default_scopes = [
		"role_list",
		keycloak_saml_client_scope.client_scope.name
	]
}
	`, testAccRealm.Realm, client, clientScope)
}

func testKeycloakSamlClientDefaultScopes_noDefaultScopes(client, clientScope string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id

	sign_documents          = false
	sign_assertions         = true
	include_authn_statement = true

	signing_certificate     = file("misc/saml-cert.pem")
	signing_private_key     = file("misc/saml-key.pem")
}

resource "keycloak_saml_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id

	description = "test description"
}
	`, testAccRealm.Realm, client, clientScope)
}

func testKeycloakSamlClientDefaultScopes_listOfScopes(client, clientScope string, listOfDefaultScopes []string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id

	sign_documents          = false
	sign_assertions         = true
	include_authn_statement = true

	signing_certificate     = file("misc/saml-cert.pem")
	signing_private_key     = file("misc/saml-key.pem")
}

resource "keycloak_saml_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id

	description = "test description"
}

resource "keycloak_saml_client_default_scopes" "default_scopes" {
	realm_id       = data.keycloak_realm.realm.id
	client_id      = keycloak_saml_client.client.id
	default_scopes = %s

	depends_on = ["keycloak_saml_client_scope.client_scope"]
}
	`, testAccRealm.Realm, client, clientScope, arrayOfStringsForTerraformResource(listOfDefaultScopes))
}

func testKeycloakSamlClientDefaultScopes_validationNoClient(client, clientScope string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client_scope" "client_scope" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id

	description = "test description"
}

resource "keycloak_saml_client_default_scopes" "default_scopes" {
	realm_id       = data.keycloak_realm.realm.id
	client_id      = "%s"
	default_scopes = [
		"role_list",
		keycloak_saml_client_scope.client_scope.name
	]
}
	`, testAccRealm.Realm, clientScope, client)
}

func testKeycloakSamlClientDefaultScopes_multipleClientScopes(client string, allClientScopes, attachedClientScopes []string) string {
	var clientScopeResources strings.Builder
	for _, clientScope := range allClientScopes {
		if strings.HasPrefix(clientScope, "terraform") {
			clientScopeResources.WriteString(fmt.Sprintf(`
resource "keycloak_saml_client_scope" "client_scope_%s" {
	name        = "%s"
	realm_id    = data.keycloak_realm.realm.id
}
		`, clientScope, clientScope))
		}
	}

	var attachedClientScopesInterpolated []string
	for _, attachedClientScope := range attachedClientScopes {
		if strings.HasPrefix(attachedClientScope, "terraform") {
			attachedClientScopesInterpolated = append(attachedClientScopesInterpolated, fmt.Sprintf("${keycloak_saml_client_scope.client_scope_%s.name}", attachedClientScope))
		} else {
			attachedClientScopesInterpolated = append(attachedClientScopesInterpolated, attachedClientScope)
		}
	}

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_saml_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id

	sign_documents          = false
	sign_assertions         = true
	include_authn_statement = true

	signing_certificate     = file("misc/saml-cert.pem")
	signing_private_key     = file("misc/saml-key.pem")
}

%s

resource "keycloak_saml_client_default_scopes" "default_scopes" {
	realm_id       = data.keycloak_realm.realm.id
	client_id      = keycloak_saml_client.client.id
	default_scopes = %s
}
	`, testAccRealm.Realm, client, clientScopeResources.String(), arrayOfStringsForTerraformResource(attachedClientScopesInterpolated))
}
