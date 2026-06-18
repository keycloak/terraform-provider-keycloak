package provider

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

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

func reverseStringMap(m map[string]string) map[string]string {
	reversed := make(map[string]string, len(m))
	for k, v := range m {
		reversed[v] = k
	}

	return reversed
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

func emptySliceIfNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
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

// suppressDiffForMultivalueAttributeOrder returns a DiffSuppressFunc that
// suppresses diffs when the order of multiple attribute values changed.
func suppressDiffForMultivalueAttributeOrder() schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, d *schema.ResourceData) bool {
		oldParts := strings.Split(old, MULTIVALUE_ATTRIBUTE_SEPARATOR)
		newParts := strings.Split(new, MULTIVALUE_ATTRIBUTE_SEPARATOR)
		slices.Sort(oldParts)
		slices.Sort(newParts)
		return strings.Join(oldParts, MULTIVALUE_ATTRIBUTE_SEPARATOR) == strings.Join(newParts, MULTIVALUE_ATTRIBUTE_SEPARATOR)
	}
}

// suppressDiffWhenNotInConfig returns a DiffSuppressFunc that suppresses diffs
// when the specified attribute is not present in the config (null).
// This allows:
// - Clearing a field with an empty string (field = "")
// - Keeping the server value when the field is not in the config
func suppressDiffWhenNotInConfig(attrName string) schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, d *schema.ResourceData) bool {
		rawConfig := d.GetRawConfig()
		if rawConfig.IsNull() || !rawConfig.IsKnown() {
			return true
		}
		configValue := rawConfig.GetAttr(attrName)
		return configValue.IsNull()
	}
}

func intPointer(i int) *int {
	return &i
}
