package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkterraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// Exercise the Terraform protocol with an ephemeral variable and a local API,
// including checking the actual password sent to Keycloak (which its GET masks).
func TestKeycloakRealmSMTPPasswordWriteOnly(t *testing.T) {
	var mu sync.Mutex
	var current keycloak.Realm
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/admin/serverinfo" {
			fmt.Fprint(w, `{"systemInfo":{"version":"26.0.0"}}`)
			return
		}
		switch r.Method {
		case http.MethodPost, http.MethodPut:
			var next keycloak.Realm
			if err := json.NewDecoder(r.Body).Decode(&next); err != nil {
				t.Error(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if next.SmtpServer.Password == "**********" {
				next.SmtpServer.Password = current.SmtpServer.Password
			}
			current = next
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			response := current
			if response.SmtpServer.Password != "" {
				response.SmtpServer.Password = "**********"
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Error(err)
			}
		case http.MethodDelete:
			current = keycloak.Realm{}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	client, err := keycloak.NewKeycloakClient(context.Background(), server.URL, "", "", "", "", "master", "", "", "test-token", "", "", "", "", false, 10, "", false, "", "", "", false, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	provider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{"keycloak_realm": resourceKeycloakRealm()},
		ConfigureContextFunc: func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return client, nil
		},
	}
	factories := map[string]func() (tfprotov5.ProviderServer, error){
		"keycloak": func() (tfprotov5.ProviderServer, error) { return provider.GRPCProvider(), nil },
	}
	check := func(password, version string) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			if current.SmtpServer.Password != password {
				return fmt.Errorf("unexpected SMTP password in API")
			}
			attrs := state.RootModule().Resources["keycloak_realm.realm"].Primary.Attributes
			if attrs["smtp_server.0.auth.0.password_wo"] != "" {
				return fmt.Errorf("write-only password persisted in state")
			}
			if version != "" && attrs["smtp_server.0.auth.0.password"] != "" {
				return fmt.Errorf("write-only password leaked into legacy password state")
			}
			if attrs["smtp_server.0.auth.0.password_wo_version"] != version {
				return fmt.Errorf("unexpected SMTP password version in state")
			}
			return nil
		}
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmSMTPPasswordWriteOnlyConfig("first-secret", "v1", "first"),
				Check:  check("first-secret", "v1"),
			},
			{
				Config:             testKeycloakRealmSMTPPasswordWriteOnlyConfig("second-secret", "v1", "first"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testKeycloakRealmSMTPPasswordWriteOnlyConfig("second-secret", "v1", "updated"),
				Check:  check("first-secret", "v1"),
			},
			{
				Config: testKeycloakRealmSMTPPasswordWriteOnlyConfig("second-secret", "v2", "updated"),
				Check:  check("second-secret", "v2"),
			},
			{
				Config: `resource "keycloak_realm" "realm" {
					realm = "smtp-test"
					smtp_server {
						host = "smtp.example.com"
						from = "admin@example.com"
						auth {
							username = "user"
							password = "legacy-secret"
						}
					}
				}`,
				Check: check("legacy-secret", ""),
			},
			{
				Config: testKeycloakRealmSMTPPasswordWriteOnlyConfig("migrated-secret", "v1", "updated"),
				Check:  check("migrated-secret", "v1"),
			},
			{
				ResourceName:      "keycloak_realm.realm",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"smtp_server.0.auth.0.password", "smtp_server.0.auth.0.password_wo_version",
				},
			},
			{
				Config:             testKeycloakRealmSMTPPasswordWriteOnlyConfig("migrated-secret", "v1", "updated"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: `resource "keycloak_realm" "realm" {
					realm = "smtp-test"
					smtp_server {
						host = "smtp.example.com"
						from = "admin@example.com"
					}
				}`,
				Check: check("", ""),
			},
		},
	})
}

func testKeycloakRealmSMTPPasswordWriteOnlyConfig(password, version, displayName string) string {
	return fmt.Sprintf(`
variable "smtp_password" {
	default = %q
	type = string
	sensitive = true
	ephemeral = true
}
resource "keycloak_realm" "realm" {
	realm = "smtp-test"
	display_name = %q
	smtp_server {
		host = "smtp.example.com"
		from = "admin@example.com"
		auth {
			username = "user"
			password_wo = var.smtp_password
			password_wo_version = %q
		}
	}
}`, password, displayName, version)
}

func TestKeycloakRealmSMTPPasswordValidation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		auth      map[string]interface{}
		wantError bool
	}{
		{"legacy", map[string]interface{}{"password": "secret"}, false},
		{"write-only", map[string]interface{}{"password_wo": "secret", "password_wo_version": "v1"}, false},
		{"missing-password", map[string]interface{}{}, true},
		{"missing-version", map[string]interface{}{"password_wo": "secret"}, true},
		{"missing-write-only-password", map[string]interface{}{"password_wo_version": "v1"}, true},
		{"conflicting-passwords", map[string]interface{}{"password": "legacy", "password_wo": "secret", "password_wo_version": "v1"}, true},
		{"legacy-with-version", map[string]interface{}{"password": "secret", "password_wo_version": "v1"}, true},
		{"empty-version", map[string]interface{}{"password_wo": "secret", "password_wo_version": ""}, true},
		{"empty-write-only-password", map[string]interface{}{"password_wo": "", "password_wo_version": "v1"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.auth["username"] = "user"
			config := sdkterraform.NewResourceConfigRaw(map[string]interface{}{
				"realm": "smtp-test",
				"smtp_server": []interface{}{map[string]interface{}{
					"host": "smtp.example.com", "from": "admin@example.com", "auth": []interface{}{tc.auth},
				}},
			})
			if diags := resourceKeycloakRealm().Validate(config); diags.HasError() != tc.wantError {
				t.Fatalf("unexpected validation result: %v", diags)
			}
		})
	}
}
