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

func TestAccKeycloakOpenidClientServiceAccountRole_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_client_service_account_role.test"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientServiceAccountRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientServiceAccountRole_basic(clientId),
				Check:  testAccCheckKeycloakOpenidClientServiceAccountRoleExists(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getKeycloakOpenidClientServiceAccountRoleImportId(resourceName),
			},
		},
	})
}

func TestAccKeycloakOpenidClientServiceAccountRole_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var serviceAccountRole = &keycloak.OpenidClientServiceAccountRole{}

	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientServiceAccountRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientServiceAccountRole_basic(clientId),
				Check:  testAccCheckKeycloakOpenidClientServiceAccountRoleFetch("keycloak_openid_client_service_account_role.test", serviceAccountRole),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenidClientServiceAccountRole(testCtx, serviceAccountRole.RealmId, serviceAccountRole.ServiceAccountUserId, serviceAccountRole.ContainerId, serviceAccountRole.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOpenidClientServiceAccountRole_basic(clientId),
				Check:  testAccCheckKeycloakOpenidClientServiceAccountRoleExists("keycloak_openid_client_service_account_role.test"),
			},
		},
	})
}

func TestAccKeycloakOpenidClientServiceAccountRole_enableAfterCreate(t *testing.T) {
	t.Parallel()
	bearerClientId := acctest.RandomWithPrefix("tf-acc")
	consumerClientId := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_openid_client_service_account_role.consumer_service_account_role"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientServiceAccountRoleDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClientServiceAccountRole_enableAfterCreate_before(bearerClientId, consumerClientId),
			},
			{
				Config: testKeycloakOpenidClientServiceAccountRole_enableAfterCreate_after(bearerClientId, consumerClientId),
				Check:  testAccCheckKeycloakOpenidClientServiceAccountRoleExists(resourceName),
			},
		},
	})
}

func testAccCheckKeycloakOpenidClientServiceAccountRoleExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakOpenidClientServiceAccountRoleFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientServiceAccountRoleFetch(resourceName string, serviceAccountRole *keycloak.OpenidClientServiceAccountRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedServiceAccountRole, err := getKeycloakOpenidClientServiceAccountRoleFromState(s, resourceName)
		if err != nil {
			return err
		}

		serviceAccountRole.ServiceAccountUserId = fetchedServiceAccountRole.ServiceAccountUserId
		serviceAccountRole.RealmId = fetchedServiceAccountRole.RealmId
		serviceAccountRole.ClientRole = fetchedServiceAccountRole.ClientRole
		serviceAccountRole.ContainerId = fetchedServiceAccountRole.ContainerId
		serviceAccountRole.Id = fetchedServiceAccountRole.Id

		return nil
	}
}

func testAccCheckKeycloakOpenidClientServiceAccountRoleDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client_service_account_role" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]
			serviceAccountUserId := rs.Primary.Attributes["service_account_user_id"]
			clientId := rs.Primary.Attributes["client_id"]
			id := strings.Split(rs.Primary.ID, "/")[1]

			serviceAccountRole, _ := keycloakClient.GetOpenidClientServiceAccountRole(testCtx, realm, serviceAccountUserId, clientId, id)
			if serviceAccountRole != nil {
				return fmt.Errorf("service account role exists")
			}
		}

		return nil
	}
}

func getKeycloakOpenidClientServiceAccountRoleFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClientServiceAccountRole, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]
	serviceAccountUserId := rs.Primary.Attributes["service_account_user_id"]
	clientId := rs.Primary.Attributes["client_id"]
	id := strings.Split(rs.Primary.ID, "/")[1]

	serviceAccountRole, err := keycloakClient.GetOpenidClientServiceAccountRole(testCtx, realm, serviceAccountUserId, clientId, id)
	if err != nil {
		return nil, fmt.Errorf("error getting service account role mapping: %s", err)
	}

	return serviceAccountRole, nil
}

func getKeycloakOpenidClientServiceAccountRoleImportId(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		realmId := rs.Primary.Attributes["realm_id"]
		serviceAccountUserId := rs.Primary.Attributes["service_account_user_id"]
		clientId := rs.Primary.Attributes["client_id"]
		roleId := strings.Split(rs.Primary.ID, "/")[1]

		return fmt.Sprintf("%s/%s/%s/%s", realmId, serviceAccountUserId, clientId, roleId), nil
	}
}

func testKeycloakOpenidClientServiceAccountRole_basic(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "test" {
	client_id                = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	access_type              = "CONFIDENTIAL"
	service_accounts_enabled = true
}

data "keycloak_openid_client" "broker" {
  realm_id  = data.keycloak_realm.realm.id
  client_id = "broker"
}

resource "keycloak_openid_client_service_account_role" "test" {
	realm_id                = data.keycloak_realm.realm.id
	client_id               = data.keycloak_openid_client.broker.id
	service_account_user_id = keycloak_openid_client.test.service_account_user_id
	role                    = "read-token"
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakOpenidClientServiceAccountRole_enableAfterCreate_before(bearerClientId, consumerClientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "bearer" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "BEARER-ONLY"
}

resource "keycloak_role" "bearer_role" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.bearer.id
	name      = "bearer-role"
}

resource "keycloak_openid_client" "consumer" {
  realm_id  = data.keycloak_realm.realm.id
  client_id = "%s"

  access_type              = "CONFIDENTIAL"
  service_accounts_enabled = false
}
	`, testAccRealm.Realm, bearerClientId, consumerClientId)
}

func testKeycloakOpenidClientServiceAccountRole_enableAfterCreate_after(bearerClientId, consumerClientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "bearer" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "BEARER-ONLY"
}

resource "keycloak_role" "bearer_role" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.bearer.id
	name      = "bearer-role"
}

resource "keycloak_openid_client" "consumer" {
  realm_id  = data.keycloak_realm.realm.id
  client_id = "%s"

  access_type              = "CONFIDENTIAL"
  service_accounts_enabled = true
}

resource "keycloak_openid_client_service_account_role" "consumer_service_account_role" {
  realm_id                = data.keycloak_realm.realm.id
  service_account_user_id = keycloak_openid_client.consumer.service_account_user_id
  client_id               = keycloak_openid_client.bearer.id
  role                    = keycloak_role.bearer_role.name
}
	`, testAccRealm.Realm, bearerClientId, consumerClientId)
}
