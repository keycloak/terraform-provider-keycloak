package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeycloakDataSourceAuthenticationFlow_basic(t *testing.T) {
	t.Parallel()

	alias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakAuthenticationFlow_basic(alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationFlowExists("keycloak_authentication_flow.flow"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_flow.flow", "id", "data.keycloak_authentication_flow.flow", "id"),
					testAccCheckDataKeycloakAuthenticationFlow("data.keycloak_authentication_flow.flow"),
				),
			},
		},
	})
}

func testAccCheckDataKeycloakAuthenticationFlow(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmID := rs.Primary.Attributes["realm_id"]

		authenticationFlow, err := keycloakClient.GetAuthenticationFlow(testCtx, realmID, id)
		if err != nil {
			return err
		}

		if authenticationFlow.Id != id {
			return fmt.Errorf("expected authenticationFlow with ID %s but got %s", id, authenticationFlow.Id)
		}

		return nil
	}
}

func TestAccKeycloakDataSourceAuthenticationExecution_wrongAlias(t *testing.T) {
	t.Parallel()
	alias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testDataSourceKeycloakAuthenticationFlow_wrongAlias(alias),
				ExpectError: regexp.MustCompile("no authentication flow found for alias .*"),
			},
		},
	})
}

func testDataSourceKeycloakAuthenticationFlow_basic(alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "flow" {
	realm_id = data.keycloak_realm.realm.id
	alias    = "%s"
}

data "keycloak_authentication_flow" "flow" {
	realm_id 			= data.keycloak_realm.realm.id
	alias   			= keycloak_authentication_flow.flow.alias

	depends_on = [
		keycloak_authentication_flow.flow,
	]
}
	`, testAccRealm.Realm, alias)
}

func testDataSourceKeycloakAuthenticationFlow_wrongAlias(alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "flow" {
	realm_id = data.keycloak_realm.realm.id
	alias    = "%s"
}

data "keycloak_authentication_flow" "flow" {
	realm_id 			= data.keycloak_realm.realm.id
	alias   			= "foo"

	depends_on = [
		keycloak_authentication_flow.flow,
	]
}
	`, testAccRealm.Realm, alias)
}
