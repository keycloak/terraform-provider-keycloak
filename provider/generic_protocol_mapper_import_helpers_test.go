package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func getGenericProtocolMapperIdForClient(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmId := rs.Primary.Attributes["realm_id"]
		clientId := rs.Primary.Attributes["client_id"]

		return fmt.Sprintf("%s/client/%s/%s", realmId, clientId, id), nil
	}
}

func getGenericProtocolMapperIdForClientScope(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmId := rs.Primary.Attributes["realm_id"]
		clientScopeId := rs.Primary.Attributes["client_scope_id"]

		return fmt.Sprintf("%s/client-scope/%s/%s", realmId, clientScopeId, id), nil
	}
}
