package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakGenericProtocolMapper() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakGenericProtocolMapperRead,

		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The realm id where the associated client or client scope exists.",
			},
			"client_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The client this protocol mapper is attached to.",
				ConflictsWith: []string{"client_scope_id"},
			},
			"client_scope_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The client scope this protocol mapper is attached to.",
				ConflictsWith: []string{"client_id"},
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the protocol mapper.",
			},
			"protocol": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The protocol of the client (openid-connect / saml).",
			},
			"protocol_mapper": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the protocol mapper.",
			},
			"config": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The configuration of the protocol mapper.",
			},
		},
	}
}

func dataSourceKeycloakGenericProtocolMapperRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientId := data.Get("client_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)
	name := data.Get("name").(string)

	if clientId == "" && clientScopeId == "" {
		return diag.Errorf("one of client_id or client_scope_id must be set")
	}

	foundMapper, err := keycloakClient.GetGenericProtocolMapperByName(ctx, realmId, clientId, clientScopeId, name)
	if err != nil {
		return diag.FromErr(err)
	}

	if foundMapper == nil {
		return diag.Errorf("protocol mapper with name %q not found", name)
	}

	data.SetId(foundMapper.Id)
	data.Set("protocol", foundMapper.Protocol)
	data.Set("protocol_mapper", foundMapper.ProtocolMapper)
	data.Set("config", foundMapper.Config)

	return nil
}
