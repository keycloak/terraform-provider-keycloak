terraform {
  required_providers {
    keycloak = {
      source  = "terraform.local/keycloak/keycloak"
      version = ">= 5.0.0"
    }
  }
}

provider "keycloak" {
  client_id     = "terraform"
  client_secret = "884e0f95-0f42-4a63-9b1f-94274655669e"
  url           = "http://localhost:8080"
  additional_headers = {
    foo = "bar"
  }
}

resource "keycloak_realm" "test" {
  realm             = "test"
  enabled           = true
  display_name      = "foo"
  display_name_html = "<b>foo</b>"

  smtp_server {
    host                  = "mysmtphost.com"
    port                  = 25
    from_display_name     = "Tom"
    from                  = "tom@myhost.com"
    reply_to_display_name = "Tom"
    reply_to              = "tom@myhost.com"
    ssl                   = true
    starttls              = true
    envelope_from         = "nottom@myhost.com"

    auth {
      username = "tom"
      password = "tom"
    }
  }

  account_theme        = "base"
  access_code_lifespan = "30m"

  internationalization {
    supported_locales = [
      "en",
      "de",
      "es",
    ]

    default_locale = "en"
  }

  security_defenses {
    headers {
      x_frame_options                     = "DENY"
      content_security_policy             = "frame-src 'self'; frame-ancestors 'self'; object-src 'none';"
      content_security_policy_report_only = ""
      x_content_type_options              = "nosniff"
      x_robots_tag                        = "none"
      x_xss_protection                    = "1; mode=block"
      strict_transport_security           = "max-age=31536000; includeSubDomains"
    }

    brute_force_detection {
      permanent_lockout                = false
      max_login_failures               = 31
      wait_increment_seconds           = 61
      quick_login_check_milli_seconds  = 1000
      minimum_quick_login_wait_seconds = 120
      max_failure_wait_seconds         = 900
      failure_reset_time_seconds       = 43200
    }
  }

  ssl_required    = "external"
  password_policy = "upperCase(1) and length(8) and forceExpiredPasswordChange(365) and notUsername"

  attributes = {
    mycustomAttribute  = "myCustomValue"
    userProfileEnabled = true
  }

  web_authn_policy {
    relying_party_entity_name = "Example"
    relying_party_id          = "keycloak.example.com"
    signature_algorithms = [
      "ES256",
      "RS256"
    ]
  }

  web_authn_passwordless_policy {
    relying_party_entity_name = "Example"
    relying_party_id          = "keycloak.example.com"
    signature_algorithms = [
      "ES256",
      "RS256"
    ]
  }
}

resource "keycloak_realm_localization" "test_translation" {
  realm_id = keycloak_realm.test.id
  locale   = "en"
  texts = {
    "test" : "translation"
  }
}

resource "keycloak_required_action" "custom-terms-and-conditions" {
  realm_id       = keycloak_realm.test.realm
  alias          = "TERMS_AND_CONDITIONS"
  default_action = true
  enabled        = true
  name           = "Custom Terms and Conditions"
}

resource "keycloak_required_action" "update-password" {
  realm_id       = keycloak_realm.test.realm
  alias          = "UPDATE_PASSWORD"
  default_action = true
  enabled        = true
  name           = "Update Password"

  config = {
    max_auth_age = "600"
  }
}

resource "keycloak_required_action" "custom-configured_totp" {
  realm_id       = keycloak_realm.test.realm
  alias          = "CONFIGURE_TOTP"
  default_action = true
  enabled        = true
  name           = "Custom configure totp"
  priority       = keycloak_required_action.custom-terms-and-conditions.priority + 15
}

resource "keycloak_required_action" "required_action" {
  realm_id = keycloak_realm.test.realm
  alias    = "webauthn-register"
  enabled  = true
  name     = "Webauthn Register"
}

resource "keycloak_group" "foo" {
  realm_id = keycloak_realm.test.id
  name     = "foo"
}

resource "keycloak_group" "nested_foo" {
  realm_id  = keycloak_realm.test.id
  parent_id = keycloak_group.foo.id
  name      = "nested-foo"
}

resource "keycloak_group" "bar" {
  realm_id = keycloak_realm.test.id
  name     = "bar"
}

resource "keycloak_user" "user" {
  realm_id = keycloak_realm.test.id
  username = "test-user"

  email      = "test-user@fakedomain.com"
  first_name = "Testy"
  last_name  = "Tester"
}

resource "keycloak_user" "another_user" {
  realm_id = keycloak_realm.test.id
  username = "another-test-user"

  email      = "another-test-user@fakedomain.com"
  first_name = "Testy"
  last_name  = "Tester"
}

resource "keycloak_user" "user_with_password" {
  realm_id = keycloak_realm.test.id
  username = "user-with-password"

  email      = "user-with-password@fakedomain.com"
  first_name = "Testy"
  last_name  = "Tester"

  initial_password {
    value     = "My password"
    temporary = false
  }
}

resource "keycloak_group_memberships" "foo_members" {
  realm_id = keycloak_realm.test.id
  group_id = keycloak_group.foo.id

  members = [
    keycloak_user.user.username,
    keycloak_user.another_user.username,
  ]
}

resource "keycloak_group" "baz" {
  realm_id = keycloak_realm.test.id
  name     = "baz"
}

resource "keycloak_default_groups" "default" {
  realm_id = keycloak_realm.test.id
  group_ids = [
    keycloak_group.baz.id
  ]
}

resource "keycloak_openid_client" "test_client" {
  client_id   = "test-openid-client"
  name        = "test-openid-client"
  realm_id    = keycloak_realm.test.id
  description = "a test openid client"

  standard_flow_enabled    = true
  service_accounts_enabled = true

  access_type = "CONFIDENTIAL"

  valid_redirect_uris = [
    "http://localhost:5555/callback",
  ]
  valid_post_logout_redirect_uris = [
    "http://localhost:5555/post-logout",
    "http://localhost:5555/post-logout3",
  ]

  client_secret = "secret"

  pkce_code_challenge_method = "plain"

  login_theme = "keycloak"

  backchannel_logout_url                     = "http://localhost:3333/backchannel"
  backchannel_logout_session_required        = true
  backchannel_logout_revoke_offline_sessions = true

  extra_config = {
    customAttribute = "a test custom value"
  }
}

resource "keycloak_openid_client_scope" "test_default_client_scope" {
  name     = "test-default-client-scope"
  realm_id = keycloak_realm.test.id

  description         = "test"
  consent_screen_text = "hello"
}

resource "keycloak_openid_client_scope" "test_optional_client_scope" {
  name     = "test-optional-client-scope"
  realm_id = keycloak_realm.test.id

  description         = "test"
  consent_screen_text = "hello"
}

resource "keycloak_openid_client_default_scopes" "default_client_scopes" {
  realm_id  = keycloak_realm.test.id
  client_id = keycloak_openid_client.test_client.id

  default_scopes = [
    "profile",
    "email",
    "roles",
    "web-origins",
    keycloak_openid_client_scope.test_default_client_scope.name,
  ]
}

resource "keycloak_openid_client_optional_scopes" "optional_client_scopes" {
  realm_id  = keycloak_realm.test.id
  client_id = keycloak_openid_client.test_client.id

  optional_scopes = [
    "address",
    "phone",
    "offline_access",
    "microprofile-jwt",
    keycloak_openid_client_scope.test_optional_client_scope.name,
  ]
}

resource "keycloak_ldap_user_federation" "openldap" {
  name     = "openldap"
  realm_id = keycloak_realm.test.id

  enabled        = true
  import_enabled = false

  username_ldap_attribute = "cn"
  rdn_ldap_attribute      = "cn"
  uuid_ldap_attribute     = "entryDN"

  user_object_classes = [
    "simpleSecurityObject",
    "organizationalRole",
  ]

  connection_url  = "ldap://openldap"
  users_dn        = "dc=example,dc=org"
  bind_dn         = "cn=admin,dc=example,dc=org"
  bind_credential = "adminpassword"

  connection_timeout = "5s"
  read_timeout       = "10s"

  kerberos {
    server_principal                         = "HTTP/keycloak.local@FOO.LOCAL"
    use_kerberos_for_password_authentication = false
    key_tab                                  = "/etc/keycloak.keytab"
    kerberos_realm                           = "FOO.LOCAL"
  }

  cache {
    policy = "NO_CACHE"
  }
}


resource "keycloak_ldap_user_federation" "openldap_no_default_mappers" {
  name     = "openldap-no-default-mappers"
  realm_id = keycloak_realm.test.id

  enabled        = true
  import_enabled = false

  username_ldap_attribute = "cn"
  rdn_ldap_attribute      = "cn"
  uuid_ldap_attribute     = "entryDN"

  user_object_classes = [
    "simpleSecurityObject",
    "organizationalRole",
  ]

  connection_url  = "ldap://openldap"
  users_dn        = "dc=example,dc=org"
  bind_dn         = "cn=admin,dc=example,dc=org"
  bind_credential = "adminpassword"

  connection_timeout = "5s"
  read_timeout       = "10s"

  kerberos {
    server_principal                         = "HTTP/keycloak.local@FOO.LOCAL"
    use_kerberos_for_password_authentication = false
    key_tab                                  = "/etc/keycloak.keytab"
    kerberos_realm                           = "FOO.LOCAL"
  }

  cache {
    policy = "NO_CACHE"
  }

  delete_default_mappers = true
}

resource "keycloak_ldap_role_mapper" "ldap_role_mapper" {
  realm_id                = keycloak_realm.test.id
  ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id
  name                    = "role-mapper"

  ldap_roles_dn            = "dc=example,dc=org"
  role_name_ldap_attribute = "cn"
  role_object_classes = [
    "groupOfNames"
  ]
  membership_attribute_type      = "DN"
  membership_ldap_attribute      = "member"
  membership_user_ldap_attribute = "cn"
  user_roles_retrieve_strategy   = "GET_ROLES_FROM_USER_MEMBEROF_ATTRIBUTE"
  memberof_ldap_attribute        = "memberOf"
}

resource "keycloak_ldap_user_attribute_mapper" "description_attr_mapper" {
  name                    = "description-mapper"
  realm_id                = keycloak_ldap_user_federation.openldap.realm_id
  ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id

  user_model_attribute = "description"
  ldap_attribute       = "description"

  always_read_value_from_ldap = false
}

resource "keycloak_ldap_user_attribute_mapper" "default_attr_mapper" {
  name                    = "defaultval-mapper"
  realm_id                = keycloak_ldap_user_federation.openldap.realm_id
  ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id

  user_model_attribute = "defaultval"
  ldap_attribute       = "defaultval"

  always_read_value_from_ldap = false
  is_mandatory_in_ldap        = true
  attribute_default_value     = "testing"
}

resource "keycloak_ldap_group_mapper" "group_mapper" {
  name                    = "group mapper"
  realm_id                = keycloak_ldap_user_federation.openldap.realm_id
  ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id

  ldap_groups_dn            = "dc=example,dc=org"
  group_name_ldap_attribute = "cn"

  group_object_classes = [
    "groupOfNames",
  ]

  membership_attribute_type      = "DN"
  membership_ldap_attribute      = "member"
  membership_user_ldap_attribute = "cn"
  memberof_ldap_attribute        = "memberOf"
}

resource "keycloak_ldap_msad_user_account_control_mapper" "msad_uac_mapper" {
  name                    = "uac-mapper1"
  realm_id                = keycloak_ldap_user_federation.openldap.realm_id
  ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id
}

resource "keycloak_ldap_msad_lds_user_account_control_mapper" "msad_lds_uac_mapper" {
  name                    = "msad-lds-uac-mapper"
  realm_id                = keycloak_ldap_user_federation.openldap.realm_id
  ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id
}

resource "keycloak_ldap_full_name_mapper" "full_name_mapper" {
  name                    = "full-name-mapper"
  realm_id                = keycloak_ldap_user_federation.openldap.realm_id
  ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id

  ldap_full_name_attribute = "cn"
  read_only                = true
}

resource "keycloak_ldap_custom_mapper" "custom_mapper" {
  name                    = "custom-mapper"
  realm_id                = keycloak_ldap_user_federation.openldap.realm_id
  ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id

  provider_id   = "msad-user-account-control-mapper"
  provider_type = "org.keycloak.storage.ldap.mappers.LDAPStorageMapper"
}

resource "keycloak_ldap_custom_mapper" "custom_mapper_with_config" {
  name                    = "custom-mapper-with-config"
  realm_id                = keycloak_ldap_user_federation.openldap.realm_id
  ldap_user_federation_id = keycloak_ldap_user_federation.openldap.id

  provider_id   = "user-attribute-ldap-mapper"
  provider_type = "org.keycloak.storage.ldap.mappers.LDAPStorageMapper"
  config = {
    "user.model.attribute" = "username"
    "ldap.attribute"       = "cn"
  }
}


resource "keycloak_custom_user_federation" "custom" {
  name        = "custom1"
  realm_id    = "master"
  provider_id = "custom"

  enabled = true
}

resource "keycloak_openid_user_attribute_protocol_mapper" "map_user_attributes_client" {
  name           = "tf-test-open-id-user-attribute-protocol-mapper-client"
  realm_id       = keycloak_realm.test.id
  client_id      = keycloak_openid_client.test_client.id
  user_attribute = "description"
  claim_name     = "description"
}

resource "keycloak_openid_user_attribute_protocol_mapper" "map_user_permissions_attributes_client" {
  name           = "tf-test-open-id-user-multivalue-attribute-protocol-mapper-client"
  realm_id       = keycloak_realm.test.id
  client_id      = keycloak_openid_client.test_client.id
  user_attribute = "permissions"
  claim_name     = "permissions"
  multivalued    = true
}


resource "keycloak_openid_user_attribute_protocol_mapper" "map_user_attributes_client_scope" {
  name            = "tf-test-open-id-user-attribute-protocol-mapper-client-scope"
  realm_id        = keycloak_realm.test.id
  client_scope_id = keycloak_openid_client_scope.test_default_client_scope.id
  user_attribute  = "foo2"
  claim_name      = "bar2"
}

resource "keycloak_openid_group_membership_protocol_mapper" "map_group_memberships_client" {
  name       = "tf-test-open-id-group-membership-protocol-mapper-client"
  realm_id   = keycloak_realm.test.id
  client_id  = keycloak_openid_client.test_client.id
  claim_name = "bar"
}

resource "keycloak_openid_group_membership_protocol_mapper" "map_group_memberships_client_scope" {
  name            = "tf-test-open-id-group-membership-protocol-mapper-client-scope"
  realm_id        = keycloak_realm.test.id
  client_scope_id = keycloak_openid_client_scope.test_optional_client_scope.id
  claim_name      = "bar2"
}

resource "keycloak_openid_full_name_protocol_mapper" "map_full_names_client" {
  name      = "tf-test-open-id-full-name-protocol-mapper-client"
  realm_id  = keycloak_realm.test.id
  client_id = keycloak_openid_client.test_client.id
}

resource "keycloak_openid_full_name_protocol_mapper" "map_full_names_client_scope" {
  name            = "tf-test-open-id-full-name-protocol-mapper-client-scope"
  realm_id        = keycloak_realm.test.id
  client_scope_id = keycloak_openid_client_scope.test_default_client_scope.id
}

resource "keycloak_openid_user_property_protocol_mapper" "map_user_properties_client" {
  name          = "tf-test-open-id-user-property-protocol-mapper-client"
  realm_id      = keycloak_realm.test.id
  client_id     = keycloak_openid_client.test_client.id
  user_property = "foo"
  claim_name    = "bar"
}

resource "keycloak_openid_user_property_protocol_mapper" "map_user_properties_client_scope" {
  name            = "tf-test-open-id-user-property-protocol-mapper-client-scope"
  realm_id        = keycloak_realm.test.id
  client_scope_id = keycloak_openid_client_scope.test_optional_client_scope.id
  user_property   = "foo2"
  claim_name      = "bar2"
}

resource "keycloak_openid_hardcoded_claim_protocol_mapper" "hardcoded_claim_client" {
  name      = "tf-test-open-id-hardcoded-claim-protocol-mapper-client"
  realm_id  = keycloak_realm.test.id
  client_id = keycloak_openid_client.test_client.id

  claim_name  = "foo"
  claim_value = "bar"
}

resource "keycloak_openid_hardcoded_claim_protocol_mapper" "hardcoded_claim_client_scope" {
  name            = "tf-test-open-id-hardcoded-claim-protocol-mapper-client-scope"
  realm_id        = keycloak_realm.test.id
  client_scope_id = keycloak_openid_client_scope.test_default_client_scope.id

  claim_name  = "foo"
  claim_value = "bar"
}

resource "keycloak_openid_user_realm_role_protocol_mapper" "user_realm_role_client" {
  name      = "tf-test-open-id-user-realm-role-claim-protocol-mapper-client"
  realm_id  = keycloak_realm.test.id
  client_id = keycloak_openid_client.test_client.id

  claim_name = "foo"
}

resource "keycloak_openid_user_realm_role_protocol_mapper" "user_realm_role_client_scope" {
  name            = "tf-test-open-id-user-realm-role-protocol-mapper-client-scope"
  realm_id        = keycloak_realm.test.id
  client_scope_id = keycloak_openid_client_scope.test_default_client_scope.id

  claim_name = "foo"
}

resource "keycloak_openid_user_client_role_protocol_mapper" "user_client_role_client" {
  name      = "tf-test-open-id-user-client-role-claim-protocol-mapper-client"
  realm_id  = keycloak_realm.test.id
  client_id = keycloak_openid_client.test_client.id

  claim_name  = "foo"
  multivalued = false

  client_id_for_role_mappings = keycloak_openid_client.bearer_only_client.client_id
  client_role_prefix          = "prefixValue"

  add_to_id_token     = true
  add_to_access_token = false
  add_to_userinfo     = false
}

resource "keycloak_openid_user_client_role_protocol_mapper" "user_client_role_client_scope" {
  name            = "tf-test-open-id-user-client-role-protocol-mapper-client-scope"
  realm_id        = keycloak_realm.test.id
  client_scope_id = keycloak_openid_client_scope.test_default_client_scope.id

  claim_name  = "foo"
  multivalued = false

  client_id_for_role_mappings = keycloak_openid_client.bearer_only_client.client_id
  client_role_prefix          = "prefixValue"

  add_to_id_token     = true
  add_to_access_token = false
  add_to_userinfo     = false
}

resource "keycloak_openid_user_session_note_protocol_mapper" "user_session_note_client" {
  name      = "tf-test-open-id-user-session-note-protocol-mapper-client"
  realm_id  = keycloak_realm.test.id
  client_id = keycloak_openid_client.test_client.id

  claim_name       = "foo"
  claim_value_type = "String"
  session_note     = "bar"

  add_to_id_token     = true
  add_to_access_token = false
}

resource "keycloak_openid_user_session_note_protocol_mapper" "user_session_note_client_scope" {
  name            = "tf-test-open-id-user-session-note-protocol-mapper-client-scope"
  realm_id        = keycloak_realm.test.id
  client_scope_id = keycloak_openid_client_scope.test_default_client_scope.id

  claim_name       = "foo2"
  claim_value_type = "String"
  session_note     = "bar2"

  add_to_id_token     = true
  add_to_access_token = false
}

resource "keycloak_openid_client" "bearer_only_client" {
  client_id   = "test-bearer-only-client"
  name        = "test-bearer-only-client"
  realm_id    = keycloak_realm.test.id
  description = "a test openid client using bearer-only"

  access_type = "BEARER-ONLY"
}

resource "keycloak_openid_audience_protocol_mapper" "audience_client_scope" {
  name            = "tf-test-openid-audience-protocol-mapper-client-scope"
  realm_id        = keycloak_realm.test.id
  client_scope_id = keycloak_openid_client_scope.test_default_client_scope.id

  add_to_id_token     = true
  add_to_access_token = false

  included_client_audience = keycloak_openid_client.bearer_only_client.client_id
}

resource "keycloak_openid_audience_protocol_mapper" "audience_client" {
  name      = "tf-test-openid-audience-protocol-mapper-client"
  realm_id  = keycloak_realm.test.id
  client_id = keycloak_openid_client.test_client.id

  add_to_id_token     = false
  add_to_access_token = true

  included_custom_audience = "foo"
}

resource "keycloak_saml_client" "saml_client" {
  realm_id  = keycloak_realm.test.id
  client_id = "test-saml-client"
  name      = "test-saml-client"

  sign_documents          = false
  sign_assertions         = true
  include_authn_statement = true

  signing_certificate = file("../provider/misc/saml-cert.pem")
  signing_private_key = file("../provider/misc/saml-key.pem")
}

resource "keycloak_saml_user_attribute_protocol_mapper" "saml_user_attribute_mapper" {
  realm_id  = keycloak_realm.test.id
  client_id = keycloak_saml_client.saml_client.id
  name      = "test-saml-user-attribute-mapper"

  user_attribute             = "user-attribute"
  friendly_name              = "friendly-name"
  saml_attribute_name        = "saml-attribute-name"
  saml_attribute_name_format = "Unspecified"
}

resource "keycloak_saml_user_property_protocol_mapper" "saml_user_property_mapper" {
  realm_id  = keycloak_realm.test.id
  client_id = keycloak_saml_client.saml_client.id
  name      = "test-saml-user-property-mapper"

  user_property              = "email"
  saml_attribute_name        = "email"
  saml_attribute_name_format = "Unspecified"
}

resource "keycloak_oidc_identity_provider" "oidc" {
  realm             = keycloak_realm.test.id
  alias             = "oidc"
  authorization_url = "https://example.com/auth"
  token_url         = "https://example.com/token"
  client_id         = "example_id"
  client_secret     = "example_token"
  default_scopes    = "openid random profile"
  sync_mode         = "FORCE"
  gui_order         = 1
}

resource "keycloak_oidc_google_identity_provider" "google" {
  realm                                   = keycloak_realm.test.id
  client_id                               = "myclientid.apps.googleusercontent.com"
  client_secret                           = "myclientsecret"
  hosted_domain                           = "mycompany.com"
  request_refresh_token                   = true
  default_scopes                          = "openid random profile"
  accepts_prompt_none_forward_from_client = false
  sync_mode                               = "FORCE"
  gui_order                               = 2
}

//This example does not work in keycloak 10, because the interfaces that our customIdp implements, have changed in the keycloak latest version.
//We need to make decide which keycloak version we going to support and test for the customIdp
//resource keycloak_oidc_identity_provider custom_oidc_idp {
//  realm             = "${keycloak_realm.test.id}"
//  provider_id       = "customIdp"
//  alias             = "custom"
//  authorization_url = "https://example.com/auth"
//  token_url         = "https://example.com/token"
//  client_id         = "example_id"
//  client_secret     = "example_token"
//
//  extra_config = {
//    dummyConfig = "dummyValue"
//  }
//}

resource "keycloak_attribute_importer_identity_provider_mapper" "oidc" {
  realm                   = keycloak_realm.test.id
  name                    = "attributeImporter"
  claim_name              = "upn"
  identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
  user_attribute          = "email"

  #KC10 support
  extra_config = {
    syncMode = "INHERIT"
  }
}

resource "keycloak_attribute_to_role_identity_provider_mapper" "oidc" {
  realm                   = keycloak_realm.test.id
  name                    = "attributeToRole"
  claim_name              = "upn"
  identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
  claim_value             = "value"
  role                    = "testRole"

  #KC10 support
  extra_config = {
    syncMode = "INHERIT"
  }
}

resource "keycloak_user_template_importer_identity_provider_mapper" "oidc" {
  realm                   = keycloak_realm.test.id
  name                    = "userTemplate"
  identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
  template                = "$${ALIAS}/$${CLAIM.upn}"

  #KC10 support
  extra_config = {
    syncMode = "INHERIT"
  }
}

resource "keycloak_hardcoded_role_identity_provider_mapper" "oidc" {
  realm                   = keycloak_realm.test.id
  name                    = "hardcodedRole"
  identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
  role                    = "testrole"

  #KC10 support
  extra_config = {
    syncMode = "INHERIT"
  }
}

resource "keycloak_hardcoded_group_identity_provider_mapper" "oidc" {
  realm                   = keycloak_realm.test.id
  name                    = "hardcodedGroup"
  identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
  group                   = "testgroup"

  #KC10 support
  extra_config = {
    syncMode = "INHERIT"
  }
}

resource "keycloak_hardcoded_attribute_identity_provider_mapper" "oidc" {
  realm                   = keycloak_realm.test.id
  name                    = "hardcodedUserSessionAttribute"
  identity_provider_alias = keycloak_oidc_identity_provider.oidc.alias
  attribute_name          = "attribute"
  attribute_value         = "value"
  user_session            = true

  #KC10 support
  extra_config = {
    syncMode = "INHERIT"
  }
}

resource "keycloak_saml_identity_provider" "saml" {
  realm                      = keycloak_realm.test.id
  alias                      = "saml"
  entity_id                  = "https://example.com/entity_id"
  single_sign_on_service_url = "https://example.com/auth"
  sync_mode                  = "FORCE"
  gui_order                  = 3
}

resource "keycloak_attribute_importer_identity_provider_mapper" "saml" {
  realm                   = keycloak_realm.test.id
  name                    = "Attribute: email"
  attribute_name          = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
  identity_provider_alias = keycloak_saml_identity_provider.saml.alias
  user_attribute          = "email"

  #KC10 support
  extra_config = {
    syncMode = "INHERIT"
  }
}

resource "keycloak_attribute_to_role_identity_provider_mapper" "saml" {
  realm                   = keycloak_realm.test.id
  name                    = "attributeToRole"
  attribute_name          = "upn"
  identity_provider_alias = keycloak_saml_identity_provider.saml.alias
  attribute_value         = "value"
  role                    = "testRole"

  #KC10 support
  extra_config = {
    syncMode = "INHERIT"
  }
}

resource "keycloak_user_template_importer_identity_provider_mapper" "saml" {
  realm                   = keycloak_realm.test.id
  name                    = "userTemplate"
  identity_provider_alias = keycloak_saml_identity_provider.saml.alias
  template                = "$${ALIAS}/$${NAMEID}"

  #KC10 support
  extra_config = {
    syncMode = "INHERIT"
  }
}

resource "keycloak_hardcoded_role_identity_provider_mapper" "saml" {
  realm                   = keycloak_realm.test.id
  name                    = "hardcodedRole"
  identity_provider_alias = keycloak_saml_identity_provider.saml.alias
  role                    = "testrole"

  #KC10 support
  extra_config = {
    syncMode = "INHERIT"
  }
}

resource "keycloak_hardcoded_group_identity_provider_mapper" "saml" {
  realm                   = keycloak_realm.test.id
  name                    = "hardcodedGroup"
  identity_provider_alias = keycloak_saml_identity_provider.saml.alias
  group                   = "testgroup"

  #KC10 support
  extra_config = {
    syncMode = "INHERIT"
  }
}

resource "keycloak_hardcoded_attribute_identity_provider_mapper" "saml" {
  realm                   = keycloak_realm.test.id
  name                    = "hardcodedAttribute"
  identity_provider_alias = keycloak_saml_identity_provider.saml.alias
  attribute_name          = "attribute"
  attribute_value         = "value"
  user_session            = false

  #KC10 support
  extra_config = {
    syncMode = "INHERIT"
  }
}

resource "keycloak_saml_identity_provider" "saml_custom" {
  realm                      = keycloak_realm.test.id
  alias                      = "custom_saml"
  provider_id                = "saml"
  entity_id                  = "https://example.com/entity_id"
  single_sign_on_service_url = "https://example.com/auth"
  sync_mode                  = "FORCE"
  gui_order                  = 4
  extra_config = {
    mycustomAttribute = "aValue"
  }
}

data "keycloak_openid_client" "broker" {
  realm_id  = keycloak_realm.test.id
  client_id = "broker"
}

data "keycloak_openid_client_authorization_policy" "default" {
  realm_id           = keycloak_realm.test.id
  resource_server_id = keycloak_openid_client.test_client_auth.resource_server_id
  name               = "default"
}

resource "keycloak_openid_client" "test_client_auth" {
  client_id   = "test-client-auth"
  name        = "test-client-auth"
  realm_id    = keycloak_realm.test.id
  description = "a test openid client"

  access_type                  = "CONFIDENTIAL"
  direct_access_grants_enabled = true
  implicit_flow_enabled        = true
  service_accounts_enabled     = true

  valid_redirect_uris = [
    "http://localhost:5555/callback",
  ]

  authorization {
    policy_enforcement_mode = "ENFORCING"
  }

  client_secret = "secret"
}

resource "keycloak_openid_client" "test_open_id_client_with_consent_text" {
  client_id   = "test_open_id_client_with_consent_text"
  name        = "test_open_id_client_with_consent_text"
  realm_id    = keycloak_realm.test.id
  description = "a test openid client that has consent text"

  standard_flow_enabled    = true
  service_accounts_enabled = true

  access_type = "CONFIDENTIAL"

  valid_redirect_uris = [
    "http://localhost:5555/callback",
  ]

  client_secret = "secret"

  pkce_code_challenge_method = "plain"

  login_theme = "keycloak"

  backchannel_logout_url                     = "http://localhost:3333/backchannel"
  backchannel_logout_session_required        = true
  backchannel_logout_revoke_offline_sessions = true

  extra_config = {
    customAttribute = "a test custom value"
  }

  consent_required          = true
  display_on_consent_screen = true
  consent_screen_text       = "some consent screen text"
}


resource "keycloak_openid_client_authorization_permission" "resource" {
  resource_server_id = keycloak_openid_client.test_client_auth.resource_server_id
  realm_id           = keycloak_realm.test.id
  name               = "test"

  policies = [
    data.keycloak_openid_client_authorization_policy.default.id,
  ]

  resources = [
    keycloak_openid_client_authorization_resource.resource.id,
  ]

  scopes = [
    keycloak_openid_client_authorization_scope.resource.id
  ]
}

resource "keycloak_openid_client_authorization_resource" "resource" {
  resource_server_id = keycloak_openid_client.test_client_auth.resource_server_id
  name               = "test-openid-client1"
  realm_id           = keycloak_realm.test.id

  uris = [
    "/endpoint/*",
  ]

  attributes = {
    "asdads" = "asdasd"
  }
}

resource "keycloak_openid_client_authorization_scope" "resource" {
  resource_server_id = keycloak_openid_client.test_client_auth.resource_server_id
  name               = "test-openid-client1"
  realm_id           = keycloak_realm.test.id
}

resource "keycloak_user" "user_with_multivalueattributes" {
  realm_id = keycloak_realm.test.id
  username = "user-with-mutivalueattributes"

  attributes = {
    "permissions" = "permission1##permission2"
  }
  initial_password {
    value     = "My password"
    temporary = false
  }
}

resource "keycloak_user" "resource" {
  realm_id = keycloak_realm.test.id
  username = "test"

  attributes = {
    "key" = "value"
  }
}

resource "keycloak_openid_client_service_account_role" "read_token" {
  realm_id                = keycloak_realm.test.id
  client_id               = data.keycloak_openid_client.broker.id
  service_account_user_id = keycloak_openid_client.test_client_auth.service_account_user_id
  role                    = "read-token"
}

resource "keycloak_authentication_flow" "browser-copy-flow" {
  alias       = "browserCopyFlow"
  realm_id    = keycloak_realm.test.id
  description = "browser based authentication"
}

resource "keycloak_authentication_execution" "browser-copy-cookie" {
  realm_id          = keycloak_realm.test.id
  parent_flow_alias = keycloak_authentication_flow.browser-copy-flow.alias
  authenticator     = "auth-cookie"
  requirement       = "ALTERNATIVE"
  priority          = 20
}

resource "keycloak_authentication_execution" "browser-copy-kerberos" {
  realm_id          = keycloak_realm.test.id
  parent_flow_alias = keycloak_authentication_flow.browser-copy-flow.alias
  authenticator     = "auth-spnego"
  requirement       = "DISABLED"
  priority          = 10
}

resource "keycloak_authentication_execution" "browser-copy-idp-redirect" {
  realm_id          = keycloak_realm.test.id
  parent_flow_alias = keycloak_authentication_flow.browser-copy-flow.alias
  authenticator     = "identity-provider-redirector"
  requirement       = "ALTERNATIVE"
  priority          = 30
}

resource "keycloak_authentication_subflow" "browser-copy-flow-forms" {
  realm_id          = keycloak_realm.test.id
  parent_flow_alias = keycloak_authentication_flow.browser-copy-flow.alias
  alias             = "browser-copy-flow-forms"
  requirement       = "ALTERNATIVE"
  priority          = 40
}

resource "keycloak_authentication_execution" "browser-copy-auth-username-password-form" {
  realm_id          = keycloak_realm.test.id
  parent_flow_alias = keycloak_authentication_subflow.browser-copy-flow-forms.alias
  authenticator     = "auth-username-password-form"
  requirement       = "REQUIRED"
  priority          = 50
}

resource "keycloak_authentication_execution" "browser-copy-otp" {
  realm_id          = keycloak_realm.test.id
  parent_flow_alias = keycloak_authentication_subflow.browser-copy-flow-forms.alias
  authenticator     = "auth-otp-form"
  requirement       = "REQUIRED"
  priority          = 60
}

resource "keycloak_authentication_execution_config" "config" {
  realm_id     = keycloak_realm.test.id
  execution_id = keycloak_authentication_execution.browser-copy-idp-redirect.id
  alias        = "idp-XXX-config"
  config = {
    defaultProvider = "idp-XXX"
  }
}

resource "keycloak_authentication_bindings" "test_bindings" {
  realm_id     = keycloak_realm.test.id
  browser_flow = keycloak_authentication_flow.browser-copy-flow.alias
}

resource "keycloak_openid_client" "client" {
  client_id   = "my-override-flow-binding-client"
  realm_id    = keycloak_realm.test.id
  access_type = "PUBLIC"
  authentication_flow_binding_overrides {
    browser_id = keycloak_authentication_flow.browser-copy-flow.id
  }
}

resource "keycloak_realm_user_profile" "userprofile" {
  realm_id = keycloak_realm.test.id

  attribute {
    name = "username"
  }

  attribute {
    name = "email"
  }

  attribute {
    name         = "field1"
    display_name = "Field 1"
    group        = "group1"

    enabled_when_scope = ["offline_access"]

    required_for_roles  = ["user"]
    required_for_scopes = ["offline_access"]

    permissions {
      view = ["admin", "user"]
      edit = ["admin", "user"]
    }

    validator {
      name = "person-name-prohibited-characters"
    }

    validator {
      name = "pattern"
      config = {
        pattern       = "^[a-z]+$"
        error_message = "Nope"
      }
    }

    annotations = {
      foo = "bar"
    }
  }

  attribute {
    name = "field2"
  }

  group {
    name                = "group1"
    display_header      = "Group 1"
    display_description = "A first group"

    annotations = {
      foo = "bar"
    }
  }

  group {
    name = "group2"
  }
}

resource "keycloak_realm_client_policy_profile" "profile" {
  name     = "my-profile"
  realm_id = keycloak_realm.test.id
  executor {
    name = "intent-client-bind-checker"
    configuration = {
      auto-configure = true
    }
  }
  executor {
    name = "secure-session"
  }
}

resource "keycloak_realm_client_policy_profile_policy" "policy" {
  name        = "my-profile-policy"
  realm_id    = keycloak_realm.test.id
  description = "Some desc"
  profiles = [
    keycloak_realm_client_policy_profile.profile.name
  ]

  condition {
    name = "client-type"
    configuration = {
      "protocol" = "openid-connect"
    }
  }

  condition {
    name = "client-attributes"
    configuration = {
      "is-negative-logic" = false
      "attributes"        = jsonencode([{ "key" : "something", "value" : "other3" }])
    }
  }
}
