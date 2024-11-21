package provider

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func randomBool() bool {
	return rand.Intn(2) == 0
}

func randomStringInSlice(slice []string) string {
	return slice[acctest.RandIntRange(0, len(slice)-1)]
}

func randomStringSliceSubset(slice []string) []string {
	var result []string

	for _, s := range slice {
		if randomBool() {
			result = append(result, s)
		}
	}

	return result
}

// Returns a slice of strings in the format ["foo", "bar"] for
// use within terraform resource definitions for acceptance tests
func arrayOfStringsForTerraformResource(parts []string) string {
	var tfStrings []string

	for _, part := range parts {
		tfStrings = append(tfStrings, fmt.Sprintf(`"%s"`, part))
	}

	return "[" + strings.Join(tfStrings, ", ") + "]"
}

func randomDurationString() string {
	return (time.Duration(acctest.RandIntRange(1, 604800)) * time.Second).String()
}

func skipIfEnvSet(t *testing.T, envs ...string) {
	for _, k := range envs {
		if os.Getenv(k) != "" {
			t.Skipf("Environment variable %s is set, skipping...", k)
		}
	}
}

func skipIfEnvNotSet(t *testing.T, envs ...string) {
	for _, k := range envs {
		if os.Getenv(k) == "" {
			t.Skipf("Environment variable %s is not set, skipping...", k)
		}
	}
}

// Skips the test if the keycloak server matches a specific major version
func skipIfVersionIsLessThanOrEqualTo(ctx context.Context, t *testing.T, keycloakClient *keycloak.KeycloakClient, version keycloak.Version) {
	ok, err := keycloakClient.VersionIsLessThanOrEqualTo(ctx, version)
	if err != nil {
		t.Errorf("error checking keycloak version: %v", err)
	}

	if ok {
		t.Skipf("keycloak server version is less than or equal to %s, skipping...", version)
	}
}

func skipIfVersionIsGreaterThanOrEqualTo(ctx context.Context, t *testing.T, keycloakClient *keycloak.KeycloakClient, version keycloak.Version) {
	ok, err := keycloakClient.VersionIsGreaterThanOrEqualTo(ctx, version)
	if err != nil {
		t.Errorf("error checking keycloak version: %v", err)
	}

	if ok {
		t.Skipf("keycloak server version is greater than or equal to %s, skipping...", version)
	}
}

func TestCheckResourceAttrNot(name, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		err := resource.TestCheckResourceAttr(name, key, value)(s)
		if err == nil {
			return fmt.Errorf("%s: Attribute '%s' expected to not equal %#v", name, key, value)
		}

		return nil
	}
}
