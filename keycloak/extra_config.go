package keycloak

import (
	"encoding/json"
	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"
	"reflect"
	"strconv"
	"strings"
)

func unmarshalExtraConfig(data []byte, reflectValue reflect.Value, extraConfig *map[string]interface{}) error {
	err := json.Unmarshal(data, extraConfig)
	if err != nil {
		return err
	}

	for i := 0; i < reflectValue.NumField(); i++ {
		structField := reflectValue.Type().Field(i)
		jsonKey := strings.Split(structField.Tag.Get("json"), ",")[0]
		if jsonKey != "-" {
			configValue, ok := (*extraConfig)[jsonKey]
			if ok {
				field := reflectValue.FieldByName(structField.Name)
				if field.IsValid() && field.CanSet() {
					if field.Kind() == reflect.String {
						field.SetString(configValue.(string))
					} else if field.Kind() == reflect.Bool {
						boolVal, err := strconv.ParseBool(configValue.(string))
						if err == nil {
							field.Set(reflect.ValueOf(types.KeycloakBoolQuoted(boolVal)))
						}
					} else if field.Kind() == reflect.TypeOf([]string{}).Kind() {
						var sliceQuoted types.KeycloakSliceQuoted
						var sliceHashDelimited types.KeycloakSliceHashDelimited

						if err = json.Unmarshal([]byte(configValue.(string)), &sliceQuoted); err == nil {
							field.Set(reflect.ValueOf(sliceQuoted))
						} else if err = sliceHashDelimited.UnmarshalJSON([]byte(configValue.(string))); err == nil {
							field.Set(reflect.ValueOf(sliceHashDelimited))
						}

					}

					delete(*extraConfig, jsonKey)
				}
			}
		}
	}

	return nil
}

func marshalExtraConfig(reflectValue reflect.Value, extraConfig map[string]interface{}) ([]byte, error) {
	out := map[string]interface{}{}

	for k, v := range extraConfig {
		out[k] = v
	}

	for i := 0; i < reflectValue.NumField(); i++ {
		jsonKey := strings.Split(reflectValue.Type().Field(i).Tag.Get("json"), ",")[0]
		if jsonKey != "-" {
			field := reflectValue.Field(i)
			if field.IsValid() && field.CanSet() {
				if field.Kind() == reflect.String {
					out[jsonKey] = field.String()
				} else if field.Kind() == reflect.Bool {
					out[jsonKey] = types.KeycloakBoolQuoted(field.Bool())
				} else if field.Kind() == reflect.TypeOf([]string{}).Kind() {
					if s, ok := field.Interface().(types.KeycloakSliceQuoted); ok {
						out[jsonKey] = s
					} else if s, ok := field.Interface().(types.KeycloakSliceHashDelimited); ok {
						out[jsonKey] = s
					}
				}
			}
		}
	}
	return json.Marshal(out)
}
