package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenIdHardcodedRoleProtocolMapper() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakOpenIdHardcodedRoleProtocolMapperCreate,
		ReadContext:   resourceKeycloakOpenIdHardcodedRoleProtocolMapperRead,
		UpdateContext: resourceKeycloakOpenIdHardcodedRoleProtocolMapperUpdate,
		DeleteContext: resourceKeycloakOpenIdHardcodedRoleProtocolMapperDelete,
		Importer: &schema.ResourceImporter{
			// import a mapper tied to a client:
			// {{realmId}}/client/{{clientId}}/{{protocolMapperId}}
			// or a client scope:
			// {{realmId}}/client-scope/{{clientScopeId}}/{{protocolMapperId}}
			StateContext: genericProtocolMapperImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A human-friendly name that will appear in the Keycloak console.",
			},
			"realm_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The realm id where the associated client or client scope exists.",
			},
			"client_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "The mapper's associated client. Cannot be used at the same time as client_scope_id.",
				ConflictsWith: []string{"client_scope_id"},
			},
			"client_scope_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "The mapper's associated client scope. Cannot be used at the same time as client_id.",
				ConflictsWith: []string{"client_id"},
			},
			"role_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func mapFromDataToOpenIdHardcodedRoleProtocolMapper(data *schema.ResourceData) *keycloak.OpenIdHardcodedRoleProtocolMapper {
	return &keycloak.OpenIdHardcodedRoleProtocolMapper{
		Id:            data.Id(),
		Name:          data.Get("name").(string),
		RealmId:       data.Get("realm_id").(string),
		ClientId:      data.Get("client_id").(string),
		ClientScopeId: data.Get("client_scope_id").(string),

		RoleId: data.Get("role_id").(string),
	}
}

func mapFromOpenIdHardcodedRoleMapperToData(mapper *keycloak.OpenIdHardcodedRoleProtocolMapper, data *schema.ResourceData) {
	data.SetId(mapper.Id)
	data.Set("name", mapper.Name)
	data.Set("realm_id", mapper.RealmId)

	if mapper.ClientId != "" {
		data.Set("client_id", mapper.ClientId)
	} else {
		data.Set("client_scope_id", mapper.ClientScopeId)
	}

	data.Set("role_id", mapper.RoleId)
}

func resourceKeycloakOpenIdHardcodedRoleProtocolMapperCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	openIdHardcodedRoleMapper := mapFromDataToOpenIdHardcodedRoleProtocolMapper(data)

	err := keycloakClient.ValidateOpenIdHardcodedRoleProtocolMapper(ctx, openIdHardcodedRoleMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewOpenIdHardcodedRoleProtocolMapper(ctx, openIdHardcodedRoleMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	mapFromOpenIdHardcodedRoleMapperToData(openIdHardcodedRoleMapper, data)

	return resourceKeycloakOpenIdHardcodedRoleProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdHardcodedRoleProtocolMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	openIdHardcodedRoleMapper, err := keycloakClient.GetOpenIdHardcodedRoleProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id())
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	mapFromOpenIdHardcodedRoleMapperToData(openIdHardcodedRoleMapper, data)

	return nil
}

func resourceKeycloakOpenIdHardcodedRoleProtocolMapperUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	openIdHardcodedRoleMapper := mapFromDataToOpenIdHardcodedRoleProtocolMapper(data)

	err := keycloakClient.ValidateOpenIdHardcodedRoleProtocolMapper(ctx, openIdHardcodedRoleMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.UpdateOpenIdHardcodedRoleProtocolMapper(ctx, openIdHardcodedRoleMapper)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakOpenIdHardcodedRoleProtocolMapperRead(ctx, data, meta)
}

func resourceKeycloakOpenIdHardcodedRoleProtocolMapperDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return diag.FromErr(keycloakClient.DeleteOpenIdHardcodedRoleProtocolMapper(ctx, realmId, clientId, clientScopeId, data.Id()))
}
