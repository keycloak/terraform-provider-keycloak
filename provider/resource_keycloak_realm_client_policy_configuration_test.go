package provider

import "testing"

func TestKeycloakRealmClientPolicyConfigurationValidation(t *testing.T) {
	for _, tc := range []struct {
		body  string
		valid bool
	}{
		{`[]`, true}, {`[{"name":"one"}]`, true}, {`[{"name":"one"},{"name":"two"}]`, true},
		{`null`, false}, {`{}`, false}, {`[null]`, false}, {`["entry"]`, false}, {`[{}]`, false}, {`[{"name":" "}]`, false},
		{`[{"name":"one"},{"name":"one"}]`, false}, {`[] {}`, false},
	} {
		_, err := clientPolicyEntries(tc.body)
		if (err == nil) != tc.valid {
			t.Errorf("validation %q: %v", tc.body, err)
		}
	}
	a := `[{"name":"one","configuration":{"number":9007199254740993,"bool":false,"array":[1,"two"]}}]`
	b := `[ {"configuration":{"array":[1,"two"],"bool":false,"number":9007199254740993},"name":"one"} ]`
	if canonicalClientPolicyJSON(a) != canonicalClientPolicyJSON(b) {
		t.Fatal("object order or whitespace caused drift")
	}
	if canonicalClientPolicyJSON(`[{"name":"one"},{"name":"two"}]`) == canonicalClientPolicyJSON(`[{"name":"two"},{"name":"one"}]`) {
		t.Fatal("array order was lost")
	}
	if canonicalClientPolicyJSON(`[] {}`) != `[] {}` {
		t.Fatal("invalid trailing JSON was silently normalized")
	}
}
