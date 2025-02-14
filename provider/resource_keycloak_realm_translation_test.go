package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func testAccCheckKeycloakRealmTranslationsDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_translation" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			language := rs.Primary.Attributes["language"]

			realmTranslation, _ := keycloakClient.GetRealmTranslations(testCtx, realm, language)
			if realmTranslation != nil {
				return fmt.Errorf("translation for realm %s", realm)
			}
		}

		return nil
	}
}

// func TestAccKeycloakRealmTranslations_Create(t *testing.T) {
// 	skipIfVersionIsGreaterThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_24)

// 	realmName := acctest.RandomWithPrefix("tf-acc")

// 	resource.Test(t, resource.TestCase{
// 		ProviderFactories: testAccProviderFactories,
// 		PreCheck:          func() { testAccPreCheck(t) },
// 		CheckDestroy:      testAccCheckKeycloakRealmTranslationsDestroy(),
// 		Steps: []resource.TestStep{
// 			{
// 				Config:      testKeycloakRealmUserProfile_userProfileEnabledNotSet(realmName),
// 				ExpectError: regexp.MustCompile("User Profile is disabled"),
// 			},
// 		},
// 	})
// }

func testKeycloakRealmTranslation_template(realm string) string {
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
		language  = "en"
		translations = {
			"k": "v"
		}
	}
		`, realm)
}

func getRealmTranslationFromState(s *terraform.State, resourceName string) (map[string]string, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	language := rs.Primary.Attributes["language"]

	realmTranslations, err := keycloakClient.GetRealmTranslations(testCtx, realm, language)
	if err != nil {
		return nil, fmt.Errorf("error getting realm user profile: %s", err)
	}
	fmt.Println("GETTING REALM TRANSLATION FROM STATE")
	fmt.Printf("Translations: %s", realmTranslations)
	return *realmTranslations, nil
}

func testAccCheckKeycloakRealmTranslationeExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getRealmTranslationFromState(s, resourceName)
		if err != nil {
			fmt.Println("Error!!!")
			return err
		}

		return nil
	}
}

func TestAccKeycloakRealmTranslation_basicEmpty(t *testing.T) {
	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmTranslationsDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmTranslation_template(realmName),
				Check:  testAccCheckKeycloakRealmTranslationeExists("keycloak_realm_translation.realm_translation"),
			},
		},
	})
}
