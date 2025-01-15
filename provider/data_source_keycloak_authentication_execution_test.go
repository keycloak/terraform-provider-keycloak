package provider

import (
	"fmt"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKeycloakDataSourceAuthenticationExecution_basic(t *testing.T) {
	t.Parallel()

	parentFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationExecutionConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakAuthenticationExecution_basic(parentFlowAlias, nil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.execution"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "id", "data.keycloak_authentication_execution.execution", "id"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "realm_id", "data.keycloak_authentication_execution.execution", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "parent_flow_alias", "data.keycloak_authentication_execution.execution", "parent_flow_alias"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "authenticator", "data.keycloak_authentication_execution.execution", "provider_id"),
					testAccCheckDataKeycloakAuthenticationExecution("data.keycloak_authentication_execution.execution"),
				),
			},
		},
	})
}

func TestAccKeycloakDataSourceAuthenticationExecution_errorNoExecutions(t *testing.T) {
	t.Parallel()
	parentFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationExecutionConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testDataSourceKeycloakAuthenticationExecution_errorNoExecutions(parentFlowAlias),
				ExpectError: regexp.MustCompile("no authentication executions found for parent flow alias .*"),
			},
		},
	})
}

func TestAccKeycloakDataSourceAuthenticationExecution_errorWrongProviderId(t *testing.T) {
	t.Parallel()
	parentFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationExecutionConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testDataSourceKeycloakAuthenticationExecution_errorWrongProviderId(parentFlowAlias, acctest.RandString(10)),
				ExpectError: regexp.MustCompile("no authentication execution under parent flow alias .* with provider id .* found"),
			},
		},
	})
}

func TestAccKeycloakDataSourceAuthenticationExecution_basicWithPriority(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_25)
	t.Parallel()

	parentFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationExecutionConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceKeycloakAuthenticationExecution_basic(parentFlowAlias, nil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.execution"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "id", "data.keycloak_authentication_execution.execution", "id"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "realm_id", "data.keycloak_authentication_execution.execution", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "priority", "data.keycloak_authentication_execution.execution", "priority"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "parent_flow_alias", "data.keycloak_authentication_execution.execution", "parent_flow_alias"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "authenticator", "data.keycloak_authentication_execution.execution", "provider_id"),
					testAccCheckDataKeycloakAuthenticationExecution("data.keycloak_authentication_execution.execution"),
				),
			},
			{
				Config: testDataSourceKeycloakAuthenticationExecution_basic(parentFlowAlias, intPointer(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.execution"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "id", "data.keycloak_authentication_execution.execution", "id"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "realm_id", "data.keycloak_authentication_execution.execution", "realm_id"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "priority", "data.keycloak_authentication_execution.execution", "priority"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "parent_flow_alias", "data.keycloak_authentication_execution.execution", "parent_flow_alias"),
					resource.TestCheckResourceAttrPair("keycloak_authentication_execution.execution", "authenticator", "data.keycloak_authentication_execution.execution", "provider_id"),
					testAccCheckDataKeycloakAuthenticationExecution("data.keycloak_authentication_execution.execution"),
				),
			},
		},
	})
}

func testAccCheckDataKeycloakAuthenticationExecution(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmID := rs.Primary.Attributes["realm_id"]
		parentFlowAlias := rs.Primary.Attributes["parent_flow_alias"]
		providerID := rs.Primary.Attributes["provider_id"]

		authenticationExecutionInfo, err := keycloakClient.GetAuthenticationExecutionInfoFromProviderId(testCtx, realmID, parentFlowAlias, providerID)
		if err != nil {
			return err
		}

		if authenticationExecutionInfo.Id != id {
			return fmt.Errorf("expected authenticationExecutionInfo with ID %s but got %s", id, authenticationExecutionInfo.Id)
		}

		return nil
	}
}

func testDataSourceKeycloakAuthenticationExecution_basic(parentFlowAlias string, priority *int) string {
	priorityAttr := ""
	if priority != nil {
		priorityAttr = fmt.Sprintf(`priority = %d`, *priority)
	}

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
	requirement       = "REQUIRED"
    %s
}

data "keycloak_authentication_execution" "execution" {
	realm_id 			= data.keycloak_realm.realm.id
	parent_flow_alias   = keycloak_authentication_flow.flow.alias
	provider_id     	= "identity-provider-redirector"

	depends_on = [
		keycloak_authentication_execution.execution,
	]
}
	`, testAccRealm.Realm, parentFlowAlias, priorityAttr)
}

func testDataSourceKeycloakAuthenticationExecution_errorNoExecutions(parentFlowAlias string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "flow" {
	realm_id = data.keycloak_realm.realm.id
	alias    = "%s"
}

data "keycloak_authentication_execution" "execution" {
	realm_id 			= data.keycloak_realm.realm.id
	parent_flow_alias   = keycloak_authentication_flow.flow.alias
	provider_id     	= "foo"

	depends_on = [
		keycloak_authentication_flow.flow,
	]
}
	`, testAccRealm.Realm, parentFlowAlias)
}

func testDataSourceKeycloakAuthenticationExecution_errorWrongProviderId(parentFlowAlias, providerId string) string {
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
	requirement       = "REQUIRED"
}

data "keycloak_authentication_execution" "execution" {
	realm_id 			= data.keycloak_realm.realm.id
	parent_flow_alias   = keycloak_authentication_flow.flow.alias
	provider_id     	= "%s"

	depends_on = [
		keycloak_authentication_execution.execution,
	]
}
	`, testAccRealm.Id, parentFlowAlias, providerId)
}
