package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakWorkflow() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakWorkflowCreate,
		ReadContext:   resourceKeycloakWorkflowRead,
		UpdateContext: resourceKeycloakWorkflowUpdate,
		DeleteContext: resourceKeycloakWorkflowDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakWorkflowImport,
		},
		Schema: map[string]*schema.Schema{
			"realm": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The realm this workflow belongs to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the workflow.",
			},
			"on": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The event that triggers the workflow. Supported values: user_created, user_removed, user_authenticated, user_federated_identity_added, user_federated_identity_removed, user_group_membership_added, user_group_membership_removed, user_role_granted, user_role_revoked.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the workflow is enabled.",
			},
			"conditions": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Expression that must be satisfied for the workflow to run.",
			},
			"cancel_in_progress": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Event that cancels an in-progress workflow execution.",
			},
			"restart_in_progress": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Event that restarts an in-progress workflow execution.",
			},
			"schedule": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Scheduling configuration for the workflow.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"after": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Interval between successive scheduled runs, as a duration string (e.g. 30s, 1d) or milliseconds.",
						},
						"batch_size": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Maximum number of resources processed per scheduled batch.",
						},
					},
				},
			},
			"state": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Runtime state of the workflow, as reported by Keycloak.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"errors": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Errors recorded for the workflow.",
						},
					},
				},
			},
			"step": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "Ordered list of steps to execute.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uses": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The step type to execute (e.g. disable-user, delete-user, notify-user).",
						},
						"after": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Delay before executing this step, as a duration string (e.g. 7d) or milliseconds (e.g. 2592000000).",
						},
						"priority": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Execution priority of the step, used to order steps that would otherwise run at the same time.",
						},
						"config": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Key-value configuration for the step.",
						},
						"scheduled_at": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Epoch timestamp in milliseconds at which the step is scheduled to execute.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Execution status of the step, as reported by Keycloak.",
						},
					},
				},
			},
		},
	}
}

func getWorkflowFromData(data *schema.ResourceData) *keycloak.Workflow {
	workflow := &keycloak.Workflow{
		Id:         data.Id(),
		Realm:      data.Get("realm").(string),
		Name:       data.Get("name").(string),
		On:         data.Get("on").(string),
		Enabled:    data.Get("enabled").(bool),
		Conditions: data.Get("conditions").(string),
	}

	cancelInProgress := data.Get("cancel_in_progress").(string)
	restartInProgress := data.Get("restart_in_progress").(string)
	if cancelInProgress != "" || restartInProgress != "" {
		workflow.Concurrency = &keycloak.WorkflowConcurrency{
			CancelInProgress:  cancelInProgress,
			RestartInProgress: restartInProgress,
		}
	}

	if v, ok := data.GetOk("schedule"); ok {
		scheduleList := v.([]interface{})
		if len(scheduleList) == 1 && scheduleList[0] != nil {
			scheduleMap := scheduleList[0].(map[string]interface{})
			workflow.Schedule = &keycloak.WorkflowSchedule{
				After:     scheduleMap["after"].(string),
				BatchSize: scheduleMap["batch_size"].(int),
			}
		}
	}

	steps := make([]keycloak.WorkflowStep, 0)
	if v, ok := data.GetOk("step"); ok {
		for _, raw := range v.([]interface{}) {
			stepMap := raw.(map[string]interface{})
			step := keycloak.WorkflowStep{
				Uses:     stepMap["uses"].(string),
				After:    stepMap["after"].(string),
				Priority: stepMap["priority"].(string),
			}
			if cfgRaw, ok := stepMap["config"]; ok {
				config := make(map[string]string)
				for k, val := range cfgRaw.(map[string]interface{}) {
					config[k] = val.(string)
				}
				step.Config = config
			}
			steps = append(steps, step)
		}
	}
	workflow.Steps = steps

	return workflow
}

func setWorkflowData(data *schema.ResourceData, workflow *keycloak.Workflow) {
	data.SetId(workflow.Id)
	data.Set("realm", workflow.Realm)
	data.Set("name", workflow.Name)
	data.Set("on", workflow.On)
	data.Set("enabled", workflow.Enabled)
	data.Set("conditions", workflow.Conditions)
	if workflow.Concurrency != nil {
		data.Set("cancel_in_progress", workflow.Concurrency.CancelInProgress)
		data.Set("restart_in_progress", workflow.Concurrency.RestartInProgress)
	} else {
		data.Set("cancel_in_progress", "")
		data.Set("restart_in_progress", "")
	}

	if workflow.Schedule != nil {
		data.Set("schedule", []interface{}{
			map[string]interface{}{
				"after":      workflow.Schedule.After,
				"batch_size": workflow.Schedule.BatchSize,
			},
		})
	} else {
		data.Set("schedule", []interface{}{})
	}

	if workflow.State != nil {
		errors := make([]interface{}, 0, len(workflow.State.Errors))
		for _, e := range workflow.State.Errors {
			errors = append(errors, e)
		}
		data.Set("state", []interface{}{
			map[string]interface{}{
				"errors": errors,
			},
		})
	} else {
		data.Set("state", []interface{}{})
	}

	steps := make([]map[string]interface{}, 0, len(workflow.Steps))
	for _, step := range workflow.Steps {
		stepMap := map[string]interface{}{
			"uses":         step.Uses,
			"after":        step.After,
			"priority":     step.Priority,
			"scheduled_at": int(step.ScheduledAt),
			"status":       step.Status,
		}
		if len(step.Config) > 0 {
			config := make(map[string]interface{})
			for k, v := range step.Config {
				config[k] = v
			}
			stepMap["config"] = config
		}
		steps = append(steps, stepMap)
	}
	data.Set("step", steps)
}

func resourceKeycloakWorkflowImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	parts := strings.Split(d.Id(), "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import format: {{realm}}/{{workflowId}}")
	}

	realm, id := parts[0], parts[1]

	_, err := keycloakClient.GetWorkflow(ctx, realm, id)
	if err != nil {
		return nil, err
	}

	d.Set("realm", realm)
	d.SetId(id)

	diagnostics := resourceKeycloakWorkflowRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("Error reading workflow: %s", diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceKeycloakWorkflowCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	workflow := getWorkflowFromData(data)

	if err := keycloakClient.NewWorkflow(ctx, workflow); err != nil {
		return diag.FromErr(err)
	}

	setWorkflowData(data, workflow)

	return resourceKeycloakWorkflowRead(ctx, data, meta)
}

func resourceKeycloakWorkflowRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realm := data.Get("realm").(string)
	id := data.Id()

	workflow, err := keycloakClient.GetWorkflow(ctx, realm, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setWorkflowData(data, workflow)

	return nil
}

func resourceKeycloakWorkflowUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	workflow := getWorkflowFromData(data)

	if err := keycloakClient.UpdateWorkflow(ctx, workflow); err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakWorkflowRead(ctx, data, meta)
}

func resourceKeycloakWorkflowDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realm := data.Get("realm").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteWorkflow(ctx, realm, id))
}
