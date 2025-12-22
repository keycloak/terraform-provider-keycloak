package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeycloakDataSourceAuthenticationSubFlow_byAlias(t *testing.T) {
	t.Parallel()

	parentFlowAlias := acctest.RandomWithPrefix("tf-acc")
	subflowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationSubFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakAuthenticationSubFlow_byAlias(parentFlowAlias, subflowAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_subflow.subflow", "id", "data.keycloak_authentication_subflow.subflow", "id"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_subflow.subflow", "alias", "data.keycloak_authentication_subflow.subflow", "alias"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_subflow.subflow", "realm_id", "data.keycloak_authentication_subflow.subflow", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_subflow.subflow", "parent_flow_alias", "data.keycloak_authentication_subflow.subflow", "parent_flow_alias"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_subflow.subflow", "provider_id", "data.keycloak_authentication_subflow.subflow", "provider_id"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_subflow.subflow", "description", "data.keycloak_authentication_subflow.subflow", "description"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_subflow.subflow", "authenticator", "data.keycloak_authentication_subflow.subflow", "authenticator"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_subflow.subflow", "requirement", "data.keycloak_authentication_subflow.subflow", "requirement"),
					testAccCheckDataKeycloakAuthenticationSubFlow("data.keycloak_authentication_subflow.subflow"),
				),
			},
		},
	})
}

func TestAccKeycloakDataSourceAuthenticationSubFlow_byId(t *testing.T) {
	t.Parallel()

	parentFlowAlias := acctest.RandomWithPrefix("tf-acc")
	subflowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationSubFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakAuthenticationSubFlow_byId(parentFlowAlias, subflowAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_subflow.subflow", "id", "data.keycloak_authentication_subflow.subflow_by_id", "id"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_subflow.subflow", "alias", "data.keycloak_authentication_subflow.subflow_by_id", "alias"),
					testAccCheckDataKeycloakAuthenticationSubFlow("data.keycloak_authentication_subflow.subflow_by_id"),
				),
			},
		},
	})
}

func TestAccKeycloakDataSourceAuthenticationSubFlow_wrongAlias(t *testing.T) {
	t.Parallel()

	parentFlowAlias := acctest.RandomWithPrefix("tf-acc")
	subflowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationSubFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testDataSourceKeycloakAuthenticationSubFlow_wrongAlias(parentFlowAlias, subflowAlias),
				ExpectError: regexp.MustCompile("no authentication subflow found for alias .*"),
			},
		},
	})
}

func testAccCheckDataKeycloakAuthenticationSubFlow(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmID := rs.Primary.Attributes["realm_id"]
		parentFlowAlias := rs.Primary.Attributes["parent_flow_alias"]

		authenticationSubFlow, err := keycloakClient.GetAuthenticationSubFlow(testCtx, realmID, parentFlowAlias, id)
		if err != nil {
			return err
		}

		if authenticationSubFlow.Id != id {
			return fmt.Errorf("expected authenticationSubFlow with ID %s but got %s", id, authenticationSubFlow.Id)
		}

		return nil
	}
}

func testDataSourceKeycloakAuthenticationSubFlow_byAlias(parentAlias, alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "flow" {
	realm_id = data.keycloak_realm.realm.id
	alias    = "%s"
}

resource "keycloak_authentication_subflow" "subflow" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	alias             = "%s"
	provider_id       = "basic-flow"
	description       = "Test subflow"
}

data "keycloak_authentication_subflow" "subflow" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	alias             = keycloak_authentication_subflow.subflow.alias

	depends_on = [
		keycloak_authentication_subflow.subflow,
	]
}
	`, testAccRealm.Realm, parentAlias, alias)
}

func testDataSourceKeycloakAuthenticationSubFlow_byId(parentAlias, alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "flow" {
	realm_id = data.keycloak_realm.realm.id
	alias    = "%s"
}

resource "keycloak_authentication_subflow" "subflow" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	alias             = "%s"
	provider_id       = "basic-flow"
	description       = "Test subflow"
}

data "keycloak_authentication_subflow" "subflow_by_id" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	id                = keycloak_authentication_subflow.subflow.id

	depends_on = [
		keycloak_authentication_subflow.subflow,
	]
}
	`, testAccRealm.Realm, parentAlias, alias)
}

func testDataSourceKeycloakAuthenticationSubFlow_wrongAlias(parentAlias, alias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "flow" {
	realm_id = data.keycloak_realm.realm.id
	alias    = "%s"
}

resource "keycloak_authentication_subflow" "subflow" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	alias             = "%s"
	provider_id       = "basic-flow"
}

data "keycloak_authentication_subflow" "subflow" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	alias             = "wrong-alias"

	depends_on = [
		keycloak_authentication_subflow.subflow,
	]
}
	`, testAccRealm.Realm, parentAlias, alias)
}
