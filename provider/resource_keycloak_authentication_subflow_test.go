package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"testing"
)

func TestAccKeycloakAuthenticationSubFlow_basic(t *testing.T) {
	t.Parallel()

	parentAuthFlowAlias := acctest.RandomWithPrefix("tf-acc")
	authFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationSubFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationSubFlow_basic(parentAuthFlowAlias, authFlowAlias),
				Check:  testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
			},
			{
				ResourceName:      "keycloak_authentication_subflow.subflow",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getSubFlowImportId("keycloak_authentication_subflow.subflow"),
			},
		},
	})
}

func TestAccKeycloakAuthenticationSubFlow_createAfterManualDestroy(t *testing.T) {
	t.Parallel()

	var authenticationSubFlow = &keycloak.AuthenticationSubFlow{}

	authParentFlowAlias := acctest.RandomWithPrefix("tf-acc")
	authFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationSubFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationSubFlow_basic(authParentFlowAlias, authFlowAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					testAccCheckKeycloakAuthenticationSubFlowFetch("keycloak_authentication_subflow.subflow", authenticationSubFlow),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteAuthenticationSubFlow(testCtx, authenticationSubFlow.RealmId, authenticationSubFlow.ParentFlowAlias, authenticationSubFlow.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakAuthenticationSubFlow_basic(authParentFlowAlias, authFlowAlias),
				Check:  testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
			},
		},
	})
}

func TestAccKeycloakAuthenticationSubFlow_updateAuthenticationSubFlow(t *testing.T) {
	t.Parallel()

	authParentFlowAlias := acctest.RandomWithPrefix("tf-acc")
	authFlowAliasBefore := acctest.RandomWithPrefix("tf-acc")
	authFlowAliasAfter := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationSubFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationSubFlow_basic(authParentFlowAlias, authFlowAliasBefore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttr("keycloak_authentication_subflow.subflow", "alias", authFlowAliasBefore),
				),
			},
			{
				Config: testKeycloakAuthenticationSubFlow_basic(authParentFlowAlias, authFlowAliasAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttr("keycloak_authentication_subflow.subflow", "alias", authFlowAliasAfter),
				),
			},
		},
	})
}

func TestAccKeycloakAuthenticationSubFlow_updateAuthenticationSubFlowRequirement(t *testing.T) {
	t.Parallel()

	authParentFlowAlias := acctest.RandomWithPrefix("tf-acc")
	authFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationSubFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationSubFlow_basic(authParentFlowAlias, authFlowAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttr("keycloak_authentication_subflow.subflow", "requirement", "DISABLED"),
				),
			},
			{
				Config: testKeycloakAuthenticationSubFlow_basicWithRequirement(authParentFlowAlias, authFlowAlias, "REQUIRED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttr("keycloak_authentication_subflow.subflow", "requirement", "REQUIRED"),
				),
			},
			{
				Config: testKeycloakAuthenticationSubFlow_basicWithRequirement(authParentFlowAlias, authFlowAlias, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttr("keycloak_authentication_subflow.subflow", "requirement", "DISABLED"),
				),
			},
		},
	})
}

func TestAccKeycloakAuthenticationSubFlow_updateAuthenticationSubFlowPriority(t *testing.T) {
	t.Parallel()

	if ok, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_25); !ok {
		t.Skip()
	}

	authParentFlowAlias := acctest.RandomWithPrefix("tf-acc")
	authFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationSubFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationSubFlow_basic(authParentFlowAlias, authFlowAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttr("keycloak_authentication_subflow.subflow", "priority", "0"),
				),
			},
			{
				Config: testKeycloakAuthenticationSubFlow_basicWithPriority(authParentFlowAlias, authFlowAlias, 111),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttr("keycloak_authentication_subflow.subflow", "priority", "111"),
				),
			},
		},
	})
}

func TestAccKeycloakAuthenticationSubFlow_createAuthenticationSubFlowPriority(t *testing.T) {
	t.Parallel()

	if ok, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_25); !ok {
		t.Skip()
	}

	authParentFlowAlias := acctest.RandomWithPrefix("tf-acc")
	authFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationSubFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationSubFlow_basicWithPriority(authParentFlowAlias, authFlowAlias, 111),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttr("keycloak_authentication_subflow.subflow", "priority", "111"),
				),
			},
		},
	})
}

func TestAccKeycloakAuthenticationSubFlowNested_updateAuthenticationPriority(t *testing.T) {
	t.Parallel()

	if ok, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_25); !ok {
		t.Skip()
	}

	authParentFlowAlias := acctest.RandomWithPrefix("tf-acc")
	authFlowAlias := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakAuthenticationSubFlowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakAuthenticationSubFlow_nested(authParentFlowAlias, authFlowAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttr("keycloak_authentication_subflow.subflow", "priority", "0"),
				),
			},
			{
				Config: testKeycloakAuthenticationSubFlow_nestedWithPriority(authParentFlowAlias, authFlowAlias, 30, 20, 10, 20, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.kerberos_execution"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution.kerberos_execution", "priority", "10"),
					testAccCheckKeycloakAuthenticationExecutionIndex("keycloak_authentication_execution.kerberos_execution", 0),

					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.cookie_execution"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution.cookie_execution", "priority", "20"),
					testAccCheckKeycloakAuthenticationExecutionIndex("keycloak_authentication_execution.cookie_execution", 1),

					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttr("keycloak_authentication_subflow.subflow", "priority", "30"),
					testAccCheckKeycloakAuthenticationSubFlowIndex("keycloak_authentication_subflow.subflow", 2),

					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.username-password-form"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution.username-password-form", "priority", "10"),
					testAccCheckKeycloakAuthenticationExecutionIndex("keycloak_authentication_execution.username-password-form", 0),

					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.otp-form"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution.otp-form", "priority", "20"),
					testAccCheckKeycloakAuthenticationExecutionIndex("keycloak_authentication_execution.otp-form", 1),
				),
			},
			{
				Config: testKeycloakAuthenticationSubFlow_nestedWithPriority(authParentFlowAlias, authFlowAlias, 30, 20, 10, 50, 40),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.kerberos_execution"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution.kerberos_execution", "priority", "10"),
					testAccCheckKeycloakAuthenticationExecutionIndex("keycloak_authentication_execution.kerberos_execution", 0),

					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.cookie_execution"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution.cookie_execution", "priority", "20"),
					testAccCheckKeycloakAuthenticationExecutionIndex("keycloak_authentication_execution.cookie_execution", 1),

					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttr("keycloak_authentication_subflow.subflow", "priority", "30"),
					testAccCheckKeycloakAuthenticationSubFlowIndex("keycloak_authentication_subflow.subflow", 2),

					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.username-password-form"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution.username-password-form", "priority", "40"),
					testAccCheckKeycloakAuthenticationExecutionIndex("keycloak_authentication_execution.username-password-form", 0),

					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.otp-form"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution.otp-form", "priority", "50"),
					testAccCheckKeycloakAuthenticationExecutionIndex("keycloak_authentication_execution.otp-form", 1),
				),
			},
			{
				Config: testKeycloakAuthenticationSubFlow_nestedWithPriority(authParentFlowAlias, authFlowAlias, 30, 10, 20, 40, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.cookie_execution"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution.cookie_execution", "priority", "10"),
					testAccCheckKeycloakAuthenticationExecutionIndex("keycloak_authentication_execution.cookie_execution", 0),

					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.kerberos_execution"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution.kerberos_execution", "priority", "20"),
					testAccCheckKeycloakAuthenticationExecutionIndex("keycloak_authentication_execution.kerberos_execution", 1),

					testAccCheckKeycloakAuthenticationSubFlowExists("keycloak_authentication_subflow.subflow"),
					resource.TestCheckResourceAttr("keycloak_authentication_subflow.subflow", "priority", "30"),
					testAccCheckKeycloakAuthenticationSubFlowIndex("keycloak_authentication_subflow.subflow", 2),

					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.otp-form"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution.otp-form", "priority", "40"),
					testAccCheckKeycloakAuthenticationExecutionIndex("keycloak_authentication_execution.otp-form", 0),

					testAccCheckKeycloakAuthenticationExecutionExists("keycloak_authentication_execution.username-password-form"),
					resource.TestCheckResourceAttr("keycloak_authentication_execution.username-password-form", "priority", "50"),
					testAccCheckKeycloakAuthenticationExecutionIndex("keycloak_authentication_execution.username-password-form", 1),
				),
			},
		},
	})
}

func testAccCheckKeycloakAuthenticationSubFlowExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getAuthenticationSubFlowFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakAuthenticationSubFlowFetch(resourceName string, authenticationSubFlow *keycloak.AuthenticationSubFlow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedAuthenticationSubFlow, err := getAuthenticationSubFlowFromState(s, resourceName)
		if err != nil {
			return err
		}

		authenticationSubFlow.Id = fetchedAuthenticationSubFlow.Id
		authenticationSubFlow.ParentFlowAlias = fetchedAuthenticationSubFlow.ParentFlowAlias
		authenticationSubFlow.RealmId = fetchedAuthenticationSubFlow.RealmId
		authenticationSubFlow.Alias = fetchedAuthenticationSubFlow.Alias

		return nil
	}
}

func testAccCheckKeycloakAuthenticationSubFlowDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_authentication_subflow" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]
			parentFlowAlias := rs.Primary.Attributes["parent_flow_alias"]

			authenticationSubFlow, _ := keycloakClient.GetAuthenticationSubFlow(testCtx, realm, parentFlowAlias, id)
			if authenticationSubFlow != nil {
				return fmt.Errorf("authentication flow with id %s still exists", id)
			}
		}

		return nil
	}
}

func getAuthenticationSubFlowFromState(s *terraform.State, resourceName string) (*keycloak.AuthenticationSubFlow, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]
	parentFlowAlias := rs.Primary.Attributes["parent_flow_alias"]

	authenticationSubFlow, err := keycloakClient.GetAuthenticationSubFlow(testCtx, realm, parentFlowAlias, id)

	if err != nil {
		return nil, fmt.Errorf("error getting authentication subflow with id %s: %s", id, err)
	}

	return authenticationSubFlow, nil
}

func getSubFlowImportId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		parentFlowAlias := rs.Primary.Attributes["parent_flow_alias"]
		realmId := rs.Primary.Attributes["realm_id"]

		return fmt.Sprintf("%s/%s/%s", realmId, parentFlowAlias, id), nil
	}
}

func testAccCheckKeycloakAuthenticationSubFlowIndex(resourceName string, idx int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realm := rs.Primary.Attributes["realm_id"]
		parentFlowAlias := rs.Primary.Attributes["parent_flow_alias"]
		providerID := rs.Primary.Attributes["authenticator"]

		authenticationExecutionInfo, err := keycloakClient.GetAuthenticationExecutionInfoFromProviderId(testCtx, realm, parentFlowAlias, providerID)
		if err != nil {
			return err
		}

		if authenticationExecutionInfo == nil {
			return fmt.Errorf("authentication flow with id %s does not exists", id)
		}

		if authenticationExecutionInfo.FlowId != id {
			return fmt.Errorf("expected authenticationExecutionInfo with FlowId %s but got %s", id, authenticationExecutionInfo.Id)
		}

		if authenticationExecutionInfo.Index != idx {
			return fmt.Errorf("expected index %d but got %d at %s", idx, authenticationExecutionInfo.Index, resourceName)
		}

		return nil
	}
}

func testKeycloakAuthenticationSubFlow_basic(parentAlias, alias string) string {
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

	alias       = "%s"
	provider_id = "basic-flow"
}
	`, testAccRealm.Realm, parentAlias, alias)
}

func testKeycloakAuthenticationSubFlow_basicWithRequirement(parentAlias, alias, requirement string) string {
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

	alias       = "%s"
	provider_id = "basic-flow"
	requirement = "%s"
}
	`, testAccRealm.Realm, parentAlias, alias, requirement)
}

func testKeycloakAuthenticationSubFlow_basicWithPriority(parentAlias, alias string, priority int) string {
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

	alias       = "%s"
	provider_id = "basic-flow"
	priority = %d
}
	`, testAccRealm.Realm, parentAlias, alias, priority)
}

func testKeycloakAuthenticationSubFlow_nested(parentAlias, alias string) string {
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

	alias       = "%s"
	provider_id = "basic-flow"
}

resource "keycloak_authentication_execution" "cookie_execution" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	authenticator     = "auth-cookie"
}

resource "keycloak_authentication_execution" "kerberos_execution" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	authenticator     = "auth-spnego"
}

resource "keycloak_authentication_execution" "otp-form" {
  realm_id          = data.keycloak_realm.realm.id
  parent_flow_alias = keycloak_authentication_subflow.subflow.alias
  authenticator     = "auth-otp-form"
}

resource "keycloak_authentication_execution" "username-password-form" {
  realm_id          = data.keycloak_realm.realm.id
  parent_flow_alias = keycloak_authentication_subflow.subflow.alias
  authenticator     = "auth-username-password-form"
}

	`, testAccRealm.Realm, parentAlias, alias)
}

func testKeycloakAuthenticationSubFlow_nestedWithPriority(parentAlias, alias string, prioritySubflow int, priorityCookie int, priorityKerberos int, priorityOtp int, priorityUserNamePassword int) string {
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

	alias       = "%s"
	provider_id = "basic-flow"
	priority     = %d
}

resource "keycloak_authentication_execution" "cookie_execution" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	authenticator     = "auth-cookie"
	priority          = %d
}

resource "keycloak_authentication_execution" "kerberos_execution" {
	realm_id          = data.keycloak_realm.realm.id
	parent_flow_alias = keycloak_authentication_flow.flow.alias
	authenticator     = "auth-spnego"
	priority          = %d
}

resource "keycloak_authentication_execution" "otp-form" {
  realm_id          = data.keycloak_realm.realm.id
  parent_flow_alias = keycloak_authentication_subflow.subflow.alias
  authenticator     = "auth-otp-form"
  priority          = %d
}

resource "keycloak_authentication_execution" "username-password-form" {
  realm_id          = data.keycloak_realm.realm.id
  parent_flow_alias = keycloak_authentication_subflow.subflow.alias
  authenticator     = "auth-username-password-form"
  priority          = %d
}

	`, testAccRealm.Realm, parentAlias, alias, prioritySubflow, priorityCookie, priorityKerberos, priorityOtp, priorityUserNamePassword)
}
