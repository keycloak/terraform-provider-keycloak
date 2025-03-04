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

func TestAccKeycloakRealmTranslation_basic(t *testing.T) {
	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmTranslationsDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmTranslation_basic(realmName),
				Check:  testAccCheckKeycloakRealmTranslationsExist("keycloak_realm_translation.realm_translation", "en", map[string]string{"k": "v"}),
			},
		},
	})
}

func TestAccKeycloakRealmTranslation_empty(t *testing.T) {
	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmTranslationsDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmTranslation_empty(realmName),
				Check:  testAccCheckKeycloakRealmTranslationsExist("keycloak_realm_translation.realm_translation", "en", map[string]string{}),
			},
		},
	})
}

// Tests creating a realm translation in a realm without localization in a non-default locale
// The translation should exist, but it won't take effect.
func TestAccKeycloakRealmTranslation_noLocalization(t *testing.T) {
	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmTranslationsDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmTranslation_noInternationalization(realmName),
				Check:  testAccCheckKeycloakRealmTranslationsExist("keycloak_realm_translation.realm_translation", "de", map[string]string{"k": "v"}),
			},
		},
	})
}

func testAccCheckKeycloakRealmTranslationsDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_translation" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			locale := rs.Primary.Attributes["locale"]

			realmTranslation, _ := keycloakClient.GetRealmTranslations(testCtx, realm, locale)
			if realmTranslation != nil {
				return fmt.Errorf("translation for realm %s", realm)
			}
		}

		return nil
	}
}

func testKeycloakRealmTranslation_basic(realm string) string {
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

	resource "keycloak_realm_translation" "realm_translation" {
		realm_id                          = keycloak_realm.realm.id
		locale  = "en"
		translations = {
			"k": "v"
		}
	}
		`, realm)
}

func testKeycloakRealmTranslation_empty(realm string) string {
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

	resource "keycloak_realm_translation" "realm_translation" {
		realm_id                          = keycloak_realm.realm.id
		locale  = "en"
		translations = {
		}
	}
		`, realm)
}

func testKeycloakRealmTranslation_noInternationalization(realm string) string {
	return fmt.Sprintf(`
	resource "keycloak_realm" "realm" {
		realm = "%s"
	}

	resource "keycloak_realm_translation" "realm_translation" {
		realm_id                          = keycloak_realm.realm.id
		locale  = "de"
		translations = {
			"k": "v"
		}
	}
		`, realm)
}

func getRealmTranslationFromState(s *terraform.State, resourceName string) (map[string]string, string, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, "", fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	locale := rs.Primary.Attributes["locale"]

	realmTranslations, err := keycloakClient.GetRealmTranslations(testCtx, realm, locale)
	if err != nil {
		return nil, "", fmt.Errorf("error getting realm user profile: %s", err)
	}
	return *realmTranslations, locale, nil
}

func testAccCheckKeycloakRealmTranslationsExist(resourceName string, expectedLocale string, expectedTranslations map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		translations, locale, err := getRealmTranslationFromState(s, resourceName)
		if err != nil {
			return err
		}
		if expectedLocale != locale {
			return fmt.Errorf("assigned and expected translations locale do not match %v != %v", locale, expectedLocale)
		}
		if !reflect.DeepEqual(translations, expectedTranslations) {
			return fmt.Errorf("assigned and expected realm translations do not match %v != %v", translations, expectedTranslations)
		}

		return nil
	}
}
