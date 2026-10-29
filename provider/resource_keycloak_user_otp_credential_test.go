package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKeycloakUserOtpCredential_basic(t *testing.T) {
	t.Parallel()

	username := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUserOtpCredential_basic(username, "JBSWY3DPEHPK3PXP", "my-label"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasOtpCredential("keycloak_user_otp_credential.otp"),
					testAccCheckKeycloakUserHasNoConfigureTotpRequiredAction("keycloak_user.user"),
					resource.TestCheckResourceAttr("keycloak_user_otp_credential.otp", "algorithm", "HmacSHA1"),
					resource.TestCheckResourceAttr("keycloak_user_otp_credential.otp", "digits", "6"),
					resource.TestCheckResourceAttr("keycloak_user_otp_credential.otp", "period", "30"),
					resource.TestCheckResourceAttr("keycloak_user_otp_credential.otp", "user_label", "my-label"),
				),
			},
			// a new secret forces a replacement
			{
				Config: testKeycloakUserOtpCredential_basic(username, "GEZDGNBVGY3TQOJQ", "my-label"),
				Check:  testAccCheckKeycloakUserHasOtpCredential("keycloak_user_otp_credential.otp"),
			},
		},
	})
}

func TestAccKeycloakUserOtpCredential_customParameters(t *testing.T) {
	t.Parallel()

	username := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_user_otp_credential" "otp" {
	realm_id  = data.keycloak_realm.realm.id
	user_id   = keycloak_user.user.id
	secret    = "JBSWY3DPEHPK3PXP"
	algorithm = "HmacSHA256"
	digits    = 8
	period    = 60
}
`, testAccRealm.Realm, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasOtpCredential("keycloak_user_otp_credential.otp"),
					resource.TestCheckResourceAttr("keycloak_user_otp_credential.otp", "algorithm", "HmacSHA256"),
					resource.TestCheckResourceAttr("keycloak_user_otp_credential.otp", "digits", "8"),
					resource.TestCheckResourceAttr("keycloak_user_otp_credential.otp", "period", "60"),
				),
			},
		},
	})
}

func TestAccKeycloakUserOtpCredential_removedOutOfBand(t *testing.T) {
	t.Parallel()

	username := acctest.RandomWithPrefix("tf-acc")
	var realmId, userId, credentialId string

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUserOtpCredential_basic(username, "JBSWY3DPEHPK3PXP", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasOtpCredential("keycloak_user_otp_credential.otp"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["keycloak_user_otp_credential.otp"]
						realmId = rs.Primary.Attributes["realm_id"]
						userId = rs.Primary.Attributes["user_id"]
						credentialId = userOtpCredentialIdFromResourceId(rs.Primary.ID)

						return nil
					},
				),
			},
			// the credential is removed directly and must be recreated on the next apply
			{
				PreConfig: func() {
					if err := keycloakClient.DeleteUserOtpCredential(testCtx, realmId, userId, credentialId); err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakUserOtpCredential_basic(username, "JBSWY3DPEHPK3PXP", ""),
				Check:  testAccCheckKeycloakUserHasOtpCredential("keycloak_user_otp_credential.otp"),
			},
			// the credential is deleted when the resource is removed
			{
				Config: testKeycloakUserOtpCredential_noCredential(username),
				Check: func(s *terraform.State) error {
					if _, err := keycloakClient.GetUserOtpCredential(testCtx, realmId, userId, credentialId); err == nil {
						return fmt.Errorf("OTP credential %s still exists", credentialId)
					}

					return nil
				},
			},
		},
	})
}

func testAccCheckKeycloakUserHasOtpCredential(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		_, err := keycloakClient.GetUserOtpCredential(testCtx, rs.Primary.Attributes["realm_id"], rs.Primary.Attributes["user_id"], userOtpCredentialIdFromResourceId(rs.Primary.ID))

		return err
	}
}

func testAccCheckKeycloakUserHasNoConfigureTotpRequiredAction(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		user, err := keycloakClient.GetUser(testCtx, rs.Primary.Attributes["realm_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		for _, action := range user.RequiredActions {
			if action == "CONFIGURE_TOTP" {
				return fmt.Errorf("user %s still has the CONFIGURE_TOTP required action", user.Username)
			}
		}

		return nil
	}
}

func testKeycloakUserOtpCredential_basic(username, secret, label string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_user" "user" {
	realm_id         = data.keycloak_realm.realm.id
	username         = "%s"
	required_actions = ["CONFIGURE_TOTP"]

	# Keycloak clears CONFIGURE_TOTP when the OTP credential is created
	lifecycle {
		ignore_changes = [required_actions]
	}
}

resource "keycloak_user_otp_credential" "otp" {
	realm_id   = data.keycloak_realm.realm.id
	user_id    = keycloak_user.user.id
	secret     = "%s"
	user_label = "%s"
}
	`, testAccRealm.Realm, username, secret, label)
}

func testKeycloakUserOtpCredential_noCredential(username string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_user" "user" {
	realm_id         = data.keycloak_realm.realm.id
	username         = "%s"
	required_actions = ["CONFIGURE_TOTP"]

	lifecycle {
		ignore_changes = [required_actions]
	}
}
	`, testAccRealm.Realm, username)
}
