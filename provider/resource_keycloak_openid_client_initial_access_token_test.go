package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mrparkers/terraform-provider-keycloak/keycloak"
	"strings"
	"testing"
)

func TestAccKeycloakOpenidClientsInitialAccessToken_basic(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientsInitialAccessTokenDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientInitialAccessToken_basic(realmName),
				Check:  testAccCheckOpenidClientInitialAccessTokenExists("keycloak_openid_client_initial_access_token.test_initial_access_token"),
			},
		},
	})
}

func testAccCheckOpenidClientInitialAccessTokenExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getOpenidClientInitialAccessTokenState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func getOpenidClientInitialAccessTokenState(s *terraform.State, resourceName string) (*keycloak.OpenidClientInitialAccessToken, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	id := rs.Primary.ID

	token, err := keycloakClient.GetClientInitialAccessToken(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting realm events config: %s", err)
	}

	return token, nil
}

func testAccCheckKeycloakOpenidClientsInitialAccessTokenDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client_initial_access_token" || strings.HasPrefix(name, "data") {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			role, _ := keycloakClient.GetClientInitialAccessToken(testCtx, realm, id)
			if role != nil {
				return fmt.Errorf("%s with id %s still exists", name, id)
			}
		}

		return nil
	}
}

func testKeycloakOpenidClientInitialAccessToken_basic(realm string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
realm = "%s"
}
resource "keycloak_openid_client_initial_access_token" "test_initial_access_token" {
  realm_id = keycloak_realm.realm.id
  token_count = 2
  expiration = 345600
}
	`, realm)
}
