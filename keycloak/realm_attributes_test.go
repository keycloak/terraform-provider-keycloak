package keycloak

import (
	"reflect"
	"testing"
)

func TestMergeRealmAttributes(t *testing.T) {
	tests := []struct {
		name    string
		current map[string]interface{}
		old     map[string]interface{}
		updated map[string]interface{}
		want    map[string]interface{}
	}{
		{
			name:    "keeps unmanaged attributes and updates managed ones",
			current: map[string]interface{}{"managed": "a", "unmanaged": "x"},
			old:     map[string]interface{}{"managed": "a"},
			updated: map[string]interface{}{"managed": "b"},
			want:    map[string]interface{}{"managed": "b", "unmanaged": "x"},
		},
		{
			name:    "removes attributes that were dropped from the configuration",
			current: map[string]interface{}{"managed": "a", "unmanaged": "x"},
			old:     map[string]interface{}{"managed": "a"},
			updated: map[string]interface{}{},
			want:    map[string]interface{}{"unmanaged": "x"},
		},
		{
			name:    "adds new attributes",
			current: map[string]interface{}{"unmanaged": "x"},
			old:     nil,
			updated: map[string]interface{}{"managed": "a"},
			want:    map[string]interface{}{"managed": "a", "unmanaged": "x"},
		},
		{
			name:    "overwrites an attribute that already exists",
			current: map[string]interface{}{"frontendUrl": "https://old"},
			old:     map[string]interface{}{},
			updated: map[string]interface{}{"frontendUrl": "https://new"},
			want:    map[string]interface{}{"frontendUrl": "https://new"},
		},
		{
			name:    "removing an attribute that is already gone is a no-op",
			current: map[string]interface{}{"unmanaged": "x"},
			old:     map[string]interface{}{"managed": "a"},
			updated: nil,
			want:    map[string]interface{}{"unmanaged": "x"},
		},
		{
			name:    "works without current attributes",
			current: nil,
			old:     nil,
			updated: map[string]interface{}{"managed": "a"},
			want:    map[string]interface{}{"managed": "a"},
		},
		{
			name:    "returns an empty map rather than nil",
			current: nil,
			old:     map[string]interface{}{"managed": "a"},
			updated: nil,
			want:    map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := copyAttributes(tt.current)
			old := copyAttributes(tt.old)
			updated := copyAttributes(tt.updated)

			got := MergeRealmAttributes(current, old, updated)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeRealmAttributes() = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(current, tt.current) || !reflect.DeepEqual(old, tt.old) || !reflect.DeepEqual(updated, tt.updated) {
				t.Errorf("MergeRealmAttributes() modified its arguments")
			}
		})
	}
}

func copyAttributes(attributes map[string]interface{}) map[string]interface{} {
	if attributes == nil {
		return nil
	}
	result := make(map[string]interface{}, len(attributes))
	for key, value := range attributes {
		result[key] = value
	}
	return result
}
