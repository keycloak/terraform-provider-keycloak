package provider

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakRealmLocalizationTexts_basic(t *testing.T) {
	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmLocalizationTextsDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmLocalizationTexts_basic(realmName),
				Check:  testAccCheckKeycloakRealmLocalizationTextsExist("keycloak_realm_localization.realm_localization", "en", map[string]string{"k": "v"}),
			},
		},
	})
}

func TestAccKeycloakRealmLocalizationTexts_empty(t *testing.T) {
	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmLocalizationTextsDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmLocalizationTexts_empty(realmName),
				Check:  testAccCheckKeycloakRealmLocalizationTextsExist("keycloak_realm_localization.realm_localization", "en", map[string]string{}),
			},
		},
	})
}

// Tests creating a realm translation in a realm without localization in a non-default locale
// The translation should exist, but it won't take effect.
func TestAccKeycloakRealmLocalizationTexts_noLocalization(t *testing.T) {
	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmLocalizationTextsDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmLocalizationTexts_noInternationalization(realmName),
				Check:  testAccCheckKeycloakRealmLocalizationTextsExist("keycloak_realm_localization.realm_localization", "de", map[string]string{"k": "v"}),
			},
		},
	})
}

func testAccCheckKeycloakRealmLocalizationTextsDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_localization" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			locale := rs.Primary.Attributes["locale"]

			realmLocalizationTexts, _ := keycloakClient.GetRealmLocalizationTexts(testCtx, realm, locale)
			if realmLocalizationTexts != nil {
				return fmt.Errorf("translation for realm %s", realm)
			}
		}

		return nil
	}
}

func testKeycloakRealmLocalizationTexts_basic(realm string) string {
	return fmt.Sprintf(`
	resource "keycloak_realm" "realm" {
		realm = "%s"
		internationalization {
			supported_locales = [
				"en"
			]
			default_locale    = "en"
		}
	}

	resource "keycloak_realm_localization" "realm_localization" {
		realm_id                          = keycloak_realm.realm.id
		locale  = "en"
		texts = {
			"k": "v"
		}
	}
		`, realm)
}

func testKeycloakRealmLocalizationTexts_empty(realm string) string {
	return fmt.Sprintf(`
	resource "keycloak_realm" "realm" {
		realm = "%s"
		internationalization {
			supported_locales = [
				"en"
			]
			default_locale    = "en"
		}
	}

	resource "keycloak_realm_localization" "realm_localization" {
		realm_id                          = keycloak_realm.realm.id
		locale  = "en"
		texts = {
		}
	}
		`, realm)
}

func testKeycloakRealmLocalizationTexts_noInternationalization(realm string) string {
	return fmt.Sprintf(`
	resource "keycloak_realm" "realm" {
		realm = "%s"
	}

	resource "keycloak_realm_localization" "realm_localization" {
		realm_id                          = keycloak_realm.realm.id
		locale  = "de"
		texts = {
			"k": "v"
		}
	}
		`, realm)
}

func getRealmLocalizationTextsFromState(s *terraform.State, resourceName string) (map[string]string, string, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, "", fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	locale := rs.Primary.Attributes["locale"]

	realmLocalizationTexts, err := keycloakClient.GetRealmLocalizationTexts(testCtx, realm, locale)
	if err != nil {
		return nil, "", fmt.Errorf("error getting realm user profile: %s", err)
	}
	return *realmLocalizationTexts, locale, nil
}

func testAccCheckKeycloakRealmLocalizationTextsExist(resourceName string, expectedLocale string, expectedLocalizationTexts map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		texts, locale, err := getRealmLocalizationTextsFromState(s, resourceName)
		if err != nil {
			return err
		}
		if expectedLocale != locale {
			return fmt.Errorf("assigned and expected texts locale do not match %v != %v", locale, expectedLocale)
		}
		if !reflect.DeepEqual(texts, expectedLocalizationTexts) {
			return fmt.Errorf("assigned and expected realm texts do not match %v != %v", texts, expectedLocalizationTexts)
		}

		return nil
	}
}
