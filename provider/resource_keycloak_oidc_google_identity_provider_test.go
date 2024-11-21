package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"regexp"
	"strconv"
	"testing"
)

/*
	note: we cannot use parallel tests for this resource as only one instance of a google identity provider can be created
	for a realm.
*/

func TestAccKeycloakOidcGoogleIdentityProvider_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcGoogleIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcGoogleIdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcGoogleIdentityProviderExists("keycloak_oidc_google_identity_provider.google"),
			},
		},
	})
}

func TestAccKeycloakOidcGoogleIdentityProvider_extraConfig(t *testing.T) {
	customConfigValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcGoogleIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcGoogleIdentityProvider_customConfig("dummyConfig", customConfigValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcGoogleIdentityProviderExists("keycloak_oidc_google_identity_provider.google_custom"),
					testAccCheckKeycloakOidcGoogleIdentityProviderHasCustomConfigValue("keycloak_oidc_google_identity_provider.google_custom", customConfigValue),
				),
			},
		},
	})
}

// ensure that extra_config keys which are covered by top-level attributes are not allowed
func TestAccKeycloakOidcGoogleIdentityProvider_extraConfigInvalid(t *testing.T) {
	customConfigValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcGoogleIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOidcGoogleIdentityProvider_customConfig("syncMode", customConfigValue),
				ExpectError: regexp.MustCompile("extra_config key \"syncMode\" is not allowed"),
			},
		},
	})
}

func TestAccKeycloakOidcGoogleIdentityProvider_createAfterManualDestroy(t *testing.T) {
	var idp = &keycloak.IdentityProvider{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcGoogleIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcGoogleIdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcGoogleIdentityProviderFetch("keycloak_oidc_google_identity_provider.google", idp),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteIdentityProvider(testCtx, idp.Realm, idp.Alias)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOidcGoogleIdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcGoogleIdentityProviderExists("keycloak_oidc_google_identity_provider.google"),
			},
		},
	})
}

func TestAccKeycloakOidcGoogleIdentityProvider_basicUpdateAll(t *testing.T) {
	firstEnabled := randomBool()

	firstOidc := &keycloak.IdentityProvider{
		Alias:   acctest.RandString(10),
		Enabled: firstEnabled,
		Config: &keycloak.IdentityProviderConfig{
			HostedDomain:                "mycompany.com",
			AcceptsPromptNoneForwFrmClt: false,
			ClientId:                    acctest.RandString(10),
			ClientSecret:                acctest.RandString(10),
			GuiOrder:                    strconv.Itoa(acctest.RandIntRange(1, 3)),
			SyncMode:                    randomStringInSlice(syncModes),
		},
	}

	secondOidc := &keycloak.IdentityProvider{
		Alias:   acctest.RandString(10),
		Enabled: !firstEnabled,
		Config: &keycloak.IdentityProviderConfig{
			HostedDomain:                "mycompany.com",
			AcceptsPromptNoneForwFrmClt: false,
			ClientId:                    acctest.RandString(10),
			ClientSecret:                acctest.RandString(10),
			GuiOrder:                    strconv.Itoa(acctest.RandIntRange(1, 3)),
			SyncMode:                    randomStringInSlice(syncModes),
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOidcGoogleIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcGoogleIdentityProvider_basicFromInterface(firstOidc),
				Check:  testAccCheckKeycloakOidcGoogleIdentityProviderExists("keycloak_oidc_google_identity_provider.google"),
			},
			{
				Config: testKeycloakOidcGoogleIdentityProvider_basicFromInterface(secondOidc),
				Check:  testAccCheckKeycloakOidcGoogleIdentityProviderExists("keycloak_oidc_google_identity_provider.google"),
			},
		},
	})
}

func testAccCheckKeycloakOidcGoogleIdentityProviderExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakOidcGoogleIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakOidcGoogleIdentityProviderFetch(resourceName string, idp *keycloak.IdentityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcGoogleIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		idp.Alias = fetchedOidc.Alias
		idp.Realm = fetchedOidc.Realm

		return nil
	}
}

func testAccCheckKeycloakOidcGoogleIdentityProviderHasCustomConfigValue(resourceName, customConfigValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcGoogleIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetchedOidc.Config.ExtraConfig["dummyConfig"].(string) != customConfigValue {
			return fmt.Errorf("expected custom oidc provider to have config with a custom key 'dummyConfig' with a value %s, but value was %s", customConfigValue, fetchedOidc.Config.ExtraConfig["dummyConfig"].(string))
		}

		return nil
	}
}

func testAccCheckKeycloakOidcGoogleIdentityProviderDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_oidc_google_identity_provider" {
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

func getKeycloakOidcGoogleIdentityProviderFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProvider, error) {
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

func testKeycloakOidcGoogleIdentityProvider_basic() string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_google_identity_provider" "google" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"
}
	`, testAccRealm.Realm)
}

func testKeycloakOidcGoogleIdentityProvider_customConfig(configKey, configValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_google_identity_provider" "google_custom" {
	realm             = data.keycloak_realm.realm.id
	provider_id       = "google"
	client_id         = "example_id"
	client_secret     = "example_token"
	extra_config      = {
		%s = "%s"
	}
}
	`, testAccRealm.Realm, configKey, configValue)
}

func testKeycloakOidcGoogleIdentityProvider_basicFromInterface(idp *keycloak.IdentityProvider) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_google_identity_provider" "google" {
	realm             						= data.keycloak_realm.realm.id
	enabled           						= %t
	hosted_domain	  						= "%s"
	accepts_prompt_none_forward_from_client	= %t
	client_id         						= "%s"
	client_secret     						= "%s"
	gui_order                               = %s
	sync_mode                               = "%s"
}
	`, testAccRealm.Realm, idp.Enabled, idp.Config.HostedDomain, idp.Config.AcceptsPromptNoneForwFrmClt, idp.Config.ClientId, idp.Config.ClientSecret, idp.Config.GuiOrder, idp.Config.SyncMode)
}
