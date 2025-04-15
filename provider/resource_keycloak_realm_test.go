package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"regexp"
	"testing"
)

func TestAccKeycloakRealm_basic(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
			{
				Config: testKeycloakRealm_notEnabled(realmName, realmDisplayName),
				Check:  testAccCheckKeycloakRealmEnabled("keycloak_realm.realm", false),
			},
			{
				Config: testKeycloakRealm_basic(realmName, fmt.Sprintf("%s-changed", realmDisplayName), realmDisplayName),
				Check:  testAccCheckKeycloakRealmDisplayName("keycloak_realm.realm", fmt.Sprintf("%s-changed", realmDisplayName)),
			},
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmDisplayNameHtml("keycloak_realm.realm", realmDisplayNameHtml),
			},
		},
	})
}

func TestAccKeycloakRealm_createAfterManualDestroy(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteRealm(testCtx, realmName)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
		},
	})
}

func TestAccKeycloakRealm_import(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
			{
				ResourceName:      "keycloak_realm.realm",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKeycloakRealm_OTP(t *testing.T) {
	realm := acctest.RandomWithPrefix("tf-acc")

	otpType := randomStringInSlice(keycloakRealmValidOTPTypes)
	otpAlgorithm := randomStringInSlice(keycloakRealmValidOTPAlgorithms)
	otpPeriod := acctest.RandIntRange(15, 45)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_WithOTP(realm, otpType, otpAlgorithm, otpPeriod),
				Check:  testAccCheckKeycloakRealmOTP("keycloak_realm.realm", otpType, otpAlgorithm, otpPeriod),
			},
		},
	})
}

func TestAccKeycloakRealm_SmtpServer(t *testing.T) {
	realm := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_WithSmtpServer(realm, "myhost.com", "My Host", "user"),
				Check:  testAccCheckKeycloakRealmSmtp("keycloak_realm.realm", "myhost.com", "My Host", "user"),
			},
			{
				Config: testKeycloakRealm_basic(realm, realm, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmSmtp("keycloak_realm.realm", "", "", ""),
			},
		},
	})
}

func TestAccKeycloakRealm_SmtpServerUpdate(t *testing.T) {
	realm := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_WithSmtpServer(realm, "myhost.com", "My Host", "user"),
				Check:  testAccCheckKeycloakRealmSmtp("keycloak_realm.realm", "myhost.com", "My Host", "user"),
			},
			{
				Config: testKeycloakRealm_WithSmtpServer(realm, "myhost2.com", "My Host2", "user2"),
				Check:  testAccCheckKeycloakRealmSmtp("keycloak_realm.realm", "myhost2.com", "My Host2", "user2"),
			},
		},
	})
}

func TestAccKeycloakRealm_SmtpServerInvalid(t *testing.T) {
	realm := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakRealm_WithSmtpServerWithoutHost(realm, "My Host"),
				ExpectError: regexp.MustCompile("The argument \"host\" is required, but no definition was found."),
			},
			{
				Config:      testKeycloakRealm_WithSmtpServerWithoutFrom(realm, "myhost.com"),
				ExpectError: regexp.MustCompile("The argument \"from\" is required, but no definition was found."),
			},
		},
	})
}

func TestAccKeycloakRealm_themes(t *testing.T) {
	realmOne := &keycloak.Realm{
		Realm:        "terraform-" + acctest.RandString(10),
		DisplayName:  "terraform-" + acctest.RandString(10),
		LoginTheme:   randomStringInSlice([]string{"base", "keycloak"}),
		AccountTheme: randomStringInSlice([]string{"base", "keycloak"}),
		AdminTheme:   randomStringInSlice([]string{"base", "keycloak"}),
		EmailTheme:   randomStringInSlice([]string{"base", "keycloak"}),
	}

	realmTwo := &keycloak.Realm{
		Realm:        realmOne.Realm,
		DisplayName:  realmOne.DisplayName,
		LoginTheme:   randomStringInSlice([]string{"base", "keycloak"}),
		AccountTheme: randomStringInSlice([]string{"base", "keycloak"}),
		AdminTheme:   randomStringInSlice([]string{"base", "keycloak"}),
		EmailTheme:   randomStringInSlice([]string{"base", "keycloak"}),
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_themes(realmOne),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
			{
				Config: testKeycloakRealm_themes(realmTwo),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
		},
	})
}

func TestAccKeycloakRealm_themesValidation(t *testing.T) {
	realm := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakRealm_themesValidation(realm, "login", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("validation error: theme \".+\" does not exist on the server"),
			},
			{
				Config:      testKeycloakRealm_themesValidation(realm, "account", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("validation error: theme \".+\" does not exist on the server"),
			},
			{
				Config:      testKeycloakRealm_themesValidation(realm, "admin", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("validation error: theme \".+\" does not exist on the server"),
			},
			{
				Config:      testKeycloakRealm_themesValidation(realm, "email", acctest.RandString(10)),
				ExpectError: regexp.MustCompile("validation error: theme \".+\" does not exist on the server"),
			},
		},
	})
}

func TestAccKeycloakRealm_InternationalizationValidation(t *testing.T) {
	realm := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakRealm_internationalizationValidationWithoutSupportedLocales(realm, "en"),
				ExpectError: regexp.MustCompile("The argument \"supported_locales\" is required, but no definition was found."),
			},
			{
				Config:      testKeycloakRealm_internationalizationValidation(realm, "en", "de"),
				ExpectError: regexp.MustCompile("validation error: DefaultLocale should be in the SupportLocales"),
			},
		},
	})
}

func TestAccKeycloakRealm_Internationalization(t *testing.T) {
	realm := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_internationalizationValidation(realm, "en", "en"),
				Check:  testAccCheckKeycloakRealmInternationalizationIsEnabled("keycloak_realm.realm", "en"),
			},
			{
				Config: testKeycloakRealm_internationalizationValidation(realm, "es", "es"),
				Check:  testAccCheckKeycloakRealmInternationalizationIsEnabled("keycloak_realm.realm", "es"),
			},
			{
				Config: testKeycloakRealm_basic(realm, realm, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmInternationalizationIsDisabled("keycloak_realm.realm"),
			},
		},
	})
}

func TestAccKeycloakRealm_InternationalizationDisabled(t *testing.T) {
	realm := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realm, realm, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmInternationalizationIsDisabled("keycloak_realm.realm"),
			},
		},
	})
}

func TestAccKeycloakRealm_loginConfigBasic(t *testing.T) {
	realm := &keycloak.Realm{
		Realm:                       "terraform-" + acctest.RandString(10),
		RegistrationAllowed:         true,
		RegistrationEmailAsUsername: true,
		EditUsernameAllowed:         randomBool(),
		ResetPasswordAllowed:        randomBool(),
		RememberMe:                  randomBool(),
		VerifyEmail:                 randomBool(),
		LoginWithEmailAllowed:       randomBool(),
		DuplicateEmailsAllowed:      false,
		SslRequired:                 "external",
	}

	updatedRealm := &keycloak.Realm{
		Realm:                       realm.Realm,
		RegistrationAllowed:         true,
		RegistrationEmailAsUsername: false,
		EditUsernameAllowed:         randomBool(),
		ResetPasswordAllowed:        randomBool(),
		RememberMe:                  randomBool(),
		VerifyEmail:                 randomBool(),
		LoginWithEmailAllowed:       false,
		DuplicateEmailsAllowed:      true,
		SslRequired:                 "all",
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_loginConfigBasic(realm),
				Check:  testKeycloakRealmLoginInfo("keycloak_realm.realm", realm),
			},
			{
				Config: testKeycloakRealm_loginConfigBasic(updatedRealm),
				Check:  testKeycloakRealmLoginInfo("keycloak_realm.realm", updatedRealm),
			},
		},
	})
}

func TestAccKeycloakRealm_loginConfigValidation(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakRealm_invalidRegistrationEmailAsUsernameAndDuplicateEmailsAllowed(realmName),
				ExpectError: regexp.MustCompile("validation error: DuplicateEmailsAllowed cannot be true if RegistrationEmailAsUsername is true"),
			},
			{
				Config:      testKeycloakRealm_invalidLoginWithEmailAllowedAndDuplicateEmailsAllowed(realmName),
				ExpectError: regexp.MustCompile("validation error: DuplicateEmailsAllowed cannot be true if LoginWithEmailAllowed is true"),
			},
			{
				Config:      testKeycloakRealm_invalidLoginWithSSLRequiredOnInvalidValue(realmName),
				ExpectError: regexp.MustCompile("validation error: SslRequired should be 'none', 'external' or 'all'"),
			},
		},
	})
}

func TestAccKeycloakRealm_tokenSettings(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
			{
				Config: testKeycloakRealm_tokenSettings(realmName),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
			// This is duplicated so another set of random value is used, effectively an update test
			{
				Config: testKeycloakRealm_tokenSettings(realmName),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
		},
	})
}

func TestAccKeycloakRealm_tokenSettingsOauth2Device(t *testing.T) {
	if ok, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_13); !ok {
		t.Skip()
	}

	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
			{
				Config: testKeycloakRealm_tokenSettingsOauth2Device(realmName),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
		},
	})
}

func TestAccKeycloakRealm_computedTokenSettings(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmExists("keycloak_realm.realm"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "sso_session_idle_timeout"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "sso_session_idle_timeout", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "sso_session_max_lifespan"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "sso_session_max_lifespan", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "offline_session_idle_timeout"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "offline_session_idle_timeout", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "offline_session_max_lifespan"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "offline_session_max_lifespan", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "client_session_idle_timeout"),
					resource.TestCheckResourceAttr("keycloak_realm.realm", "client_session_idle_timeout", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "client_session_max_lifespan"),
					resource.TestCheckResourceAttr("keycloak_realm.realm", "client_session_max_lifespan", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "access_token_lifespan"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "access_token_lifespan", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "access_token_lifespan_for_implicit_flow"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "access_token_lifespan_for_implicit_flow", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "access_code_lifespan"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "access_code_lifespan", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "access_code_lifespan_login"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "access_code_lifespan_login", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "access_code_lifespan_user_action"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "access_code_lifespan_user_action", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "action_token_generated_by_user_lifespan"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "action_token_generated_by_user_lifespan", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "action_token_generated_by_admin_lifespan"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "action_token_generated_by_admin_lifespan", "0s"),
				),
			},
		},
	})
}

func TestAccKeycloakRealm_oauth2DeviceSettings(t *testing.T) {
	if ok, _ := keycloakClient.VersionIsGreaterThanOrEqualTo(testCtx, keycloak.Version_13); !ok {
		t.Skip()
	}

	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmExists("keycloak_realm.realm"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "oauth2_device_code_lifespan"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "oauth2_device_code_lifespan", "0s"),

					resource.TestCheckResourceAttrSet("keycloak_realm.realm", "oauth2_device_polling_interval"),
					TestCheckResourceAttrNot("keycloak_realm.realm", "oauth2_device_polling_interval", "0"),
				),
			},
		},
	})
}

func TestAccKeycloakRealm_securityDefensesHeaders(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmSecurityDefensesHeaders("keycloak_realm.realm", "SAMEORIGIN"),
			},
			{
				Config: testKeycloakRealm_securityDefensesHeaders(realmName, realmDisplayName, "SAMEORIGIN"),
				Check:  testAccCheckKeycloakRealmSecurityDefensesHeaders("keycloak_realm.realm", "SAMEORIGIN"),
			},
			{
				Config: testKeycloakRealm_securityDefensesHeaders(realmName, realmDisplayName, "DENY"),
				Check:  testAccCheckKeycloakRealmSecurityDefensesHeaders("keycloak_realm.realm", "DENY"),
			},
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmSecurityDefensesHeaders("keycloak_realm.realm", "SAMEORIGIN"),
			},
		},
	})
}

func TestAccKeycloakRealm_securityDefensesBruteForceDetection(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetection("keycloak_realm.realm", false),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetectionFailureFactor("keycloak_realm.realm", 30),
				),
			},
			{
				Config: testKeycloakRealm_securityDefensesBruteForceDetection(realmName, realmDisplayName, 33),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetection("keycloak_realm.realm", true),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetectionFailureFactor("keycloak_realm.realm", 33),
				),
			},
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetection("keycloak_realm.realm", false),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetectionFailureFactor("keycloak_realm.realm", 30),
				),
			},
		},
	})
}

func TestAccKeycloakRealm_securityDefenses(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmSecurityDefensesHeaders("keycloak_realm.realm", "SAMEORIGIN"),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetection("keycloak_realm.realm", false),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetectionFailureFactor("keycloak_realm.realm", 30),
				),
			},
			{
				Config: testKeycloakRealm_securityDefenses(realmName, realmDisplayName, "DENY", 33),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmSecurityDefensesHeaders("keycloak_realm.realm", "DENY"),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetection("keycloak_realm.realm", true),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetectionFailureFactor("keycloak_realm.realm", 33),
				),
			},
			{
				Config: testKeycloakRealm_securityDefensesHeaders(realmName, realmDisplayName, "SAMEORIGIN"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmSecurityDefensesHeaders("keycloak_realm.realm", "SAMEORIGIN"),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetection("keycloak_realm.realm", false),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetectionFailureFactor("keycloak_realm.realm", 30),
				),
			},
			{
				Config: testKeycloakRealm_securityDefensesBruteForceDetection(realmName, realmDisplayName, 31),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmSecurityDefensesHeaders("keycloak_realm.realm", "SAMEORIGIN"),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetection("keycloak_realm.realm", true),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetectionFailureFactor("keycloak_realm.realm", 31),
				),
			},
			{
				Config: testKeycloakRealm_securityDefenses(realmName, realmDisplayName, "DENY", 37),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmSecurityDefensesHeaders("keycloak_realm.realm", "DENY"),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetection("keycloak_realm.realm", true),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetectionFailureFactor("keycloak_realm.realm", 37),
				),
			},
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmSecurityDefensesHeaders("keycloak_realm.realm", "SAMEORIGIN"),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetection("keycloak_realm.realm", false),
					testAccCheckKeycloakRealmSecurityDefensesBruteForceDetectionFailureFactor("keycloak_realm.realm", 30),
				),
			},
		},
	})
}

func TestAccKeycloakRealm_passwordPolicy(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")
	passwordPolicyStringValid1 := "upperCase(1) and length(8) and forceExpiredPasswordChange(365) and notUsername"
	passwordPolicyStringValid2 := "upperCase(1) and length(8)"
	passwordPolicyStringValid3 := "lowerCase(2)"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmPasswordPolicy("keycloak_realm.realm", ""),
			},
			{
				Config: testKeycloakRealm_passwordPolicy(realmName, realmDisplayName, passwordPolicyStringValid1),
				Check:  testAccCheckKeycloakRealmPasswordPolicy("keycloak_realm.realm", passwordPolicyStringValid1),
			},
			{
				Config: testKeycloakRealm_passwordPolicy(realmName, realmDisplayName, passwordPolicyStringValid2),
				Check:  testAccCheckKeycloakRealmPasswordPolicy("keycloak_realm.realm", passwordPolicyStringValid2),
			},
			{
				Config: testKeycloakRealm_passwordPolicy(realmName, realmDisplayName, passwordPolicyStringValid3),
				Check:  testAccCheckKeycloakRealmPasswordPolicy("keycloak_realm.realm", passwordPolicyStringValid3),
			},
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmPasswordPolicy("keycloak_realm.realm", ""),
			},
		},
	})
}

func TestAccKeycloakRealm_browserFlow(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")
	newBrowserFlow := "registration"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmBrowserFlow("keycloak_realm.realm", "browser"),
			},
			{
				Config: testKeycloakRealm_browserFlow(realmName, realmDisplayName, newBrowserFlow),
				Check:  testAccCheckKeycloakRealmBrowserFlow("keycloak_realm.realm", newBrowserFlow),
			},
			// when a realm binding argument is unset, it will remain the same
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmBrowserFlow("keycloak_realm.realm", newBrowserFlow),
			},
		},
	})
}

func TestAccKeycloakRealm_customAttribute(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	key := acctest.RandomWithPrefix("tf-acc")
	value := acctest.RandomWithPrefix("tf-acc")
	value2 := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_withCustomAttribute(realmName, key, value),
				Check:  testAccCheckKeycloakRealmCustomAttribute("keycloak_realm.realm", key, value),
			},
			{
				Config: testKeycloakRealm_withCustomAttribute(realmName, key, value2),
				Check:  testAccCheckKeycloakRealmCustomAttribute("keycloak_realm.realm", key, value2),
			},
		},
	})
}

func TestAccKeycloakRealm_passwordPolicyInvalid(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")
	passwordPolicyStringInvalid1 := "unknownpolicy(1) and length(8) and forceExpiredPasswordChange(365) and notUsername"
	passwordPolicyStringInvalid2 := "lowerCase(1) and length(8) and unknownpolicy(365) and notUsername"
	passwordPolicyStringInvalid3 := "unknownpolicy(2)"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmPasswordPolicy("keycloak_realm.realm", ""),
			},
			{
				Config:      testKeycloakRealm_passwordPolicy(realmName, realmDisplayName, passwordPolicyStringInvalid1),
				ExpectError: regexp.MustCompile("validation error: password-policy .+ does not exist on the server, installed providers: .+"),
			},
			{
				Config:      testKeycloakRealm_passwordPolicy(realmName, realmDisplayName, passwordPolicyStringInvalid2),
				ExpectError: regexp.MustCompile("validation error: password-policy .+ does not exist on the server, installed providers: .+"),
			},
			{
				Config:      testKeycloakRealm_passwordPolicy(realmName, realmDisplayName, passwordPolicyStringInvalid3),
				ExpectError: regexp.MustCompile("validation error: password-policy .+ does not exist on the server, installed providers: .+"),
			},
		},
	})
}

func TestAccKeycloakRealm_createWithInternalId(t *testing.T) {
	realmName := "my-realm"
	internalId := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_basicInternalId(realmName, internalId),
				Check: func(state *terraform.State) error {
					realmWithInternalId, err := keycloakClient.GetRealm(testCtx, realmName)
					if err != nil {
						return err
					}

					if realmWithInternalId.Id != internalId {
						return fmt.Errorf("expected realm with internal id %s but got %s", internalId, realmWithInternalId.Id)
					}

					return nil
				},
			},
		},
	})
}

func TestAccKeycloakRealm_internalId(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	internalId := acctest.RandomWithPrefix("tf-acc")
	realm := &keycloak.Realm{
		Realm: realmName,
		Id:    internalId,
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				ResourceName:  "keycloak_realm.realm",
				ImportStateId: realmName,
				ImportState:   true,
				Config:        testKeycloakRealm_basic(realmName, "foo", "<b>foo</b>"),
				PreConfig: func() {
					err := keycloakClient.NewRealm(testCtx, realm)
					if err != nil {
						t.Fatal(err)
					}
				},
				Check: testAccCheckKeycloakRealmWithInternalId(realmName, internalId),
			},
		},
	})
}

func TestAccKeycloakRealm_default_client_scopes(t *testing.T) {

	realmName := acctest.RandomWithPrefix("tf-acc")
	defaultDefaultClientScope := []string{"profile"}
	defaultOptionalClientScope := []string{"email", "roles"}

	realm := &keycloak.Realm{
		Realm: realmName,
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				ResourceName:  "keycloak_realm.realm",
				ImportStateId: realmName,
				ImportState:   true,
				Config:        testKeycloakRealm_default_client_scopes(realmName, defaultDefaultClientScope, defaultOptionalClientScope),
				PreConfig: func() {
					err := keycloakClient.NewRealm(testCtx, realm)
					if err != nil {
						t.Fatal(err)
					}
				},
				Check: testAccCheckKeycloakRealm_default_client_scopes(realmName, defaultDefaultClientScope, defaultOptionalClientScope),
			},
		},
	})

	// test empty default client scope configuration
	realmName2 := acctest.RandomWithPrefix("tf-acc")
	var defaultDefaultClientScope2 []string  // deliberately empty
	var defaultOptionalClientScope2 []string // deliberately empty

	realm2 := &keycloak.Realm{
		Realm: realmName2,
	}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				ResourceName:  "keycloak_realm.realm",
				ImportStateId: realmName2,
				ImportState:   true,
				Config:        testKeycloakRealm_default_client_scopes(realmName2, defaultDefaultClientScope2, defaultOptionalClientScope2),
				PreConfig: func() {
					err := keycloakClient.NewRealm(testCtx, realm2)
					if err != nil {
						t.Fatal(err)
					}
				},
				Check: testAccCheckKeycloakRealm_default_client_scopes(realmName2, defaultDefaultClientScope2, defaultOptionalClientScope2),
			},
		},
	})
}

func testKeycloakRealm_default_client_scopes(realm string, defaultDefaultClientScopes []string, defaultOptionalClientScopes []string) string {

	defaultDefaultClientScopesString := fmt.Sprintf("%s", arrayOfStringsForTerraformResource(defaultDefaultClientScopes))
	defaultOptionalClientScopesString := fmt.Sprintf("%s", arrayOfStringsForTerraformResource(defaultOptionalClientScopes))

	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm                          = "%s"
	enabled                        = true
	default_default_client_scopes  = %s
	default_optional_client_scopes = %s
}
	`, realm, defaultDefaultClientScopesString, defaultOptionalClientScopesString)
}

func testAccCheckKeycloakRealm_default_client_scopes(resourceName string, defaultDefaultClientScope, defaultOptionalClientScope []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if len(defaultDefaultClientScope) == 0 {
			if len(realm.DefaultDefaultClientScopes) != 0 {
				return fmt.Errorf("expected realm %s to have empty default default client scopes but was %s", realm.Realm, realm.DefaultDefaultClientScopes)
			}
		} else {
			for _, expectedScope := range defaultDefaultClientScope {
				found := false
				for _, s := range realm.DefaultDefaultClientScopes {
					if expectedScope == s {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("expected realm %s to have default default client scopes with value %s but was %s", realm.Realm, defaultDefaultClientScope, realm.DefaultDefaultClientScopes)
				}
			}
		}

		if len(defaultOptionalClientScope) == 0 {
			if len(realm.DefaultOptionalClientScopes) != 0 {
				return fmt.Errorf("expected realm %s to have empty default optional client scopes but was %s", realm.Realm, realm.DefaultOptionalClientScopes)
			}
		} else {
			for _, expectedScope := range defaultOptionalClientScope {
				found := false
				for _, s := range realm.DefaultOptionalClientScopes {
					if expectedScope == s {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("expected realm %s to have default optional client scopes with value %s but was %s", realm.Realm, defaultOptionalClientScope, realm.DefaultOptionalClientScopes)
				}
			}
		}

		return nil
	}
}

func TestAccKeycloakRealm_webauthn(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayName := acctest.RandomWithPrefix("tf-acc")
	realmDisplayNameHtml := acctest.RandomWithPrefix("tf-acc")
	rpName := acctest.RandomWithPrefix("tf-acc")
	rpId := acctest.RandomWithPrefix("tf-acc")
	attestationConveyancePreference := randomStringInSlice([]string{"none", "indirect", "not specified"})
	authenticatorAttachment := randomStringInSlice([]string{"platform", "cross-platform", "not specified"})
	requireResidentKey := randomStringInSlice([]string{"Yes", "No", "not specified"})
	userVerificationRequirement := randomStringInSlice([]string{"not specified", "required", "preferred", "discouraged"})
	signatureAlgorithms := randomStringSliceSubset([]string{"ES256", "ES384", "ES512", "RS256", "ES384", "ES512"})
	avoidSameAuthenticatorRegister := randomBool()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealm_webauthn_policy(realmName, realmDisplayName, realmDisplayNameHtml, rpName, rpId, attestationConveyancePreference, authenticatorAttachment, requireResidentKey, userVerificationRequirement, signatureAlgorithms, avoidSameAuthenticatorRegister),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
			{
				Config: testKeycloakRealm_webauthn_passwordless_policy(realmName, realmDisplayName, realmDisplayNameHtml, rpName, rpId, attestationConveyancePreference, authenticatorAttachment, requireResidentKey, userVerificationRequirement, signatureAlgorithms, avoidSameAuthenticatorRegister),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
			{
				Config: testKeycloakRealm_basic(realmName, realmDisplayName, realmDisplayNameHtml),
				Check:  testAccCheckKeycloakRealmExists("keycloak_realm.realm"),
			},
		},
	})
}

func testKeycloakRealmLoginInfo(resourceName string, realm *keycloak.Realm) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realmFromState, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realmFromState.Realm != realm.Realm {
			return fmt.Errorf("expected realm in state to have name %s, but was %s", realmFromState.Realm, realm.Realm)
		}

		if realmFromState.RegistrationAllowed != realm.RegistrationAllowed {
			return fmt.Errorf("expected realm %s to have registration_allowed set to %t, but was %t", realm.Realm, realm.RegistrationAllowed, realmFromState.RegistrationAllowed)
		}

		if realmFromState.RegistrationEmailAsUsername != realm.RegistrationEmailAsUsername {
			return fmt.Errorf("expected realm %s to have registration_email_as_username set to %t, but was %t", realm.Realm, realm.RegistrationEmailAsUsername, realmFromState.RegistrationEmailAsUsername)
		}

		if realmFromState.EditUsernameAllowed != realm.EditUsernameAllowed {
			return fmt.Errorf("expected realm %s to have edit_username_allowed set to %t, but was %t", realm.Realm, realm.EditUsernameAllowed, realmFromState.EditUsernameAllowed)
		}

		if realmFromState.ResetPasswordAllowed != realm.ResetPasswordAllowed {
			return fmt.Errorf("expected realm %s to have reset_password_allowed set to %t, but was %t", realm.Realm, realm.ResetPasswordAllowed, realmFromState.ResetPasswordAllowed)
		}

		if realmFromState.RememberMe != realm.RememberMe {
			return fmt.Errorf("expected realm %s to have remember_me set to %t, but was %t", realm.Realm, realm.RememberMe, realmFromState.RememberMe)
		}

		if realmFromState.VerifyEmail != realm.VerifyEmail {
			return fmt.Errorf("expected realm %s to have verify_email set to %t, but was %t", realm.Realm, realm.VerifyEmail, realmFromState.VerifyEmail)
		}

		if realmFromState.LoginWithEmailAllowed != realm.LoginWithEmailAllowed {
			return fmt.Errorf("expected realm %s to have login_with_email_allowed set to %t, but was %t", realm.Realm, realm.LoginWithEmailAllowed, realmFromState.LoginWithEmailAllowed)
		}

		if realmFromState.DuplicateEmailsAllowed != realm.DuplicateEmailsAllowed {
			return fmt.Errorf("expected realm %s to have duplicate_emails_allowed set to %t, but was %t", realm.Realm, realm.DuplicateEmailsAllowed, realmFromState.DuplicateEmailsAllowed)
		}

		if realmFromState.SslRequired != realm.SslRequired {
			return fmt.Errorf("expected realm %s to have ssl_required set to %s, but was %s", realm.Realm, realm.SslRequired, realmFromState.SslRequired)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakRealmEnabled(resourceName string, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realm.Enabled != enabled {
			return fmt.Errorf("expected realm %s to have enabled set to %t, but was %t", realm.Realm, enabled, realm.Enabled)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmDisplayName(resourceName string, displayName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realm.DisplayName != displayName {
			return fmt.Errorf("expected realm %s to have display name set to %s, but was %s", realm.Realm, displayName, realm.DisplayName)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmDisplayNameHtml(resourceName string, displayNameHtml string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realm.DisplayNameHtml != displayNameHtml {
			return fmt.Errorf("expected realm %s to have display name html set to %s, but was %s", realm.Realm, displayNameHtml, realm.DisplayNameHtml)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmSmtp(resourceName, host, from, user string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realm.SmtpServer.Host != host {
			return fmt.Errorf("expected realm %s to have smtp host set to %s, but was %s", realm.Realm, host, realm.SmtpServer.Host)
		}

		if realm.SmtpServer.From != from {
			return fmt.Errorf("expected realm %s to have smtp from set to %s, but was %s", realm.Realm, from, realm.SmtpServer.From)
		}

		if realm.SmtpServer.User != user {
			return fmt.Errorf("expected realm %s to have smtp user set to %s, but was %s", realm.Realm, user, realm.SmtpServer.User)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmOTP(resourceName, otpType, algorithm string, period int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realm.OTPPolicyType != otpType {
			return fmt.Errorf("expected realm %s to have OTP type set to %s, but was %s", realm.Realm, otpType, realm.OTPPolicyType)
		}

		if realm.OTPPolicyAlgorithm != algorithm {
			return fmt.Errorf("expected realm %s to have OTP algorithm set to %s, but was %s", realm.Realm, algorithm, realm.OTPPolicyAlgorithm)
		}

		if realm.OTPPolicyPeriod != period {
			return fmt.Errorf("expected realm %s to have OTP period set to %d, but was %d", realm.Realm, period, realm.OTPPolicyPeriod)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmInternationalizationIsEnabled(resourceName string, defaultLocale string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if !realm.InternationalizationEnabled {
			return fmt.Errorf("expected realm %s to have internationalization enabled but was disabled", realm.Realm)
		}

		if realm.DefaultLocale != defaultLocale {
			return fmt.Errorf("expected realm %s to have defaultLocale set to %s, but was %s", realm.Realm, defaultLocale, realm.DefaultLocale)
		}

		if !contains(realm.SupportLocales, defaultLocale) {
			return fmt.Errorf("expected realm %s to contain defaultLocale %s, but was %s", realm.Realm, defaultLocale, realm.SupportLocales)
		}
		return nil
	}
}

func testAccCheckKeycloakRealmInternationalizationIsDisabled(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realm.InternationalizationEnabled {
			return fmt.Errorf("expected realm %s to have internationalization disabled but was enabled", realm.Realm)
		}
		return nil
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func testAccCheckKeycloakRealmDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm" {
				continue
			}

			realmName := rs.Primary.ID
			realm, _ := keycloakClient.GetRealm(testCtx, realmName)
			if realm != nil {
				return fmt.Errorf("realm %s still exists", realmName)
			}
		}

		return nil
	}
}

func getRealmFromState(s *terraform.State, resourceName string) (*keycloak.Realm, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realmName := rs.Primary.Attributes["realm"]

	realm, err := keycloakClient.GetRealm(testCtx, realmName)
	if err != nil {
		return nil, fmt.Errorf("error getting realm %s: %s", realmName, err)
	}

	return realm, nil
}

func testAccCheckKeycloakRealmSecurityDefensesHeaders(resourceName, xFrameOptions string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realm.BrowserSecurityHeaders.XFrameOptions != xFrameOptions {
			return fmt.Errorf("expected realm %s to have BrowserSecurityHeaders xFrameOptions set to %s, but was %s", realm.Realm, xFrameOptions, realm.BrowserSecurityHeaders.XFrameOptions)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmSecurityDefensesBruteForceDetection(resourceName string, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realm.BruteForceProtected != enabled {
			return fmt.Errorf("expected realm %s to have BruteForceProtection set to %t, but was %t", realm.Realm, enabled, realm.BruteForceProtected)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmSecurityDefensesBruteForceDetectionFailureFactor(resourceName string, maxLoginFailures int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realm.FailureFactor != maxLoginFailures {
			return fmt.Errorf("expected realm %s to have FailureFactor set to %d, but was %d", realm.Realm, maxLoginFailures, realm.FailureFactor)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmPasswordPolicy(resourceName, passwordPolicy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realm.PasswordPolicy != passwordPolicy {
			return fmt.Errorf("expected realm %s to have passwordPolicy %s, but was %s", realm.Realm, passwordPolicy, realm.PasswordPolicy)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmBrowserFlow(resourceName, browserFlow string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if *realm.BrowserFlow != browserFlow {
			return fmt.Errorf("expected realm %s to have browserFlow binding %s, but was %s", realm.Realm, browserFlow, *realm.BrowserFlow)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmCustomAttribute(resourceName, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realm.Attributes[key] != value {
			return fmt.Errorf("expected realm %s to have an attribute %s with value %s but was %s", realm.Realm, key, value, realm.Attributes[key])
		}

		return nil
	}
}

func testAccCheckKeycloakRealmWithInternalId(resourceName, id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realm, err := getRealmFromState(s, resourceName)
		if err != nil {
			return err
		}

		if realm.Id != id {
			return fmt.Errorf("expected realm %s to have an internal id with value %s but was %s", realm.Realm, id, realm.Id)
		}

		return nil
	}
}

func testKeycloakRealm_basic(realm, realmDisplayName, realmDisplayNameHtml string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        		= "%s"
	enabled     	 	= true
	display_name 		= "%s"
	display_name_html 	= "%s"
}
	`, realm, realmDisplayName, realmDisplayNameHtml)
}

func testKeycloakRealm_WithSmtpServer(realm, host, from, user string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
	enabled = true
	display_name = "%s"
	smtp_server {
		host = "%s"
		port = 25
		from_display_name = "Tom"
		from = "%s"
		reply_to_display_name = "Tom"
		reply_to = "tom@myhost.com"
		ssl = true
		starttls = true
		envelope_from = "nottom@myhost.com"
		auth {
			username = "%s"
			password = "tom"
		}
	}
}
	`, realm, realm, host, from, user)
}

func testKeycloakRealm_WithOTP(realm, otpType, algorithm string, period int) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm   = "%s"
	enabled = true

	otp_policy {
		type      = "%s"
		algorithm = "%s"
		period    = %d
	}
}
	`, realm, otpType, algorithm, period)
}

func testKeycloakRealm_WithSmtpServerWithoutHost(realm, from string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
	enabled = true
	display_name = "%s"
	smtp_server {
		port = 25
		from_display_name = "Tom"
		from = "%s"
		reply_to_display_name = "Tom"
		reply_to = "tom@myhost.com"
		ssl = true
		starttls = true
		envelope_from = "nottom@myhost.com"
		auth {
			username = "tom"
			password = "tom"
		}
	}
}
	`, realm, realm, from)
}

func testKeycloakRealm_WithSmtpServerWithoutFrom(realm, host string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
	enabled = true
	display_name = "%s"
	smtp_server {
		host = "%s"
		port = 25
		from_display_name = "Tom"
		reply_to_display_name = "Tom"
		reply_to = "tom@myhost.com"
		ssl = true
		starttls = true
		envelope_from = "nottom@myhost.com"
		auth {
			username = "tom"
			password = "tom"
		}
	}
}
	`, realm, realm, host)
}

func testKeycloakRealm_themes(realm *keycloak.Realm) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"
	enabled      = true
	display_name = "%s"

	login_theme   = "%s"
	account_theme = "%s"
	admin_theme   = "%s"
	email_theme   = "%s"
}
	`, realm.Realm, realm.DisplayName, realm.LoginTheme, realm.AccountTheme, realm.AdminTheme, realm.EmailTheme)
}

func testKeycloakRealm_themesValidation(realm, theme, value string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"
	enabled      = true
	display_name = "%s"

	%s_theme     = "%s"
}
	`, realm, realm, theme, value)
}

func testKeycloakRealm_internationalizationValidation(realm, supportedLocale, defaultLocale string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"
	enabled      = true
	display_name = "%s"
	internationalization {
		supported_locales	= ["nl", "%s", "fr"]
		default_locale		= "%s"
	}
}
	`, realm, realm, supportedLocale, defaultLocale)
}

func testKeycloakRealm_internationalizationValidationWithoutSupportedLocales(realm, defaultLocale string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"
	enabled      = true
	display_name = "%s"
	internationalization {
		default_locale		= "%s"
	}

}
	`, realm, realm, defaultLocale)
}

func testKeycloakRealm_notEnabled(realm, realmDisplayName string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"
	enabled      = false
	display_name = "%s"
}
	`, realm, realmDisplayName)
}

func testKeycloakRealm_loginConfigBasic(realm *keycloak.Realm) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"

	registration_allowed           = "%t"
	registration_email_as_username = "%t"
	edit_username_allowed          = "%t"
	reset_password_allowed         = "%t"
	remember_me                    = "%t"
	verify_email                   = "%t"
	login_with_email_allowed       = "%t"
	duplicate_emails_allowed       = "%t"
	ssl_required       			   = "%s"
}
	`, realm.Realm, realm.RegistrationAllowed, realm.RegistrationEmailAsUsername, realm.EditUsernameAllowed, realm.ResetPasswordAllowed, realm.RememberMe, realm.VerifyEmail, realm.LoginWithEmailAllowed, realm.DuplicateEmailsAllowed, realm.SslRequired)
}

func testKeycloakRealm_invalidRegistrationEmailAsUsernameAndDuplicateEmailsAllowed(realm string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm                          = "%s"

	registration_allowed           = true
	registration_email_as_username = true
	duplicate_emails_allowed       = true
}
	`, realm)
}

func testKeycloakRealm_invalidLoginWithEmailAllowedAndDuplicateEmailsAllowed(realm string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm                          = "%s"

	login_with_email_allowed       = true
	duplicate_emails_allowed       = true
}
	`, realm)
}

func testKeycloakRealm_invalidLoginWithSSLRequiredOnInvalidValue(realm string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm         = "%s"
	ssl_required  = "somethingElse"
}
	`, realm)
}

func testKeycloakRealm_tokenSettings(realm string) string {
	defaultSignatureAlgorithm := "RS256"
	ssoSessionIdleTimeout := randomDurationString()
	ssoSessionMaxLifespan := randomDurationString()
	ssoSessionIdleTimeoutRememberMe := randomDurationString()
	ssoSessionMaxLifespanRememberMe := randomDurationString()
	offlineSessionIdleTimeout := randomDurationString()
	offlineSessionMaxLifespan := randomDurationString()
	clientSessionIdleTimeout := randomDurationString()
	clientSessionMaxLifespan := randomDurationString()
	accessTokenLifespan := randomDurationString()
	accessTokenLifespanForImplicitFlow := randomDurationString()
	accessCodeLifespan := randomDurationString()
	accessCodeLifespanLogin := randomDurationString()
	accessCodeLifespanUserAction := randomDurationString()
	actionTokenGeneratedByUserLifespan := randomDurationString()
	actionTokenGeneratedByAdminLifespan := randomDurationString()

	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm                                    = "%s"
	enabled                                  = true
	display_name                             = "%s"

	default_signature_algorithm              = "%s"
	sso_session_idle_timeout                 = "%s"
	sso_session_max_lifespan                 = "%s"
	sso_session_idle_timeout_remember_me     = "%s"
	sso_session_max_lifespan_remember_me     = "%s"
	offline_session_idle_timeout             = "%s"
	offline_session_max_lifespan             = "%s"
	offline_session_max_lifespan_enabled     = true
	client_session_idle_timeout              = "%s"
	client_session_max_lifespan              = "%s"
	access_token_lifespan                    = "%s"
	access_token_lifespan_for_implicit_flow  = "%s"
	access_code_lifespan                     = "%s"
	access_code_lifespan_login               = "%s"
	access_code_lifespan_user_action         = "%s"
	action_token_generated_by_user_lifespan  = "%s"
	action_token_generated_by_admin_lifespan = "%s"
}
	`, realm, realm, defaultSignatureAlgorithm, ssoSessionIdleTimeout, ssoSessionMaxLifespan, ssoSessionIdleTimeoutRememberMe, ssoSessionMaxLifespanRememberMe, offlineSessionIdleTimeout, offlineSessionMaxLifespan, clientSessionIdleTimeout, clientSessionMaxLifespan, accessTokenLifespan, accessTokenLifespanForImplicitFlow, accessCodeLifespan, accessCodeLifespanLogin, accessCodeLifespanUserAction, actionTokenGeneratedByUserLifespan, actionTokenGeneratedByAdminLifespan)
}

func testKeycloakRealm_tokenSettingsOauth2Device(realm string) string {
	oauth2DeviceCodeLifespan := randomDurationString()
	oauth2DevicePollingInterval := "10"

	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm                                    = "%s"
	enabled                                  = true
	display_name                             = "%s"

	oauth2_device_code_lifespan              = "%s"
	oauth2_device_polling_interval           = "%s"
}
	`, realm, realm, oauth2DeviceCodeLifespan, oauth2DevicePollingInterval)
}

func testKeycloakRealm_securityDefensesHeaders(realm, realmDisplayName, xFrameOptions string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"
	enabled      = true
	display_name = "%s"
	security_defenses {
    	headers {
			x_frame_options = "%s"
			content_security_policy = "frame-src 'self'; frame-ancestors 'self'; object-src 'none';"
			content_security_policy_report_only = ""
			x_content_type_options = "nosniff"
			x_robots_tag = "none"
			x_xss_protection = "1; mode=block"
			strict_transport_security = "max-age=31536000; includeSubDomains"
		}
	}
}
	`, realm, realmDisplayName, xFrameOptions)
}

func testKeycloakRealm_securityDefensesBruteForceDetection(realm, realmDisplayName string, maxLoginFailures int) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"
	enabled      = true
	display_name = "%s"
	security_defenses {
    	brute_force_detection {
            permanent_lockout                 = false
      		max_login_failures                = %d
      		wait_increment_seconds            = 60
      		quick_login_check_milli_seconds   = 1000
      		minimum_quick_login_wait_seconds  = 60
      		max_failure_wait_seconds          = 900
      		failure_reset_time_seconds        = 43200
        }
	}
}
	`, realm, realmDisplayName, maxLoginFailures)
}

func testKeycloakRealm_securityDefenses(realm, realmDisplayName, xFrameOptions string, maxLoginFailures int) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"
	enabled      = true
	display_name = "%s"
	security_defenses {
    	headers {
			x_frame_options = "%s"
			content_security_policy = "frame-src 'self'; frame-ancestors 'self'; object-src 'none';"
			content_security_policy_report_only = ""
			x_content_type_options = "nosniff"
			x_robots_tag = "none"
			x_xss_protection = "1; mode=block"
			strict_transport_security = "max-age=31536000; includeSubDomains"
			referrer_policy = "origin"
		}
		brute_force_detection {
            permanent_lockout                 = false
      		max_login_failures                = %d
      		wait_increment_seconds            = 60
      		quick_login_check_milli_seconds   = 1000
      		minimum_quick_login_wait_seconds  = 60
      		max_failure_wait_seconds          = 900
      		failure_reset_time_seconds        = 43200
        }
	}
}
	`, realm, realmDisplayName, xFrameOptions, maxLoginFailures)
}

func testKeycloakRealm_passwordPolicy(realm, realmDisplayName, passwordPolicy string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"
	enabled      = true
	display_name = "%s"
	password_policy = "%s"
}
	`, realm, realmDisplayName, passwordPolicy)
}

func testKeycloakRealm_browserFlow(realm, realmDisplayName, browserFlow string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"
	enabled      = true
	display_name = "%s"
	browser_flow = "%s"
}
	`, realm, realmDisplayName, browserFlow)
}

func testKeycloakRealm_withCustomAttribute(realm, key, value string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"
	enabled      = true
	attributes   = {
		%s = "%s"
	}
}
	`, realm, key, value)
}

func testKeycloakRealm_webauthn_policy(realm, realmDisplayName, realmDisplayNameHtml, rpName, rpId, attestationConveyancePreference, authenticatorAttachment, requireResidentKey, userVerificationRequirement string, signatureAlgorithms []string, avoidSameAuthenticatorRegister bool) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        		= "%s"
	enabled     	 	= true
	display_name 		= "%s"
	display_name_html 	= "%s"

	web_authn_policy {
		relying_party_entity_name         = "%s"
		relying_party_id                  = "%s"
		signature_algorithms              = %s

		attestation_conveyance_preference = "%s"
		authenticator_attachment          = "%s"
		avoid_same_authenticator_register = %t
		require_resident_key              = "%s"
		user_verification_requirement     = "%s"
	}
}
	`, realm, realmDisplayName, realmDisplayNameHtml, rpName, rpId, arrayOfStringsForTerraformResource(signatureAlgorithms), attestationConveyancePreference, authenticatorAttachment, avoidSameAuthenticatorRegister, requireResidentKey, userVerificationRequirement)
}

func testKeycloakRealm_webauthn_passwordless_policy(realm, realmDisplayName, realmDisplayNameHtml, rpName, rpId, attestationConveyancePreference, authenticatorAttachment, requireResidentKey, userVerificationRequirement string, signatureAlgorithms []string, avoidSameAuthenticatorRegister bool) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        		= "%s"
	enabled     	 	= true
	display_name 		= "%s"
	display_name_html 	= "%s"

	web_authn_passwordless_policy {
		relying_party_entity_name         = "%s"
		relying_party_id                  = "%s"
		signature_algorithms              = %s

		attestation_conveyance_preference = "%s"
		authenticator_attachment          = "%s"
		avoid_same_authenticator_register = %t
		require_resident_key              = "%s"
		user_verification_requirement     = "%s"
	}
}
	`, realm, realmDisplayName, realmDisplayNameHtml, rpName, rpId, arrayOfStringsForTerraformResource(signatureAlgorithms), attestationConveyancePreference, authenticatorAttachment, avoidSameAuthenticatorRegister, requireResidentKey, userVerificationRequirement)
}

func testKeycloakRealm_basicInternalId(realm, internalId string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        		= "%s"
	enabled     	 	= true

	internal_id = "%s"
}
	`, realm, internalId)
}
