package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"testing"
)

func TestAccKeycloakAuthenticationFlow_basic(t *testing.T) {
	t.Parallel()
	authFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationFlow_basic(authFlowAlias),
				Check:  testAccCheckKeycloakAuthenticationFlowExists("keycloak_authentication_flow.flow"),
			},
			{
				ResourceName:        "keycloak_authentication_flow.flow",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakAuthenticationFlow_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var authenticationFlow = &keycloak.AuthenticationFlow{}

	authFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationFlow_basic(authFlowAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationFlowExists("keycloak_authentication_flow.flow"),
					testAccCheckKeycloakAuthenticationFlowFetch("keycloak_authentication_flow.flow", authenticationFlow),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteAuthenticationFlow(testCtx, authenticationFlow.RealmId, authenticationFlow.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakAuthenticationFlow_basic(authFlowAlias),
				Check:  testAccCheckKeycloakAuthenticationFlowExists("keycloak_authentication_flow.flow"),
			},
		},
	})
}

func TestAccKeycloakAuthenticationFlow_updateAuthenticationFlow(t *testing.T) {
	t.Parallel()

	authFlowAliasBefore := acctest.RandomWithPrefix("tf-acc")
	authFlowAliasAfter := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationFlow_basic(authFlowAliasBefore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationFlowExists("keycloak_authentication_flow.flow"),
					resource.TestCheckResourceAttr("keycloak_authentication_flow.flow", "alias", authFlowAliasBefore),
				),
			},
			{
				Config: testKeycloakAuthenticationFlow_basic(authFlowAliasAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationFlowExists("keycloak_authentication_flow.flow"),
					resource.TestCheckResourceAttr("keycloak_authentication_flow.flow", "alias", authFlowAliasAfter),
				),
			},
		},
	})
}

func TestAccKeycloakAuthenticationFlow_updateRealm(t *testing.T) {
	t.Parallel()

	authFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationFlow_updateRealmBefore(authFlowAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationFlowExists("keycloak_authentication_flow.flow"),
					testAccCheckKeycloakAuthenticationFlowBelongsToRealm("keycloak_authentication_flow.flow", testAccRealm.Realm),
				),
			},
			{
				Config: testKeycloakAuthenticationFlow_updateRealmAfter(authFlowAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationFlowExists("keycloak_authentication_flow.flow"),
					testAccCheckKeycloakAuthenticationFlowBelongsToRealm("keycloak_authentication_flow.flow", testAccRealmTwo.Realm),
				),
			},
		},
	})
}

func testAccCheckKeycloakAuthenticationFlowExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getAuthenticationFlowFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakAuthenticationFlowFetch(resourceName string, authenticationFlow *keycloak.AuthenticationFlow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedAuthenticationFlow, err := getAuthenticationFlowFromState(s, resourceName)
		if err != nil {
			return err
		}

		authenticationFlow.Id = fetchedAuthenticationFlow.Id
		authenticationFlow.RealmId = fetchedAuthenticationFlow.RealmId

		return nil
	}
}

func testAccCheckKeycloakAuthenticationFlowBelongsToRealm(resourceName, realm string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		authenticationFlow, err := getAuthenticationFlowFromState(s, resourceName)
		if err != nil {
			return err
		}

		if authenticationFlow.RealmId != realm {
			return fmt.Errorf("expected authentication flow with id %s to have realm_id of %s, but got %s", authenticationFlow.Id, realm, authenticationFlow.RealmId)
		}

		return nil
	}
}

func testAccCheckKeycloakAuthenticationFlowDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_authentication_flow" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			authenticationFlow, _ := keycloakClient.GetAuthenticationFlow(testCtx, realm, id)
			if authenticationFlow != nil {
				return fmt.Errorf("authentication flow with id %s still exists", id)
			}
		}

		return nil
	}
}

func getAuthenticationFlowFromState(s *terraform.State, resourceName string) (*keycloak.AuthenticationFlow, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	authenticationFlow, err := keycloakClient.GetAuthenticationFlow(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting authentication flow with id %s: %s", id, err)
	}

	return authenticationFlow, nil
}

func testKeycloakAuthenticationFlow_basic(alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "flow" {
	realm_id = data.keycloak_realm.realm.id
	alias    = "%s"
}
	`, testAccRealm.Realm, alias)
}

func testKeycloakAuthenticationFlow_updateRealmBefore(alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_1" {
	realm = "%s"
}

data "keycloak_realm" "realm_2" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "flow" {
	alias    = "%s"
	realm_id = data.keycloak_realm.realm_1.id
}
	`, testAccRealm.Realm, testAccRealmTwo.Realm, alias)
}

func testKeycloakAuthenticationFlow_updateRealmAfter(alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_1" {
	realm = "%s"
}

data "keycloak_realm" "realm_2" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "flow" {
	alias    = "%s"
	realm_id = data.keycloak_realm.realm_2.id
}
	`, testAccRealm.Realm, testAccRealmTwo.Realm, alias)
}
