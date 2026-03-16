package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakWorkflow() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakWorkflowRead,
		Schema: map[string]*schema.Schema{
			"realm": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The realm this workflow belongs to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the workflow.",
			},
			"on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"conditions": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cancel_in_progress": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"restart_in_progress": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"step": {
				Type:       schema.TypeList,
				Computed:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uses": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"after": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"config": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceKeycloakWorkflowRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realm := data.Get("realm").(string)
	name := data.Get("name").(string)

	workflow, err := keycloakClient.GetWorkflowByName(ctx, realm, name)
	if err != nil {
		return diag.FromErr(err)
	}

	setWorkflowData(data, workflow)

	return nil
}
