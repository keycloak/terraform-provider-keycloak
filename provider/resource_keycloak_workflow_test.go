package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakWorkflow_basic(t *testing.T) {
	workflowName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakWorkflowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakWorkflow_basic(workflowName),
				Check:  testAccCheckKeycloakWorkflowExists("keycloak_workflow.workflow"),
			},
		},
	})
}

func TestAccKeycloakWorkflow_update(t *testing.T) {
	workflowName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakWorkflowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakWorkflow_basic(workflowName),
				Check:  testAccCheckKeycloakWorkflowExists("keycloak_workflow.workflow"),
			},
			{
				Config: testKeycloakWorkflow_withConditions(workflowName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakWorkflowExists("keycloak_workflow.workflow"),
					resource.TestCheckResourceAttr("keycloak_workflow.workflow", "conditions", "!has-role('some-role')"),
				),
			},
		},
	})
}

func TestAccKeycloakWorkflow_createAfterManualDestroy(t *testing.T) {
	var workflow = &keycloak.Workflow{}

	workflowName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakWorkflowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakWorkflow_basic(workflowName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakWorkflowExists("keycloak_workflow.workflow"),
					testAccCheckKeycloakWorkflowFetch("keycloak_workflow.workflow", workflow),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteWorkflow(testCtx, workflow.Realm, workflow.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakWorkflow_basic(workflowName),
				Check:  testAccCheckKeycloakWorkflowExists("keycloak_workflow.workflow"),
			},
		},
	})
}

func TestAccKeycloakWorkflow_import(t *testing.T) {
	workflowName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakWorkflowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakWorkflow_basic(workflowName),
			},
			{
				ResourceName:      "keycloak_workflow.workflow",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["keycloak_workflow.workflow"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["realm"], rs.Primary.ID), nil
				},
			},
		},
	})
}

func testAccCheckKeycloakWorkflowExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getWorkflowFromState(s, resourceName)
		return err
	}
}

func testAccCheckKeycloakWorkflowDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_workflow" || strings.HasPrefix(name, "data") {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm"]

			workflow, _ := keycloakClient.GetWorkflow(testCtx, realm, id)
			if workflow != nil {
				return fmt.Errorf("%s with id %s still exists", name, id)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakWorkflowFetch(resourceName string, workflow *keycloak.Workflow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetched, err := getWorkflowFromState(s, resourceName)
		if err != nil {
			return err
		}

		workflow.Id = fetched.Id
		workflow.Realm = fetched.Realm
		workflow.Name = fetched.Name

		return nil
	}
}

func getWorkflowFromState(s *terraform.State, resourceName string) (*keycloak.Workflow, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm"]

	workflow, err := keycloakClient.GetWorkflow(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting workflow with id %s: %s", id, err)
	}

	return workflow, nil
}

func testKeycloakWorkflow_basic(name string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_workflow" "workflow" {
	realm   = data.keycloak_realm.realm.id
	name    = "%s"
	on      = "USER_LOGIN"
	enabled = true

	step {
		uses = "disable-user"
		after = "86400000"
	}
}
`, testAccRealm.Realm, name)
}

func testKeycloakWorkflow_withConditions(name string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_workflow" "workflow" {
	realm      = data.keycloak_realm.realm.id
	name       = "%s"
	on         = "USER_LOGIN"
	enabled    = true
	conditions = "!has-role('some-role')"

	step {
		uses  = "disable-user"
		after = "86400000"
	}
}
`, testAccRealm.Realm, name)
}
