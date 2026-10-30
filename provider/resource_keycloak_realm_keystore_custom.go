package provider

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakRealmKeystoreCustom() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmKeystoreCustomCreate,
		ReadContext:   resourceKeycloakRealmKeystoreCustomRead,
		UpdateContext: resourceKeycloakRealmKeystoreCustomUpdate,
		DeleteContext: resourceKeycloakRealmKeystoreCustomDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakRealmKeystoreCustomImport,
		},
		CustomizeDiff: customizeDiffRealmKeystoreCustom,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Display name of provider when linked in admin console.",
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"provider_id": {
				Type:     schema.TypeString,
				Required: true,
				// Keycloak ignores a changed providerId on update
				ForceNew:    true,
				Description: "The id of the key provider, as returned by the `getId` implementation of its `KeyProviderFactory`.",
			},
			"config": {
				Type:     schema.TypeMap,
				Optional: true,
				// Computed so CustomizeDiff can suppress order-only changes of multi-valued keys via SetNew,
				// which only operates on computed keys
				Computed:    true,
				Description: "Config of the key provider. Only keys that the key provider declares as config properties are allowed.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func getRealmKeystoreCustomConfig(raw interface{}) map[string]string {
	config := map[string]string{}
	if m, ok := raw.(map[string]interface{}); ok {
		for k, v := range m {
			config[k] = v.(string)
		}
	}

	return config
}

func getRealmKeystoreCustomFromData(data *schema.ResourceData) *keycloak.RealmKeystoreCustom {
	return &keycloak.RealmKeystoreCustom{
		Id:         data.Id(),
		Name:       data.Get("name").(string),
		RealmId:    data.Get("realm_id").(string),
		ProviderId: data.Get("provider_id").(string),
		Config:     getRealmKeystoreCustomConfig(data.Get("config")),
	}
}

func setRealmKeystoreCustomData(data *schema.ResourceData, keystore *keycloak.RealmKeystoreCustom, config map[string]string) {
	data.SetId(keystore.Id)

	data.Set("name", keystore.Name)
	data.Set("realm_id", keystore.RealmId)
	data.Set("provider_id", keystore.ProviderId)
	data.Set("config", config)
}

// mergeRealmKeystoreCustomConfig returns the config to store in the state. It only contains the keys
// that are already in the state: key providers may add declared keys on their own (e.g. a generated
// certificate), which would otherwise show up as a permanent diff. A key that was removed outside of
// Terraform is left out, so it shows up as a diff.
//
// Keycloak masks the values of secret properties, so a masked value keeps the value from the state.
// Multi-valued keys keep the value from the state if they only differ in the order of their elements.
func mergeRealmKeystoreCustomConfig(prior, remote map[string]string, properties map[string]keycloak.ComponentTypeProperty) map[string]string {
	config := map[string]string{}
	for k, priorValue := range prior {
		remoteValue, ok := remote[k]
		if !ok {
			continue
		}

		property := properties[k]
		switch {
		case property.Secret && remoteValue == keycloak.RealmKeystoreSecretValue:
			config[k] = priorValue
		case property.IsMultiValued() && equalCommaSeparatedSet(priorValue, remoteValue):
			config[k] = priorValue
		default:
			config[k] = remoteValue
		}
	}

	return config
}

func resourceKeycloakRealmKeystoreCustomCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	keystore := getRealmKeystoreCustomFromData(data)

	err := keycloakClient.ValidateRealmKeystoreCustom(ctx, keystore)
	if err != nil {
		return diag.FromErr(err)
	}

	err = keycloakClient.NewRealmKeystoreCustom(ctx, keystore)
	if err != nil {
		return diag.FromErr(err)
	}

	setRealmKeystoreCustomData(data, keystore, keystore.Config)

	return resourceKeycloakRealmKeystoreCustomRead(ctx, data, meta)
}

func resourceKeycloakRealmKeystoreCustomRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	keystore, err := keycloakClient.GetRealmKeystoreCustom(ctx, realmId, id)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	properties, err := keycloakClient.RealmKeystoreCustomConfigProperties(ctx, keystore.ProviderId)
	if err != nil {
		return diag.FromErr(err)
	}

	config := mergeRealmKeystoreCustomConfig(getRealmKeystoreCustomConfig(data.Get("config")), keystore.Config, properties)

	setRealmKeystoreCustomData(data, keystore, config)

	return nil
}

func resourceKeycloakRealmKeystoreCustomUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	keystore := getRealmKeystoreCustomFromData(data)

	err := keycloakClient.ValidateRealmKeystoreCustom(ctx, keystore)
	if err != nil {
		return diag.FromErr(err)
	}

	// Keycloak keeps config keys that are missing from an update. Only changed keys are sent, so Keycloak
	// keeps unchanged values as they are stored, and removed keys are sent without a value.
	oldRaw, _ := data.GetChange("config")
	oldConfig := getRealmKeystoreCustomConfig(oldRaw)
	changedConfig := map[string]string{}
	for k, v := range keystore.Config {
		if oldValue, ok := oldConfig[k]; !ok || oldValue != v {
			changedConfig[k] = v
		}
	}
	var removedKeys []string
	for k := range oldConfig {
		if _, ok := keystore.Config[k]; !ok {
			removedKeys = append(removedKeys, k)
		}
	}
	keystore.Config = changedConfig

	err = keycloakClient.UpdateRealmKeystoreCustom(ctx, keystore, removedKeys)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmKeystoreCustomRead(ctx, data, meta)
}

func resourceKeycloakRealmKeystoreCustomDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)
	id := data.Id()

	return diag.FromErr(keycloakClient.DeleteRealmKeystoreCustom(ctx, realmId, id))
}

func resourceKeycloakRealmKeystoreCustomImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	parts := strings.Split(d.Id(), "/")

	var realmId, id string
	switch {
	case len(parts) == 2:
		// Format: {realmId}/{keystoreId}
		realmId, id = parts[0], parts[1]
	case len(parts) >= 3:
		// Format: {realmId}/{name}/{providerId}. Keycloak creates components with server-generated ids,
		// so this format lets users import them with static values in an import block.
		realmId = parts[0]
		name := strings.Join(parts[1:len(parts)-1], "/")
		providerId := parts[len(parts)-1]

		keystores, err := keycloakClient.GetRealmKeystoresCustom(ctx, realmId, name, providerId)
		if err != nil {
			return nil, fmt.Errorf("error fetching key providers: %w", err)
		}

		if len(keystores) == 0 {
			return nil, fmt.Errorf("no key provider found with name=%q, providerId=%q in realm %q", name, providerId, realmId)
		}
		if len(keystores) > 1 {
			return nil, fmt.Errorf("%d key providers found with name=%q, providerId=%q in realm %q, import by id instead", len(keystores), name, providerId, realmId)
		}

		id = keystores[0].Id
	default:
		return nil, fmt.Errorf("Invalid import. Supported import formats: {{realmId}}/{{keystoreId}}, {{realmId}}/{{name}}/{{providerId}}")
	}

	keystore, err := keycloakClient.GetRealmKeystoreCustom(ctx, realmId, id)
	if err != nil {
		return nil, err
	}

	// Read only keeps the config keys that are in the state, so the import takes all of them
	setRealmKeystoreCustomData(d, keystore, keystore.Config)

	return []*schema.ResourceData{d}, nil
}

// plannedRealmKeystoreCustomConfig returns the config to plan. Multi-valued keys keep the prior value if
// they only differ in the order of their elements, so a reordered config doesn't cause an update.
func plannedRealmKeystoreCustomConfig(prior, config map[string]string, properties map[string]keycloak.ComponentTypeProperty) map[string]string {
	planned := map[string]string{}
	for k, v := range config {
		planned[k] = v
		if priorValue, ok := prior[k]; ok && properties[k].IsMultiValued() && equalCommaSeparatedSet(priorValue, v) {
			planned[k] = priorValue
		}
	}

	return planned
}

// customizeDiffRealmKeystoreCustom validates the planned config against the server info, so a missing
// key provider or an undeclared config key fails the plan instead of the apply, and suppresses
// order-only changes of multi-valued keys.
func customizeDiffRealmKeystoreCustom(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	// config is computed, so a config that is left out or empty would keep the keys from the state
	// (or be unknown on create) instead of removing them
	configOmitted := false
	if rawConfig := diff.GetRawConfig(); !rawConfig.IsNull() {
		c := rawConfig.GetAttr("config")
		configOmitted = c.IsNull() || (c.IsKnown() && c.LengthInt() == 0)
	}

	if !diff.NewValueKnown("provider_id") || (!configOmitted && !diff.NewValueKnown("config")) {
		return nil
	}

	keycloakClient := meta.(*keycloak.KeycloakClient)

	providerId := diff.Get("provider_id").(string)
	oldRaw, newRaw := diff.GetChange("config")
	config := getRealmKeystoreCustomConfig(newRaw)
	if configOmitted {
		config = map[string]string{}
	}

	err := keycloakClient.ValidateRealmKeystoreCustom(ctx, &keycloak.RealmKeystoreCustom{
		ProviderId: providerId,
		Config:     config,
	})
	if err != nil {
		return err
	}

	properties, err := keycloakClient.RealmKeystoreCustomConfigProperties(ctx, providerId)
	if err != nil {
		return err
	}

	planned := plannedRealmKeystoreCustomConfig(getRealmKeystoreCustomConfig(oldRaw), config, properties)
	if !configOmitted && reflect.DeepEqual(planned, getRealmKeystoreCustomConfig(newRaw)) {
		return nil
	}

	return diff.SetNew("config", planned)
}
