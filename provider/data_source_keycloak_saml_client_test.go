package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccKeycloakDataSourceSamlClient_basic(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.keycloak_saml_client.test"
	resourceName := "keycloak_saml_client.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakSamlClientConfig(clientId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id", resourceName, "client_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "realm_id", resourceName, "realm_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "include_authn_statement", resourceName, "include_authn_statement"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sign_documents", resourceName, "sign_documents"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sign_assertions", resourceName, "sign_assertions"),
					resource.TestCheckResourceAttrPair(dataSourceName, "encrypt_assertions", resourceName, "encrypt_assertions"),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_signature_required", resourceName, "client_signature_required"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signature_key_name", resourceName, "signature_key_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "force_post_binding", resourceName, "force_post_binding"),
					resource.TestCheckResourceAttrPair(dataSourceName, "front_channel_logout", resourceName, "front_channel_logout"),
					resource.TestCheckResourceAttrPair(dataSourceName, "force_name_id_format", resourceName, "force_name_id_format"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signature_algorithm", resourceName, "signature_algorithm"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name_id_format", resourceName, "name_id_format"),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_url", resourceName, "root_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "valid_redirect_uris", resourceName, "valid_redirect_uris"),
					resource.TestCheckResourceAttrPair(dataSourceName, "base_url", resourceName, "base_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "master_saml_processing_url", resourceName, "master_saml_processing_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "encryption_certificate", resourceName, "encryption_certificate"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_certificate", resourceName, "signing_certificate"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_private_key", resourceName, "signing_private_key"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idp_initiated_sso_url_name", resourceName, "idp_initiated_sso_url_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idp_initiated_sso_relay_state", resourceName, "idp_initiated_sso_relay_state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "assertion_consumer_post_url", resourceName, "assertion_consumer_post_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "assertion_consumer_redirect_url", resourceName, "assertion_consumer_redirect_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "logout_service_post_binding_url", resourceName, "logout_service_post_binding_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "logout_service_redirect_binding_url", resourceName, "logout_service_redirect_binding_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "full_scope_allowed", resourceName, "full_scope_allowed"),
					resource.TestCheckResourceAttrPair(dataSourceName, "login_theme", resourceName, "login_theme"),
				),
			},
		},
	})
}

func testAccKeycloakSamlClientConfig(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource keycloak_saml_client test {
	client_id = "%s"
	realm_id  = data.keycloak_realm.realm.id
}

data keycloak_saml_client test {
	client_id = keycloak_saml_client.test.client_id
	realm_id  = data.keycloak_realm.realm.id

	depends_on = [
		keycloak_saml_client.test
	]
}
`, testAccRealm.Realm, clientId)
}
