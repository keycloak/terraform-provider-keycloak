package keycloak

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"
)

// Everything the client logs at DEBUG level while it logs in, refreshes its token and talks to the admin API.
func TestDebugLogsDoNotContainSecrets(t *testing.T) {
	// request bodies are indented when TF_LOG is DEBUG
	t.Setenv("TF_LOG", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/protocol/openid-connect/token"):
			w.Write([]byte(`{"access_token":"ACCESS_TOKEN_SECRET","refresh_token":"REFRESH_TOKEN_SECRET","token_type":"Bearer","expires_in":60}`))
		case strings.HasSuffix(r.URL.Path, "/serverinfo"):
			w.Write([]byte(`{"systemInfo":{"version":"26.0.0"}}`))
		case strings.HasSuffix(r.URL.Path, "/client-secret"):
			w.Write([]byte(`{"type":"secret","value":"ROTATED_SECRET"}`))
		case strings.HasSuffix(r.URL.Path, "/localization/en/greeting"):
			w.WriteHeader(http.StatusNoContent)
		default:
			w.Write([]byte(`{"id":"abc","clientId":"app","secret":"RETURNED_SECRET"}`))
		}
	}))
	t.Cleanup(server.Close)

	var output bytes.Buffer
	ctx := tflogtest.RootLogger(context.Background(), &output)

	client, err := NewKeycloakClient(ctx, server.URL, "", "", "my-client", "CLIENT_SECRET_VALUE", "master", "admin", "PASSWORD_VALUE", "", "RS256", "", "", "", true, 30, "", false, "", "", "", false, map[string]string{}, "26.0.0")
	if err != nil {
		t.Fatalf("NewKeycloakClient: %v", err)
	}
	if err := client.Refresh(ctx); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	federation := map[string]interface{}{
		"name":   "ldap",
		"config": map[string][]string{"bindCredential": {"BIND_CREDENTIAL_VALUE"}, "bindDn": {"cn=admin"}},
	}
	if _, _, err := client.post(ctx, "/realms/test/components", federation); err != nil {
		t.Fatalf("post: %v", err)
	}
	if err := client.put(ctx, "/realms/test/users/abc/reset-password", PasswordCredentials{Type: "password", Value: "USER_PASSWORD_VALUE"}); err != nil {
		t.Fatalf("put: %v", err)
	}
	if _, err := client.getRaw(ctx, "/realms/test/clients/abc", nil); err != nil {
		t.Fatalf("get client: %v", err)
	}
	if _, err := client.getRaw(ctx, "/realms/test/clients/abc/client-secret", nil); err != nil {
		t.Fatalf("get client secret: %v", err)
	}
	if err := client.putPlain(ctx, "/realms/test/localization/en/greeting", "PLAIN_TEXT_VALUE"); err != nil {
		t.Fatalf("putPlain: %v", err)
	}

	providedToken := &KeycloakClient{
		baseUrl: server.URL, httpClient: server.Client(), accessTokenProvided: true,
		clientCredentials: &ClientCredentials{AccessToken: "PROVIDED_ACCESS_TOKEN", TokenType: "bearer"},
	}
	if err := providedToken.login(ctx); err != nil {
		t.Fatalf("login with a provided access token: %v", err)
	}

	logs := output.String()
	for _, secret := range []string{
		"CLIENT_SECRET_VALUE", "PASSWORD_VALUE", "ACCESS_TOKEN_SECRET", "REFRESH_TOKEN_SECRET", "PROVIDED_ACCESS_TOKEN",
		"BIND_CREDENTIAL_VALUE", "USER_PASSWORD_VALUE", "RETURNED_SECRET", "ROTATED_SECRET", "PLAIN_TEXT_VALUE",
	} {
		if strings.Contains(logs, secret) {
			t.Errorf("%q is logged", secret)
		}
	}

	entries, err := tflogtest.MultilineJSONDecode(&output)
	if err != nil {
		t.Fatalf("cannot decode the logs: %v", err)
	}
	logged := map[string]bool{}
	for _, entry := range entries {
		for _, field := range []string{"request", "response", "body"} {
			if value, ok := entry[field].(string); ok {
				logged[entry["@message"].(string)+": "+value] = true
			}
		}
	}

	// what is no secret is still there to debug with
	for _, want := range []string{
		"Login request: client_id=my-client&client_secret=[REDACTED]&grant_type=password&password=[REDACTED]&username=admin",
		"Refresh request: client_id=my-client&client_secret=[REDACTED]&grant_type=password&password=[REDACTED]&username=admin",
		`Login response: {"access_token":"[REDACTED]","expires_in":60,"refresh_token":"[REDACTED]","token_type":"Bearer"}`,
		`Refresh response: {"access_token":"[REDACTED]","expires_in":60,"refresh_token":"[REDACTED]","token_type":"Bearer"}`,
		`Sending request: {"config":{"bindCredential":["[REDACTED]"],"bindDn":["cn=admin"]},"name":"ldap"}`,
		`Sending request: {"temporary":false,"type":"password","value":"[REDACTED]"}`,
		`Received response: {"clientId":"app","id":"abc","secret":"[REDACTED]"}`,
		`Received response: {"type":"secret","value":"[REDACTED]"}`,
		"Sending request: [16 bytes omitted: not a JSON object or array]",
	} {
		if !logged[want] {
			t.Errorf("missing log entry %q in:\n%s", want, logs)
		}
	}
}

func TestRedactJSONBody(t *testing.T) {
	for name, test := range map[string]struct {
		body   string
		indent bool
		want   string
	}{
		"realm settings that only look like secrets are kept": {
			body: `{"passwordPolicy":"length(8)","resetPasswordAllowed":true,"resetCredentialsFlow":"reset credentials","accessTokenLifespan":300,"webAuthnPolicyPasswordlessRpId":"example.com","smtpServer":{"host":"smtp","password":"s3cret","authTokenClientSecret":"s3cret"}}`,
			want: `{"accessTokenLifespan":300,"passwordPolicy":"length(8)","resetCredentialsFlow":"reset credentials","resetPasswordAllowed":true,"smtpServer":{"authTokenClientSecret":"[REDACTED]","host":"smtp","password":"[REDACTED]"},"webAuthnPolicyPasswordlessRpId":"example.com"}`,
		},
		"client attributes and tokens": {
			body: `[{"registrationAccessToken":"s3cret","attributes":{"saml.signing.private.key":"s3cret","client.secret.rotated":"s3cret","access.token.lifespan":"60","use.refresh.tokens":"true"}}]`,
			want: `[{"attributes":{"access.token.lifespan":"60","client.secret.rotated":"[REDACTED]","saml.signing.private.key":"[REDACTED]","use.refresh.tokens":"true"},"registrationAccessToken":"[REDACTED]"}]`,
		},
		"component config": {
			body: `{"config":{"keystorePassword":["s3cret"],"privateKey":["s3cret"],"usePasswordModifyExtendedOp":["true"],"secretSize":["64"]}}`,
			want: `{"config":{"keystorePassword":["[REDACTED]"],"privateKey":["[REDACTED]"],"secretSize":["64"],"usePasswordModifyExtendedOp":["true"]}}`,
		},
		"everything below a sensitive key": {
			body: `{"credentials":[{"type":"otp","secretData":"s3cret","userLabel":"s3cret","createdDate":1}]}`,
			want: `{"credentials":[{"createdDate":1,"secretData":"[REDACTED]","type":"[REDACTED]","userLabel":"[REDACTED]"}]}`,
		},
		"an empty secret stays empty": {
			body: `{"secret":"","url":"https://example.com/?a=b&c=<d>"}`,
			want: `{"secret":"","url":"https://example.com/?a=b&c=<d>"}`,
		},
		"indentation": {
			body:   `{"secret":"s3cret"}`,
			indent: true,
			want:   "{\n    \"secret\": \"[REDACTED]\"\n}",
		},
		"a JSON string":            {body: `"s3cret"`, want: "[8 bytes omitted: not a JSON object or array]"},
		"a truncated JSON object":  {body: `{"secret":"s3cret"`, want: "[18 bytes omitted: not a JSON object or array]"},
		"a body that is not JSON":  {body: "s3cret", want: "[6 bytes omitted: not a JSON object or array]"},
		"data after a JSON object": {body: `{"a":1} s3cret`, want: `{"a":1}`},
	} {
		t.Run(name, func(t *testing.T) {
			if got := redactJSONBody([]byte(test.body), test.indent); got != test.want {
				t.Errorf("got  %s\nwant %s", got, test.want)
			}
		})
	}
}
