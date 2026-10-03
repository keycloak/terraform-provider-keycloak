package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
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
	var writes int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/admin/serverinfo" {
			fmt.Fprint(w, `{"systemInfo":{"version":"26.8.0"}}`)
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
				// Keycloak 26.8 removes masked credentials when the destination changes.
				old, new := current.SmtpServer, next.SmtpServer
				if old.Port == "" {
					old.Port = "25"
				}
				if new.Port == "" {
					new.Port = "25"
				}
				next.SmtpServer.Password = ""
				if old.Host == new.Host && old.Port == new.Port && old.Ssl == new.Ssl &&
					old.StartTls == new.StartTls && old.From == new.From && old.User == new.User &&
					old.AuthTokenUrl == new.AuthTokenUrl && old.AuthTokenClientId == new.AuthTokenClientId && old.AuthTokenScope == new.AuthTokenScope {
					next.SmtpServer.Password = current.SmtpServer.Password
				}
			}
			current = next
			writes++
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

	steps := []resource.TestStep{
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
			Config: strings.ReplaceAll(testKeycloakRealmSMTPPasswordWriteOnlyConfig("second-secret", "v2", "updated"), `from = "admin@example.com"`, "from = \"admin@example.com\"\nfrom_display_name = \"Updated sender\""),
			Check:  check("second-secret", "v2"),
		},
	}
	// Each destination change must fail without a version bump, then succeed
	// with the real password when a new version is provided.
	for i, change := range []struct{ old, new string }{
		{`host = "smtp.example.com"`, `host = "other.example.com"`},
		{`host = "smtp.example.com"`, "host = \"smtp.example.com\"\nport = \"587\""},
		{`host = "smtp.example.com"`, "host = \"smtp.example.com\"\nssl = true"},
		{`host = "smtp.example.com"`, "host = \"smtp.example.com\"\nstarttls = true"},
		{`from = "admin@example.com"`, `from = "other@example.com"`},
		{`username = "user"`, `username = "other-user"`},
	} {
		version := fmt.Sprintf("destination-%d", i)
		previousVersion := "v2"
		if i > 0 {
			previousVersion = fmt.Sprintf("reset-%d", i-1)
		}
		changed := func(v string) string {
			return strings.ReplaceAll(testKeycloakRealmSMTPPasswordWriteOnlyConfig("second-secret", v, "updated"), change.old, change.new)
		}
		steps = append(steps,
			resource.TestStep{Config: changed(previousVersion), PlanOnly: true, ExpectError: regexp.MustCompile("SMTP destination settings changed")},
			resource.TestStep{Config: changed(version), Check: check("second-secret", version)},
			resource.TestStep{Config: testKeycloakRealmSMTPPasswordWriteOnlyConfig("second-secret", fmt.Sprintf("reset-%d", i), "updated"), Check: check("second-secret", fmt.Sprintf("reset-%d", i))},
		)
	}
	steps = append(steps, []resource.TestStep{
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
	}...)
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: factories,
		Steps:                    steps,
	})

	// Verify that the emulator itself rejects a mask on a changed destination,
	// rather than silently hiding this regression with unconditional preservation.
	base := keycloak.Realm{Realm: "smtp-test", SmtpServer: keycloak.SmtpServer{
		Host: "smtp.example.com", From: "admin@example.com", User: "user", Auth: true, AuthType: "basic", Password: "secret",
	}}
	if err := client.NewRealm(context.Background(), &base); err != nil {
		t.Fatal(err)
	}
	masked := base
	masked.SmtpServer.Password = "**********"
	if err := client.UpdateRealm(context.Background(), &masked); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	if current.SmtpServer.Password != "secret" {
		t.Error("emulator did not preserve password with unchanged destination")
	}
	mu.Unlock()
	masked.SmtpServer.Host = "other.example.com"
	if err := client.UpdateRealm(context.Background(), &masked); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	if current.SmtpServer.Password != "" {
		t.Error("emulator preserved masked password with changed destination")
	}
	mu.Unlock()

	// Bypass planning and simulate destination drift before Apply. The provider
	// must refuse to write even when the Terraform diff only changes display_name.
	if err := client.UpdateRealm(context.Background(), &base); err != nil {
		t.Fatal(err)
	}
	data := schema.TestResourceDataRaw(t, resourceKeycloakRealm().Schema, map[string]interface{}{
		"realm": "smtp-test", "display_name": "original",
		"smtp_server": []interface{}{map[string]interface{}{
			"host": "other.example.com", "from": "admin@example.com",
			"auth": []interface{}{map[string]interface{}{"username": "user", "password_wo_version": "v1"}},
		}},
	})
	data.SetId("smtp-test")
	data = resourceKeycloakRealm().Data(data.State())
	if err := data.Set("display_name", "updated"); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	before := writes
	mu.Unlock()
	diags := resourceKeycloakRealmUpdate(context.Background(), data, client)
	if !diags.HasError() || !strings.Contains(diags[0].Summary, "SMTP destination settings changed") {
		t.Fatalf("expected SMTP destination error, got %v", diags)
	}
	mu.Lock()
	defer mu.Unlock()
	if writes != before || current.SmtpServer.Password != "secret" {
		t.Error("unsafe apply modified SMTP settings or password")
	}
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
