package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// socialIdentityProvidersWithClientSecret lists the identity provider resources that use
// identityProviderClientSecretSchema, with the arguments each one requires besides client_id.
var socialIdentityProvidersWithClientSecret = []struct {
	resourceType string
	extraConfig  string
}{
	{resourceType: "keycloak_oidc_facebook_identity_provider"},
	{resourceType: "keycloak_oidc_github_identity_provider"},
	{resourceType: "keycloak_oidc_google_identity_provider"},
	{resourceType: "keycloak_oidc_microsoft_identity_provider"},
	{resourceType: "keycloak_oidc_openshift_v4_identity_provider", extraConfig: `base_url = "https://openshift.example.com"`},
}

func TestAccKeycloakSocialIdentityProvider_clientSecretWriteOnly(t *testing.T) {
	for _, idp := range socialIdentityProvidersWithClientSecret {
		t.Run(idp.resourceType, func(t *testing.T) {
			alias := acctest.RandomWithPrefix("tf-acc")
			resourceName := idp.resourceType + ".idp"
			secret := `client_secret_wo = "$${vault.secret-1}"
	client_secret_wo_version = "1"`
			rotatedSecret := `client_secret_wo = "$${vault.secret-2}"
	client_secret_wo_version = "2"`

			resource.Test(t, resource.TestCase{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				PreCheck:                 func() { testAccPreCheck(t) },
				CheckDestroy:             testAccCheckKeycloakSocialIdentityProviderDestroy(idp.resourceType),
				Steps: []resource.TestStep{
					{
						Config: testKeycloakSocialIdentityProvider_clientSecret(idp.resourceType, alias, idp.extraConfig, secret, "openid"),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckKeycloakSocialIdentityProviderClientSecret(resourceName, "${vault.secret-1}"),
							resource.TestCheckNoResourceAttr(resourceName, "client_secret"),
							resource.TestCheckNoResourceAttr(resourceName, "client_secret_wo"),
							resource.TestCheckResourceAttr(resourceName, "client_secret_wo_version", "1"),
						),
					},
					{
						// An update that does not change client_secret_wo_version keeps the secret in Keycloak.
						Config: testKeycloakSocialIdentityProvider_clientSecret(idp.resourceType, alias, idp.extraConfig, secret, "openid email"),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckKeycloakSocialIdentityProviderClientSecret(resourceName, "${vault.secret-1}"),
							resource.TestCheckResourceAttr(resourceName, "default_scopes", "openid email"),
							resource.TestCheckResourceAttr(resourceName, "client_secret_wo_version", "1"),
						),
					},
					{
						Config:   testKeycloakSocialIdentityProvider_clientSecret(idp.resourceType, alias, idp.extraConfig, secret, "openid email"),
						PlanOnly: true,
					},
					{
						Config: testKeycloakSocialIdentityProvider_clientSecret(idp.resourceType, alias, idp.extraConfig, rotatedSecret, "openid email"),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckKeycloakSocialIdentityProviderClientSecret(resourceName, "${vault.secret-2}"),
							resource.TestCheckNoResourceAttr(resourceName, "client_secret"),
							resource.TestCheckResourceAttr(resourceName, "client_secret_wo_version", "2"),
						),
					},
				},
			})
		})
	}
}

// Moving an existing identity provider from client_secret to client_secret_wo sets the
// secret through the write-only argument and removes it from the state.
func TestAccKeycloakSocialIdentityProvider_clientSecretToWriteOnly(t *testing.T) {
	for _, idp := range socialIdentityProvidersWithClientSecret {
		t.Run(idp.resourceType, func(t *testing.T) {
			alias := acctest.RandomWithPrefix("tf-acc")
			resourceName := idp.resourceType + ".idp"

			resource.Test(t, resource.TestCase{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				PreCheck:                 func() { testAccPreCheck(t) },
				CheckDestroy:             testAccCheckKeycloakSocialIdentityProviderDestroy(idp.resourceType),
				Steps: []resource.TestStep{
					{
						Config: testKeycloakSocialIdentityProvider_clientSecret(idp.resourceType, alias, idp.extraConfig, `client_secret = "$${vault.secret-1}"`, "openid"),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckKeycloakSocialIdentityProviderClientSecret(resourceName, "${vault.secret-1}"),
							resource.TestCheckResourceAttr(resourceName, "client_secret", "${vault.secret-1}"),
						),
					},
					{
						Config: testKeycloakSocialIdentityProvider_clientSecret(idp.resourceType, alias, idp.extraConfig, `client_secret_wo = "$${vault.secret-2}"
	client_secret_wo_version = "1"`, "openid"),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckKeycloakSocialIdentityProviderClientSecret(resourceName, "${vault.secret-2}"),
							// The SDK keeps the removed argument in the state as an empty string.
							resource.TestCheckResourceAttr(resourceName, "client_secret", ""),
							resource.TestCheckResourceAttr(resourceName, "client_secret_wo_version", "1"),
						),
					},
				},
			})
		})
	}
}

// client_secret_wo can come from a value that is only known during apply.
func TestAccKeycloakSocialIdentityProvider_clientSecretWriteOnlyFromComputedValue(t *testing.T) {
	for _, idp := range socialIdentityProvidersWithClientSecret {
		t.Run(idp.resourceType, func(t *testing.T) {
			alias := acctest.RandomWithPrefix("tf-acc")
			resourceName := idp.resourceType + ".idp"
			secret := fmt.Sprintf(`client_secret_wo = keycloak_openid_client.secret_source.client_secret
	client_secret_wo_version = "1"
}

resource "keycloak_openid_client" "secret_source" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s-source"
	access_type = "CONFIDENTIAL"`, alias)

			resource.Test(t, resource.TestCase{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				PreCheck:                 func() { testAccPreCheck(t) },
				CheckDestroy:             testAccCheckKeycloakSocialIdentityProviderDestroy(idp.resourceType),
				Steps: []resource.TestStep{
					{
						Config: testKeycloakSocialIdentityProvider_clientSecret(idp.resourceType, alias, idp.extraConfig, secret, "openid"),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckKeycloakSocialIdentityProviderClientSecret(resourceName, ""),
							resource.TestCheckNoResourceAttr(resourceName, "client_secret"),
						),
					},
				},
			})
		})
	}
}

func TestAccKeycloakSocialIdentityProvider_clientSecretValidation(t *testing.T) {
	for _, idp := range socialIdentityProvidersWithClientSecret {
		t.Run(idp.resourceType, func(t *testing.T) {
			alias := acctest.RandomWithPrefix("tf-acc")

			resource.Test(t, resource.TestCase{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				PreCheck:                 func() { testAccPreCheck(t) },
				Steps: []resource.TestStep{
					{
						Config:      testKeycloakSocialIdentityProvider_clientSecret(idp.resourceType, alias, idp.extraConfig, "", "openid"),
						ExpectError: regexp.MustCompile(`one of\s+` + "`client_secret,client_secret_wo`" + `\s+must be\s+specified`),
					},
					{
						Config: testKeycloakSocialIdentityProvider_clientSecret(idp.resourceType, alias, idp.extraConfig, `client_secret = "secret-1"
	client_secret_wo = "secret-2"
	client_secret_wo_version = "1"`, "openid"),
						ExpectError: regexp.MustCompile(`conflicts with`),
					},
					{
						Config:      testKeycloakSocialIdentityProvider_clientSecret(idp.resourceType, alias, idp.extraConfig, `client_secret_wo = "secret-2"`, "openid"),
						ExpectError: regexp.MustCompile(`Missing required argument`),
					},
				},
			})
		})
	}
}

// testAccCheckKeycloakSocialIdentityProviderClientSecret checks the client secret stored in Keycloak.
// Keycloak masks secrets but returns vault references as they are, so with a vault reference
// the exact value is compared. An empty want only checks that a secret is set.
func testAccCheckKeycloakSocialIdentityProviderClientSecret(resourceName, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		idp, err := keycloakClient.GetIdentityProvider(testCtx, rs.Primary.Attributes["realm"], rs.Primary.Attributes["alias"])
		if err != nil {
			return err
		}
		if idp.Config.ClientSecret == "" {
			return fmt.Errorf("expected a client secret on identity provider %s, but it was empty", idp.Alias)
		}
		if want != "" && idp.Config.ClientSecret != want {
			return fmt.Errorf("expected client secret %q on identity provider %s, got %q", want, idp.Alias, idp.Config.ClientSecret)
		}

		return nil
	}
}

func testAccCheckKeycloakSocialIdentityProviderDestroy(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}

			idp, _ := keycloakClient.GetIdentityProvider(testCtx, rs.Primary.Attributes["realm"], rs.Primary.ID)
			if idp != nil {
				return fmt.Errorf("%s with id %s still exists", resourceType, rs.Primary.ID)
			}
		}

		return nil
	}
}

func testKeycloakSocialIdentityProvider_clientSecret(resourceType, alias, extraConfig, clientSecretConfig, defaultScopes string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "%s" "idp" {
	realm          = data.keycloak_realm.realm.id
	alias          = "%s"
	client_id      = "example_id"
	default_scopes = "%s"
	%s
	%s
}
	`, testAccRealm.Realm, resourceType, alias, defaultScopes, extraConfig, clientSecretConfig)
}
