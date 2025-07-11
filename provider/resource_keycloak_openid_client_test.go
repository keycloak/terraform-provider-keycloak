package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/keycloak/terraform-provider-keycloak/keycloak/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakOpenidClient_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_basic(clientId),
				Check:  testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
			},
			{
				ResourceName:            "keycloak_openid_client.client",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     testAccRealm.Realm + "/",
				ImportStateVerifyIgnore: []string{"exclude_session_state_from_auth_response", "exclude_issuer_from_auth_response"},
			},
		},
	})
}

func TestAccKeycloakOpenidClient_basic_with_consent(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_basic_with_consent(clientId),
				Check:  testAccCheckKeycloakOpenidClientExistsWithCorrectConsentSettings("keycloak_openid_client.client"),
			},
			{
				ResourceName:            "keycloak_openid_client.client",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     testAccRealm.Realm + "/",
				ImportStateVerifyIgnore: []string{"exclude_session_state_from_auth_response", "exclude_issuer_from_auth_response"},
			},
		},
	})
}

func TestAccKeycloakOpenidClient_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var client = &keycloak.OpenidClient{}

	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_basic(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
					testAccCheckKeycloakOpenidClientFetch("keycloak_openid_client.client", client, false),
				),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteOpenidClient(testCtx, client.RealmId, client.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOpenidClient_basic(clientId),
				Check:  testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_updateRealm(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_updateRealmBefore(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
					testAccCheckKeycloakOpenidClientBelongsToRealm("keycloak_openid_client.client", testAccRealm.Realm),
				),
			},
			{
				Config: testKeycloakOpenidClient_updateRealmAfter(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
					testAccCheckKeycloakOpenidClientBelongsToRealm("keycloak_openid_client.client", testAccRealmTwo.Realm),
				),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_accessType(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_accessType(clientId, "CONFIDENTIAL"),
				Check:  testAccCheckKeycloakOpenidClientAccessType("keycloak_openid_client.client", false, false),
			},
			{
				Config: testKeycloakOpenidClient_accessType(clientId, "PUBLIC"),
				Check:  testAccCheckKeycloakOpenidClientAccessType("keycloak_openid_client.client", true, false),
			},
			{
				Config: testKeycloakOpenidClient_accessType(clientId, "BEARER-ONLY"),
				Check:  testAccCheckKeycloakOpenidClientAccessType("keycloak_openid_client.client", false, true),
			},
		},
	})
}
func TestAccKeycloakOpenidClient_clientAuthenticatorType(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_clientAuthenticatorType(clientId, "client-secret"),
				Check:  testAccCheckKeycloakOpenidClientAuthenticatorType("keycloak_openid_client.client", "client-secret"),
			},
			{
				Config: testKeycloakOpenidClient_clientAuthenticatorType(clientId, "client-jwt"),
				Check:  testAccCheckKeycloakOpenidClientAuthenticatorType("keycloak_openid_client.client", "client-jwt"),
			},
			{
				Config: testKeycloakOpenidClient_clientAuthenticatorType(clientId, "client-secret-jwt"),
				Check:  testAccCheckKeycloakOpenidClientAuthenticatorType("keycloak_openid_client.client", "client-secret-jwt"),
			},
			{
				Config: testKeycloakOpenidClient_clientAuthenticatorType(clientId, "client-x509"),
				Check:  testAccCheckKeycloakOpenidClientAuthenticatorType("keycloak_openid_client.client", "client-x509"),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_updateInPlace(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	enabled := randomBool()
	standardFlowEnabled := randomBool()
	implicitFlowEnabled := randomBool()
	directAccessGrantsEnabled := randomBool()
	serviceAccountsEnabled := randomBool()

	if !standardFlowEnabled {
		implicitFlowEnabled = !standardFlowEnabled
	}

	rootUrlBefore := "http://localhost:2222/" + acctest.RandString(20)
	openidClientBefore := &keycloak.OpenidClient{
		ClientId:                  clientId,
		Name:                      acctest.RandString(10),
		Enabled:                   enabled,
		Description:               acctest.RandString(50),
		ClientSecret:              acctest.RandString(10),
		StandardFlowEnabled:       standardFlowEnabled,
		ImplicitFlowEnabled:       implicitFlowEnabled,
		DirectAccessGrantsEnabled: directAccessGrantsEnabled,
		ServiceAccountsEnabled:    serviceAccountsEnabled,
		ValidRedirectUris:         []string{acctest.RandString(10), acctest.RandString(10), acctest.RandString(10), acctest.RandString(10)},
		WebOrigins:                []string{acctest.RandString(10), acctest.RandString(10), acctest.RandString(10)},
		AdminUrl:                  acctest.RandString(20),
		BaseUrl:                   "http://localhost:2222/" + acctest.RandString(20),
		RootUrl:                   &rootUrlBefore,
		Attributes: keycloak.OpenidClientAttributes{
			BackchannelLogoutUrl:                 "http://localhost:3333/backchannel",
			BackchannelLogoutSessionRequired:     types.KeycloakBoolQuoted(randomBool()),
			BackchannelLogoutRevokeOfflineTokens: types.KeycloakBoolQuoted(randomBool()),
			PostLogoutRedirectUris:               []string{acctest.RandString(10), acctest.RandString(10), acctest.RandString(10), acctest.RandString(10)},
		},
	}

	standardFlowEnabled, implicitFlowEnabled = implicitFlowEnabled, standardFlowEnabled

	rootUrlAfter := "http://localhost:2222/" + acctest.RandString(20)
	openidClientAfter := &keycloak.OpenidClient{
		ClientId:                  clientId,
		Name:                      acctest.RandString(10),
		Enabled:                   !enabled,
		Description:               acctest.RandString(50),
		ClientSecret:              acctest.RandString(10),
		StandardFlowEnabled:       standardFlowEnabled,
		ImplicitFlowEnabled:       implicitFlowEnabled,
		DirectAccessGrantsEnabled: !directAccessGrantsEnabled,
		ServiceAccountsEnabled:    !serviceAccountsEnabled,
		ValidRedirectUris:         []string{acctest.RandString(10), acctest.RandString(10)},
		WebOrigins:                []string{acctest.RandString(10), acctest.RandString(10), acctest.RandString(10), acctest.RandString(10), acctest.RandString(10)},
		AdminUrl:                  acctest.RandString(20),
		BaseUrl:                   "http://localhost:2222/" + acctest.RandString(20),
		RootUrl:                   &rootUrlAfter,
		Attributes: keycloak.OpenidClientAttributes{
			BackchannelLogoutUrl:                 "http://localhost:3333/backchannel",
			BackchannelLogoutSessionRequired:     types.KeycloakBoolQuoted(randomBool()),
			BackchannelLogoutRevokeOfflineTokens: types.KeycloakBoolQuoted(randomBool()),
			PostLogoutRedirectUris:               []string{acctest.RandString(10), acctest.RandString(10)},
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_fromInterface(openidClientBefore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
				),
			},
			{
				Config: testKeycloakOpenidClient_fromInterface(openidClientAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
				),
			},
			{
				Config: testKeycloakOpenidClient_basic(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
				),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_backChannel(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	backchannelLogoutUrl := fmt.Sprintf("https://%s.com", acctest.RandString(10))
	backchannelLogoutSessionRequired := randomBool()
	backchannelLogoutRevokeOfflineSessions := !backchannelLogoutSessionRequired

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_backchannel(clientId, backchannelLogoutUrl, backchannelLogoutSessionRequired, backchannelLogoutRevokeOfflineSessions),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
					testAccCheckKeycloakOpenidClientHasBackchannelSettings("keycloak_openid_client.client", backchannelLogoutUrl, backchannelLogoutSessionRequired, backchannelLogoutRevokeOfflineSessions),
				),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_frontChannel(t *testing.T) {
	t.Parallel()

	clientId := acctest.RandomWithPrefix("tf-acc")
	frontchannelLogoutUrl := fmt.Sprintf("https://%s.com/logout", acctest.RandString(10))
	frontchannelLogoutEnabled := true

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_frontchannel(clientId, frontchannelLogoutUrl, frontchannelLogoutEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
					testAccCheckKeycloakOpenidClientHasFrontchannelSettings("keycloak_openid_client.client", frontchannelLogoutUrl, frontchannelLogoutEnabled),
				),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_AccessToken_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	accessTokenLifespan := "1800"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_AccessToken_basic(clientId, accessTokenLifespan),
				Check:  testAccCheckKeycloakOpenidClientExistsWithCorrectLifespan("keycloak_openid_client.client", accessTokenLifespan),
			},
			{
				ResourceName:            "keycloak_openid_client.client",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     testAccRealm.Realm + "/",
				ImportStateVerifyIgnore: []string{"exclude_session_state_from_auth_response", "exclude_issuer_from_auth_response"},
			},
		},
	})
}

func TestAccKeycloakOpenidClient_ClientTimeouts_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	offlineSessionIdleTimeout := "1800"
	offlineSessionMaxLifespan := "1900"
	sessionIdleTimeout := "2000"
	sessionMaxLifespan := "2100"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_ClientTimeouts(clientId,
					offlineSessionIdleTimeout, offlineSessionMaxLifespan, sessionIdleTimeout, sessionMaxLifespan),
				Check: testAccCheckKeycloakOpenidClientExistsWithCorrectClientTimeouts("keycloak_openid_client.client",
					offlineSessionIdleTimeout, offlineSessionMaxLifespan, sessionIdleTimeout, sessionMaxLifespan,
				),
			},
			{
				ResourceName:            "keycloak_openid_client.client",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     testAccRealm.Realm + "/",
				ImportStateVerifyIgnore: []string{"exclude_session_state_from_auth_response", "exclude_issuer_from_auth_response"},
			},
		},
	})
}

func TestAccKeycloakOpenidClient_Device_basic(t *testing.T) {
	if ok, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_13); !ok {
		t.Skip()
	}

	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	oauth2DeviceCodeLifespan := "300"
	oauth2DevicePollingInterval := "60"
	oauth2DeviceAuthorizationGrantEnabled := true

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_oauth2DeviceTimes(clientId,
					oauth2DeviceCodeLifespan, oauth2DevicePollingInterval, oauth2DeviceAuthorizationGrantEnabled,
				),
				Check: testAccCheckKeycloakOpenidClientOauth2Device("keycloak_openid_client.client",
					oauth2DeviceCodeLifespan, oauth2DevicePollingInterval, oauth2DeviceAuthorizationGrantEnabled,
				),
			},
			{
				ResourceName:            "keycloak_openid_client.client",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     testAccRealm.Realm + "/",
				ImportStateVerifyIgnore: []string{"exclude_session_state_from_auth_response", "exclude_issuer_from_auth_response"},
			},
		},
	})
}

func TestAccKeycloakOpenidClient_secret(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	clientSecret := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_basic(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
					testAccCheckKeycloakOpenidClientHasNonEmptyClientSecret("keycloak_openid_client.client"),
				),
			},
			{
				Config: testKeycloakOpenidClient_secret(clientId, clientSecret),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
					testAccCheckKeycloakOpenidClientHasClientSecret("keycloak_openid_client.client", clientSecret),
				),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_secretWriteOnly(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	clientSecretWO := acctest.RandomWithPrefix("tf-acc")
	clientSecretWOUpdated := acctest.RandomWithPrefix("tf-acc")
	clientSecretExplicit := acctest.RandomWithPrefix("tf-acc")
	clientSecretWOVersion := 1

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				// test CREATION of the client_secret via write-only attribute
				Config: testKeycloakOpenidClient_secretWriteOnly(clientId, clientSecretWO, clientSecretWOVersion),
				Check: resource.ComposeTestCheckFunc(
					// assert openid client against the Keycloak's API response (value SHOULD be the new one)
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
					testAccCheckKeycloakOpenidClientHasClientSecret("keycloak_openid_client.client", clientSecretWO),

					// assert openid client against the Terraform state (client_secret value SHOULD NOT be stored in state)
					resource.TestCheckNoResourceAttr("keycloak_openid_client.client", "client_secret"),
					resource.TestCheckResourceAttr("keycloak_openid_client.client", "client_secret_wo_version", strconv.Itoa(clientSecretWOVersion)),
				),
			},
			{
				// test NO UPDATE of client_secret when the client_secret_wo_version is NOT MODIFIED
				Config: testKeycloakOpenidClient_secretWriteOnly(clientId, clientSecretWOUpdated, clientSecretWOVersion),
				Check: resource.ComposeTestCheckFunc(
					// assert openid client against the Keycloak's API response (value should be the same as before)
					testAccCheckKeycloakOpenidClientHasClientSecret("keycloak_openid_client.client", clientSecretWO),

					// assert openid client against the Terraform state (client_secret value SHOULD NOT be stored in state)
					resource.TestCheckNoResourceAttr("keycloak_openid_client.client", "client_secret"),
				),
			},
			{
				// test UPDATE of client_secret when the client_secret_wo_version is MODIFIED
				Config: testKeycloakOpenidClient_secretWriteOnly(clientId, clientSecretWOUpdated, clientSecretWOVersion+1),
				Check: resource.ComposeTestCheckFunc(
					// assert openid client against the Keycloak's API response (value SHOULD be the new one)
					testAccCheckKeycloakOpenidClientHasClientSecret("keycloak_openid_client.client", clientSecretWOUpdated),

					// assert openid client against the Terraform state (client_secret value SHOULD NOT be stored in state)
					resource.TestCheckNoResourceAttr("keycloak_openid_client.client", "client_secret"),
					resource.TestCheckResourceAttr("keycloak_openid_client.client", "client_secret_wo_version", strconv.Itoa(clientSecretWOVersion+1)),
				),
			},
			{
				// test UPDATE of client_secret when write-only attrs are exchanged for explicit client_secret
				Config: testKeycloakOpenidClient_secret(clientId, clientSecretExplicit),
				Check: resource.ComposeTestCheckFunc(
					// assert openid client against the Keycloak's API response (value SHOULD be the new one)
					testAccCheckKeycloakOpenidClientHasClientSecret("keycloak_openid_client.client", clientSecretExplicit),

					// assert openid client against the Terraform state (client_secret value SHOULD be stored in state)
					resource.TestCheckResourceAttr("keycloak_openid_client.client", "client_secret", clientSecretExplicit),
					resource.TestCheckResourceAttr("keycloak_openid_client.client", "client_secret_wo_version", strconv.Itoa(0)),
				),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_redirectUrisValidation(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	accessType := randomStringInSlice([]string{"PUBLIC", "CONFIDENTIAL"})

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenidClient_invalidRedirectUris(clientId, accessType, true, false),
				ExpectError: regexp.MustCompile("validation error: standard \\(authorization code\\) and implicit flows require at least one valid redirect uri"),
			},
			{
				Config:      testKeycloakOpenidClient_invalidRedirectUris(clientId, accessType, false, true),
				ExpectError: regexp.MustCompile("validation error: standard \\(authorization code\\) and implicit flows require at least one valid redirect uri"),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_publicClientCredentialsValidation(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenidClient_invalidPublicClientWithClientCredentials(clientId),
				ExpectError: regexp.MustCompile("validation error: service accounts \\(client credentials flow\\) cannot be enabled on public clients"),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_bearerClientNoGrantsValidation(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenidClient_bearerOnlyClientsCannotIssueTokens(clientId, true, false, false, false),
				ExpectError: regexp.MustCompile("validation error: Keycloak cannot issue tokens for bearer-only clients; no oauth2 flows can be enabled for this client"),
			},
			{
				Config:      testKeycloakOpenidClient_bearerOnlyClientsCannotIssueTokens(clientId, false, true, false, false),
				ExpectError: regexp.MustCompile("validation error: Keycloak cannot issue tokens for bearer-only clients; no oauth2 flows can be enabled for this client"),
			},
			{
				Config:      testKeycloakOpenidClient_bearerOnlyClientsCannotIssueTokens(clientId, false, false, true, false),
				ExpectError: regexp.MustCompile("validation error: Keycloak cannot issue tokens for bearer-only clients; no oauth2 flows can be enabled for this client"),
			},
			{
				Config:      testKeycloakOpenidClient_bearerOnlyClientsCannotIssueTokens(clientId, false, false, false, true),
				ExpectError: regexp.MustCompile("validation error: Keycloak cannot issue tokens for bearer-only clients; no oauth2 flows can be enabled for this client"),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_pkceCodeChallengeMethod(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenidClient_pkceChallengeMethod(clientId, "invalidMethod"),
				ExpectError: regexp.MustCompile(`expected pkce_code_challenge_method to be one of \[\"" "plain" "S256"\], got invalidMethod`),
			},
			{
				Config: testKeycloakOpenidClient_omitPkceChallengeMethod(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod("keycloak_openid_client.client", ""),
					testAccCheckKeycloakOpenidClientHasExcludeSessionStateFromAuthResponse("keycloak_openid_client.client", false),
					testAccCheckKeycloakOpenidClientHasExcludeIssuerFromAuthResponse("keycloak_openid_client.client", false),
				),
			},
			{
				Config: testKeycloakOpenidClient_pkceChallengeMethod(clientId, "plain"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod("keycloak_openid_client.client", "plain"),
					testAccCheckKeycloakOpenidClientHasExcludeSessionStateFromAuthResponse("keycloak_openid_client.client", false),
					testAccCheckKeycloakOpenidClientHasExcludeIssuerFromAuthResponse("keycloak_openid_client.client", false),
				),
			},
			{
				Config: testKeycloakOpenidClient_pkceChallengeMethod(clientId, "S256"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod("keycloak_openid_client.client", "S256"),
					testAccCheckKeycloakOpenidClientHasExcludeSessionStateFromAuthResponse("keycloak_openid_client.client", false),
					testAccCheckKeycloakOpenidClientHasExcludeIssuerFromAuthResponse("keycloak_openid_client.client", false),
				),
			},
			{
				Config: testKeycloakOpenidClient_pkceChallengeMethod(clientId, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod("keycloak_openid_client.client", ""),
					testAccCheckKeycloakOpenidClientHasExcludeSessionStateFromAuthResponse("keycloak_openid_client.client", false),
					testAccCheckKeycloakOpenidClientHasExcludeIssuerFromAuthResponse("keycloak_openid_client.client", false),
				),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_excludeSessionStateFromAuthResponse(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_omitExcludeSessionStateFromAuthResponse(clientId, "plain"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasExcludeSessionStateFromAuthResponse("keycloak_openid_client.client", false),
					testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod("keycloak_openid_client.client", "plain"),
				),
			},
			{
				Config: testKeycloakOpenidClient_excludeSessionStateFromAuthResponse(clientId, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasExcludeSessionStateFromAuthResponse("keycloak_openid_client.client", false),
					testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod("keycloak_openid_client.client", ""),
				),
			},
			{
				Config: testKeycloakOpenidClient_excludeSessionStateFromAuthResponse(clientId, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasExcludeSessionStateFromAuthResponse("keycloak_openid_client.client", true),
					testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod("keycloak_openid_client.client", ""),
				),
			},
			{
				Config: testKeycloakOpenidClient_excludeSessionStateFromAuthResponse(clientId, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasExcludeSessionStateFromAuthResponse("keycloak_openid_client.client", false),
					testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod("keycloak_openid_client.client", ""),
				),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_excludeIssuerFromAuthResponse(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_omitExcludeIssuerFromAuthResponse(clientId, "plain"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasExcludeIssuerFromAuthResponse("keycloak_openid_client.client", false),
					testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod("keycloak_openid_client.client", "plain"),
				),
			},
			{
				Config: testKeycloakOpenidClient_excludeIssuerFromAuthResponse(clientId, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasExcludeIssuerFromAuthResponse("keycloak_openid_client.client", false),
					testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod("keycloak_openid_client.client", ""),
				),
			},
			{
				Config: testKeycloakOpenidClient_excludeIssuerFromAuthResponse(clientId, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasExcludeIssuerFromAuthResponse("keycloak_openid_client.client", true),
					testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod("keycloak_openid_client.client", ""),
				),
			},
			{
				Config: testKeycloakOpenidClient_excludeIssuerFromAuthResponse(clientId, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientHasExcludeIssuerFromAuthResponse("keycloak_openid_client.client", false),
					testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod("keycloak_openid_client.client", ""),
				),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_authenticationFlowBindingOverrides(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_authenticationFlowBindingOverrides(clientId),
				Check:  testAccCheckKeycloakOpenidClientAuthenticationFlowBindingOverrides("keycloak_openid_client.client", "keycloak_authentication_flow.another_flow"),
			},
			{
				Config: testKeycloakOpenidClient_withoutAuthenticationFlowBindingOverrides(clientId),
				Check:  testAccCheckKeycloakOpenidClientAuthenticationFlowBindingOverrides("keycloak_openid_client.client", ""),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_loginTheme(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")
	loginThemeKeycloak := "keycloak"
	loginThemeBase := "base"
	loginThemeRandom := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_loginTheme(clientId, loginThemeKeycloak),
				Check:  testAccCheckKeycloakOpenidClientLoginTheme("keycloak_openid_client.client", loginThemeKeycloak),
			},
			{
				Config: testKeycloakOpenidClient_loginTheme(clientId, loginThemeBase),
				Check:  testAccCheckKeycloakOpenidClientLoginTheme("keycloak_openid_client.client", loginThemeBase),
			},
			{
				Config:      testKeycloakOpenidClient_loginTheme(clientId, loginThemeRandom),
				ExpectError: regexp.MustCompile("validation error: theme \".+\" does not exist on the server"),
			},
			{
				Config: testKeycloakOpenidClient_loginTheme(clientId, loginThemeKeycloak),
				Check:  testAccCheckKeycloakOpenidClientLoginTheme("keycloak_openid_client.client", loginThemeKeycloak),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_import(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientNotDestroyed(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenidClient_import("non-existing-client", true),
				ExpectError: regexp.MustCompile("Error: openid client with name non-existing-client does not exist"),
			},
			{
				Config: testKeycloakOpenidClient_import("account", true),
				Check:  testAccCheckKeycloakOpenidClientExistsWithEnabledStatus("keycloak_openid_client.client", true),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_useRefreshTokens(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_useRefreshTokens(clientId, true),
				Check:  testAccCheckKeycloakOpenidClientUseRefreshTokens("keycloak_openid_client.client", true),
			},
			{
				Config: testKeycloakOpenidClient_useRefreshTokens(clientId, false),
				Check:  testAccCheckKeycloakOpenidClientUseRefreshTokens("keycloak_openid_client.client", false),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_enableStandardTokenExchange(t *testing.T) {
	if ok, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_26_2); !ok {
		t.Skip()
	}
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_enableStandardTokenExchange(clientId, true),
				Check:  testAccCheckKeycloakOpenidClientEnableStandardTokenExchange("keycloak_openid_client.client", true),
			},
			{
				Config: testKeycloakOpenidClient_enableStandardTokenExchange(clientId, false),
				Check:  testAccCheckKeycloakOpenidClientEnableStandardTokenExchange("keycloak_openid_client.client", false),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_allowRefreshTokenInStandardExchange(t *testing.T) {
	if ok, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_26_2); !ok {
		t.Skip()
	}
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_allowRefreshTokenInStandardExchange(clientId, "SAME_SESSION"),
				Check:  testAccCheckKeycloakOpenidClientAllowRefreshTokenInStandardExchange("keycloak_openid_client.client", "SAME_SESSION"),
			},
			{
				Config: testKeycloakOpenidClient_allowRefreshTokenInStandardExchange(clientId, ""),
				Check:  testAccCheckKeycloakOpenidClientAllowRefreshTokenInStandardExchange("keycloak_openid_client.client", ""),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_useRefreshTokensClientCredentials(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_useRefreshTokensClientCredentials(clientId, true),
				Check:  testAccCheckKeycloakOpenidClientUseRefreshTokensClientCredentials("keycloak_openid_client.client", true),
			},
			{
				Config: testKeycloakOpenidClient_useRefreshTokensClientCredentials(clientId, false),
				Check:  testAccCheckKeycloakOpenidClientUseRefreshTokensClientCredentials("keycloak_openid_client.client", false),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_extraConfig(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_extraConfig(clientId, map[string]string{
					"key1": "value1",
					"key2": "value2",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExtraConfig("keycloak_openid_client.client", "key1", "value1"),
					testAccCheckKeycloakOpenidClientExtraConfig("keycloak_openid_client.client", "key2", "value2"),
				),
			},
			{
				Config: testKeycloakOpenidClient_extraConfig(clientId, map[string]string{
					"key2": "value2",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExtraConfig("keycloak_openid_client.client", "key2", "value2"),
					testAccCheckKeycloakOpenidClientExtraConfigMissing("keycloak_openid_client.client", "key1"),
				),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_extraConfigInvalid(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOpenidClient_extraConfig(clientId, map[string]string{"login_theme": "keycloak"}),
				ExpectError: regexp.MustCompile(`extra_config key "login_theme" is not allowed`),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_oauth2DeviceAuthorizationGrantEnabled(t *testing.T) {
	if ok, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_13); !ok {
		t.Skip()
	}

	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakOpenidClientDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOpenidClient_oauth2DeviceAuthorizationGrantEnabled(clientId, true),
				Check:  testAccCheckKeycloakOpenidClientOauth2DeviceAuthorizationGrantEnabled("keycloak_openid_client.client", true),
			},
			{
				Config: testKeycloakOpenidClient_oauth2DeviceAuthorizationGrantEnabled(clientId, false),
				Check:  testAccCheckKeycloakOpenidClientOauth2DeviceAuthorizationGrantEnabled("keycloak_openid_client.client", false),
			},
		},
	})
}

func TestAccKeycloakOpenidClient_secretRegenerated(t *testing.T) {
	clientId := acctest.RandomWithPrefix("tf-acc")
	var client = &keycloak.OpenidClient{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckKeycloakRoleDestroy(),
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{ // Create client with client_secret_regenerate_when_changed = null
				Config: testKeycloakOpenidClient_basic(clientId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
					testAccCheckKeycloakOpenidClientFetch("keycloak_openid_client.client", client, true),
				),
			},
			{ // Set client_secret_regenerate_when_changed = "initial-value" and validate that secret wasn't regenerated
				Config: testKeycloakOpenidClient_basicWithSecretRegenerate(clientId, "initial-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
					resource.TestCheckResourceAttr("keycloak_openid_client.client", "client_secret_regenerate_when_changed.rotation", "initial-value"),
					testAccCheckKeycloakOpenidClientExistsWithRegeneratedSecret("keycloak_openid_client.client", client),
					testAccCheckKeycloakOpenidClientFetch("keycloak_openid_client.client", client, true),
				),
			},
			{ // Set client_secret_regenerate_when_changed = "second-value" and validate that secret was regenerated
				Config: testKeycloakOpenidClient_basicWithSecretRegenerate(clientId, "second-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol("keycloak_openid_client.client"),
					resource.TestCheckResourceAttr("keycloak_openid_client.client", "client_secret_regenerate_when_changed.rotation", "second-value"),
					testAccCheckKeycloakOpenidClientExistsWithRegeneratedSecret("keycloak_openid_client.client", client),
					testAccCheckKeycloakOpenidClientFetch("keycloak_openid_client.client", client, true),
				),
			},
		},
	})
}

func testAccCheckKeycloakOpenidClientExistsWithCorrectProtocol(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Protocol != "openid-connect" {
			return fmt.Errorf("expected openid client to have openid-connect protocol, but got %s", client.Protocol)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientExistsWithCorrectConsentSettings(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.ConsentRequired != true {
			return fmt.Errorf("expected openid client to have ConsentRequired %v, but got %v", true, client.ConsentRequired)
		}

		if client.Attributes.DisplayOnConsentScreen != true {
			return fmt.Errorf("expected openid client to have DisplayClientOnConsentScreen %v, but got %v", true, client.Attributes.DisplayOnConsentScreen)
		}

		if client.Attributes.ConsentScreenText != "some consent screen text" {
			return fmt.Errorf("expected openid client to have ConsentScreenText %v, but got %v", "some consent screen text", client.Attributes.ConsentScreenText)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientHasBackchannelSettings(resourceName, backchannelLogoutUrl string, backchannelLogoutSessionRequired, backchannelLogoutRevokeOfflineSessions bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.BackchannelLogoutUrl != backchannelLogoutUrl {
			return fmt.Errorf("expected openid client to have backchannel logout url %s, got %s", backchannelLogoutUrl, client.Attributes.BackchannelLogoutUrl)
		}

		if bool(client.Attributes.BackchannelLogoutSessionRequired) != backchannelLogoutSessionRequired {
			return fmt.Errorf("expected openid client to have backchannel session required bool %t, got %t", backchannelLogoutSessionRequired, bool(client.Attributes.BackchannelLogoutSessionRequired))
		}

		if bool(client.Attributes.BackchannelLogoutRevokeOfflineTokens) != backchannelLogoutRevokeOfflineSessions {
			return fmt.Errorf("expected openid client to have backchannel revoke offline sessions bool %t, got %t", backchannelLogoutRevokeOfflineSessions, bool(client.Attributes.BackchannelLogoutRevokeOfflineTokens))
		}

		return nil
	}
}
func testAccCheckKeycloakOpenidClientHasFrontchannelSettings(resourceName, frontChannelLogoutUrl string, frontChannelLogoutEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.FrontchannelLogoutUrl != frontChannelLogoutUrl {
			return fmt.Errorf("expected openid client to have frontchannel logout url %s, got %s", frontChannelLogoutUrl, client.Attributes.FrontchannelLogoutUrl)
		}

		if client.FrontChannelLogoutEnabled != frontChannelLogoutEnabled {
			return fmt.Errorf("expected openid client to have frontchannel enabled bool %t, got %t", frontChannelLogoutEnabled, client.FrontChannelLogoutEnabled)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientExistsWithCorrectLifespan(resourceName string, accessTokenLifespan string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.AccessTokenLifespan != accessTokenLifespan {
			return fmt.Errorf("expected openid client to have access token lifespan set to %s, but got %s", accessTokenLifespan, client.Attributes.AccessTokenLifespan)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientExistsWithCorrectClientTimeouts(resourceName string,
	offlineSessionIdleTimeout string, offlineSessionMaxLifespan string,
	sessionIdleTimeout string, sessionMaxLifespan string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.ClientOfflineSessionIdleTimeout != offlineSessionIdleTimeout {
			return fmt.Errorf("expected openid client to have client offline session idle timeout set to %s, but got %s", offlineSessionIdleTimeout, client.Attributes.ClientOfflineSessionIdleTimeout)
		}

		if client.Attributes.ClientOfflineSessionMaxLifespan != offlineSessionMaxLifespan {
			return fmt.Errorf("expected openid client to have client offline session max lifespan set to %s, but got %s", offlineSessionMaxLifespan, client.Attributes.ClientOfflineSessionMaxLifespan)
		}

		if client.Attributes.ClientSessionIdleTimeout != sessionIdleTimeout {
			return fmt.Errorf("expected openid client to have client session idle timeout set to %s, but got %s", sessionIdleTimeout, client.Attributes.ClientSessionIdleTimeout)
		}

		if client.Attributes.ClientSessionMaxLifespan != sessionMaxLifespan {
			return fmt.Errorf("expected openid client to have client session max lifespan set to %s, but got %s", sessionMaxLifespan, client.Attributes.ClientSessionMaxLifespan)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientOauth2Device(resourceName string,
	oauth2DeviceCodeLifespan string, Oauth2DevicePollingInterval string, oauth2DeviceAuthorizationGrantEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.Oauth2DeviceAuthorizationGrantEnabled != types.KeycloakBoolQuoted(oauth2DeviceAuthorizationGrantEnabled) {
			return fmt.Errorf("expected openid client to have device authorizationen granted enabled set to %t, but got %v", oauth2DeviceAuthorizationGrantEnabled, client.Attributes.Oauth2DeviceAuthorizationGrantEnabled)
		}

		if client.Attributes.Oauth2DeviceCodeLifespan != oauth2DeviceCodeLifespan {
			return fmt.Errorf("expected openid client to have device code lifespan set to %s, but got %s", oauth2DeviceCodeLifespan, client.Attributes.Oauth2DeviceCodeLifespan)
		}

		if client.Attributes.Oauth2DevicePollingInterval != Oauth2DevicePollingInterval {
			return fmt.Errorf("expected openid client to have device polling interval set to %s, but got %s", Oauth2DevicePollingInterval, client.Attributes.Oauth2DevicePollingInterval)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientFetch(resourceName string, client *keycloak.OpenidClient, withSecret bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedClient, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		client.Id = fetchedClient.Id
		client.RealmId = fetchedClient.RealmId

		if withSecret {
			client.ClientSecret = fetchedClient.ClientSecret
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientAccessType(resourceName string, public, bearer bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.PublicClient != public {
			return fmt.Errorf("expected openid client to have public set to %t, but got %t", public, client.PublicClient)
		}

		if client.BearerOnly != bearer {
			return fmt.Errorf("expected openid client to have bearer set to %t, but got %t", bearer, client.BearerOnly)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientAuthenticatorType(resourceName string, authType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.ClientAuthenticatorType != authType {
			return fmt.Errorf("expected openid client to have client_authenticator_type set to %s, but got %s", authType, client.ClientAuthenticatorType)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientBelongsToRealm(resourceName, realm string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.RealmId != realm {
			return fmt.Errorf("expected openid client %s to have realm_id of %s, but got %s", client.ClientId, realm, client.RealmId)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientHasClientSecret(resourceName, secret string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.ClientSecret != secret {
			return fmt.Errorf("expected openid client %s to have a client secret value of %s, but got %s", client.ClientId, secret, client.ClientSecret)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientHasNonEmptyClientSecret(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.ClientSecret == "" {
			return fmt.Errorf("expected openid client %s to have non empty secret value", client.ClientId)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			client, _ := keycloakClient.GetOpenidClient(testCtx, realm, id)
			if client != nil {
				return fmt.Errorf("openid client %s still exists", id)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientNotDestroyed() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_openid_client" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			client, _ := keycloakClient.GetOpenidClient(testCtx, realm, id)
			if client == nil {
				return fmt.Errorf("openid client %s dost not exists", id)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientHasPkceCodeChallengeMethod(resourceName, pkceCodeChallengeMethod string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.PkceCodeChallengeMethod != pkceCodeChallengeMethod {
			return fmt.Errorf("expected openid client %s to have pkce code challenge method value of %s, but got %s", client.ClientId, pkceCodeChallengeMethod, client.Attributes.PkceCodeChallengeMethod)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientHasExcludeSessionStateFromAuthResponse(resourceName string, excludeSessionStateFromAuthResponse types.KeycloakBoolQuoted) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.ExcludeSessionStateFromAuthResponse != excludeSessionStateFromAuthResponse {
			return fmt.Errorf("expected openid client %s to have exclude_session_state_from_auth_response value of %t, but got %t", client.ClientId, excludeSessionStateFromAuthResponse, client.Attributes.ExcludeSessionStateFromAuthResponse)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientHasExcludeIssuerFromAuthResponse(resourceName string, excludeIssuerFromAuthResponse types.KeycloakBoolQuoted) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.ExcludeIssuerFromAuthResponse != excludeIssuerFromAuthResponse {
			return fmt.Errorf("expected openid client %s to have exclude_issuer_from_auth_response value of %t, but got %t", client.ClientId, excludeIssuerFromAuthResponse, client.Attributes.ExcludeIssuerFromAuthResponse)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientAuthenticationFlowBindingOverrides(resourceName, flowResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if flowResourceName == "" {
			if client.AuthenticationFlowBindingOverrides.BrowserId != "" {
				return fmt.Errorf("expected openid client to have browserId set to empty, but got %s", client.AuthenticationFlowBindingOverrides.BrowserId)
			}

			if client.AuthenticationFlowBindingOverrides.DirectGrantId != "" {
				return fmt.Errorf("expected openid client to have directGrantId set to empty, but got %s", client.AuthenticationFlowBindingOverrides.DirectGrantId)
			}

		} else {
			flow, err := getAuthenticationFlowFromState(s, flowResourceName)
			if err != nil {
				return err
			}

			if client.AuthenticationFlowBindingOverrides.BrowserId != flow.Id {
				return fmt.Errorf("expected openid client to have browserId set to %s, but got %s", flow.Id, client.AuthenticationFlowBindingOverrides.BrowserId)
			}

			if client.AuthenticationFlowBindingOverrides.DirectGrantId != flow.Id {
				return fmt.Errorf("expected openid client to have directGrantId set to %s, but got %s", flow.Id, client.AuthenticationFlowBindingOverrides.DirectGrantId)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientLoginTheme(resourceName string, loginTheme string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.LoginTheme != loginTheme {
			return fmt.Errorf("expected openid client to have login theme set to %s, but got %s", loginTheme, client.Attributes.LoginTheme)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientUseRefreshTokens(resourceName string, useRefreshTokens bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.UseRefreshTokens != types.KeycloakBoolQuoted(useRefreshTokens) {
			return fmt.Errorf("expected openid client to have use refresh tokens set to %t, but got %v", useRefreshTokens, client.Attributes.UseRefreshTokens)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientEnableStandardTokenExchange(resourceName string, enableStandardTokenExchange bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.StandardTokenExchangeEnabled != types.KeycloakBoolQuoted(enableStandardTokenExchange) {
			return fmt.Errorf("expected openid client to have enable standard token exchange set to %t, but got %v", enableStandardTokenExchange, client.Attributes.StandardTokenExchangeEnabled)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientAllowRefreshTokenInStandardExchange(resourceName string, AllowRefreshTokenInStandardTokenExchange string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.AllowRefreshTokenInStandardTokenExchange != AllowRefreshTokenInStandardTokenExchange {
			return fmt.Errorf("expected openid client to have enable standard token exchange set to %s, but got %v", AllowRefreshTokenInStandardTokenExchange, client.Attributes.AllowRefreshTokenInStandardTokenExchange)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientUseRefreshTokensClientCredentials(resourceName string, useRefreshTokensClientCredentials bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.UseRefreshTokensClientCredentials != types.KeycloakBoolQuoted(useRefreshTokensClientCredentials) {
			return fmt.Errorf("expected openid client to have use refresh tokens client credentials set to %t, but got %v", useRefreshTokensClientCredentials, client.Attributes.UseRefreshTokensClientCredentials)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientOauth2DeviceAuthorizationGrantEnabled(resourceName string, oauth2DeviceAuthorizationGrantEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.Oauth2DeviceAuthorizationGrantEnabled != types.KeycloakBoolQuoted(oauth2DeviceAuthorizationGrantEnabled) {
			return fmt.Errorf("expected openid client to have device authorization grant enabled set to %t, but got %v", oauth2DeviceAuthorizationGrantEnabled, client.Attributes.Oauth2DeviceAuthorizationGrantEnabled)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientExtraConfig(resourceName string, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Attributes.ExtraConfig[key] != value {
			return fmt.Errorf("expected openid client to have attribute %v set to %v, but got %v", key, value, client.Attributes.ExtraConfig[key])
		}

		return nil
	}
}

// check that a particular extra config key is missing
func testAccCheckKeycloakOpenidClientExtraConfigMissing(resourceName string, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if val, ok := client.Attributes.ExtraConfig[key]; ok {
			// keycloak 13+ will remove attributes if set to empty string. on older versions, we'll just check if this value is empty
			if versionOk, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_13); !versionOk {
				if val != "" {
					return fmt.Errorf("expected openid client to have empty attribute %v", key)
				}

				return nil
			}

			return fmt.Errorf("expected openid client to not have attribute %v", key)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientExistsWithEnabledStatus(resourceName string, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if client.Enabled != enabled {
			return fmt.Errorf("expected openid client to have enabled status %t, but got %t", enabled, client.Enabled)
		}

		return nil
	}
}

func testAccCheckKeycloakOpenidClientExistsWithRegeneratedSecret(resourceName string, client *keycloak.OpenidClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := getOpenidClientFromState(s, resourceName)
		if err != nil {
			return err
		}

		if c.ClientSecret == client.ClientSecret {
			return fmt.Errorf("expected openid client to have new secret value but got same")
		}

		return nil
	}
}

func getOpenidClientFromState(s *terraform.State, resourceName string) (*keycloak.OpenidClient, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	client, err := keycloakClient.GetOpenidClient(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting openid client %s: %s", id, err)
	}

	return client, nil
}

func testKeycloakOpenidClient_basic(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakOpenidClient_basicWithSecretRegenerate(clientId string, regenerateValue string) string {
	return fmt.Sprintf(`
 data "keycloak_realm" "realm" {
 	realm = "%s"
 }
 resource "keycloak_openid_client" "client" {
 	client_id   = "%s"
 	realm_id    = data.keycloak_realm.realm.id
 	access_type = "CONFIDENTIAL"
 	client_secret_regenerate_when_changed = {
 		rotation = "%s"
 	}
 }
 	`, testAccRealm.Realm, clientId, regenerateValue)
}

func testKeycloakOpenidClient_basic_with_consent(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   				= "%s"
	realm_id    				= data.keycloak_realm.realm.id
	access_type 				= "CONFIDENTIAL"
	consent_required            = true
	display_on_consent_screen	= true
	consent_screen_text         = "some consent screen text"
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakOpenidClient_AccessToken_basic(clientId, accessTokenLifespan string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   		  = "%s"
	realm_id    		  = data.keycloak_realm.realm.id
	access_type 		  = "CONFIDENTIAL"
	access_token_lifespan = "%s"
}
	`, testAccRealm.Realm, clientId, accessTokenLifespan)
}

func testKeycloakOpenidClient_ClientTimeouts(clientId,
	offlineSessionIdleTimeout string, offlineSessionMaxLifespan string,
	sessionIdleTimeout string, sessionMaxLifespan string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   		  = "%s"
	realm_id    		  = data.keycloak_realm.realm.id
	access_type 		  = "CONFIDENTIAL"

	client_offline_session_idle_timeout = "%s"
	client_offline_session_max_lifespan = "%s"
	client_session_idle_timeout         = "%s"
	client_session_max_lifespan         = "%s"
}
	`, testAccRealm.Realm, clientId, offlineSessionIdleTimeout, offlineSessionMaxLifespan, sessionIdleTimeout, sessionMaxLifespan)
}

func testKeycloakOpenidClient_accessType(clientId, accessType string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "%s"
}
	`, testAccRealm.Realm, clientId, accessType)
}

func testKeycloakOpenidClient_clientAuthenticatorType(clientId, authType string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	realm_id                  = data.keycloak_realm.realm.id
	client_id                 = "%s"
	access_type               = "CONFIDENTIAL"
	client_authenticator_type = "%s"
}
	`, testAccRealm.Realm, clientId, authType)
}

func testKeycloakOpenidClient_pkceChallengeMethod(clientId, pkceChallengeMethod string) string {

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
	pkce_code_challenge_method = "%s"
}
	`, testAccRealm.Realm, clientId, pkceChallengeMethod)
}

func testKeycloakOpenidClient_excludeSessionStateFromAuthResponse(clientId string, excludeSessionStateFromAuthResponse bool) string {

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
	exclude_session_state_from_auth_response = %t
}
	`, testAccRealm.Realm, clientId, excludeSessionStateFromAuthResponse)
}

func testKeycloakOpenidClient_excludeIssuerFromAuthResponse(clientId string, excludeIssuerFromAuthResponse bool) string {

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
	exclude_issuer_from_auth_response = %t
}
	`, testAccRealm.Realm, clientId, excludeIssuerFromAuthResponse)
}

func testKeycloakOpenidClient_omitPkceChallengeMethod(clientId string) string {

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakOpenidClient_omitExcludeSessionStateFromAuthResponse(clientId, pkceChallengeMethod string) string {

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
    pkce_code_challenge_method = "%s"
}
	`, testAccRealm.Realm, clientId, pkceChallengeMethod)
}

func testKeycloakOpenidClient_omitExcludeIssuerFromAuthResponse(clientId, pkceChallengeMethod string) string {

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
    pkce_code_challenge_method = "%s"
}
	`, testAccRealm.Realm, clientId, pkceChallengeMethod)
}

func testKeycloakOpenidClient_updateRealmBefore(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_1" {
	realm = "%s"
}

data "keycloak_realm" "realm_2" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm_1.id
	access_type = "BEARER-ONLY"
}
	`, testAccRealm.Realm, testAccRealmTwo.Realm, clientId)
}

func testKeycloakOpenidClient_updateRealmAfter(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm_1" {
	realm = "%s"
}

data "keycloak_realm" "realm_2" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm_2.id
	access_type = "BEARER-ONLY"
}
	`, testAccRealm.Realm, testAccRealmTwo.Realm, clientId)
}

func testKeycloakOpenidClient_fromInterface(openidClient *keycloak.OpenidClient) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id                    = "%s"
	realm_id                     = data.keycloak_realm.realm.id
	name                         = "%s"
	enabled                      = %t
	description                  = "%s"

	access_type                  = "CONFIDENTIAL"
	client_secret                = "%s"

	standard_flow_enabled        = %t
	implicit_flow_enabled        = %t
	direct_access_grants_enabled = %t
	service_accounts_enabled     = %t

	valid_redirect_uris             = %s
	valid_post_logout_redirect_uris = %s
	web_origins                     = %s
	admin_url					    = "%s"
	base_url                        = "%s"
	root_url                        = "%s"

	backchannel_logout_url                     = "%s"
	backchannel_logout_session_required        = %t
	backchannel_logout_revoke_offline_sessions = %t
}
	`, testAccRealm.Realm, openidClient.ClientId, openidClient.Name, openidClient.Enabled, openidClient.Description, openidClient.ClientSecret, openidClient.StandardFlowEnabled, openidClient.ImplicitFlowEnabled, openidClient.DirectAccessGrantsEnabled, openidClient.ServiceAccountsEnabled, arrayOfStringsForTerraformResource(openidClient.ValidRedirectUris), arrayOfStringsForTerraformResource(openidClient.Attributes.PostLogoutRedirectUris), arrayOfStringsForTerraformResource(openidClient.WebOrigins), openidClient.AdminUrl, openidClient.BaseUrl, *openidClient.RootUrl, openidClient.Attributes.BackchannelLogoutUrl, openidClient.Attributes.BackchannelLogoutSessionRequired, openidClient.Attributes.BackchannelLogoutRevokeOfflineTokens)
}

func testKeycloakOpenidClient_backchannel(clientId, backchannelLogoutUrl string, backchannelLogoutSessionRequired, backchannelLogoutRevokeOfflineSessions bool) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"

	backchannel_logout_url                     = "%s"
	backchannel_logout_session_required        = %t
	backchannel_logout_revoke_offline_sessions = %t
}
	`, testAccRealm.Realm, clientId, backchannelLogoutUrl, backchannelLogoutSessionRequired, backchannelLogoutRevokeOfflineSessions)
}

func testKeycloakOpenidClient_frontchannel(clientId, frontchannelLogoutUrl string, frontchannelLogoutEnabled bool) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"

	frontchannel_logout_url     = "%s"
	frontchannel_logout_enabled = %t
}
	`, testAccRealm.Realm, clientId, frontchannelLogoutUrl, frontchannelLogoutEnabled)
}

func testKeycloakOpenidClient_secret(clientId, clientSecret string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id     = "%s"
	realm_id      = data.keycloak_realm.realm.id
	access_type   = "CONFIDENTIAL"
	client_secret = "%s"
}
	`, testAccRealm.Realm, clientId, clientSecret)
}

func testKeycloakOpenidClient_secretWriteOnly(clientId, clientSecretWriteOnly string, clientSecretWriteOnlyVersion int) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id                = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	access_type              = "CONFIDENTIAL"
	client_secret_wo         = "%s"
	client_secret_wo_version = "%d"
}
	`, testAccRealm.Realm, clientId, clientSecretWriteOnly, clientSecretWriteOnlyVersion)
}

func testKeycloakOpenidClient_invalidRedirectUris(clientId, accessType string, standardFlowEnabled, implicitFlowEnabled bool) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id             = "%s"
	realm_id              = data.keycloak_realm.realm.id
	access_type           = "%s"

	standard_flow_enabled = %t
	implicit_flow_enabled = %t
}
	`, testAccRealm.Realm, clientId, accessType, standardFlowEnabled, implicitFlowEnabled)
}

func testKeycloakOpenidClient_invalidPublicClientWithClientCredentials(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id                = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	access_type              = "PUBLIC"

	service_accounts_enabled = true
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakOpenidClient_bearerOnlyClientsCannotIssueTokens(clientId string, standardFlowEnabled, implicitFlowEnabled, directAccessGrantsEnabled, serviceAccountsEnabled bool) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id                    = "%s"
	realm_id                     = data.keycloak_realm.realm.id
	access_type                  = "BEARER-ONLY"

	standard_flow_enabled        = %t
	implicit_flow_enabled        = %t
	direct_access_grants_enabled = %t
	service_accounts_enabled     = %t
}
	`, testAccRealm.Realm, clientId, standardFlowEnabled, implicitFlowEnabled, directAccessGrantsEnabled, serviceAccountsEnabled)
}

func testKeycloakOpenidClient_authenticationFlowBindingOverrides(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "another_flow" {
  alias    = "anotherFlow"
  realm_id = data.keycloak_realm.realm.id
  description = "this is another flow"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "PUBLIC"
	authentication_flow_binding_overrides {
		browser_id = "${keycloak_authentication_flow.another_flow.id}"
		direct_grant_id = "${keycloak_authentication_flow.another_flow.id}"
	}
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakOpenidClient_withoutAuthenticationFlowBindingOverrides(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_authentication_flow" "another_flow" {
  alias    = "anotherFlow"
  realm_id = data.keycloak_realm.realm.id
  description = "this is another flow"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "PUBLIC"
}
	`, testAccRealm.Realm, clientId)
}

func testKeycloakOpenidClient_loginTheme(clientId, loginTheme string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "PUBLIC"
	login_theme = "%s"
}
	`, testAccRealm.Realm, clientId, loginTheme)
}

func testKeycloakOpenidClient_useRefreshTokens(clientId string, useRefreshTokens bool) string {

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
	use_refresh_tokens = %t
}
	`, testAccRealm.Realm, clientId, useRefreshTokens)
}

func testKeycloakOpenidClient_enableStandardTokenExchange(clientId string, enableStandardTokenExchange bool) string {

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
	standard_token_exchange_enabled = %t
}
	`, testAccRealm.Realm, clientId, enableStandardTokenExchange)
}

func testKeycloakOpenidClient_allowRefreshTokenInStandardExchange(clientId string, allowRefreshTokenInStandardTokenExchange string) string {

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
	standard_token_exchange_enabled = true
	allow_refresh_token_in_standard_token_exchange = "%s"
}
	`, testAccRealm.Realm, clientId, allowRefreshTokenInStandardTokenExchange)
}

func testKeycloakOpenidClient_useRefreshTokensClientCredentials(clientId string, useRefreshTokensClientCredentials bool) string {

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
	use_refresh_tokens_client_credentials = %t
}
	`, testAccRealm.Realm, clientId, useRefreshTokensClientCredentials)
}

func testKeycloakOpenidClient_extraConfig(clientId string, extraConfig map[string]string) string {
	var sb strings.Builder
	sb.WriteString("{\n")
	for k, v := range extraConfig {
		sb.WriteString(fmt.Sprintf("\t\t\"%s\" = \"%s\"\n", k, v))
	}
	sb.WriteString("}")

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "CONFIDENTIAL"
	extra_config = %s
}
	`, testAccRealm.Realm, clientId, sb.String())
}

func testKeycloakOpenidClient_oauth2DeviceAuthorizationGrantEnabled(clientId string, oauth2DeviceAuthorizationGrantEnabled bool) string {

	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   							  = "%s"
	realm_id    							  = data.keycloak_realm.realm.id
	access_type 							  = "CONFIDENTIAL"
	oauth2_device_authorization_grant_enabled = %t
}
	`, testAccRealm.Realm, clientId, oauth2DeviceAuthorizationGrantEnabled)
}

func testKeycloakOpenidClient_oauth2DeviceTimes(clientId, oauth2DeviceCodeLifespan, oauth2DevicePollingInterval string, oauth2DeviceAuthorizationGrantEnabled bool) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   			 					= "%s"
	realm_id    		     					= data.keycloak_realm.realm.id
	access_type 			 					= "CONFIDENTIAL"
	oauth2_device_authorization_grant_enabled 	= %t
	oauth2_device_code_lifespan 				= "%s"
	oauth2_device_polling_interval 				= "%s"
}
	`, testAccRealm.Realm, clientId, oauth2DeviceAuthorizationGrantEnabled, oauth2DeviceCodeLifespan, oauth2DevicePollingInterval)
}

func testKeycloakOpenidClient_import(clientId string, enabled bool) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "client" {
	client_id   = "%s"
	realm_id    = data.keycloak_realm.realm.id
	access_type = "PUBLIC"
	root_url    = ""
	enabled     = %t
	import      = true
}
	`, testAccRealm.Realm, clientId, enabled)
}
