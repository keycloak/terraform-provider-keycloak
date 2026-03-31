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

func TestAccKeycloakOidcOpenshiftV4IdentityProvider_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcOpenshiftV4IdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcOpenshiftV4IdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcOpenshiftV4IdentityProviderExists("keycloak_oidc_openshift_v4_identity_provider.openshift_v4"),
			},
		},
	})
}

func TestAccKeycloakOidcOpenshiftV4IdentityProvider_customAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcOpenshiftV4IdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_openshift_v4_identity_provider" "openshift_v4" {
	realm         = data.keycloak_realm.realm.id
	client_id     = "example_id"
	client_secret = "example_token"
	base_url      = "https://openshift.example.com:8443"

	alias = "example"
}
	`, testAccRealm.Realm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcOpenshiftV4IdentityProviderExists("keycloak_oidc_openshift_v4_identity_provider.openshift_v4"),
					resource.TestCheckResourceAttr("keycloak_oidc_openshift_v4_identity_provider.openshift_v4", "alias", "example"),
				),
			},
		},
	})
}

func TestAccKeycloakOidcOpenshiftV4IdentityProvider_customDisplayName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcOpenshiftV4IdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_openshift_v4_identity_provider" "openshift_v4" {
	realm         = data.keycloak_realm.realm.id
	client_id     = "example_id"
	client_secret = "example_token"
	base_url      = "https://openshift.example.com:8443"

	display_name = "Example OpenShift v4"
}
	`, testAccRealm.Realm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcOpenshiftV4IdentityProviderExists("keycloak_oidc_openshift_v4_identity_provider.openshift_v4"),
					resource.TestCheckResourceAttr("keycloak_oidc_openshift_v4_identity_provider.openshift_v4", "display_name", "Example OpenShift v4"),
				),
			},
		},
	})
}

func TestAccKeycloakOidcOpenshiftV4IdentityProvider_extraConfig(t *testing.T) {
	customConfigValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcOpenshiftV4IdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcOpenshiftV4IdentityProvider_customConfig("dummyConfig", customConfigValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcOpenshiftV4IdentityProviderExists("keycloak_oidc_openshift_v4_identity_provider.openshift_v4_custom"),
					testAccCheckKeycloakOidcOpenshiftV4IdentityProviderHasCustomConfigValue("keycloak_oidc_openshift_v4_identity_provider.openshift_v4_custom", customConfigValue),
				),
			},
		},
	})
}

// ensure that extra_config keys which are covered by top-level attributes are not allowed
func TestAccKeycloakOidcOpenshiftV4IdentityProvider_extraConfigInvalid(t *testing.T) {
	customConfigValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcOpenshiftV4IdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOidcOpenshiftV4IdentityProvider_customConfig("syncMode", customConfigValue),
				ExpectError: regexp.MustCompile("extra_config key \"syncMode\" is not allowed"),
			},
		},
	})
}

func TestAccKeycloakOidcOpenshiftV4IdentityProvider_createAfterManualDestroy(t *testing.T) {
	var idp = &keycloak.IdentityProvider{}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcOpenshiftV4IdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcOpenshiftV4IdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcOpenshiftV4IdentityProviderFetch("keycloak_oidc_openshift_v4_identity_provider.openshift_v4", idp),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteIdentityProvider(testCtx, idp.Realm, idp.Alias)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOidcOpenshiftV4IdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcOpenshiftV4IdentityProviderExists("keycloak_oidc_openshift_v4_identity_provider.openshift_v4"),
			},
		},
	})
}

func TestAccKeycloakOidcOpenshiftV4IdentityProvider_basicUpdateAll(t *testing.T) {
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
			BaseUrl:      "https://openshift.example.com:8443",
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
			BaseUrl:      "https://openshift2.example.com:8443",
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcOpenshiftV4IdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcOpenshiftV4IdentityProvider_basicFromInterface(firstOidc),
				Check:  testAccCheckKeycloakOidcOpenshiftV4IdentityProviderExists("keycloak_oidc_openshift_v4_identity_provider.openshift_v4"),
			},
			{
				Config: testKeycloakOidcOpenshiftV4IdentityProvider_basicFromInterface(secondOidc),
				Check:  testAccCheckKeycloakOidcOpenshiftV4IdentityProviderExists("keycloak_oidc_openshift_v4_identity_provider.openshift_v4"),
			},
		},
	})
}

func testAccCheckKeycloakOidcOpenshiftV4IdentityProviderExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakOidcOpenshiftV4IdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakOidcOpenshiftV4IdentityProviderFetch(resourceName string, idp *keycloak.IdentityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcOpenshiftV4IdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		idp.Alias = fetchedOidc.Alias
		idp.Realm = fetchedOidc.Realm

		return nil
	}
}

func testAccCheckKeycloakOidcOpenshiftV4IdentityProviderHasCustomConfigValue(resourceName, customConfigValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcOpenshiftV4IdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetchedOidc.Config.ExtraConfig["dummyConfig"].(string) != customConfigValue {
			return fmt.Errorf("expected custom oidc provider to have config with a custom key 'dummyConfig' with a value %s, but value was %s", customConfigValue, fetchedOidc.Config.ExtraConfig["dummyConfig"].(string))
		}

		return nil
	}
}

func testAccCheckKeycloakOidcOpenshiftV4IdentityProviderDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_oidc_openshift_v4_identity_provider" {
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

func getKeycloakOidcOpenshiftV4IdentityProviderFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProvider, error) {
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

func testKeycloakOidcOpenshiftV4IdentityProvider_basic() string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_openshift_v4_identity_provider" "openshift_v4" {
	realm         = data.keycloak_realm.realm.id
	client_id     = "example_id"
	client_secret = "example_token"
	base_url      = "https://openshift.example.com:8443"
}
	`, testAccRealm.Realm)
}

func testKeycloakOidcOpenshiftV4IdentityProvider_customConfig(configKey, configValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_openshift_v4_identity_provider" "openshift_v4_custom" {
	realm         = data.keycloak_realm.realm.id
	provider_id   = "openshift-v4"
	client_id     = "example_id"
	client_secret = "example_token"
	base_url      = "https://openshift.example.com:8443"
	extra_config  = {
		%s = "%s"
	}
}
	`, testAccRealm.Realm, configKey, configValue)
}

func testKeycloakOidcOpenshiftV4IdentityProvider_basicFromInterface(idp *keycloak.IdentityProvider) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_openshift_v4_identity_provider" "openshift_v4" {
	realm              = data.keycloak_realm.realm.id
	enabled            = %t
	base_url           = "%s"
	client_id          = "%s"
	client_secret      = "%s"
	gui_order          = %s
	sync_mode          = "%s"
	hide_on_login_page = %t
}
	`, testAccRealm.Realm, idp.Enabled, idp.Config.BaseUrl, idp.Config.ClientId, idp.Config.ClientSecret, idp.Config.GuiOrder, idp.Config.SyncMode, idp.HideOnLogin)
}
