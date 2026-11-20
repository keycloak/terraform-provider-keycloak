package provider

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestOpenidClientAuthorizationResourceAttributesRoundTrip(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceKeycloakOpenidClientAuthorizationResource().Schema, map[string]interface{}{})
	keycloakAttributes := map[string][]string{"single": {"value"}, "multi": {"a", "b"}}

	err := setOpenidClientAuthorizationResourceData(data, &keycloak.OpenidClientAuthorizationResource{
		Id: "id", RealmId: "realm", ResourceServerId: "server", Name: "name", Attributes: keycloakAttributes,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]interface{}{"single": "value", "multi": "a,b"}
	if got := data.Get("attributes"); !reflect.DeepEqual(got, want) {
		t.Fatalf("attributes in state = %#v, want %#v", got, want)
	}
	if got := getOpenidClientAuthorizationResourceFromData(data).Attributes; !reflect.DeepEqual(got, keycloakAttributes) {
		t.Fatalf("attributes sent to Keycloak = %#v, want %#v", got, keycloakAttributes)
	}
}
