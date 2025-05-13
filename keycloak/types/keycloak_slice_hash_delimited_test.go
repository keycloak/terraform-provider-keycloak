package types

import (
	"reflect"
	"testing"
)

func TestKeycloakSliceHashDelimited_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		s       KeycloakSliceHashDelimited
		want    []byte
		wantErr bool
	}{
		{"should render null for nil slice", nil, []byte("null"), false},
		{"should render single item", KeycloakSliceHashDelimited{"https://app/redirect1"}, []byte("\"https://app/redirect1\""), false},
		{"should render two items", KeycloakSliceHashDelimited{"https://app/redirect1", "https://app/redirect2"}, []byte("\"https://app/redirect1##https://app/redirect2\""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
