package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mrparkers/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakOpenIdClientInitialAccessToken() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceKeycloakOpenIdClientInitialAccessTokenRead,
		CreateContext: resourceKeycloakOpenIdClientInitialAccessTokenCreate,
		DeleteContext: resourceKeycloakOpenIdClientInitialAccessTokenDelete,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"token_count": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"remaining_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"expiration": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"token_value": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceKeycloakOpenIdClientInitialAccessTokenRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	tokens, err := keycloakClient.GetClientInitialAccessTokens(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	id := data.Id()
	list := *tokens
	for i := range list {
		if list[i].Id == id {
			data.SetId(id)
			data.Set("realm_id", list[i].RealmId)
			data.Set("token_count", list[i].Count)
			data.Set("remaining_count", list[i].RemainingCount)
			data.Set("expiration", list[i].Expiration)
			data.Set("token_value", list[i].Token)
		}
	}
	return nil
}

func resourceKeycloakOpenIdClientInitialAccessTokenCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	initialAccessToken := getOpenidClientInitialAccessTokenFromData(data)
	createdToken, err := keycloakClient.NewOpenidClientInitialAccessToken(ctx, initialAccessToken)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(createdToken.Id)
	data.Set("realm_id", createdToken.RealmId)
	data.Set("token_count", createdToken.Count)
	data.Set("remaining_count", createdToken.RemainingCount)
	data.Set("expiration", createdToken.Expiration)
	data.Set("token_value", createdToken.Token)

	return nil
}

func resourceKeycloakOpenIdClientInitialAccessTokenDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteClientInitialAccessToken(ctx, realmId, id))
}

func getOpenidClientInitialAccessTokenFromData(data *schema.ResourceData) *keycloak.OpenidClientInitialAccessToken {
	initialAccessToken := &keycloak.OpenidClientInitialAccessToken{
		Id:         data.Id(),
		RealmId:    data.Get("realm_id").(string),
		Count:      data.Get("token_count").(int),
		Expiration: data.Get("expiration").(int),
	}

	return initialAccessToken
}
