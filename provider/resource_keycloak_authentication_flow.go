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

func resourceKeycloakAuthenticationFlow() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakAuthenticationFlowCreate,
		ReadContext:   resourceKeycloakAuthenticationFlowRead,
		DeleteContext: resourceKeycloakAuthenticationFlowDelete,
		UpdateContext: resourceKeycloakAuthenticationFlowUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakAuthenticationFlowImport,
		},
		Schema: map[string]*schema.Schema{
			"realm_id": {
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
				ValidateFunc: validation.StringInSlice([]string{"basic-flow", "client-flow"}, false), //it seems toplevel can only one of these and not 'form-flow'
				Optional:     true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func mapFromDataToAuthenticationFlow(data *schema.ResourceData) *keycloak.AuthenticationFlow {
	authenticationFlow := &keycloak.AuthenticationFlow{
		Id:          data.Id(),
		RealmId:     data.Get("realm_id").(string),
		Alias:       data.Get("alias").(string),
		ProviderId:  data.Get("provider_id").(string),
		Description: data.Get("description").(string),
	}

	return authenticationFlow
}

func mapFromAuthenticationFlowToData(data *schema.ResourceData, authenticationFlow *keycloak.AuthenticationFlow) {
	data.SetId(authenticationFlow.Id)
	data.Set("realm_id", authenticationFlow.RealmId)
	data.Set("alias", authenticationFlow.Alias)
	data.Set("provider_id", authenticationFlow.ProviderId)
	data.Set("description", authenticationFlow.Description)
}

func mapFromAuthenticationFlowInfoToData(data *schema.ResourceData, authenticationFlow *keycloak.AuthenticationFlow) {
	data.SetId(authenticationFlow.Id)
	data.Set("realm_id", authenticationFlow.RealmId)
	data.Set("alias", authenticationFlow.Alias)
}

func resourceKeycloakAuthenticationFlowCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	authenticationFlow := mapFromDataToAuthenticationFlow(data)

	err := keycloakClient.NewAuthenticationFlow(ctx, authenticationFlow)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromAuthenticationFlowToData(data, authenticationFlow)

	return resourceKeycloakAuthenticationFlowRead(ctx, data, meta)
}

func resourceKeycloakAuthenticationFlowRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	authenticationFlow, err := keycloakClient.GetAuthenticationFlow(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromAuthenticationFlowToData(data, authenticationFlow)
	return nil
}

func resourceKeycloakAuthenticationFlowUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	authenticationFlow := mapFromDataToAuthenticationFlow(data)

	err := keycloakClient.UpdateAuthenticationFlow(ctx, authenticationFlow)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromAuthenticationFlowToData(data, authenticationFlow)
	return nil
}

func resourceKeycloakAuthenticationFlowDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteAuthenticationFlow(ctx, realmId, id))
}

func resourceKeycloakAuthenticationFlowImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{authenticationFlowId}}")
	}

	_, err := keycloakClient.GetAuthenticationFlow(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", parts[0])
	d.SetId(parts[1])

	diagnostics := resourceKeycloakAuthenticationFlowRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
