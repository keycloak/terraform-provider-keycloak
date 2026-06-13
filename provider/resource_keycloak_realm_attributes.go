package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakRealmAttributes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmAttributesCreate,
		ReadContext:   resourceKeycloakRealmAttributesRead,
		UpdateContext: resourceKeycloakRealmAttributesUpdate,
		DeleteContext: resourceKeycloakRealmAttributesDelete,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"attributes": {
				Type:        schema.TypeMap,
				Required:    true,
				Description: "A map of attributes for the realm",
			},
		},
	}
}
func resourceKeycloakRealmAttributesCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	attributes := mapFromDataToRealmAttributes(data)

	err := keycloakClient.UpdateRealmAttributes(ctx, realmId, attributes)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(realmId)

	return resourceKeycloakRealmAttributesRead(ctx, data, meta)
}

func resourceKeycloakRealmAttributesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	attributes, err := keycloakClient.GetRealmAttributes(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}
	if attributes == nil {
		return nil
	}

	mapFromRealmAttributesToData(attributes, data)

	return nil
}

func resourceKeycloakRealmAttributesUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	attributes := mapFromDataToRealmAttributes(data)

	err := keycloakClient.UpdateRealmAttributes(ctx, realmId, attributes)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmAttributesRead(ctx, data, meta)
}

func resourceKeycloakRealmAttributesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	attributes := &keycloak.RealmAttributes{
		Attributes: map[string]interface{}{},
	}

	err := keycloakClient.UpdateRealmAttributes(ctx, realmId, attributes)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId("")

	return nil
}

func mapFromDataToRealmAttributes(data *schema.ResourceData) *keycloak.RealmAttributes {
	return &keycloak.RealmAttributes{
		RealmId:    data.Get("realm_id").(string),
		Attributes: data.Get("attributes").(map[string]interface{}),
	}
}

func mapFromRealmAttributesToData(attributes *keycloak.RealmAttributes, data *schema.ResourceData) {
	_attributes := map[string]interface{}{}
	if v, ok := data.GetOk("attributes"); ok {
		for key := range v.(map[string]interface{}) {
			_attributes[key] = attributes.Attributes[key]
		}
	}
	if err := data.Set("attributes", _attributes); err != nil {
		panic(fmt.Sprintf("Failed to set attributes: %v", err))
	}
}
