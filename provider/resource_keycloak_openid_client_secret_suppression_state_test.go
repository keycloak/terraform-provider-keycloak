package provider

import (
	"context"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkterraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"testing"
)

func TestKeycloakOpenidClient_suppressionDefaultFromLegacyState(t *testing.T) {
	data, err := schema.InternalMap(resourceKeycloakOpenidClient().Schema).Data(&sdkterraform.InstanceState{
		ID:         "legacy-client",
		Attributes: map[string]string{"client_id": "legacy", "realm_id": "example", "access_type": "PUBLIC"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &keycloak.OpenidClient{Id: "legacy-client", ClientId: "legacy", RealmId: "example", PublicClient: true}
	if err := setOpenidClientData(context.Background(), nil, data, client); err != nil {
		t.Fatal(err)
	}
	value, exists := data.State().Attributes["suppress_client_secret"]
	if !exists || value != "false" {
		t.Fatalf("refresh must persist false for older state, got %q (present=%t)", value, exists)
	}
}

func TestKeycloakOpenidClient_dataSourceWithoutSuppressionOption(t *testing.T) {
	data, err := schema.InternalMap(dataSourceKeycloakOpenidClient().Schema).Data(&sdkterraform.InstanceState{
		ID:         "existing-client",
		Attributes: map[string]string{"client_id": "existing", "realm_id": "example", "access_type": "PUBLIC"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &keycloak.OpenidClient{Id: "existing-client", ClientId: "existing", RealmId: "example", PublicClient: true}
	if err := setOpenidClientData(context.Background(), nil, data, client); err != nil {
		t.Fatal(err)
	}
	if _, exists := data.State().Attributes["suppress_client_secret"]; exists {
		t.Fatal("data source must not acquire the resource-only suppression field")
	}
}

func TestClientSecretSuppressionClearsUnreadableState(t *testing.T) {
	for _, suppress := range []bool{false, true} {
		data := schema.TestResourceDataRaw(t, resourceKeycloakOpenidClient().Schema, map[string]interface{}{"realm_id": "fixture", "client_id": "client", "client_secret": "old-secret", "suppress_client_secret": suppress})
		client := &keycloak.OpenidClient{Id: "client", RealmId: "fixture", ClientId: "client", ClientSecretUnreadable: true}
		if err := setOpenidClientData(context.Background(), nil, data, client); err != nil {
			t.Fatal(err)
		}
		want := "old-secret"
		if suppress {
			want = ""
		}
		if data.Get("client_secret") != want {
			t.Fatalf("suppress=%t did not preserve upstream 403 semantics: got %q", suppress, data.Get("client_secret"))
		}
	}
}

func TestSuppressionPlanClearsStaleComputedSecret(t *testing.T) {
	for _, plaintext := range []interface{}{nil, ""} {
		config := map[string]interface{}{"realm_id": "fixture", "client_id": "client", "access_type": "CONFIDENTIAL", "suppress_client_secret": true}
		if plaintext != nil {
			config["client_secret"] = plaintext
		}
		state := &sdkterraform.InstanceState{ID: "client", Attributes: map[string]string{"realm_id": "fixture", "client_id": "client", "access_type": "CONFIDENTIAL", "client_secret": "stale-secret", "suppress_client_secret": "false"}}
		rawSecret := cty.NullVal(cty.String)
		if plaintext != nil {
			rawSecret = cty.StringVal(plaintext.(string))
		}
		state.RawConfig = cty.ObjectVal(map[string]cty.Value{"client_secret": rawSecret})
		diff, err := resourceKeycloakOpenidClient().Diff(context.Background(), state, sdkterraform.NewResourceConfigRaw(config), nil)
		if err != nil {
			t.Fatal(err)
		}
		secret := diff.Attributes["client_secret"]
		if secret == nil || secret.New != "" {
			t.Fatalf("stale secret retained in plan: %#v", secret)
		}
		data := schema.TestResourceDataRaw(t, resourceKeycloakOpenidClient().Schema, map[string]interface{}{"realm_id": "fixture", "client_id": "client", "client_secret": "stale-secret", "suppress_client_secret": true})
		client, err := getOpenidClientFromData(data)
		if err != nil {
			t.Fatal(err)
		}
		if client.ClientSecret != "" {
			t.Fatal("stale computed secret included in update payload")
		}
	}
}
