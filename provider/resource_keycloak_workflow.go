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
				Description: "The event that triggers the workflow (e.g. USER_LOGIN, USER_ADD, USER_GROUP_MEMBERSHIP_ADD, USER_ROLE_ADD).",
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
							Description: "Delay in milliseconds before executing this step.",
						},
						"config": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Key-value configuration for the step.",
						},
					},
				},
			},
		},
	}
}

func getWorkflowFromData(data *schema.ResourceData) *keycloak.Workflow {
	workflow := &keycloak.Workflow{
		Id:                data.Id(),
		Realm:             data.Get("realm").(string),
		Name:              data.Get("name").(string),
		On:                data.Get("on").(string),
		Enabled:           data.Get("enabled").(bool),
		Conditions:        data.Get("conditions").(string),
		CancelInProgress:  data.Get("cancel_in_progress").(string),
		RestartInProgress: data.Get("restart_in_progress").(string),
	}

	steps := make([]keycloak.WorkflowStep, 0)
	if v, ok := data.GetOk("step"); ok {
		for _, raw := range v.([]interface{}) {
			stepMap := raw.(map[string]interface{})
			step := keycloak.WorkflowStep{
				Uses:  stepMap["uses"].(string),
				After: stepMap["after"].(string),
			}
			if cfgRaw, ok := stepMap["config"]; ok {
				config := make(map[string][]string)
				for k, val := range cfgRaw.(map[string]interface{}) {
					config[k] = []string{val.(string)}
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
	data.Set("cancel_in_progress", workflow.CancelInProgress)
	data.Set("restart_in_progress", workflow.RestartInProgress)

	steps := make([]map[string]interface{}, 0, len(workflow.Steps))
	for _, step := range workflow.Steps {
		stepMap := map[string]interface{}{
			"uses":  step.Uses,
			"after": step.After,
		}
		if len(step.Config) > 0 {
			config := make(map[string]interface{})
			for k, vals := range step.Config {
				if len(vals) > 0 {
					config[k] = vals[0]
				}
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
		return nil, fmt.Errorf("invalid import. Supported import format: {{realm}}/{{workflowId}}")
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
		return nil, fmt.Errorf("error reading workflow: %s", diagnostics[0].Summary)
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
