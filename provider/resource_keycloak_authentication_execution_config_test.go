package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakAuthenticationExecutionConfig_basic(t *testing.T) {
	t.Parallel()

	flowAlias := acctest.RandomWithPrefix("tf-acc")
	configAlias := acctest.RandomWithPrefix("tf-acc")
	configProviderOne := acctest.RandomWithPrefix("tf-acc")
	configProviderTwo := acctest.RandomWithPrefix("tf-acc")

	var config1, config2 keycloak.AuthenticationExecutionConfig

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakAuthenticationExecutionConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakAuthenticationExecutionConfig(flowAlias, configAlias, configProviderOne),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationExecutionConfigExists("keycloak_authentication_execution_config.config", &config1),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "realm_id", testAccRealm.Realm),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "alias", configAlias),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "config.%", "1"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "config.defaultProvider", configProviderOne),
				),
			},
			{
				Config: testAccKeycloakAuthenticationExecutionConfig(flowAlias, configAlias, configProviderTwo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationExecutionConfigExists("keycloak_authentication_execution_config.config", &config2),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "realm_id", testAccRealm.Realm),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "alias", configAlias),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "config.%", "1"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "config.defaultProvider", configProviderTwo),
					testAccCheckKeycloakAuthenticationExecutionConfigForceNew(&config1, &config2, false),
				),
			},
		},
	})
}

func TestAccKeycloakAuthenticationExecutionConfig_updateForcesNew(t *testing.T) {
	t.Parallel()

	flowAlias := acctest.RandomWithPrefix("tf-acc")
	configAliasOne := acctest.RandomWithPrefix("tf-acc")
	configAliasTwo := acctest.RandomWithPrefix("tf-acc")
	configProvider := acctest.RandomWithPrefix("tf-acc")

	var config1, config2 keycloak.AuthenticationExecutionConfig

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakAuthenticationExecutionConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakAuthenticationExecutionConfig(flowAlias, configAliasOne, configProvider),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationExecutionConfigExists("keycloak_authentication_execution_config.config", &config1),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "realm_id", testAccRealm.Realm),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "alias", configAliasOne),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "config.%", "1"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "config.defaultProvider", configProvider),
				),
			},
			{
				Config: testAccKeycloakAuthenticationExecutionConfig(flowAlias, configAliasTwo, configProvider),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationExecutionConfigExists("keycloak_authentication_execution_config.config", &config2),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "realm_id", testAccRealm.Realm),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "alias", configAliasTwo),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "config.%", "1"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "config.defaultProvider", configProvider),
					testAccCheckKeycloakAuthenticationExecutionConfigForceNew(&config1, &config2, true),
				),
			},
		},
	})
}

func TestAccKeycloakAuthenticationExecutionConfig_import(t *testing.T) {
	t.Parallel()

	flowAlias := acctest.RandomWithPrefix("tf-acc")
	configAlias := acctest.RandomWithPrefix("tf-acc")
	configProvider := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakAuthenticationExecutionConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakAuthenticationExecutionConfig(flowAlias, configAlias, configProvider),
			},
			{
				ResourceName:      "keycloak_authentication_execution_config.config",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getExecutionConfigImportId("keycloak_authentication_execution_config.config"),
			},
		},
	})
}

func getExecutionConfigImportId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found", resourceName)
		}

		realmId := rs.Primary.Attributes["realm_id"]
		executionId := rs.Primary.Attributes["execution_id"]
		id := rs.Primary.ID

		return fmt.Sprintf("%s/%s/%s", realmId, executionId, id), nil
	}
}

func testAccCheckKeycloakAuthenticationExecutionConfigExists(resourceName string, config *keycloak.AuthenticationExecutionConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		config.RealmId = rs.Primary.Attributes["realm_id"]
		config.ExecutionId = rs.Primary.Attributes["execution_id"]
		config.Id = rs.Primary.ID

		if err := keycloakClient.GetAuthenticationExecutionConfig(testCtx, config); err != nil {
			return fmt.Errorf("error fetching authentication execution config: %v", err)
		}

		return nil
	}
}

func testAccCheckKeycloakAuthenticationExecutionConfigDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "keycloak_authentication_execution_config" {
			continue
		}

		config := &keycloak.AuthenticationExecutionConfig{
			RealmId: rs.Primary.Attributes["realm_id"],
			Id:      rs.Primary.ID,
		}
		if err := keycloakClient.GetAuthenticationExecutionConfig(testCtx, config); err == nil {
			return fmt.Errorf("authentication execution config still exists")
		} else if !keycloak.ErrorIs404(err) {
			return fmt.Errorf("could not fetch authentication execution config: %v", err)
		}
	}

	return nil
}

func testAccCheckKeycloakAuthenticationExecutionConfigForceNew(old, new *keycloak.AuthenticationExecutionConfig, wantNew bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if wantNew {
			if old.Id == new.Id {
				return fmt.Errorf("expecting authentication execution config ID to differ, got %+v and %+v", old, new)
			}
		} else {
			if old.Id != new.Id {
				return fmt.Errorf("expecting authentication execution config ID to be equal, got %+v and %+v", old, new)
			}
		}
		return nil
	}
}

func testAccKeycloakAuthenticationExecutionConfig(flowAlias, configAlias, configProvider string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "flow" {
	realm_id = data.keycloak_realm.realm.id
	alias    = "%s"
}

resource "keycloak_authentication_execution" "execution" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	authenticator     = "identity-provider-redirector"
}

resource "keycloak_authentication_execution_config" "config" {
	realm_id     = data.keycloak_realm.realm.id
	execution_id = keycloak_authentication_execution.execution.id
	alias        = "%s"
	config = {
		defaultProvider = "%s"
	}
}`, testAccRealm.Realm, flowAlias, configAlias, configProvider)
}

// Keycloak >= 26.6.7 / 26.7.3 returns secret config properties masked as "**********".
// The provider must keep the configured value in state so plans stay clean and secret rotation is still detected.
func TestAccKeycloakAuthenticationExecutionConfig_maskedSecret(t *testing.T) {
	t.Parallel()

	flowAlias := acctest.RandomWithPrefix("tf-acc")
	configAlias := acctest.RandomWithPrefix("tf-acc")
	secretOne := acctest.RandomWithPrefix("tf-acc-secret")
	secretTwo := acctest.RandomWithPrefix("tf-acc-secret")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakAuthenticationExecutionConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakAuthenticationExecutionConfig_recaptcha(flowAlias, configAlias, secretOne),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "config.site.key", "site-key"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "config.secret.key", secretOne),
					testAccCheckKeycloakAuthenticationExecutionConfigSecretMasked("keycloak_authentication_execution_config.config", "secret.key", secretOne),
				),
			},
			{
				// no changes -> plan must be empty even though Keycloak returns the masked secret
				Config:   testAccKeycloakAuthenticationExecutionConfig_recaptcha(flowAlias, configAlias, secretOne),
				PlanOnly: true,
			},
			{
				// rotating the secret must still be detected and applied
				Config: testAccKeycloakAuthenticationExecutionConfig_recaptcha(flowAlias, configAlias, secretTwo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("keycloak_authentication_execution_config.config", "config.secret.key", secretTwo),
					testAccCheckKeycloakAuthenticationExecutionConfigSecretMasked("keycloak_authentication_execution_config.config", "secret.key", secretTwo),
				),
			},
		},
	})
}

// testAccCheckKeycloakAuthenticationExecutionConfigSecretMasked asserts that Keycloak returns the given secret config key
// either masked (Keycloak >= 26.6.7 / 26.7.3) or, on older versions that do not mask secrets yet, with the expected value.
func testAccCheckKeycloakAuthenticationExecutionConfigSecretMasked(resourceName, key, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var config keycloak.AuthenticationExecutionConfig
		if err := testAccCheckKeycloakAuthenticationExecutionConfigExists(resourceName, &config)(s); err != nil {
			return err
		}
		value, ok := config.Config[key]
		if !ok {
			return fmt.Errorf("expected config key %s to exist in Keycloak, got %+v", key, config.Config)
		}
		if value != keycloakMaskedSecret && value != expected {
			return fmt.Errorf("expected config key %s to be %q or masked, got %q", key, expected, value)
		}
		return nil
	}
}

func testAccKeycloakAuthenticationExecutionConfig_recaptcha(flowAlias, configAlias, secret string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "flow" {
	realm_id = data.keycloak_realm.realm.id
	alias    = "%s"
}

resource "keycloak_authentication_subflow" "form" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	alias             = "%s-form"
	provider_id       = "form-flow"
	authenticator     = "registration-page-form"
	requirement       = "REQUIRED"
}

resource "keycloak_authentication_execution" "execution" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_subflow.form.alias
	authenticator     = "registration-recaptcha-action"
	requirement       = "REQUIRED"
}

resource "keycloak_authentication_execution_config" "config" {
	realm_id     = data.keycloak_realm.realm.id
	execution_id = keycloak_authentication_execution.execution.id
	alias        = "%s"
	config = {
		"site.key"   = "site-key"
		"secret.key" = "%s"
	}
}`, testAccRealm.Realm, flowAlias, flowAlias, configAlias, secret)
}
