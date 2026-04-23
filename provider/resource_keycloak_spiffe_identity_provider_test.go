package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakSpiffeIdentityProvider_basic(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_5)
	t.Parallel()

	spiffeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakSpiffeIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakSpiffeIdentityProvider_basic(testAccRealm.Realm, spiffeName, "spiffe://example.org", "https://example.com/bundle"),
				Check:  testAccCheckKeycloakSpiffeIdentityProviderExists("keycloak_spiffe_identity_provider.spiffe"),
			},
		},
	})
}

func TestAccKeycloakSpiffeIdentityProvider_insecureBundleEndpoint(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_5)
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	realm := &keycloak.Realm{
		Realm:       realmName,
		SslRequired: "none",
	}

	spiffeName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakSpiffeIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					err := keycloakClient.NewRealm(testCtx, realm)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakSpiffeIdentityProvider_basic(realmName, spiffeName, "spiffe://example.org", "http://example.com/bundle"),
				Check:  testAccCheckKeycloakSpiffeIdentityProviderExists("keycloak_spiffe_identity_provider.spiffe"),
			},
		},
	})
}

func testKeycloakSpiffeIdentityProvider_basic(realm, alias, trustDomain, bundleEndpoint string) string {
	return fmt.Sprintf(`
		data "keycloak_realm" "realm" {
			realm = "%s"
		}

		resource "keycloak_spiffe_identity_provider" "spiffe" {
			realm           = data.keycloak_realm.realm.id
			alias           = "%s"
			trust_domain    = "%s"
			bundle_endpoint = "%s"
		}
			`, realm, alias, trustDomain, bundleEndpoint)
}

func testAccCheckKeycloakSpiffeIdentityProviderExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		idp, err := getKeycloakSpiffeIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		// SPIFFE identity provider should always be hidden on login page
		if idp.HideOnLogin != true {
			return fmt.Errorf("error checking if spiffe identity provider is hidden on login page: expected true but got %t", idp.HideOnLogin)
		}

		return nil
	}
}

func getKeycloakSpiffeIdentityProviderFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProvider, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm"]
	alias := rs.Primary.Attributes["alias"]

	spiffe, err := keycloakClient.GetIdentityProvider(testCtx, realm, alias)
	if err != nil {
		return nil, fmt.Errorf("error getting spiffe identity provider config with alias %s: %s", alias, err)
	}

	return spiffe, nil
}

func testAccCheckKeycloakSpiffeIdentityProviderDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_spiffe_identity_provider" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm"]

			spiffe, _ := keycloakClient.GetIdentityProvider(testCtx, realm, id)
			if spiffe != nil {
				return fmt.Errorf("spiffe config with id %s still exists", id)
			}
		}

		return nil
	}
}
