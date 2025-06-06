package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func keys(data map[string]string) []string {
	var result []string
	for k := range data {
		result = append(result, k)
	}
	return result
}

func mapKeyFromValue(m map[string]string, value string) (string, bool) {
	for k, v := range m {
		if v == value {
			return k, true
		}
	}

	return "", false
}

func mergeSchemas(a map[string]*schema.Schema, b map[string]*schema.Schema) map[string]*schema.Schema {
	result := a
	for k, v := range b {
		result[k] = v
	}
	return result
}

// Converts duration string to an int representing the number of seconds, which is used by the Keycloak API
// Ex: "1h" => 3600
func getSecondsFromDurationString(s string) (int, error) {
	duration, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}

	return int(duration.Seconds()), nil
}

// Converts number of seconds from Keycloak API to a duration string used by the provider
// Ex: 3600 => "1h0m0s"
func getDurationStringFromSeconds(seconds int) string {
	return (time.Duration(seconds) * time.Second).String()
}

// This will suppress the Terraform diff when comparing duration strings.
// As long as both strings represent the same number of seconds, it makes no difference to the Keycloak API
func suppressDurationStringDiff(_, old, new string, _ *schema.ResourceData) bool {
	if old == "" || new == "" {
		return false
	}

	oldDuration, _ := time.ParseDuration(old)
	newDuration, _ := time.ParseDuration(new)

	return oldDuration.Seconds() == newDuration.Seconds()
}

func handleNotFoundError(ctx context.Context, err error, data *schema.ResourceData) diag.Diagnostics {
	if keycloak.ErrorIs404(err) {
		tflog.Warn(ctx, "Removing resource from state as it no longer exists", map[string]interface{}{
			"id": data.Id(),
		})
		data.SetId("")

		return nil
	}

	return diag.FromErr(err)
}

func interfaceSliceToStringSlice(iv []interface{}) []string {
	var sv []string
	for _, i := range iv {
		sv = append(sv, i.(string))
	}

	return sv
}

func stringArrayDifference(a, b []string) []string {
	var aWithoutB []string

	for _, s := range a {
		if !stringSliceContains(b, s) {
			aWithoutB = append(aWithoutB, s)
		}
	}

	return aWithoutB
}

func stringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func stringPointer(s string) *string {
	return &s
}

func intPointer(i int) *int {
	return &i
}

// requiredWithoutAll returns a validator which checks that the attribute at `argument` exists
// if none of the attributes in `checkExists` exist in the configuration
func requiredWithoutAll(key cty.Path, checkExists []cty.Path) schema.ValidateRawResourceConfigFunc {
	return func(ctx context.Context, req schema.ValidateResourceConfigFuncRequest, resp *schema.ValidateResourceConfigFuncResponse) {
		// Skip validation for null or unknown values
		if req.RawConfig.IsNull() || !req.RawConfig.IsKnown() {
			return
		}

		// Check if any of the checkExists attributes exist
		anyExists := false
		for _, path := range checkExists {
			val, err := path.Apply(req.RawConfig)
			if err == nil && !val.IsNull() && val.IsKnown() {
				anyExists = true
				break
			}
		}

		// If any exist, the argument is not required
		if anyExists {
			return
		}

		val, err := key.Apply(req.RawConfig)
		if err != nil || val.IsNull() {
			resp.Diagnostics = append(resp.Diagnostics, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Required attribute not set",
				Detail:   fmt.Sprintf("The attribute %s is required when none of the following attributes are specified: %v", key, checkExists),
			})
		}
	}
}
