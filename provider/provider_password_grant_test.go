package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"testing"
)

func TestAccKeycloakProvider_passwordGrant(t *testing.T) {
	skipIfEnvNotSet(t, "KEYCLOAK_TEST_PASSWORD_GRANT")

	t.Parallel()

	os.Setenv("KEYCLOAK_USER", "keycloak")
	os.Setenv("KEYCLOAK_PASSWORD", "password")

	defer func() {
		os.Unsetenv("KEYCLOAK_USER")
		os.Unsetenv("KEYCLOAK_PASSWORD")
	}()

	provider := KeycloakProvider(keycloakClient)

	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"keycloak": func() (*schema.Provider, error) {
				return provider, nil
			},
		},
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_basic(clientId),
			},
		},
	})
}
