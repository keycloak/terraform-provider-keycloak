package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mrparkers/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenidDefaultClientScope() *schema.Resource {
	return &schema.Resource{
		Create: resourceKeycloakOpenidDefaultClientScopeCreate,
		Read:   resourceKeycloakOpenidDefaultClientScopesRead,
		Delete: resourceKeycloakOpenidDefaultClientScopeDelete,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_scope_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_scope_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceKeycloakOpenidDefaultClientScopeCreate(data *schema.ResourceData, meta interface{}) error {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return keycloakClient.PutOpenidRealmDefaultClientScope(realmId, clientScopeId)
}

func resourceKeycloakOpenidDefaultClientScopesRead(data *schema.ResourceData, meta interface{}) error {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	clientScope, err := keycloakClient.GetOpenidRealmDefaultClientScope(realmId, clientScopeId)
	if err != nil {
		return handleNotFoundError(err, data)
	}

	data.Set("client_scope_id", clientScope.Id)
	data.Set("client_scope_name", clientScope.Name)

	return nil
}

func resourceKeycloakOpenidDefaultClientScopeDelete(data *schema.ResourceData, meta interface{}) error {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	clientScopeId := data.Get("client_scope_id").(string)

	return keycloakClient.DeleteOpenidRealmDefaultClientScope(realmId, clientScopeId)
}
