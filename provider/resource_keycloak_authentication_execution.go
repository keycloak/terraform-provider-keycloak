package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakAuthenticationExecution() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakAuthenticationExecutionCreate,
		ReadContext:   resourceKeycloakAuthenticationExecutionRead,
		DeleteContext: resourceKeycloakAuthenticationExecutionDelete,
		UpdateContext: resourceKeycloakAuthenticationExecutionUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakAuthenticationExecutionImport,
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
			"authenticator": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"requirement": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"REQUIRED", "ALTERNATIVE", "OPTIONAL", "CONDITIONAL", "DISABLED"}, false), //OPTIONAL is removed from 8.0.0 onwards
				Default:      "DISABLED",
			},
			"priority": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func mapFromDataToAuthenticationExecution(data *schema.ResourceData) *keycloak.AuthenticationExecution {
	authenticationExecution := &keycloak.AuthenticationExecution{
		Id:              data.Id(),
		RealmId:         data.Get("realm_id").(string),
		ParentFlowAlias: data.Get("parent_flow_alias").(string),
		Authenticator:   data.Get("authenticator").(string),
		Requirement:     data.Get("requirement").(string),
		Priority:        data.Get("priority").(int),
	}

	return authenticationExecution
}

func mapFromAuthenticationExecutionToData(ctx context.Context, keycloakClient *keycloak.KeycloakClient, data *schema.ResourceData, authenticationExecution *keycloak.AuthenticationExecution) error {

	data.SetId(authenticationExecution.Id)

	data.Set("realm_id", authenticationExecution.RealmId)
	data.Set("parent_flow_alias", authenticationExecution.ParentFlowAlias)
	data.Set("authenticator", authenticationExecution.Authenticator)
	data.Set("requirement", authenticationExecution.Requirement)

	versionOk, err := keycloakClient.VersionIsGreaterThanOrEqualTo(ctx, keycloak.Version_25)
	if err != nil {
		return err
	}

	if versionOk {
		data.Set("priority", authenticationExecution.Priority)
	}

	return nil
}

func mapFromAuthenticationExecutionInfoToData(ctx context.Context, keycloakClient *keycloak.KeycloakClient, data *schema.ResourceData, authenticationExecutionInfo *keycloak.AuthenticationExecutionInfo) error {
	data.SetId(authenticationExecutionInfo.Id)

	data.Set("realm_id", authenticationExecutionInfo.RealmId)
	data.Set("parent_flow_alias", authenticationExecutionInfo.ParentFlowAlias)

	versionOk, err := keycloakClient.VersionIsGreaterThanOrEqualTo(ctx, keycloak.Version_25)
	if err != nil {
		return err
	}

	if versionOk {
		data.Set("priority", authenticationExecutionInfo.Priority)
	}

	return nil
}

func resourceKeycloakAuthenticationExecutionCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	authenticationExecution := mapFromDataToAuthenticationExecution(data)

	err := keycloakClient.NewAuthenticationExecution(ctx, authenticationExecution)
	if err != nil {
		return diag.FromErr(err)
	}

	err = mapFromAuthenticationExecutionToData(ctx, keycloakClient, data, authenticationExecution)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakAuthenticationExecutionRead(ctx, data, meta)
}

func resourceKeycloakAuthenticationExecutionRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	parentFlowAlias := data.Get("parent_flow_alias").(string)
	id := data.Id()

	authenticationExecution, err := keycloakClient.GetAuthenticationExecution(ctx, realmId, parentFlowAlias, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	err = mapFromAuthenticationExecutionToData(ctx, keycloakClient, data, authenticationExecution)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakAuthenticationExecutionUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	authenticationExecution := mapFromDataToAuthenticationExecution(data)

	err := keycloakClient.UpdateAuthenticationExecution(ctx, authenticationExecution)
	if err != nil {
		return diag.FromErr(err)
	}

	err = mapFromAuthenticationExecutionToData(ctx, keycloakClient, data, authenticationExecution)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakAuthenticationExecutionDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteAuthenticationExecution(ctx, realmId, id))
}

func resourceKeycloakAuthenticationExecutionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")

	if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{parentFlowAlias}}/{{authenticationExecutionId}}")
	}

	_, err := keycloakClient.GetAuthenticationExecution(ctx, parts[0], parts[1], parts[2])
	if err != nil {
		return nil, err
	}

	d.Set("realm_id", parts[0])
	d.Set("parent_flow_alias", parts[1])
	d.SetId(parts[2])

	diagnostics := resourceKeycloakAuthenticationExecutionRead(ctx, d, meta)
	if diagnostics.HasError() {
		return nil, errors.New(diagnostics[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
