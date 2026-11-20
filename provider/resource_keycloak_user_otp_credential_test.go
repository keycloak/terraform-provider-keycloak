package provider

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
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
					testAccCheckKeycloakUserUntouched("keycloak_user.user"),
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasOtpCredential("keycloak_user_otp_credential.otp"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["keycloak_user_otp_credential.otp"]
						recreatedId := userOtpCredentialIdFromResourceId(rs.Primary.ID)
						if recreatedId == credentialId {
							return fmt.Errorf("expected the OTP credential to be recreated with a new ID")
						}
						credentialId = recreatedId

						return nil
					},
				),
			},
			// the credential is deleted when the resource is removed
			{
				Config: testKeycloakUserOtpCredential_noCredential(username),
				Check: func(s *terraform.State) error {
					_, err := keycloakClient.GetUserOtpCredential(testCtx, realmId, userId, credentialId)
					if err == nil {
						return fmt.Errorf("OTP credential %s still exists", credentialId)
					}
					if !keycloak.ErrorIs404(err) {
						return fmt.Errorf("could not check whether OTP credential %s was deleted: %w", credentialId, err)
					}

					return nil
				},
			},
		},
	})
}

// Users can rename their credentials in the account console, which must not replace the credential.
func TestAccKeycloakUserOtpCredential_labelChangedOutOfBand(t *testing.T) {
	t.Parallel()

	username := acctest.RandomWithPrefix("tf-acc")
	var realmId, userId, credentialId string

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUserOtpCredential_basic(username, "JBSWY3DPEHPK3PXP", "my-label"),
				Check: func(s *terraform.State) error {
					rs := s.RootModule().Resources["keycloak_user_otp_credential.otp"]
					realmId = rs.Primary.Attributes["realm_id"]
					userId = rs.Primary.Attributes["user_id"]
					credentialId = userOtpCredentialIdFromResourceId(rs.Primary.ID)

					return nil
				},
			},
			{
				PreConfig: func() {
					if err := testAccSetCredentialLabel(realmId, userId, credentialId, "renamed by the user"); err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakUserOtpCredential_basic(username, "JBSWY3DPEHPK3PXP", "my-label"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: func(s *terraform.State) error {
					rs := s.RootModule().Resources["keycloak_user_otp_credential.otp"]
					if id := userOtpCredentialIdFromResourceId(rs.Primary.ID); id != credentialId {
						return fmt.Errorf("OTP credential was replaced: %s -> %s", credentialId, id)
					}

					return nil
				},
			},
		},
	})
}

// testAccSetCredentialLabel renames a credential the way a user does in the account console. The admin
// endpoint takes plain text, which the client of the provider does not send.
func testAccSetCredentialLabel(realmId, userId, credentialId, label string) error {
	baseUrl := os.Getenv("KEYCLOAK_URL")

	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	form.Add("client_id", os.Getenv("KEYCLOAK_CLIENT_ID"))
	form.Add("client_secret", os.Getenv("KEYCLOAK_CLIENT_SECRET"))

	tokenResponse, err := keycloakClient.GetHttpClient().PostForm(fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", baseUrl, os.Getenv("KEYCLOAK_REALM")), form)
	if err != nil {
		return err
	}
	defer tokenResponse.Body.Close()

	var token struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(tokenResponse.Body).Decode(&token); err != nil || token.AccessToken == "" {
		return fmt.Errorf("could not get an access token (status %d): %v", tokenResponse.StatusCode, err)
	}

	request, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/admin/realms/%s/users/%s/credentials/%s/userLabel", baseUrl, realmId, userId, credentialId), strings.NewReader(label))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token.AccessToken)
	request.Header.Set("Content-Type", "text/plain")

	response, err := keycloakClient.GetHttpClient().Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("could not rename the credential (status %d): %s", response.StatusCode, string(body))
	}

	return nil
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

// Creating the credential must leave the rest of the user alone, including the required actions. The plan
// after the apply is checked for drift by the test framework, this also checks the user in Keycloak.
func testAccCheckKeycloakUserUntouched(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		user, err := keycloakClient.GetUser(testCtx, rs.Primary.Attributes["realm_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if user.Email != user.Username+"@example.com" || user.FirstName != "Test" || user.LastName != "User" || !user.Enabled {
			return fmt.Errorf("user fields changed: %+v", user)
		}

		if len(user.RequiredActions) != 2 || user.RequiredActions[0] != "CONFIGURE_TOTP" || user.RequiredActions[1] != "UPDATE_PASSWORD" {
			return fmt.Errorf("required actions changed: %v", user.RequiredActions)
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
	email            = "%s@example.com"
	first_name       = "Test"
	last_name        = "User"
	required_actions = ["CONFIGURE_TOTP", "UPDATE_PASSWORD"]
}

resource "keycloak_user_otp_credential" "otp" {
	realm_id   = data.keycloak_realm.realm.id
	user_id    = keycloak_user.user.id
	secret     = "%s"
	user_label = "%s"
}
	`, testAccRealm.Realm, username, username, secret, label)
}

func testKeycloakUserOtpCredential_noCredential(username string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_user" "user" {
	realm_id         = data.keycloak_realm.realm.id
	username         = "%s"
	email            = "%s@example.com"
	first_name       = "Test"
	last_name        = "User"
	required_actions = ["CONFIGURE_TOTP", "UPDATE_PASSWORD"]
}
	`, testAccRealm.Realm, username, username)
}

func TestAccKeycloakUserOtpCredential_invalidSecret(t *testing.T) {
	t.Parallel()

	username := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakUserOtpCredential_basic(username, "not base32!", ""),
				ExpectError: regexp.MustCompile("must be a non-empty, valid Base32 string"),
			},
			{
				Config:      testKeycloakUserOtpCredential_basic(username, "A", ""),
				ExpectError: regexp.MustCompile("must be a non-empty, valid Base32 string"),
			},
			{
				Config:      testKeycloakUserOtpCredential_basic(username, "JBSWY3DP========", ""),
				ExpectError: regexp.MustCompile("wrong amount of = padding"),
			},
			{
				Config:      testKeycloakUserOtpCredential_basic(username, "JBSWY3DP=", ""),
				ExpectError: regexp.MustCompile("wrong amount of = padding"),
			},
		},
	})
}

// The secret can only be checked by logging in: the direct grant flow asks for a one-time password
// of a user that has an OTP credential.
func TestAccKeycloakUserOtpCredential_login(t *testing.T) {
	t.Parallel()

	username := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")

	tests := []struct {
		name      string
		secret    string
		algorithm string
		digits    int
		period    int
	}{
		{"defaults", "JBSWY3DPEHPK3PXP", "HmacSHA1", 6, 30},
		{"padded", "JBSWY3DPEHPK3PXPJBSWY3DPEE======", "HmacSHA1", 6, 30},
		{"lowercase", "jbswy3dpehpk3pxp", "HmacSHA1", 6, 30},
		{"custom", "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", "HmacSHA256", 8, 60},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			user := fmt.Sprintf("%s-%s", username, test.name)

			resource.Test(t, resource.TestCase{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				PreCheck:                 func() { testAccPreCheck(t) },
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	realm_id                     = data.keycloak_realm.realm.id
	client_id                    = "%s"
	access_type                  = "PUBLIC"
	direct_access_grants_enabled = true
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username   = "%s"
	email      = "%s@example.com"
	first_name = "Test"
	last_name  = "User"

	initial_password {
		value = "password"
	}
}

resource "keycloak_user_otp_credential" "otp" {
	realm_id  = data.keycloak_realm.realm.id
	user_id   = keycloak_user.user.id
	secret    = "%s"
	algorithm = "%s"
	digits    = %d
	period    = %d
}
`, testAccRealm.Realm, clientId+test.name, user, user, test.secret, test.algorithm, test.digits, test.period),
						Check: testAccCheckKeycloakUserOtpLogin(user, "password", clientId+test.name, test.secret, test.algorithm, test.digits, test.period),
					},
				},
			})
		})
	}
}

func testAccCheckKeycloakUserOtpLogin(username, password, clientId, secret, algorithm string, digits, period int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// without a one-time password the login must fail, otherwise this test proves nothing
		if err := testAccKeycloakPasswordGrant(username, password, clientId, ""); err == nil {
			return fmt.Errorf("user %s could log in without a one-time password", username)
		}

		code, err := totpCode(secret, algorithm, digits, period, time.Now())
		if err != nil {
			return err
		}

		return testAccKeycloakPasswordGrant(username, password, clientId, code)
	}
}

func testAccKeycloakPasswordGrant(username, password, clientId, totp string) error {
	resourceUrl := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", os.Getenv("KEYCLOAK_URL"), testAccRealm.Realm)

	form := url.Values{}
	form.Add("username", username)
	form.Add("password", password)
	form.Add("client_id", clientId)
	form.Add("grant_type", "password")
	if totp != "" {
		form.Add("totp", totp)
	}

	request, err := http.NewRequest(http.MethodPost, resourceUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := keycloakClient.GetHttpClient().Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("login of %s failed with status %d: %s", username, response.StatusCode, string(body))
	}

	return nil
}

// totpCode is RFC 6238 for a Base32 secret.
func totpCode(secret, algorithm string, digits, period int, at time.Time) (string, error) {
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		key, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
		if err != nil {
			return "", err
		}
	}

	var newHash func() hash.Hash
	switch algorithm {
	case "HmacSHA1":
		newHash = sha1.New
	case "HmacSHA256":
		newHash = sha256.New
	case "HmacSHA512":
		newHash = sha512.New
	default:
		return "", fmt.Errorf("unsupported algorithm %s", algorithm)
	}

	counter := make([]byte, 8)
	binary.BigEndian.PutUint64(counter, uint64(at.Unix()/int64(period)))

	mac := hmac.New(newHash, key)
	mac.Write(counter)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff

	modulus := uint32(1)
	for i := 0; i < digits; i++ {
		modulus *= 10
	}

	return fmt.Sprintf("%0*d", digits, value%modulus), nil
}

// With a write-only secret nothing secret ends up in the state, and a new version replaces the credential.
func TestAccKeycloakUserOtpCredential_writeOnly(t *testing.T) {
	t.Parallel()

	username := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")
	secret := "JBSWY3DPEHPK3PXP"
	newSecret := "GEZDGNBVGY3TQOJQ"

	var credentialId string

	checkNoSecretInState := func(s *terraform.State) error {
		for key, value := range s.RootModule().Resources["keycloak_user_otp_credential.otp"].Primary.Attributes {
			if strings.Contains(value, secret) || strings.Contains(value, newSecret) {
				return fmt.Errorf("the secret is stored in the state in %s", key)
			}
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUserOtpCredential_writeOnly(username, clientId, secret, "v1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasOtpCredential("keycloak_user_otp_credential.otp"),
					resource.TestCheckNoResourceAttr("keycloak_user_otp_credential.otp", "secret"),
					resource.TestCheckNoResourceAttr("keycloak_user_otp_credential.otp", "secret_wo"),
					resource.TestCheckResourceAttr("keycloak_user_otp_credential.otp", "secret_wo_version", "v1"),
					checkNoSecretInState,
					testAccCheckKeycloakUserOtpLogin(username, "password", clientId, secret, "HmacSHA1", 6, 30),
					func(s *terraform.State) error {
						credentialId = userOtpCredentialIdFromResourceId(s.RootModule().Resources["keycloak_user_otp_credential.otp"].Primary.ID)

						return nil
					},
				),
			},
			{
				// the same version keeps the credential, even when the secret in the configuration differs
				Config: testKeycloakUserOtpCredential_writeOnly(username, clientId, newSecret, "v1"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				Config: testKeycloakUserOtpCredential_writeOnly(username, clientId, newSecret, "v2"),
				Check: resource.ComposeTestCheckFunc(
					checkNoSecretInState,
					testAccCheckKeycloakUserOtpLogin(username, "password", clientId, newSecret, "HmacSHA1", 6, 30),
					func(s *terraform.State) error {
						if id := userOtpCredentialIdFromResourceId(s.RootModule().Resources["keycloak_user_otp_credential.otp"].Primary.ID); id == credentialId {
							return fmt.Errorf("expected a new version to replace the OTP credential")
						}

						return nil
					},
				),
			},
		},
	})
}

func TestAccKeycloakUserOtpCredential_secretArguments(t *testing.T) {
	t.Parallel()

	username := acctest.RandomWithPrefix("tf-acc")

	config := func(arguments string) string {
		return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
}

resource "keycloak_user_otp_credential" "otp" {
	realm_id = data.keycloak_realm.realm.id
	user_id  = keycloak_user.user.id
	%s
}
`, testAccRealm.Realm, username, arguments)
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      config(""),
				ExpectError: regexp.MustCompile("one of `secret,secret_wo` must be specified"),
			},
			{
				Config:      config(`secret = "JBSWY3DPEHPK3PXP"` + "\n" + `secret_wo = "JBSWY3DPEHPK3PXP"` + "\n" + `secret_wo_version = "v1"`),
				ExpectError: regexp.MustCompile("conflicts with"),
			},
			{
				Config:      config(`secret_wo = "JBSWY3DPEHPK3PXP"`),
				ExpectError: regexp.MustCompile("secret_wo_version"),
			},
			{
				Config:      config(`secret_wo = "not base32!"` + "\n" + `secret_wo_version = "v1"`),
				ExpectError: regexp.MustCompile("must be a non-empty, valid Base32 string"),
			},
		},
	})
}

func testKeycloakUserOtpCredential_writeOnly(username, clientId, secret, version string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	realm_id                     = data.keycloak_realm.realm.id
	client_id                    = "%s"
	access_type                  = "PUBLIC"
	direct_access_grants_enabled = true
}

resource "keycloak_user" "user" {
	realm_id   = data.keycloak_realm.realm.id
	username   = "%s"
	email      = "%s@example.com"
	first_name = "Test"
	last_name  = "User"

	initial_password {
		value = "password"
	}
}

resource "keycloak_user_otp_credential" "otp" {
	realm_id          = data.keycloak_realm.realm.id
	user_id           = keycloak_user.user.id
	secret_wo         = "%s"
	secret_wo_version = "%s"
}
`, testAccRealm.Realm, clientId, username, username, secret, version)
}

// Renaming the realm must not replace the credential: it is addressed through realm_internal_id, and the realm
// in its id follows the rename.
func TestAccKeycloakUserOtpCredential_renameRealm(t *testing.T) {
	t.Parallel()

	realmName := acctest.RandomWithPrefix("tf-acc")
	renamedRealmName := acctest.RandomWithPrefix("tf-acc-renamed")
	username := acctest.RandomWithPrefix("tf-acc")

	var credentialId string

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUserOtpCredential_realm(realmName, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserHasOtpCredential("keycloak_user_otp_credential.otp"),
					resource.TestCheckResourceAttrPair("keycloak_user_otp_credential.otp", "realm_internal_id", "keycloak_realm.realm", "internal_id"),
					func(s *terraform.State) error {
						credentialId = userOtpCredentialIdFromResourceId(s.RootModule().Resources["keycloak_user_otp_credential.otp"].Primary.ID)

						return nil
					},
				),
			},
			{
				Config: testKeycloakUserOtpCredential_realm(renamedRealmName, username),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("keycloak_user_otp_credential.otp", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccRefreshClientToken,
					func(s *terraform.State) error {
						// realm_id in the state only follows the rename on the next refresh, so look the credential up by the new name
						rs := s.RootModule().Resources["keycloak_user_otp_credential.otp"]
						if id := userOtpCredentialIdFromResourceId(rs.Primary.ID); id != credentialId {
							return fmt.Errorf("OTP credential was replaced: %s -> %s", credentialId, id)
						}

						_, err := keycloakClient.GetUserOtpCredential(testCtx, renamedRealmName, rs.Primary.Attributes["user_id"], credentialId)

						return err
					},
				),
			},
			{
				// the realm in the id and realm_id follow the rename on the next read
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("keycloak_user_otp_credential.otp", "realm_id", renamedRealmName),
					func(s *terraform.State) error {
						id := s.RootModule().Resources["keycloak_user_otp_credential.otp"].Primary.ID
						if !strings.HasPrefix(id, renamedRealmName+"/") || userOtpCredentialIdFromResourceId(id) != credentialId {
							return fmt.Errorf("expected the id %s to follow the rename to %s", id, renamedRealmName)
						}

						return nil
					},
				),
			},
		},
	})
}

func testKeycloakUserOtpCredential_realm(realm, username string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_user" "user" {
	realm_internal_id = keycloak_realm.realm.internal_id
	username          = "%s"
}

resource "keycloak_user_otp_credential" "otp" {
	realm_internal_id = keycloak_realm.realm.internal_id
	user_id           = keycloak_user.user.id
	secret            = "JBSWY3DPEHPK3PXP"
}
`, realm, username)
}
