package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakLdapUserFederation_basic(t *testing.T) {
	t.Parallel()
	ldapName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserFederation_basic(ldapName),
				Check:  testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap"),
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_import(t *testing.T) {
	t.Parallel()
	ldapName := acctest.RandomWithPrefix("tf-acc")

	bindCredentialForImport := "admin"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserFederation_basic(ldapName),
				Check:  testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap"),
			},
			{
				ResourceName:      "keycloak_ldap_user_federation.openldap",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getLdapUserFederationImportId("keycloak_ldap_user_federation.openldap", bindCredentialForImport),
			},
			{
				Config: testKeycloakLdapUserFederation_noAuth(ldapName),
				Check:  testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap_no_auth"),
			},
			{
				ResourceName:        "keycloak_ldap_user_federation.openldap_no_auth",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: testAccRealmUserFederation.Realm + "/",
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_createAfterManualDestroy(t *testing.T) {
	t.Parallel()
	var ldap = &keycloak.LdapUserFederation{}

	ldapName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserFederation_basic(ldapName),
				Check:  testAccCheckKeycloakLdapUserFederationFetch("keycloak_ldap_user_federation.openldap", ldap),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteLdapUserFederation(testCtx, ldap.RealmId, ldap.Id)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakLdapUserFederation_basic(ldapName),
				Check:  testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap"),
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_basicUpdateRealm(t *testing.T) {
	t.Parallel()
	ldapName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserFederation_basic(ldapName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap"),
					resource.TestCheckResourceAttr("keycloak_ldap_user_federation.openldap", "realm_id", testAccRealmUserFederation.Realm),
				),
			},
			{
				Config: testKeycloakLdapUserFederation_basic(ldapName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap"),
					resource.TestCheckResourceAttr("keycloak_ldap_user_federation.openldap", "realm_id", testAccRealmUserFederation.Realm),
				),
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_deleteDefaultMappers(t *testing.T) {
	t.Parallel()
	ldapName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserFederation_deleteDefaultMappers(ldapName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap"),
					testAccCheckKeycloakLdapUserFederationHasNoDefaultMappers("keycloak_ldap_user_federation.openldap"),
				),
			},
		},
	})
}

func generateRandomLdapKerberos(enabled bool) *keycloak.LdapUserFederation {
	connectionTimeout, _ := keycloak.GetDurationStringFromMilliseconds(strconv.Itoa(acctest.RandIntRange(1, 3600) * 1000))
	readTimeout, _ := keycloak.GetDurationStringFromMilliseconds(strconv.Itoa(acctest.RandIntRange(1, 3600) * 1000))

	evictionDay := acctest.RandIntRange(0, 6)
	evictionHour := acctest.RandIntRange(0, 23)
	evictionMinute := acctest.RandIntRange(0, 59)

	return &keycloak.LdapUserFederation{
		RealmId:                              testAccRealmUserFederation.Realm,
		Name:                                 "terraform-" + acctest.RandString(10),
		Enabled:                              enabled,
		UsernameLDAPAttribute:                acctest.RandString(10),
		UuidLDAPAttribute:                    acctest.RandString(10),
		UserObjectClasses:                    []string{acctest.RandString(10), acctest.RandString(10), acctest.RandString(10)},
		ConnectionUrl:                        "ldap://" + acctest.RandString(10),
		UsersDn:                              acctest.RandString(10),
		BindDn:                               acctest.RandString(10),
		BindCredential:                       acctest.RandString(10),
		SearchScope:                          randomStringInSlice([]string{"ONE_LEVEL", "SUBTREE"}),
		StartTls:                             true,
		UsePasswordModifyExtendedOp:          true,
		ValidatePasswordPolicy:               true,
		UseTruststoreSpi:                     randomStringInSlice([]string{"ALWAYS", "ONLY_FOR_LDAPS", "NEVER"}),
		ConnectionTimeout:                    connectionTimeout,
		ReadTimeout:                          readTimeout,
		Pagination:                           true,
		BatchSizeForSync:                     acctest.RandIntRange(50, 10000),
		FullSyncPeriod:                       acctest.RandIntRange(1, 3600),
		ChangedSyncPeriod:                    acctest.RandIntRange(1, 3600),
		CachePolicy:                          randomStringInSlice([]string{"DEFAULT", "EVICT_DAILY", "EVICT_WEEKLY", "MAX_LIFESPAN", "NO_CACHE"}),
		ServerPrincipal:                      acctest.RandString(10),
		UseKerberosForPasswordAuthentication: randomBool(),
		AllowKerberosAuthentication:          true,
		KeyTab:                               acctest.RandString(10),
		KerberosRealm:                        acctest.RandString(10),
		MaxLifespan:                          randomStringInSlice([]string{"1h", "2h", "3h"}),
		EvictionDay:                          &evictionDay,
		EvictionHour:                         &evictionHour,
		EvictionMinute:                       &evictionMinute,
		EditMode:                             "WRITABLE",
	}
}

func checkMatchingNestedKey(resourcePath string, blockName string, fieldInBlock string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("Could not find resource %s", resourcePath)
		}

		matchExpression := fmt.Sprintf(`%s\.\d\.mappings\.\d+\.%s`, blockName, fieldInBlock)

		for k, v := range r.Primary.Attributes {
			if isMatch, _ := regexp.Match(matchExpression, []byte(k)); isMatch {
				if v == value {
					return nil
				}

				return fmt.Errorf("Value for attribute %s.%s does match: %s != %s", blockName, fieldInBlock, v, value)
			}
		}

		return nil
	}
}

func TestAccKeycloakLdapUserFederation_basicUpdateKerberosSettings(t *testing.T) {
	t.Parallel()
	firstLdap := generateRandomLdapKerberos(true)
	secondLdap := generateRandomLdapKerberos(false)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserFederation_basicFromInterface(firstLdap),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap"),
					resource.TestCheckResourceAttr("keycloak_ldap_user_federation.openldap", "realm_id", firstLdap.RealmId),
					checkMatchingNestedKey("keycloak_ldap_user_federation.openldap", "kerberos", "kerberos_realm", firstLdap.KerberosRealm),
					checkMatchingNestedKey("keycloak_ldap_user_federation.openldap", "kerberos", "server_principal", firstLdap.ServerPrincipal),
					checkMatchingNestedKey("keycloak_ldap_user_federation.openldap", "kerberos", "use_kerberos_for_password_authentication", strconv.FormatBool(firstLdap.UseKerberosForPasswordAuthentication)),
					checkMatchingNestedKey("keycloak_ldap_user_federation.openldap", "kerberos", "key_tab", firstLdap.KeyTab),
				),
			},
			{
				Config: testKeycloakLdapUserFederation_basicFromInterface(secondLdap),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap"),
					resource.TestCheckResourceAttr("keycloak_ldap_user_federation.openldap", "realm_id", secondLdap.RealmId),
					checkMatchingNestedKey("keycloak_ldap_user_federation.openldap", "kerberos", "kerberos_realm", secondLdap.KerberosRealm),
					checkMatchingNestedKey("keycloak_ldap_user_federation.openldap", "kerberos", "server_principal", secondLdap.ServerPrincipal),
					checkMatchingNestedKey("keycloak_ldap_user_federation.openldap", "kerberos", "use_kerberos_for_password_authentication", strconv.FormatBool(secondLdap.UseKerberosForPasswordAuthentication)),
					checkMatchingNestedKey("keycloak_ldap_user_federation.openldap", "kerberos", "key_tab", secondLdap.KeyTab),
				),
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_basicUpdateAll(t *testing.T) {
	t.Parallel()
	firstEnabled := randomBool()
	firstStartTls := randomBool()
	firstUsePasswordModifyExtendedOp := randomBool()
	firstValidatePasswordPolicy := randomBool()
	firstPagination := randomBool()

	firstTrustEmail := randomBool()
	secondTrustEmail := !firstTrustEmail

	firstConnectionTimeout, _ := keycloak.GetDurationStringFromMilliseconds(strconv.Itoa(acctest.RandIntRange(1, 3600) * 1000))
	secondConnectionTimeout, _ := keycloak.GetDurationStringFromMilliseconds(strconv.Itoa(acctest.RandIntRange(1, 3600) * 1000))
	firstReadTimeout, _ := keycloak.GetDurationStringFromMilliseconds(strconv.Itoa(acctest.RandIntRange(1, 3600) * 1000))
	secondReadTimeout, _ := keycloak.GetDurationStringFromMilliseconds(strconv.Itoa(acctest.RandIntRange(1, 3600) * 1000))

	evictionDay := acctest.RandIntRange(0, 6)
	evictionHour := acctest.RandIntRange(0, 23)
	evictionMinute := acctest.RandIntRange(0, 59)

	firstLdap := &keycloak.LdapUserFederation{
		Name:                                 "terraform-" + acctest.RandString(10),
		Enabled:                              firstEnabled,
		UsernameLDAPAttribute:                acctest.RandString(10),
		UuidLDAPAttribute:                    acctest.RandString(10),
		UserObjectClasses:                    []string{acctest.RandString(10), acctest.RandString(10), acctest.RandString(10)},
		ConnectionUrl:                        "ldap://" + acctest.RandString(10),
		UsersDn:                              acctest.RandString(10),
		BindDn:                               acctest.RandString(10),
		BindCredential:                       acctest.RandString(10),
		SearchScope:                          randomStringInSlice([]string{"ONE_LEVEL", "SUBTREE"}),
		StartTls:                             firstStartTls,
		UsePasswordModifyExtendedOp:          firstUsePasswordModifyExtendedOp,
		ValidatePasswordPolicy:               firstValidatePasswordPolicy,
		TrustEmail:                           firstTrustEmail,
		UseTruststoreSpi:                     randomStringInSlice([]string{"ALWAYS", "ONLY_FOR_LDAPS", "NEVER"}),
		ConnectionTimeout:                    firstConnectionTimeout,
		ReadTimeout:                          firstReadTimeout,
		Pagination:                           firstPagination,
		BatchSizeForSync:                     acctest.RandIntRange(50, 10000),
		FullSyncPeriod:                       acctest.RandIntRange(1, 3600),
		ChangedSyncPeriod:                    acctest.RandIntRange(1, 3600),
		CachePolicy:                          randomStringInSlice([]string{"DEFAULT", "EVICT_DAILY", "EVICT_WEEKLY", "MAX_LIFESPAN", "NO_CACHE"}),
		ServerPrincipal:                      acctest.RandString(10),
		UseKerberosForPasswordAuthentication: randomBool(),
		AllowKerberosAuthentication:          randomBool(),
		KeyTab:                               acctest.RandString(10),
		KerberosRealm:                        acctest.RandString(10),
		MaxLifespan:                          randomStringInSlice([]string{"1h", "2h", "3h"}),
		EvictionDay:                          &evictionDay,
		EvictionHour:                         &evictionHour,
		EvictionMinute:                       &evictionMinute,
		EditMode:                             "WRITABLE",
	}

	evictionDay = acctest.RandIntRange(0, 6)
	evictionHour = acctest.RandIntRange(0, 23)
	evictionMinute = acctest.RandIntRange(0, 59)

	secondLdap := &keycloak.LdapUserFederation{
		Name:                                 "terraform-" + acctest.RandString(10),
		Enabled:                              !firstEnabled,
		UsernameLDAPAttribute:                acctest.RandString(10),
		UuidLDAPAttribute:                    acctest.RandString(10),
		UserObjectClasses:                    []string{acctest.RandString(10)},
		ConnectionUrl:                        "ldap://" + acctest.RandString(10),
		UsersDn:                              acctest.RandString(10),
		BindDn:                               acctest.RandString(10),
		BindCredential:                       acctest.RandString(10),
		SearchScope:                          randomStringInSlice([]string{"ONE_LEVEL", "SUBTREE"}),
		StartTls:                             !firstStartTls,
		UsePasswordModifyExtendedOp:          !firstUsePasswordModifyExtendedOp,
		ValidatePasswordPolicy:               !firstValidatePasswordPolicy,
		TrustEmail:                           secondTrustEmail,
		UseTruststoreSpi:                     randomStringInSlice([]string{"ALWAYS", "ONLY_FOR_LDAPS", "NEVER"}),
		ConnectionTimeout:                    secondConnectionTimeout,
		ReadTimeout:                          secondReadTimeout,
		Pagination:                           !firstPagination,
		BatchSizeForSync:                     acctest.RandIntRange(50, 10000),
		FullSyncPeriod:                       acctest.RandIntRange(1, 3600),
		ChangedSyncPeriod:                    acctest.RandIntRange(1, 3600),
		CachePolicy:                          randomStringInSlice([]string{"DEFAULT", "EVICT_DAILY", "EVICT_WEEKLY", "MAX_LIFESPAN", "NO_CACHE"}),
		ServerPrincipal:                      acctest.RandString(10),
		UseKerberosForPasswordAuthentication: randomBool(),
		AllowKerberosAuthentication:          randomBool(),
		KeyTab:                               acctest.RandString(10),
		KerberosRealm:                        acctest.RandString(10),
		MaxLifespan:                          randomStringInSlice([]string{"1h", "2h", "3h"}),
		EvictionDay:                          &evictionDay,
		EvictionHour:                         &evictionHour,
		EvictionMinute:                       &evictionMinute,
		EditMode:                             "WRITABLE",
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserFederation_basicFromInterface(firstLdap),
				Check:  testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap"),
			},
			{
				Config: testKeycloakLdapUserFederation_basicFromInterface(secondLdap),
				Check:  testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap"),
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_unsetTimeoutDurationStrings(t *testing.T) {
	t.Parallel()
	ldapName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserFederation_basicWithTimeouts(ldapName),
				Check:  testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap"),
			},
			{
				Config: testKeycloakLdapUserFederation_basic(ldapName),
				Check:  testAccCheckKeycloakLdapUserFederationExists("keycloak_ldap_user_federation.openldap"),
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_editModeValidation(t *testing.T) {
	t.Parallel()
	ldapName := acctest.RandomWithPrefix("tf-acc")
	editMode := randomStringInSlice(keycloakLdapUserFederationEditModes)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapUserFederation_basicWithAttrValidation("edit_mode", ldapName, acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected edit_mode to be one of .+ got .+"),
			},
			{
				Config: testKeycloakLdapUserFederation_basicWithAttrValidation("edit_mode", ldapName, editMode),
				Check:  resource.TestCheckResourceAttr("keycloak_ldap_user_federation.openldap", "edit_mode", editMode),
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_vendorValidation(t *testing.T) {
	t.Parallel()
	ldapName := acctest.RandomWithPrefix("tf-acc")
	vendor := randomStringInSlice(keycloakLdapUserFederationVendors)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapUserFederation_basicWithAttrValidation("vendor", ldapName, acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected vendor to be one of .+ got .+"),
			},
			{
				Config: testKeycloakLdapUserFederation_basicWithAttrValidation("vendor", ldapName, vendor),
				Check:  resource.TestCheckResourceAttr("keycloak_ldap_user_federation.openldap", "vendor", vendor),
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_searchScopeValidation(t *testing.T) {
	t.Parallel()
	ldapName := acctest.RandomWithPrefix("tf-acc")
	searchScope := randomStringInSlice(keycloakLdapUserFederationSearchScopes)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapUserFederation_basicWithAttrValidation("search_scope", ldapName, acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected search_scope to be one of .+ got .+"),
			},
			{
				Config: testKeycloakLdapUserFederation_basicWithAttrValidation("search_scope", ldapName, searchScope),
				Check:  resource.TestCheckResourceAttr("keycloak_ldap_user_federation.openldap", "search_scope", searchScope),
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_useTrustStoreValidation(t *testing.T) {
	t.Parallel()
	ldapName := acctest.RandomWithPrefix("tf-acc")
	useTrustStore := randomStringInSlice(keycloakLdapUserFederationTruststoreSpiSettings)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapUserFederation_basicWithAttrValidation("use_truststore_spi", ldapName, acctest.RandString(10)),
				ExpectError: regexp.MustCompile("expected use_truststore_spi to be one of .+ got .+"),
			},
			{
				Config: testKeycloakLdapUserFederation_basicWithAttrValidation("use_truststore_spi", ldapName, useTrustStore),
				Check:  resource.TestCheckResourceAttr("keycloak_ldap_user_federation.openldap", "use_truststore_spi", useTrustStore),
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_bindValidation(t *testing.T) {
	t.Parallel()
	ldapName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapUserFederation_noBindCredentialValidation(ldapName),
				ExpectError: regexp.MustCompile("validation error: authentication requires both BindDN and BindCredential to be set"),
			},
			{
				Config:      testKeycloakLdapUserFederation_nobindDnValidation(ldapName),
				ExpectError: regexp.MustCompile("validation error: authentication requires both BindDN and BindCredential to be set"),
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_syncPeriodValidation(t *testing.T) {
	t.Parallel()
	ldapName := acctest.RandomWithPrefix("tf-acc")

	validSyncPeriod := acctest.RandIntRange(1, 3600)
	invalidNegativeSyncPeriod := -acctest.RandIntRange(1, 3600)
	invalidZeroSyncPeriod := 0

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakLdapUserFederation_basicWithSyncPeriod(ldapName, validSyncPeriod, invalidNegativeSyncPeriod),
				ExpectError: regexp.MustCompile(`expected .+ to be either -1 \(disabled\), or greater than zero`),
			},
			{
				Config:      testKeycloakLdapUserFederation_basicWithSyncPeriod(ldapName, invalidNegativeSyncPeriod, validSyncPeriod),
				ExpectError: regexp.MustCompile(`expected .+ to be either -1 \(disabled\), or greater than zero`),
			},
			{
				Config:      testKeycloakLdapUserFederation_basicWithSyncPeriod(ldapName, validSyncPeriod, invalidZeroSyncPeriod),
				ExpectError: regexp.MustCompile(`expected .+ to be either -1 \(disabled\), or greater than zero`),
			},
			{
				Config:      testKeycloakLdapUserFederation_basicWithSyncPeriod(ldapName, invalidZeroSyncPeriod, validSyncPeriod),
				ExpectError: regexp.MustCompile(`expected .+ to be either -1 \(disabled\), or greater than zero`),
			},
			{
				Config: testKeycloakLdapUserFederation_basicWithSyncPeriod(ldapName, validSyncPeriod, validSyncPeriod),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("keycloak_ldap_user_federation.openldap", "full_sync_period", strconv.Itoa(validSyncPeriod)),
					resource.TestCheckResourceAttr("keycloak_ldap_user_federation.openldap", "changed_sync_period", strconv.Itoa(validSyncPeriod)),
				),
			},
		},
	})
}

func TestAccKeycloakLdapUserFederation_bindCredential(t *testing.T) {
	t.Parallel()
	ldapName := acctest.RandomWithPrefix("tf-acc")
	firstBindCredential := acctest.RandomWithPrefix("tf-acc")
	secondBindCredential := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakLdapUserFederationDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakLdapUserFederation_bindCredential(ldapName, firstBindCredential),
				Check:  resource.TestCheckResourceAttr("keycloak_ldap_user_federation.openldap", "bind_credential", firstBindCredential),
			},
			{
				Config: testKeycloakLdapUserFederation_bindCredential(ldapName, secondBindCredential),
				Check:  resource.TestCheckResourceAttr("keycloak_ldap_user_federation.openldap", "bind_credential", secondBindCredential),
			},
		},
	})
}

func testAccCheckKeycloakLdapUserFederationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getLdapUserFederationFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakLdapUserFederationHasNoDefaultMappers(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ldap, err := getLdapUserFederationFromState(s, resourceName)
		if err != nil {
			return err
		}

		mappers, err := keycloakClient.GetLdapUserFederationMappers(testCtx, ldap.RealmId, ldap.Id)
		if err != nil {
			return err
		}

		if len(*mappers) != 0 {
			return fmt.Errorf("expected ldap user federation to have zero mappers, found %d", len(*mappers))
		}

		return nil
	}
}

func testAccCheckKeycloakLdapUserFederationFetch(resourceName string, ldap *keycloak.LdapUserFederation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedLdap, err := getLdapUserFederationFromState(s, resourceName)
		if err != nil {
			return err
		}

		ldap.Id = fetchedLdap.Id
		ldap.RealmId = fetchedLdap.RealmId

		return nil
	}
}

func testAccCheckKeycloakLdapUserFederationDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_ldap_user_federation" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm_id"]

			ldap, _ := keycloakClient.GetLdapUserFederation(testCtx, realm, id)
			if ldap != nil {
				return fmt.Errorf("ldap config with id %s still exists", id)
			}
		}

		return nil
	}
}

func getLdapUserFederationFromState(s *terraform.State, resourceName string) (*keycloak.LdapUserFederation, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	id := rs.Primary.ID
	realm := rs.Primary.Attributes["realm_id"]

	ldap, err := keycloakClient.GetLdapUserFederation(testCtx, realm, id)
	if err != nil {
		return nil, fmt.Errorf("error getting ldap config with id %s: %s", id, err)
	}

	return ldap, nil
}

func getLdapUserFederationImportId(resourceName, bindCredential string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		realmId := rs.Primary.Attributes["realm_id"]

		return fmt.Sprintf("%s/%s/%s", realmId, id, bindCredential), nil
	}
}

func testKeycloakLdapUserFederation_basic(ldap string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"
}
	`, testAccRealmUserFederation.Realm, ldap)
}

func testKeycloakLdapUserFederation_basicFromInterface(ldap *keycloak.LdapUserFederation) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                     = "%s"
	realm_id                 = data.keycloak_realm.realm.id

	enabled                  = %t

	username_ldap_attribute  = "%s"
	rdn_ldap_attribute       = "%s"
	uuid_ldap_attribute      = "%s"
	user_object_classes      = %s
	connection_url           = "%s"
	users_dn                 = "%s"
	bind_dn                  = "%s"
	bind_credential          = "%s"
	search_scope             = "%s"

	edit_mode                       = "%s"
	start_tls                       = %t
	use_password_modify_extended_op = %t
	validate_password_policy        = %t
	trust_email                     = %t
	use_truststore_spi              = "%s"
	connection_timeout              = "%s"
	read_timeout                    = "%s"
	pagination                      = %t

	batch_size_for_sync      = %d
	full_sync_period         = %d
	changed_sync_period      = %d

	kerberos {
		server_principal                         = "%s"
		use_kerberos_for_password_authentication = %t
		key_tab                                  = "%s"
		kerberos_realm                           = "%s"
	}

	cache {
		policy               = "%s"
		max_lifespan         = "%s"
		eviction_day         = %d
		eviction_hour        = %d
		eviction_minute      = %d
	}
}
	`, testAccRealmUserFederation.Realm, ldap.Name, ldap.Enabled, ldap.UsernameLDAPAttribute, ldap.RdnLDAPAttribute, ldap.UuidLDAPAttribute, arrayOfStringsForTerraformResource(ldap.UserObjectClasses), ldap.ConnectionUrl, ldap.UsersDn, ldap.BindDn, ldap.BindCredential, ldap.SearchScope, ldap.EditMode, ldap.StartTls, ldap.UsePasswordModifyExtendedOp, ldap.ValidatePasswordPolicy, ldap.TrustEmail, ldap.UseTruststoreSpi, ldap.ConnectionTimeout, ldap.ReadTimeout, ldap.Pagination, ldap.BatchSizeForSync, ldap.FullSyncPeriod, ldap.ChangedSyncPeriod, ldap.ServerPrincipal, ldap.UseKerberosForPasswordAuthentication, ldap.KeyTab, ldap.KerberosRealm, ldap.CachePolicy, ldap.MaxLifespan, *ldap.EvictionDay, *ldap.EvictionHour, *ldap.EvictionMinute)
}

func testKeycloakLdapUserFederation_basicWithAttrValidation(attr, ldap, val string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	%s                      = "%s"

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"
}
	`, testAccRealmUserFederation.Realm, ldap, attr, val)
}

func testKeycloakLdapUserFederation_nobindDnValidation(ldap string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	bind_credential         = "admin"

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
}
	`, testAccRealmUserFederation.Realm, ldap)
}

func testKeycloakLdapUserFederation_noBindCredentialValidation(ldap string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	bind_dn                 = "cn=admin,dc=example,dc=org"

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
}
	`, testAccRealmUserFederation.Realm, ldap)
}

func testKeycloakLdapUserFederation_basicWithSyncPeriod(ldap string, fullSyncPeriod, changedSyncPeriod int) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"

	full_sync_period        = %d
	changed_sync_period     = %d
}
	`, testAccRealmUserFederation.Realm, ldap, fullSyncPeriod, changedSyncPeriod)
}

func testKeycloakLdapUserFederation_basicWithTimeouts(ldap string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"

	connection_timeout      = "10s"
	read_timeout            = "5s"
}
	`, testAccRealmUserFederation.Realm, ldap)
}

func testKeycloakLdapUserFederation_bindCredential(ldap, bindCredential string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "%s"
}
	`, testAccRealmUserFederation.Realm, ldap, bindCredential)
}

func testKeycloakLdapUserFederation_noAuth(ldap string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap_no_auth" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
}
	`, testAccRealmUserFederation.Realm, ldap)
}

func testKeycloakLdapUserFederation_deleteDefaultMappers(ldap string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_ldap_user_federation" "openldap" {
	name                    = "%s"
	realm_id                = data.keycloak_realm.realm.id

	enabled                 = true

	username_ldap_attribute = "cn"
	rdn_ldap_attribute      = "cn"
	uuid_ldap_attribute     = "entryDN"
	user_object_classes     = [
		"simpleSecurityObject",
		"organizationalRole"
	]
	connection_url          = "ldap://openldap"
	users_dn                = "dc=example,dc=org"
	bind_dn                 = "cn=admin,dc=example,dc=org"
	bind_credential         = "admin"

	delete_default_mappers = true
}
	`, testAccRealmUserFederation.Realm, ldap)
}
