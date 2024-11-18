package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func dataSourceKeycloakRealmKeys() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeycloakRealmKeysRead,
		Schema: map[string]*schema.Schema{
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"algorithms": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"status": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"algorithm": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"certificate": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"provider_id": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"provider_priority": {
							Type:     schema.TypeInt,
							Computed: true,
							Optional: true,
						},
						"kid": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"public_key": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func flattenRealmKeys(realmKeys []keycloak.Key) []map[string]interface{} {
	keyMap := make([]map[string]interface{}, 0)
	for _, key := range realmKeys {
		element := make(map[string]interface{})
		if key.Algorithm != nil {
			element["algorithm"] = key.Algorithm
		}
		if key.Certificate != nil {
			element["certificate"] = key.Certificate
		}
		if key.ProviderId != nil {
			element["provider_id"] = key.ProviderId
		}
		if key.ProviderPriority != nil {
			element["provider_priority"] = key.ProviderPriority
		}
		if key.Kid != nil {
			element["kid"] = key.Kid
		}
		if key.PublicKey != nil {
			element["public_key"] = key.PublicKey
		}
		if key.Status != nil {
			element["status"] = key.Status
		}
		if key.Type != nil {
			element["type"] = key.Type
		}

		keyMap = append(keyMap, element)
	}
	return keyMap
}

func setRealmKeysData(data *schema.ResourceData, keys *keycloak.Keys) error {
	data.SetId(data.Get("realm_id").(string))

	err := data.Set("keys", flattenRealmKeys(keys.Keys))
	if err != nil {
		return fmt.Errorf("could not set 'keys' with values '%+v'\n%+v", keys.Keys, err)
	}

	return nil
}

func dataSourceKeycloakRealmKeysRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)

	realmId := data.Get("realm_id").(string)

	keys, err := keycloakClient.GetRealmKeys(ctx, realmId)
	if err != nil {
		return handleNotFoundError(ctx, err, data)
	}

	if filterStatus, ok := data.GetOk("status"); ok {
		keys.Keys = filterKeys(keys.Keys, "status", filterStatus.(*schema.Set))
	}

	if filterAlgorithm, ok := data.GetOk("algorithms"); ok {
		keys.Keys = filterKeys(keys.Keys, "algorithms", filterAlgorithm.(*schema.Set))
	}

	if len(keys.Keys) == 0 {
		return diag.Diagnostics{{
			Summary:  "Your query returned no results. Please change your search criteria and try again.",
			Severity: diag.Error,
		}}
	}

	return diag.FromErr(setRealmKeysData(data, keys))
}

func filterKeys(allValues []keycloak.Key, filterAttribute string, allowedValues *schema.Set) []keycloak.Key {
	var result []keycloak.Key
	var keyValue string

	for _, key := range allValues {
		switch filterAttribute {
		case "status":
			keyValue = StringValue(key.Status)
		case "algorithms":
			keyValue = StringValue(key.Algorithm)
		}

		if Contains(allowedValues.List(), keyValue) {
			result = append(result, key)
		}
	}

	return result
}

// Contains checks if the array contains the value
func Contains(array []interface{}, value interface{}) bool {
	for _, element := range array {
		if element == value {
			return true
		}
	}
	return false
}

// StringValue returns the value of the string pointer passed in or "" if the pointer is nil.
func StringValue(value *string) string {
	if value != nil {
		return *value
	}
	return ""
}
