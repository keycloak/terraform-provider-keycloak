package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeycloakDataSourceRealmKeys_basic(t *testing.T) {
	t.Parallel()
	dataSourceName := "data.keycloak_realm_keys.test_keys"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakRealmKeysConfig(),
				Check:  testKeycloakRealmKeysCheck_basic(dataSourceName),
			},
		},
	})
}

func TestAccKeycloakDataSourceRealmKeys_filterByAlgorithms(t *testing.T) {
	t.Parallel()
	dataSourceName := "data.keycloak_realm_keys.test_keys"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakRealmKeysConfig_filterByAlgorithms(),
				Check:  testKeycloakRealmKeysCheck_filterByAlgorithms(dataSourceName),
			},
		},
	})
}

func getRealmKeysUsingState(state *terraform.State, resourceName string) (*terraform.ResourceState, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found in TF state: %s ", resourceName)
	}

	return rs, nil
}

func testKeycloakRealmKeysCheck_basic(dataSourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		datasourceState, err := getRealmKeysUsingState(state, dataSourceName)
		if err != nil {
			return err
		}

		if len(datasourceState.Primary.Attributes["keys.#"]) == 0 {
			return fmt.Errorf("no key exists")
		}

		return nil
	}
}

func testKeycloakRealmKeysCheck_filterByAlgorithms(dataSourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		datasourceState, err := getRealmKeysUsingState(state, dataSourceName)
		if err != nil {
			return err
		}

		if len(datasourceState.Primary.Attributes["keys.#"]) == 0 {
			return fmt.Errorf("no key exists")
		}

		algorithm := datasourceState.Primary.Attributes["keys.0.algorithm"]
		if algorithm != "AES" && algorithm != "RS256" {
			return fmt.Errorf("filtering by algorithm returned '%s', but this value was not part of the valid values", algorithm)
		}

		return nil
	}
}

func testAccKeycloakRealmKeysConfig() string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_realm_keys" "test_keys" {
	realm_id = data.keycloak_realm.realm.id
}
`, testAccRealm.Realm)
}

func testAccKeycloakRealmKeysConfig_filterByAlgorithms() string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_realm_keys" "test_keys" {
	realm_id   = data.keycloak_realm.realm.id
	algorithms = ["RS256", "AES"]
}
`, testAccRealm.Realm)
}
