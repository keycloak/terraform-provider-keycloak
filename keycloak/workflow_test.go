package keycloak

import (
	"encoding/json"
	"strings"
	"testing"
)

// Regression test for https://github.com/keycloak/terraform-provider-keycloak/issues/1622:
// concurrency settings must be marshaled into a nested "concurrency" object with kebab-case
// keys, not as flat "cancelInProgress"/"restartInProgress" fields that Keycloak ignores.
func TestWorkflowConcurrencyMarshaling(t *testing.T) {
	workflow := &Workflow{
		Name:    "test-workflow",
		On:      "user_authenticated",
		Enabled: true,
		Steps:   []WorkflowStep{{Uses: "disable-user", After: "300000"}},
		Concurrency: &WorkflowConcurrency{
			RestartInProgress: "true",
		},
	}

	data, err := json.Marshal(workflow)
	if err != nil {
		t.Fatalf("failed to marshal workflow: %s", err)
	}

	body := string(data)

	if !strings.Contains(body, `"concurrency":{"restart-in-progress":"true"}`) {
		t.Errorf("expected nested concurrency object with kebab-case keys, got: %s", body)
	}

	for _, flat := range []string{`"restartInProgress"`, `"cancelInProgress"`, `"restart_in_progress"`} {
		if strings.Contains(body, flat) {
			t.Errorf("expected concurrency settings to be nested, but found flat field %s in: %s", flat, body)
		}
	}
}

// Workflows without concurrency settings must omit the "concurrency" key entirely.
func TestWorkflowConcurrencyOmittedWhenEmpty(t *testing.T) {
	workflow := &Workflow{
		Name:    "test-workflow",
		On:      "user_authenticated",
		Enabled: true,
		Steps:   []WorkflowStep{{Uses: "disable-user", After: "300000"}},
	}

	data, err := json.Marshal(workflow)
	if err != nil {
		t.Fatalf("failed to marshal workflow: %s", err)
	}

	if body := string(data); strings.Contains(body, "concurrency") {
		t.Errorf("expected no concurrency key when concurrency is unset, got: %s", body)
	}
}

// The nested concurrency object returned by Keycloak must unmarshal back into the struct,
// so reads and imports populate the concurrency settings.
func TestWorkflowConcurrencyUnmarshaling(t *testing.T) {
	body := `{
		"name": "test-workflow",
		"on": "user_authenticated",
		"enabled": true,
		"concurrency": {"cancel-in-progress": "true", "restart-in-progress": "false"},
		"steps": [{"uses": "disable-user", "after": "300000"}]
	}`

	var workflow Workflow
	if err := json.Unmarshal([]byte(body), &workflow); err != nil {
		t.Fatalf("failed to unmarshal workflow: %s", err)
	}

	if workflow.Concurrency == nil {
		t.Fatal("expected concurrency to be populated, got nil")
	}
	if workflow.Concurrency.CancelInProgress != "true" {
		t.Errorf("expected cancel-in-progress %q, got %q", "true", workflow.Concurrency.CancelInProgress)
	}
	if workflow.Concurrency.RestartInProgress != "false" {
		t.Errorf("expected restart-in-progress %q, got %q", "false", workflow.Concurrency.RestartInProgress)
	}
}

// The schedule settings must be marshaled into a nested "schedule" object with kebab-case keys,
// matching Keycloak's WorkflowScheduleRepresentation.
func TestWorkflowScheduleMarshaling(t *testing.T) {
	workflow := &Workflow{
		Name:    "test-workflow",
		On:      "user_authenticated",
		Enabled: true,
		Steps:   []WorkflowStep{{Uses: "disable-user", After: "300000"}},
		Schedule: &WorkflowSchedule{
			After:     "86400000",
			BatchSize: 100,
		},
	}

	data, err := json.Marshal(workflow)
	if err != nil {
		t.Fatalf("failed to marshal workflow: %s", err)
	}

	if body := string(data); !strings.Contains(body, `"schedule":{"after":"86400000","batch-size":100}`) {
		t.Errorf("expected nested schedule object with kebab-case keys, got: %s", body)
	}
}

// Workflows without a schedule must omit the "schedule" key entirely, and read-only "state"
// must never be sent on writes.
func TestWorkflowScheduleAndStateOmittedWhenEmpty(t *testing.T) {
	workflow := &Workflow{
		Name:    "test-workflow",
		On:      "user_authenticated",
		Enabled: true,
		Steps:   []WorkflowStep{{Uses: "disable-user", After: "300000"}},
	}

	data, err := json.Marshal(workflow)
	if err != nil {
		t.Fatalf("failed to marshal workflow: %s", err)
	}

	body := string(data)
	if strings.Contains(body, "schedule") {
		t.Errorf("expected no schedule key when schedule is unset, got: %s", body)
	}
	if strings.Contains(body, "state") {
		t.Errorf("expected no state key when state is unset, got: %s", body)
	}
}

// The nested schedule and runtime state/step fields returned by Keycloak must unmarshal back
// into the struct so reads and imports populate them.
func TestWorkflowScheduleAndStateUnmarshaling(t *testing.T) {
	body := `{
		"name": "test-workflow",
		"on": "user_authenticated",
		"enabled": true,
		"schedule": {"after": "86400000", "batch-size": 50},
		"state": {"errors": ["boom"]},
		"steps": [{"uses": "disable-user", "after": "300000", "priority": "10", "scheduledAt": 1700000000000, "status": "SCHEDULED"}]
	}`

	var workflow Workflow
	if err := json.Unmarshal([]byte(body), &workflow); err != nil {
		t.Fatalf("failed to unmarshal workflow: %s", err)
	}

	if workflow.Schedule == nil {
		t.Fatal("expected schedule to be populated, got nil")
	}
	if workflow.Schedule.After != "86400000" {
		t.Errorf("expected schedule after %q, got %q", "86400000", workflow.Schedule.After)
	}
	if workflow.Schedule.BatchSize != 50 {
		t.Errorf("expected schedule batch-size %d, got %d", 50, workflow.Schedule.BatchSize)
	}

	if workflow.State == nil || len(workflow.State.Errors) != 1 || workflow.State.Errors[0] != "boom" {
		t.Errorf("expected state errors [boom], got %+v", workflow.State)
	}

	if len(workflow.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(workflow.Steps))
	}
	step := workflow.Steps[0]
	if step.Priority != "10" {
		t.Errorf("expected step priority %q, got %q", "10", step.Priority)
	}
	if step.ScheduledAt != 1700000000000 {
		t.Errorf("expected step scheduledAt %d, got %d", int64(1700000000000), step.ScheduledAt)
	}
	if step.Status != "SCHEDULED" {
		t.Errorf("expected step status %q, got %q", "SCHEDULED", step.Status)
	}
}
