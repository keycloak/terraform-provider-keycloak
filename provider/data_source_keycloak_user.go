package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakUserRead,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email_verified": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"first_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"required_actions": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"federated_identity": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceKeycloakUserRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmID := data.Get("realm_id").(string)
	username := data.Get("username").(string)

	user, err := keycloakClient.GetUserByUsername(ctx, realmID, username)
	if err != nil {
		return diag.FromErr(err)
	}
	if user == nil {
		return diag.Errorf("user with username %s not found", username)
	}

	mapFromUserToData(data, user)

	return nil
}
