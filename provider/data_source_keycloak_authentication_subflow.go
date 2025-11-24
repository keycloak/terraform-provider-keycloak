package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakAuthenticationSubflow() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakAuthenticationSubflowRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"parent_flow_alias": {
				Type:     schema.TypeString,
				Required: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"provider_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authenticator": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"requirement": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"priority": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceKeycloakAuthenticationSubflowRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmID := data.Get("realm_id").(string)
	parentFlowAlias := data.Get("parent_flow_alias").(string)

	var authenticationSubFlow *keycloak.AuthenticationSubFlow
	var err error

	// Try to fetch by id first if provided
	if id, ok := data.GetOk("id"); ok {
		authenticationSubFlow, err = keycloakClient.GetAuthenticationSubFlow(ctx, realmID, parentFlowAlias, id.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	} else if alias, ok := data.GetOk("alias"); ok {
		// Otherwise fetch by alias
		authenticationSubFlow, err = keycloakClient.GetAuthenticationSubFlowFromAlias(ctx, realmID, parentFlowAlias, alias.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		return diag.FromErr(fmt.Errorf("either 'id' or 'alias' must be specified"))
	}

	err = mapFromAuthenticationSubFlowToDataSource(ctx, keycloakClient, data, authenticationSubFlow)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func mapFromAuthenticationSubFlowToDataSource(ctx context.Context, keycloakClient *keycloak.KeycloakClient, data *schema.ResourceData, authenticationSubFlow *keycloak.AuthenticationSubFlow) error {
	data.SetId(authenticationSubFlow.Id)
	data.Set("realm_id", authenticationSubFlow.RealmId)
	data.Set("parent_flow_alias", authenticationSubFlow.ParentFlowAlias)
	data.Set("alias", authenticationSubFlow.Alias)
	data.Set("provider_id", authenticationSubFlow.ProviderId)
	data.Set("description", authenticationSubFlow.Description)
	data.Set("authenticator", authenticationSubFlow.Authenticator)
	data.Set("requirement", authenticationSubFlow.Requirement)

	versionOk, err := keycloakClient.VersionIsGreaterThanOrEqualTo(ctx, keycloak.Version_25)
	if err != nil {
		return err
	}

	if versionOk {
		data.Set("priority", authenticationSubFlow.Priority)
	}

	return nil
}
