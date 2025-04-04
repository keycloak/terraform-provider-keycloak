package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"strings"
	"testing"
)

func TestAccKeycloakOpenidClientServiceAccountRealmRole_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_client_service_account_realm_role.test"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientServiceAccountRealmRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientServiceAccountRealmRole_basic(clientId),
				Check:  testAccCheckKeycloakOpenidClientServiceAccountRealmRoleExists(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getKeycloakOpenidClientServiceAccountRealmRoleImportId(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenidClientServiceAccountRealmRole_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var serviceAccountRole = &keycloak.OpenidClientServiceAccountRealmRole{}

	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientServiceAccountRealmRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientServiceAccountRealmRole_basic(clientId),
				Check:  testAccCheckKeycloakOpenidClientServiceAccountRealmRoleFetch("keycloak_openid_client_service_account_realm_role.test", serviceAccountRole),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenidClientServiceAccountRealmRole(testCtx, serviceAccountRole.RealmId, serviceAccountRole.ServiceAccountUserId, serviceAccountRole.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOpenidClientServiceAccountRealmRole_basic(clientId),
				Check:  testAccCheckKeycloakOpenidClientServiceAccountRealmRoleExists("keycloak_openid_client_service_account_realm_role.test"),
			},
		},
	})
}

func testAccCheckKeycloakOpenidClientServiceAccountRealmRoleExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakOpenidClientServiceAccountRealmRoleFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientServiceAccountRealmRoleFetch(resourceName string, serviceAccountRole *keycloak.OpenidClientServiceAccountRealmRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedServiceAccountRole, err := getKeycloakOpenidClientServiceAccountRealmRoleFromState(s, resourceName)
		if err != nil {
			return err
		}

		serviceAccountRole.ServiceAccountUserId = fetchedServiceAccountRole.ServiceAccountUserId
		serviceAccountRole.RealmId = fetchedServiceAccountRole.RealmId
		serviceAccountRole.Id = fetchedServiceAccountRole.Id

		return nil
	}
}

func testAccCheckKeycloakOpenidClientServiceAccountRealmRoleDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client_service_account_realm_role" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			serviceAccountUserId := rs.Primary.Attributes["service_account_user_id"]
			id := strings.Split(rs.Primary.ID, "/")[1]

			serviceAccountRole, _ := keycloakClient.GetOpenidClientServiceAccountRealmRole(testCtx, realm, serviceAccountUserId, id)
			if serviceAccountRole != nil {
				return fmt.Errorf("service account role exists")
			}
		}

		return nil
	}
}

func getKeycloakOpenidClientServiceAccountRealmRoleFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClientServiceAccountRealmRole, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	serviceAccountUserId := rs.Primary.Attributes["service_account_user_id"]
	id := strings.Split(rs.Primary.ID, "/")[1]

	serviceAccountRole, err := keycloakClient.GetOpenidClientServiceAccountRealmRole(testCtx, realm, serviceAccountUserId, id)
	if err != nil {
		return nil, fmt.Errorf("error getting service account role mapping: %s", err)
	}

	return serviceAccountRole, nil
}

func getKeycloakOpenidClientServiceAccountRealmRoleImportId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		serviceAccountRole, err := getKeycloakOpenidClientServiceAccountRealmRoleFromState(s, resourceName)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s/%s/%s", serviceAccountRole.RealmId, serviceAccountRole.ServiceAccountUserId, serviceAccountRole.Id), nil
	}
}

func testKeycloakOpenidClientServiceAccountRealmRole_basic(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource keycloak_openid_client test {
	client_id                = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	access_type              = "CONFIDENTIAL"
	service_accounts_enabled = true
}

resource keycloak_openid_client_service_account_realm_role test {
	service_account_user_id = keycloak_openid_client.test.service_account_user_id
	realm_id                = data.keycloak_realm.realm.id
	role                    = "offline_access"
}
	`, testAccRealm.Realm, clientId)
}
