package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
)

/*
	note: we cannot use parallel tests for this resource as only one instance of a Facebook identity provider can be created
	for a realm.
*/

func TestAccKeycloakOidcFacebookIdentityProvider_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcFacebookIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcFacebookIdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcFacebookIdentityProviderExists("keycloak_oidc_facebook_identity_provider.facebook"),
			},
		},
	})
}

func TestAccKeycloakOidcFacebookIdentityProvider_customAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcFacebookIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_facebook_identity_provider" "facebook" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"

	alias = "example"
}
	`, testAccRealm.Realm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcFacebookIdentityProviderExists("keycloak_oidc_facebook_identity_provider.facebook"),
					resource.TestCheckResourceAttr("keycloak_oidc_facebook_identity_provider.facebook", "alias", "example"),
				),
			},
		},
	})
}

func TestAccKeycloakOidcFacebookIdentityProvider_customDisplayName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcFacebookIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_facebook_identity_provider" "facebook" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"

	display_name = "Example Facebook"
}
	`, testAccRealm.Realm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcFacebookIdentityProviderExists("keycloak_oidc_facebook_identity_provider.facebook"),
					resource.TestCheckResourceAttr("keycloak_oidc_facebook_identity_provider.facebook", "display_name", "Example Facebook"),
				),
			},
		},
	})
}

func TestAccKeycloakOidcFacebookIdentityProvider_extraConfig(t *testing.T) {
	customConfigValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcFacebookIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcFacebookIdentityProvider_customConfig("dummyConfig", customConfigValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcFacebookIdentityProviderExists("keycloak_oidc_facebook_identity_provider.facebook_custom"),
					testAccCheckKeycloakOidcFacebookIdentityProviderHasCustomConfigValue("keycloak_oidc_facebook_identity_provider.facebook_custom", customConfigValue),
				),
			},
		},
	})
}

// ensure that extra_config keys which are covered by top-level attributes are not allowed
func TestAccKeycloakOidcFacebookIdentityProvider_extraConfigInvalid(t *testing.T) {
	customConfigValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcFacebookIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOidcFacebookIdentityProvider_customConfig("syncMode", customConfigValue),
				ExpectError: regexp.MustCompile("extra_config key \"syncMode\" is not allowed"),
			},
		},
	})
}

func TestAccKeycloakOidcFacebookIdentityProvider_linkOrganization(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26)

	organizationName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcFacebookIdentityProvider_linkOrganization(organizationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcFacebookIdentityProviderExists("keycloak_oidc_facebook_identity_provider.facebook"),
					testAccCheckKeycloakOidcFacebookIdentityProviderLinkOrganization("keycloak_oidc_facebook_identity_provider.facebook"),
				),
			},
		},
	})
}

func TestAccKeycloakOidcFacebookIdentityProvider_createAfterManualDestroy(t *testing.T) {
	var idp = &keycloak.IdentityProvider{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcFacebookIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcFacebookIdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcFacebookIdentityProviderFetch("keycloak_oidc_facebook_identity_provider.facebook", idp),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteIdentityProvider(testCtx, idp.Realm, idp.Alias)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOidcFacebookIdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcFacebookIdentityProviderExists("keycloak_oidc_facebook_identity_provider.facebook"),
			},
		},
	})
}

func TestAccKeycloakOidcFacebookIdentityProvider_basicUpdateAll(t *testing.T) {
	firstEnabled := randomBool()
	firstHideOnLogin := randomBool()

	firstOidc := &keycloak.IdentityProvider{
		Alias:       acctest.RandString(10),
		Enabled:     firstEnabled,
		HideOnLogin: firstHideOnLogin,
		Config: &keycloak.IdentityProviderConfig{
			FetchedFields:               "picture",
			AcceptsPromptNoneForwFrmClt: false,
			ClientId:                    acctest.RandString(10),
			ClientSecret:                acctest.RandString(10),
			GuiOrder:                    strconv.Itoa(acctest.RandIntRange(1, 3)),
			SyncMode:                    randomStringInSlice(syncModes),
			HideOnLoginPage:             types.KeycloakBoolQuoted(firstHideOnLogin),
		},
	}

	secondOidc := &keycloak.IdentityProvider{
		Alias:       acctest.RandString(10),
		Enabled:     !firstEnabled,
		HideOnLogin: !firstHideOnLogin,
		Config: &keycloak.IdentityProviderConfig{
			FetchedFields:               "picture",
			AcceptsPromptNoneForwFrmClt: false,
			ClientId:                    acctest.RandString(10),
			ClientSecret:                acctest.RandString(10),
			GuiOrder:                    strconv.Itoa(acctest.RandIntRange(1, 3)),
			SyncMode:                    randomStringInSlice(syncModes),
			HideOnLoginPage:             types.KeycloakBoolQuoted(!firstHideOnLogin),
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcFacebookIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcFacebookIdentityProvider_basicFromInterface(firstOidc),
				Check:  testAccCheckKeycloakOidcFacebookIdentityProviderExists("keycloak_oidc_facebook_identity_provider.facebook"),
			},
			{
				Config: testKeycloakOidcFacebookIdentityProvider_basicFromInterface(secondOidc),
				Check:  testAccCheckKeycloakOidcFacebookIdentityProviderExists("keycloak_oidc_facebook_identity_provider.facebook"),
			},
		},
	})
}

func testAccCheckKeycloakOidcFacebookIdentityProviderExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakOidcFacebookIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakOidcFacebookIdentityProviderFetch(resourceName string, idp *keycloak.IdentityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcFacebookIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		idp.Alias = fetchedOidc.Alias
		idp.Realm = fetchedOidc.Realm

		return nil
	}
}

func testAccCheckKeycloakOidcFacebookIdentityProviderHasCustomConfigValue(resourceName, customConfigValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcFacebookIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetchedOidc.Config.ExtraConfig["dummyConfig"].(string) != customConfigValue {
			return fmt.Errorf("expected custom oidc provider to have config with a custom key 'dummyConfig' with a value %s, but value was %s", customConfigValue, fetchedOidc.Config.ExtraConfig["dummyConfig"].(string))
		}

		return nil
	}
}

func testAccCheckKeycloakOidcFacebookIdentityProviderLinkOrganization(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetchedOidc.OrganizationId == "" {
			return fmt.Errorf("expected custom oidc provider to be linked with an organization, but it was not")
		}

		return nil
	}
}

func testAccCheckKeycloakOidcFacebookIdentityProviderDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_oidc_facebook_identity_provider" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm"]

			idp, _ := keycloakClient.GetIdentityProvider(testCtx, realm, id)
			if idp != nil {
				return fmt.Errorf("oidc config with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakOidcFacebookIdentityProviderFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProvider, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm"]
	alias := rs.Primary.Attributes["alias"]

	idp, err := keycloakClient.GetIdentityProvider(testCtx, realm, alias)
	if err != nil {
		return nil, fmt.Errorf("error getting oidc identity provider config with alias %s: %s", alias, err)
	}

	return idp, nil
}

func testKeycloakOidcFacebookIdentityProvider_basic() string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_facebook_identity_provider" "facebook" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"
}
	`, testAccRealm.Realm)
}

func testKeycloakOidcFacebookIdentityProvider_customConfig(configKey, configValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_facebook_identity_provider" "facebook_custom" {
	realm             = data.keycloak_realm.realm.id
	provider_id       = "facebook"
	client_id         = "example_id"
	client_secret     = "example_token"
	extra_config      = {
		%s = "%s"
	}
}
	`, testAccRealm.Realm, configKey, configValue)
}

func testKeycloakOidcFacebookIdentityProvider_basicFromInterface(idp *keycloak.IdentityProvider) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_facebook_identity_provider" "facebook" {
	realm                                   = data.keycloak_realm.realm.id
	enabled                                 = %t
	fetched_fields                          = "%s"
	accepts_prompt_none_forward_from_client	= %t
	client_id                               = "%s"
	client_secret                           = "%s"
	gui_order                               = %s
	sync_mode                               = "%s"
	hide_on_login_page                      = %t
}
	`, testAccRealm.Realm, idp.Enabled, idp.Config.FetchedFields, idp.Config.AcceptsPromptNoneForwFrmClt, idp.Config.ClientId, idp.Config.ClientSecret, idp.Config.GuiOrder, idp.Config.SyncMode, bool(idp.Config.HideOnLoginPage))
}

func testKeycloakOidcFacebookIdentityProvider_linkOrganization(organizationName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "org" {
	realm   = data.keycloak_realm.realm.id
	name    = "%s"
	enabled = true

	domain {
		name     = "example.com"
		verified = true
 	}
}

resource "keycloak_oidc_facebook_identity_provider" "facebook" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"

	organization_id                 = keycloak_organization.org.id
	org_domain                      = "example.com"
	org_redirect_mode_email_matches = true
}
	`, testAccRealm.Realm, organizationName)
}
