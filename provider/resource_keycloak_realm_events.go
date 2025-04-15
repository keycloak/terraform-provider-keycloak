package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakRealmEvents() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmEventsCreate,
		ReadContext:   resourceKeycloakRealmEventsRead,
		DeleteContext: resourceKeycloakRealmEventsDelete,
		UpdateContext: resourceKeycloakRealmEventsUpdate,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"admin_events_details_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},
			"admin_events_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},
			"enabled_event_types": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: false,
			},
			"events_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},
			"events_expiration": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
			},
			"events_listeners": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: false,
			},
		},
	}
}

func getRealmEventsConfigFromData(data *schema.ResourceData) *keycloak.RealmEventsConfig {
	enabledEventTypes := make([]string, 0)
	eventsListeners := make([]string, 0)

	if v, ok := data.GetOk("enabled_event_types"); ok {
		for _, enabledEventType := range v.(*schema.Set).List() {
			enabledEventTypes = append(enabledEventTypes, enabledEventType.(string))
		}
	}

	if v, ok := data.GetOk("events_listeners"); ok {
		for _, eventsListener := range v.(*schema.Set).List() {
			eventsListeners = append(eventsListeners, eventsListener.(string))
		}
	}

	realmEventsConfig := &keycloak.RealmEventsConfig{
		AdminEventsDetailsEnabled: data.Get("admin_events_details_enabled").(bool),
		AdminEventsEnabled:        data.Get("admin_events_enabled").(bool),
		EnabledEventTypes:         enabledEventTypes,
		EventsEnabled:             data.Get("events_enabled").(bool),
		EventsExpiration:          data.Get("events_expiration").(int),
		EventsListeners:           eventsListeners,
	}

	return realmEventsConfig
}

func setRealmEventsConfigData(data *schema.ResourceData, realmEventsConfig *keycloak.RealmEventsConfig) {
	data.Set("admin_events_details_enabled", realmEventsConfig.AdminEventsDetailsEnabled)
	data.Set("admin_events_enabled", realmEventsConfig.AdminEventsEnabled)
	data.Set("events_enabled", realmEventsConfig.EventsEnabled)
	data.Set("events_expiration", realmEventsConfig.EventsExpiration)
	data.Set("events_listeners", realmEventsConfig.EventsListeners)

	if _, ok := data.GetOk("enabled_event_types"); ok {
		data.Set("enabled_event_types", realmEventsConfig.EnabledEventTypes)
	}
}

func resourceKeycloakRealmEventsCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	realmId := data.Get("realm_id").(string)
	data.SetId(realmId)

	diagnostics := resourceKeycloakRealmEventsUpdate(ctx, data, meta)
	if diagnostics.HasError() {
		return diagnostics
	}

	return resourceKeycloakRealmEventsRead(ctx, data, meta)
}

func resourceKeycloakRealmEventsRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	realmEventsConfig, err := keycloakClient.GetRealmEventsConfig(ctx, realmId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	setRealmEventsConfigData(data, realmEventsConfig)

	return nil
}

func resourceKeycloakRealmEventsDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)

	// The realm events config cannot be deleted, so instead we set it back to its "zero" values.
	realmEventsConfig := &keycloak.RealmEventsConfig{}

	err := keycloakClient.UpdateRealmEventsConfig(ctx, realmId, realmEventsConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmEventsUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	realmEventsConfig := getRealmEventsConfigFromData(data)

	err := keycloakClient.UpdateRealmEventsConfig(ctx, realmId, realmEventsConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	setRealmEventsConfigData(data, realmEventsConfig)

	return nil
}
