package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestAccKeycloakUser_basic(t *testing.T) {
	t.Parallel()
	username := acctest.RandomWithPrefix("tf-acc")
	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_user.user"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUser_basic(username, attributeName, attributeValue),
				Check:  testAccCheckKeycloakUserExists(resourceName),
			},
			{
				ResourceName:        resourceName,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealm.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakUser_withInitialPassword(t *testing.T) {
	t.Parallel()
	username := acctest.RandomWithPrefix("tf-acc")
	password := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_user.user"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUser_initialPassword(username, password, clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserExists(resourceName),
					testAccCheckKeycloakUserInitialPasswordLogin(username, password, clientId),
				),
			},
		},
	})
}

func TestAccKeycloakUser_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var user = &keycloak.User{}

	username := acctest.RandomWithPrefix("tf-acc")
	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")
	resourceName := "keycloak_user.user"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUser_basic(username, attributeName, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserExists(resourceName),
					testAccCheckKeycloakUserFetch(resourceName, user),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteUser(testCtx, user.RealmId, user.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakUser_basic(username, attributeName, attributeValue),
				Check:  testAccCheckKeycloakUserExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakUser_updateUsername(t *testing.T) {
	t.Parallel()
	usernameOne := acctest.RandomWithPrefix("tf-acc")
	usernameTwo := acctest.RandomWithPrefix("tf-acc")
	attributeName := acctest.RandomWithPrefix("tf-acc")
	attributeValue := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_user.user"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUser_basic(usernameOne, attributeName, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "username", usernameOne),
				),
			},
			{
				Config: testKeycloakUser_basic(usernameTwo, attributeName, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "username", usernameTwo),
				),
			},
		},
	})
}

func TestAccKeycloakUser_updateWithInitialPasswordChangeDoesNotReset(t *testing.T) {
	t.Parallel()
	username := acctest.RandomWithPrefix("tf-acc")
	passwordOne := acctest.RandomWithPrefix("tf-acc")
	passwordTwo := acctest.RandomWithPrefix("tf-acc")
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUser_initialPassword(username, passwordOne, clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserInitialPasswordLogin(username, passwordOne, clientId),
				),
			},
			{
				Config: testKeycloakUser_initialPassword(username, passwordTwo, clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserInitialPasswordLogin(username, passwordOne, clientId),
				),
			},
		},
	})
}

func TestAccKeycloakUser_updateInPlace(t *testing.T) {
	t.Parallel()
	userOne := &keycloak.User{
		RealmId:       "terraform-" + acctest.RandString(10),
		Username:      "terraform-user-" + acctest.RandString(10),
		Email:         fmt.Sprintf("%s@gmail.com", acctest.RandString(10)),
		FirstName:     acctest.RandString(10),
		LastName:      acctest.RandString(10),
		Enabled:       randomBool(),
		EmailVerified: randomBool(),
	}

	userTwo := &keycloak.User{
		RealmId:       userOne.RealmId,
		Username:      userOne.Username,
		Email:         fmt.Sprintf("%s@gmail.com", acctest.RandString(10)),
		FirstName:     acctest.RandString(10),
		LastName:      acctest.RandString(10),
		Enabled:       randomBool(),
		EmailVerified: !userOne.EmailVerified,
	}

	resourceName := "keycloak_user.user"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUser_fromInterface(userOne),
				Check:  testAccCheckKeycloakUserExists(resourceName),
			},
			{
				Config: testKeycloakUser_fromInterface(userTwo),
				Check:  testAccCheckKeycloakUserExists(resourceName),
			},
		},
	})
}

func TestAccKeycloakUser_unsetOptionalAttributes(t *testing.T) {
	t.Parallel()
	attributeName := acctest.RandomWithPrefix("tf-acc")
	userWithOptionalAttributes := &keycloak.User{
		RealmId:   "terraform-" + acctest.RandString(10),
		Username:  "terraform-user-" + acctest.RandString(10),
		Email:     fmt.Sprintf("%s@gmail.com", acctest.RandString(10)),
		FirstName: acctest.RandString(10),
		LastName:  acctest.RandString(10),
		Enabled:   randomBool(),
		Attributes: map[string][]string{
			attributeName: {
				acctest.RandString(230),
				acctest.RandString(12),
			},
		},
	}

	resourceName := "keycloak_user.user"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUser_fromInterface(userWithOptionalAttributes),
				Check:  testAccCheckKeycloakUserExists(resourceName),
			},
			{
				Config: testKeycloakUser_basic(userWithOptionalAttributes.Username, attributeName, strings.Join(userWithOptionalAttributes.Attributes[attributeName], "")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "email", ""),
					resource.TestCheckResourceAttr(resourceName, "first_name", ""),
					resource.TestCheckResourceAttr(resourceName, "last_name", ""),
				),
			},
		},
	})
}

func TestAccKeycloakUser_validateLowercaseUsernames(t *testing.T) {
	t.Parallel()
	username := "terraform-user-" + strings.ToUpper(acctest.RandString(10))
	attributeName := "terraform-attribute-" + acctest.RandString(10)
	attributeValue := acctest.RandString(250)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakUser_basic(username, attributeName, attributeValue),
				ExpectError: regexp.MustCompile("expected username .+ to be all lowercase"),
			},
		},
	})
}

func TestAccKeycloakUser_federatedLink(t *testing.T) {
	sourceUserName := acctest.RandomWithPrefix("tf-acc")
	sourceUserName2 := acctest.RandomWithPrefix("tf-acc")
	destinationRealmName := acctest.RandomWithPrefix("tf-acc")

	resourceName := "keycloak_user.destination_user"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakUserDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakUser_FederationLink(sourceUserName, destinationRealmName),
				Check:  testAccCheckKeycloakUserHasFederationLinkWithSourceUserName(resourceName, sourceUserName),
			},
			{
				Config: testKeycloakUser_FederationLink(sourceUserName2, destinationRealmName),
				Check:  testAccCheckKeycloakUserHasFederationLinkWithSourceUserName(resourceName, sourceUserName2),
			},
		},
	})
}

func testAccCheckKeycloakUserHasFederationLinkWithSourceUserName(resourceName, sourceUserName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedUser, err := getUserFromState(s, resourceName)
		if err != nil {
			return err
		}

		var found = false
		for _, federatedIdentity := range fetchedUser.FederatedIdentities {
			if federatedIdentity.UserName == sourceUserName {
				found = true
			}
			if !found {
				return fmt.Errorf("user had unexpected federatedLink %s or unexpected username %s", federatedIdentity.IdentityProvider, federatedIdentity.UserName)
			}
		}

		if !found {
			return fmt.Errorf("user had no federatedLink, but one was expected")
		}

		return nil
	}
}

func testAccCheckKeycloakUserExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getUserFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakUserFetch(resourceName string, user *keycloak.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedUser, err := getUserFromState(s, resourceName)
		if err != nil {
			return err
		}

		user.Id = fetchedUser.Id
		user.RealmId = fetchedUser.RealmId

		return nil
	}
}

func testAccCheckKeycloakUserInitialPasswordLogin(username, password, clientId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		httpClient := &http.Client{}

		resourceUrl := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", os.Getenv("KEYCLOAK_URL"), testAccRealm.Realm)

		form := url.Values{}
		form.Add("username", username)
		form.Add("password", password)
		form.Add("client_id", clientId)
		form.Add("grant_type", "password")

		request, err := http.NewRequest(http.MethodPost, resourceUrl, strings.NewReader(form.Encode()))
		if err != nil {
			return err
		}
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		response, err := httpClient.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(response.Body)
			return fmt.Errorf("user with username %s cannot login with password %s\n body: %s", username, password, string(body))
		}

		return nil
	}
}

func testAccCheckKeycloakUserDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_user" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			user, _ := keycloakClient.GetUser(testCtx, realm, id)
			if user != nil {
				return fmt.Errorf("user with id %s still exists", id)
			}
		}

		return nil
	}
}

func getUserFromState(s *terraform.State, resourceName string) (*keycloak.User, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	user, err := keycloakClient.GetUser(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user with id %s: %s", id, err)
	}

	return user, nil
}

func testKeycloakUser_basic(username, attributeName, attributeValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_user" "user" {
	realm_id = data.keycloak_realm.realm.id
	username = "%s"
	attributes = {
		"%s" = "%s"
	}
}
	`, testAccRealm.Realm, username, attributeName, attributeValue)
}

func testKeycloakUser_initialPassword(username string, password string, clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	realm_id                     = data.keycloak_realm.realm.id
	client_id                    = "%s"

	name                         = "test client"
	enabled                      = true

	access_type                  = "PUBLIC"
	direct_access_grants_enabled = true
}

resource "keycloak_user" "user" {
	realm_id         = data.keycloak_realm.realm.id
	username         = "%s"
	initial_password {
		value = "%s"
		temporary = false
	}
}
	`, testAccRealm.Realm, clientId, username, password)
}

func testKeycloakUser_fromInterface(user *keycloak.User) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_user" "user" {
	realm_id       = data.keycloak_realm.realm.id
	username       = "%s"

	email          = "%s"
	first_name     = "%s"
	last_name      = "%s"
	enabled        = %t
	email_verified = "%t"
}
	`, testAccRealm.Realm, user.Username, user.Email, user.FirstName, user.LastName, user.Enabled, user.EmailVerified)
}

func testKeycloakUser_FederationLink(sourceRealmUserName, destinationRealmId string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "source_realm" {
  realm   = "source_test_realm"
  enabled = true
}

resource "keycloak_openid_client" "destination_client" {
  realm_id                 = "${keycloak_realm.source_realm.id}"
  client_id                = "destination_client"
  client_secret            = "secret"
  access_type              = "CONFIDENTIAL"
  standard_flow_enabled    = true
  valid_redirect_uris = [
    "http://localhost:8080/*",
  ]
}

resource "keycloak_user" "source_user" {
  realm_id   = "${keycloak_realm.source_realm.id}"
  username   = "%s"
  initial_password {
    value     = "source"
    temporary = false
  }
}

resource "keycloak_realm" "destination_realm" {
  realm   = "%s"
  enabled = true
}

resource keycloak_oidc_identity_provider source_oidc_idp {
  realm              = "${keycloak_realm.destination_realm.id}"
  alias              = "source"
  authorization_url  = "http://localhost:8080/auth/realms/${keycloak_realm.source_realm.id}/protocol/openid-connect/auth"
  token_url          = "http://localhost:8080/auth/realms/${keycloak_realm.source_realm.id}/protocol/openid-connect/token"
  client_id          = "${keycloak_openid_client.destination_client.client_id}"
  client_secret      = "${keycloak_openid_client.destination_client.client_secret}"
  default_scopes     = "openid"
}

resource "keycloak_user" "destination_user" {
  realm_id   = "${keycloak_realm.destination_realm.id}"
  username   = "my_destination_username"
  federated_identity {
    identity_provider = "${keycloak_oidc_identity_provider.source_oidc_idp.alias}"
    user_id           = "${keycloak_user.source_user.id}"
    user_name         = "${keycloak_user.source_user.username}"
  }
}
	`, sourceRealmUserName, destinationRealmId)
}
