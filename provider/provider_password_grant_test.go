package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
		ProtoV5ProviderFactories: protoV5ProviderFactories(provider),
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_basic(clientId),
			},
		},
	})
}
