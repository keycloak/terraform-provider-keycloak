package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"strings"
)

func resourceKeycloakAuthenticationSubFlow() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakAuthenticationSubFlowCreate,
		ReadContext:   resourceKeycloakAuthenticationSubFlowRead,
		DeleteContext: resourceKeycloakAuthenticationSubFlowDelete,
		UpdateContext: resourceKeycloakAuthenticationSubFlowUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakAuthenticationSubFlowImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"parent_flow_alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provider_id": {
				Type:         schema.TypeString,
				Default:      "basic-flow",
				ValidateFunc: validation.StringInSlice([]string{"basic-flow", "form-flow", "client-flow"}, false),
				Optional:     true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			//execution parts of the subflow
			"authenticator": {
				Type:        schema.TypeString,
				Description: "Might be needed to be set with certain custom subflow with specific authenticator, in general this will remain empty",
				Optional:    true,
				ForceNew:    true,
			},
			"requirement": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"REQUIRED", "ALTERNATIVE", "OPTIONAL", "CONDITIONAL", "DISABLED"}, false), //OPTIONAL is removed from 8.0.0 onwards
				Default:      "DISABLED",
			},
		},
	}
}

func mapFromDataToAuthenticationSubFlow(data *schema.ResourceData) *keycloak.AuthenticationSubFlow {
	authenticationSubFlow := &keycloak.AuthenticationSubFlow{
		Id:              data.Id(),
		RealmId:         data.Get("realm_id").(string),
		ParentFlowAlias: data.Get("parent_flow_alias").(string),
		Alias:           data.Get("alias").(string),
		ProviderId:      data.Get("provider_id").(string),
		Description:     data.Get("description").(string),
		Authenticator:   data.Get("authenticator").(string),
		Requirement:     data.Get("requirement").(string),
	}

	return authenticationSubFlow
}

func mapFromAuthenticationSubFlowToData(data *schema.ResourceData, authenticationSubFlow *keycloak.AuthenticationSubFlow) {
	data.SetId(authenticationSubFlow.Id)
	data.Set("realm_id", authenticationSubFlow.RealmId)
	data.Set("parent_flow_alias", authenticationSubFlow.ParentFlowAlias)
	data.Set("alias", authenticationSubFlow.Alias)
	data.Set("provider_id", authenticationSubFlow.ProviderId)
	data.Set("description", authenticationSubFlow.Description)
	data.Set("authenticator", authenticationSubFlow.Authenticator)
	data.Set("requirement", authenticationSubFlow.Requirement)
}

func resourceKeycloakAuthenticationSubFlowCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	authenticationFlow := mapFromDataToAuthenticationSubFlow(data)

	err := keycloakClient.NewAuthenticationSubFlow(ctx, authenticationFlow)
	if err != nil {
		return diag.FromErr(err)
	}
	mapFromAuthenticationSubFlowToData(data, authenticationFlow)
	return resourceKeycloakAuthenticationSubFlowRead(ctx, data, meta)
}

func resourceKeycloakAuthenticationSubFlowRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	parentFlowAlias := data.Get("parent_flow_alias").(string)
	id := data.Id()

	authenticationFlow, err := keycloakClient.GetAuthenticationSubFlow(ctx, realmId, parentFlowAlias, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}
	mapFromAuthenticationSubFlowToData(data, authenticationFlow)
	return nil
}

func resourceKeycloakAuthenticationSubFlowUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	authenticationFlow := mapFromDataToAuthenticationSubFlow(data)

	err := keycloakClient.UpdateAuthenticationSubFlow(ctx, authenticationFlow)
	if err != nil {
		return diag.FromErr(err)
	}
	mapFromAuthenticationSubFlowToData(data, authenticationFlow)
	return nil
}

func resourceKeycloakAuthenticationSubFlowDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	parentFlowAlias := data.Get("parent_flow_alias").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteAuthenticationSubFlow(ctx, realmId, parentFlowAlias, id))
}

func resourceKeycloakAuthenticationSubFlowImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")

	if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{parentFlowAlias}}/{{authenticationSubFlowId}}")
	}

	_, err := keycloakClient.GetAuthenticationSubFlow(ctx, parts[0], parts[1], parts[2])
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", parts[0])
	d.Set("parent_flow_alias", parts[1])
	d.SetId(parts[2])

	diagnostics := resourceKeycloakAuthenticationSubFlowRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
