package provider

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"reflect"
	"strings"
)

func getExtraConfigFromData(data *schema.ResourceData) map[string]interface{} {
	extraConfig := map[string]interface{}{}
	if v, ok := data.GetOk("extra_config"); ok {
		for key, value := range v.(map[string]interface{}) {
			extraConfig[key] = value
		}

		// check if extra config attribute has been deleted.
		// it's not enough to simply unset the attribute - we have to explicitly set
		// it to empty string in order to remove this on the Keycloak side
		if data.HasChange("extra_config") && !data.IsNewResource() {
			oldConfig, newConfig := data.GetChange("extra_config")
			newConfigMap := newConfig.(map[string]interface{})

			for oldKey := range oldConfig.(map[string]interface{}) {
				if _, ok := newConfigMap[oldKey]; !ok {
					extraConfig[oldKey] = ""
				}
			}
		}
	}

	return extraConfig
}

func setExtraConfigData(data *schema.ResourceData, extraConfig map[string]interface{}) {
	newExtraConfig := map[string]interface{}{}
	extraConfigFromState := getExtraConfigFromData(data)

	for k, v := range extraConfig {
		// when saving back to state, don't persist empty attributes that we're trying to remove from Keycloak
		if s, ok := v.(string); ok && s == "" {
			continue
		}

		// also, we don't want extra_config to be computed, so don't set anything that wasn't originally set by the user in the first place
		if _, ok := extraConfigFromState[k]; !ok {
			continue
		}

		newExtraConfig[k] = v
	}

	data.Set("extra_config", newExtraConfig)
}

// validateExtraConfig takes a reflect.Value type to check its JSON schema in order to validate that extra_config
// doesn't contain any attributes that could have been specified within the official schema
func validateExtraConfig(reflectValue reflect.Value) schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics

		extraConfig := v.(map[string]interface{})

		for i := 0; i < reflectValue.NumField(); i++ {
			field := reflectValue.Field(i)
			jsonKey := strings.Split(reflectValue.Type().Field(i).Tag.Get("json"), ",")[0]

			if jsonKey != "-" && field.CanSet() {
				if _, ok := extraConfig[jsonKey]; ok {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Invalid extra_config key",
						Detail:   fmt.Sprintf(`extra_config key "%s" is not allowed, as it conflicts with a top-level schema attribute`, jsonKey),
						AttributePath: append(path, cty.IndexStep{
							Key: cty.StringVal(jsonKey),
						}),
					})
				}
			}
		}

		return diags
	}
}
