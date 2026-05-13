package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// crpExists is a shared TestCheckFunc that verifies a client_registration_policy_*
// resource has been created in Keycloak.
func crpExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found in state: %s", resourceName)
		}
		_, err := keycloakClient.GetClientRegistrationPolicy(testCtx, rs.Primary.Attributes["realm_id"], rs.Primary.ID)
		return err
	}
}

// crpDestroy returns a CheckDestroy that asserts every resource of the given
// type has actually been removed from Keycloak.
func crpDestroy(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
			policy, _ := keycloakClient.GetClientRegistrationPolicy(testCtx, rs.Primary.Attributes["realm_id"], rs.Primary.ID)
			if policy != nil {
				return fmt.Errorf("%s policy %s still exists", resourceType, rs.Primary.ID)
			}
		}
		return nil
	}
}

// crpImportId returns an ImportStateIdFunc that builds the canonical
// {realmId}/{componentId} import id from a resource in state.
func crpImportId(resourceName string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found in state: %s", resourceName)
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["realm_id"], rs.Primary.ID), nil
	}
}
