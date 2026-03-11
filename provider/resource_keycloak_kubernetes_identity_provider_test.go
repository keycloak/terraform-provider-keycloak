package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakKubernetesIdentityProvider_basic(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_5)
	t.Parallel()

	kubernetesName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakKubernetesIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakKubernetesIdentityProvider_basic(testAccRealm.Realm, kubernetesName, "https://example.com/issuer"),
				Check:  testAccCheckKeycloakKubernetesIdentityProviderExists("keycloak_kubernetes_identity_provider.kubernetes"),
			},
		},
	})
}

func TestAccKeycloakKubernetesIdentityProvider_insecureIssuer(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_5)
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	realm := &keycloak.Realm{
		Realm:       realmName,
		SslRequired: "none",
	}

	kubernetesName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakKubernetesIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					err := keycloakClient.NewRealm(testCtx, realm)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakKubernetesIdentityProvider_basic(realmName, kubernetesName, "http://example.com/issuer"),
				Check:  testAccCheckKeycloakKubernetesIdentityProviderExists("keycloak_kubernetes_identity_provider.kubernetes"),
			},
		},
	})
}

func testKeycloakKubernetesIdentityProvider_basic(realm, alias, issuer string) string {
	return fmt.Sprintf(`
		data "keycloak_realm" "realm" {
			realm = "%s"
		}

		resource "keycloak_kubernetes_identity_provider" "kubernetes" {
			realm  = data.keycloak_realm.realm.id
			alias  = "%s"
			issuer = "%s"
		}
			`, realm, alias, issuer)
}

func testAccCheckKeycloakKubernetesIdentityProviderExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		idp, err := getKeycloakKubernetesIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		// Kubernetes identity provider should always be hidden on login page
		if idp.HideOnLogin != true {
			return fmt.Errorf("error checking if kubernetes identity provider is hidden on login page: expected true but got %t", idp.HideOnLogin)
		}

		return nil
	}
}

func getKeycloakKubernetesIdentityProviderFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProvider, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm"]
	alias := rs.Primary.Attributes["alias"]

	kubernetes, err := keycloakClient.GetIdentityProvider(testCtx, realm, alias)
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes identity provider config with alias %s: %s", alias, err)
	}

	return kubernetes, nil
}

func testAccCheckKeycloakKubernetesIdentityProviderDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_kubernetes_identity_provider" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm"]

			kubernetes, _ := keycloakClient.GetIdentityProvider(testCtx, realm, id)
			if kubernetes != nil {
				return fmt.Errorf("kubernetes config with id %s still exists", id)
			}
		}

		return nil
	}
}
