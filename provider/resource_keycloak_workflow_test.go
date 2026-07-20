package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakWorkflow_basic(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_5)
	workflowName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakWorkflowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakWorkflow_basic(workflowName),
				Check:  testAccCheckKeycloakWorkflowExists("keycloak_workflow.workflow"),
			},
		},
	})
}

func TestAccKeycloakWorkflow_update(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_5)
	workflowName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakWorkflowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakWorkflow_basic(workflowName),
				Check:  testAccCheckKeycloakWorkflowExists("keycloak_workflow.workflow"),
			},
			{
				Config: testKeycloakWorkflow_withConditions(workflowName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakWorkflowExists("keycloak_workflow.workflow"),
					resource.TestCheckResourceAttr("keycloak_workflow.workflow", "conditions", "has-role('some-role')"),
				),
			},
		},
	})
}

// Regression test for https://github.com/keycloak/terraform-provider-keycloak/issues/1622:
// restart_in_progress / cancel_in_progress must be persisted by Keycloak (sent as a nested
// "concurrency" object), not silently dropped.
func TestAccKeycloakWorkflow_concurrency(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_5)
	workflowName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakWorkflowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakWorkflow_withConcurrency(workflowName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakWorkflowExists("keycloak_workflow.workflow"),
					resource.TestCheckResourceAttr("keycloak_workflow.workflow", "restart_in_progress", "true"),
					testAccCheckKeycloakWorkflowHasConcurrency("keycloak_workflow.workflow", "", "true"),
				),
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

func TestAccKeycloakWorkflow_schedule(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_5)
	workflowName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakWorkflowDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakWorkflow_withSchedule(workflowName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakWorkflowExists("keycloak_workflow.workflow"),
					resource.TestCheckResourceAttr("keycloak_workflow.workflow", "schedule.0.after", "86400000"),
					resource.TestCheckResourceAttr("keycloak_workflow.workflow", "schedule.0.batch_size", "100"),
					testAccCheckKeycloakWorkflowHasSchedule("keycloak_workflow.workflow", "86400000", 100),
				),
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

func TestAccKeycloakWorkflow_createAfterManualDestroy(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_5)
	var workflow = &keycloak.Workflow{}

	workflowName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakWorkflowDestroy(),
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
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_5)
	workflowName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakWorkflowDestroy(),
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

// testAccCheckKeycloakWorkflowHasConcurrency verifies that Keycloak actually persisted the
// concurrency settings, by reading the workflow back from the API rather than from state.
func testAccCheckKeycloakWorkflowHasConcurrency(resourceName, expectedCancel, expectedRestart string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		workflow, err := getWorkflowFromState(s, resourceName)
		if err != nil {
			return err
		}

		if workflow.Concurrency == nil {
			return fmt.Errorf("expected concurrency to be persisted, but it was nil")
		}
		if workflow.Concurrency.CancelInProgress != expectedCancel {
			return fmt.Errorf("expected cancel-in-progress %q, got %q", expectedCancel, workflow.Concurrency.CancelInProgress)
		}
		if workflow.Concurrency.RestartInProgress != expectedRestart {
			return fmt.Errorf("expected restart-in-progress %q, got %q", expectedRestart, workflow.Concurrency.RestartInProgress)
		}

		return nil
	}
}

// testAccCheckKeycloakWorkflowHasSchedule verifies that Keycloak persisted the schedule settings
// by reading the workflow back from the API.
func testAccCheckKeycloakWorkflowHasSchedule(resourceName, expectedAfter string, expectedBatchSize int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		workflow, err := getWorkflowFromState(s, resourceName)
		if err != nil {
			return err
		}

		if workflow.Schedule == nil {
			return fmt.Errorf("expected schedule to be persisted, but it was nil")
		}
		if workflow.Schedule.After != expectedAfter {
			return fmt.Errorf("expected schedule after %q, got %q", expectedAfter, workflow.Schedule.After)
		}
		if workflow.Schedule.BatchSize != expectedBatchSize {
			return fmt.Errorf("expected schedule batch-size %d, got %d", expectedBatchSize, workflow.Schedule.BatchSize)
		}

		return nil
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
	on      = "user_authenticated"
	enabled = true

	step {
		uses = "disable-user"
		after = "86400000"
	}
}
`, testAccRealm.Realm, name)
}

func testKeycloakWorkflow_withConcurrency(name string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_workflow" "workflow" {
	realm               = data.keycloak_realm.realm.id
	name                = "%s"
	on                  = "user_authenticated"
	enabled             = true
	restart_in_progress = "true"

	step {
		uses  = "disable-user"
		after = "86400000"
	}
}
`, testAccRealm.Realm, name)
}

func testKeycloakWorkflow_withSchedule(name string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_workflow" "workflow" {
	realm   = data.keycloak_realm.realm.id
	name    = "%s"
	on      = "user_authenticated"
	enabled = true

	schedule {
		after      = "86400000"
		batch_size = 100
	}

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
	on         = "user_authenticated"
	enabled    = true
	conditions = "has-role('some-role')"

	step {
		uses  = "disable-user"
		after = "86400000"
	}
}
`, testAccRealm.Realm, name)
}
