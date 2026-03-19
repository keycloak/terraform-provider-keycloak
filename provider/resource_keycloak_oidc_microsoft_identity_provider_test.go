package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

/*
	note: we cannot use parallel tests for this resource as only one instance of a Microsoft identity provider can be created
	for a realm.
*/

func TestAccKeycloakOidcMicrosoftIdentityProvider_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcMicrosoftIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcMicrosoftIdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcMicrosoftIdentityProviderExists("keycloak_oidc_microsoft_identity_provider.microsoft"),
			},
		},
	})
}

func TestAccKeycloakOidcMicrosoftIdentityProvider_customAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcMicrosoftIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_microsoft_identity_provider" "microsoft" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"

	alias = "example"
}
	`, testAccRealm.Realm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcMicrosoftIdentityProviderExists("keycloak_oidc_microsoft_identity_provider.microsoft"),
					resource.TestCheckResourceAttr("keycloak_oidc_microsoft_identity_provider.microsoft", "alias", "example"),
				),
			},
		},
	})
}

func TestAccKeycloakOidcMicrosoftIdentityProvider_customDisplayName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcMicrosoftIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_microsoft_identity_provider" "microsoft" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"

	display_name = "Example Microsoft"
}
	`, testAccRealm.Realm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcMicrosoftIdentityProviderExists("keycloak_oidc_microsoft_identity_provider.microsoft"),
					resource.TestCheckResourceAttr("keycloak_oidc_microsoft_identity_provider.microsoft", "display_name", "Example Microsoft"),
				),
			},
		},
	})
}

func TestAccKeycloakOidcMicrosoftIdentityProvider_extraConfig(t *testing.T) {
	customConfigValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcMicrosoftIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcMicrosoftIdentityProvider_customConfig("dummyConfig", customConfigValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcMicrosoftIdentityProviderExists("keycloak_oidc_microsoft_identity_provider.microsoft_custom"),
					testAccCheckKeycloakOidcMicrosoftIdentityProviderHasCustomConfigValue("keycloak_oidc_microsoft_identity_provider.microsoft_custom", customConfigValue),
				),
			},
		},
	})
}

// ensure that extra_config keys which are covered by top-level attributes are not allowed
func TestAccKeycloakOidcMicrosoftIdentityProvider_extraConfigInvalid(t *testing.T) {
	customConfigValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcMicrosoftIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOidcMicrosoftIdentityProvider_customConfig("syncMode", customConfigValue),
				ExpectError: regexp.MustCompile("extra_config key \"syncMode\" is not allowed"),
			},
		},
	})
}

func TestAccKeycloakOidcMicrosoftIdentityProvider_linkOrganization(t *testing.T) {

	organizationName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcMicrosoftIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcMicrosoftIdentityProvider_linkOrganization(organizationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcMicrosoftIdentityProviderExists("keycloak_oidc_microsoft_identity_provider.microsoft"),
					testAccCheckKeycloakOidcMicrosoftIdentityProviderLinkOrganization("keycloak_oidc_microsoft_identity_provider.microsoft"),
				),
			},
		},
	})
}

func TestAccKeycloakOidcMicrosoftIdentityProvider_createAfterManualDestroy(t *testing.T) {
	var idp = &keycloak.IdentityProvider{}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcMicrosoftIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcMicrosoftIdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcMicrosoftIdentityProviderFetch("keycloak_oidc_microsoft_identity_provider.microsoft", idp),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteIdentityProvider(testCtx, idp.Realm, idp.Alias)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOidcMicrosoftIdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcMicrosoftIdentityProviderExists("keycloak_oidc_microsoft_identity_provider.microsoft"),
			},
		},
	})
}

func TestAccKeycloakOidcMicrosoftIdentityProvider_basicUpdateAll(t *testing.T) {
	firstEnabled := randomBool()
	firstHideOnLogin := randomBool()

	firstOidc := &keycloak.IdentityProvider{
		Alias:       acctest.RandString(10),
		Enabled:     firstEnabled,
		HideOnLogin: firstHideOnLogin,
		Config: &keycloak.IdentityProviderConfig{
			ClientId:     acctest.RandString(10),
			ClientSecret: acctest.RandString(10),
			GuiOrder:     strconv.Itoa(acctest.RandIntRange(1, 3)),
			SyncMode:     randomStringInSlice(syncModes),
		},
	}

	secondOidc := &keycloak.IdentityProvider{
		Alias:       acctest.RandString(10),
		Enabled:     !firstEnabled,
		HideOnLogin: !firstHideOnLogin,
		Config: &keycloak.IdentityProviderConfig{
			ClientId:     acctest.RandString(10),
			ClientSecret: acctest.RandString(10),
			GuiOrder:     strconv.Itoa(acctest.RandIntRange(1, 3)),
			SyncMode:     randomStringInSlice(syncModes),
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcMicrosoftIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcMicrosoftIdentityProvider_basicFromInterface(firstOidc),
				Check:  testAccCheckKeycloakOidcMicrosoftIdentityProviderExists("keycloak_oidc_microsoft_identity_provider.microsoft"),
			},
			{
				Config: testKeycloakOidcMicrosoftIdentityProvider_basicFromInterface(secondOidc),
				Check:  testAccCheckKeycloakOidcMicrosoftIdentityProviderExists("keycloak_oidc_microsoft_identity_provider.microsoft"),
			},
		},
	})
}

func testAccCheckKeycloakOidcMicrosoftIdentityProviderExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakOidcMicrosoftIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakOidcMicrosoftIdentityProviderFetch(resourceName string, idp *keycloak.IdentityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcMicrosoftIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		idp.Alias = fetchedOidc.Alias
		idp.Realm = fetchedOidc.Realm

		return nil
	}
}

func testAccCheckKeycloakOidcMicrosoftIdentityProviderHasCustomConfigValue(resourceName, customConfigValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcMicrosoftIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetchedOidc.Config.ExtraConfig["dummyConfig"].(string) != customConfigValue {
			return fmt.Errorf("expected custom oidc provider to have config with a custom key 'dummyConfig' with a value %s, but value was %s", customConfigValue, fetchedOidc.Config.ExtraConfig["dummyConfig"].(string))
		}

		return nil
	}
}

func testAccCheckKeycloakOidcMicrosoftIdentityProviderLinkOrganization(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcMicrosoftIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetchedOidc.OrganizationId == "" {
			return fmt.Errorf("expected custom oidc provider to be linked with an organization, but it was not")
		}

		return nil
	}
}

func testAccCheckKeycloakOidcMicrosoftIdentityProviderDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_oidc_microsoft_identity_provider" {
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

func getKeycloakOidcMicrosoftIdentityProviderFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProvider, error) {
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

func testKeycloakOidcMicrosoftIdentityProvider_basic() string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_microsoft_identity_provider" "microsoft" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"
}
	`, testAccRealm.Realm)
}

func testKeycloakOidcMicrosoftIdentityProvider_customConfig(configKey, configValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_microsoft_identity_provider" "microsoft_custom" {
	realm             = data.keycloak_realm.realm.id
	provider_id       = "microsoft"
	client_id         = "example_id"
	client_secret     = "example_token"
	extra_config      = {
		%s = "%s"
	}
}
	`, testAccRealm.Realm, configKey, configValue)
}

func testKeycloakOidcMicrosoftIdentityProvider_basicFromInterface(idp *keycloak.IdentityProvider) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_microsoft_identity_provider" "microsoft" {
	realm             						= data.keycloak_realm.realm.id
	enabled           						= %t
	client_id         						= "%s"
	client_secret     						= "%s"
	gui_order                     = %s
	sync_mode                     = "%s"
	hide_on_login_page            = %t
}
	`, testAccRealm.Realm, idp.Enabled, idp.Config.ClientId, idp.Config.ClientSecret, idp.Config.GuiOrder, idp.Config.SyncMode, idp.HideOnLogin)
}

func testKeycloakOidcMicrosoftIdentityProvider_linkOrganization(organizationName string) string {
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

resource "keycloak_oidc_microsoft_identity_provider" "microsoft" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"

	organization_id   				= keycloak_organization.org.id
	org_domain		  				= "example.com"
	org_redirect_mode_email_matches = true
}
	`, testAccRealm.Realm, organizationName)
}
