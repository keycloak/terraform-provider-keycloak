resource "keycloak_realm" "roles_example" {
  realm   = "roles-example"
  enabled = true
}

// API Client and roles

resource "keycloak_openid_client" "pet_api" {
  realm_id  = keycloak_realm.roles_example.id
  client_id = "pet-api"
  name      = "pet-api"

  enabled = true

  access_type = "BEARER-ONLY"
}

// Optional client scope for mapping additional client role
resource "keycloak_openid_client_scope" "extended_pet_details" {
  realm_id  = keycloak_realm.roles_example.id
  name      = "extended-pet-details"
  description = "Optional scope offering additional information when getting pets"
}

resource "keycloak_role" "pet_api_create_pet" {
  name        = "create-pet"
  realm_id    = keycloak_realm.roles_example.id
  client_id   = keycloak_openid_client.pet_api.id
  description = "Ability to create a new pet"
}

resource "keycloak_role" "pet_api_update_pet" {
  name        = "update-pet"
  realm_id    = keycloak_realm.roles_example.id
  client_id   = keycloak_openid_client.pet_api.id
  description = "Ability to update a pet"
}

resource "keycloak_role" "pet_api_read_pet" {
  name        = "read-pet"
  realm_id    = keycloak_realm.roles_example.id
  client_id   = keycloak_openid_client.pet_api.id
  description = "Ability to read / list pets"
}

resource "keycloak_role" "pet_api_delete_pet" {
  name        = "delete-pet"
  realm_id    = keycloak_realm.roles_example.id
  client_id   = keycloak_openid_client.pet_api.id
  description = "Ability to delete a pet"
}

resource "keycloak_role" "pet_api_read_pet_details" {
  name        = "read-pet-with-details"
  realm_id    = keycloak_realm.roles_example.id
  client_id   = keycloak_openid_client.pet_api.id
  description = "Ability to read / list pets with further details"
}

// Map a role from the "pet_api" api client to the "extended_pet_details" client scope
resource "keycloak_generic_role_mapper" "pet_api_read_pet_details_role_mapping" {
  realm_id  = keycloak_realm.roles_example.id
  client_scope_id = keycloak_openid_client_scope.extended_pet_details.id
  role_id   = keycloak_role.pet_api_read_pet_details.id
}

resource "keycloak_role" "pet_api_admin" {
  name      = "admin"
  realm_id  = keycloak_realm.roles_example.id
  client_id = keycloak_openid_client.pet_api.id

  composite_roles = [
    keycloak_role.pet_api_create_pet.id,
    keycloak_role.pet_api_delete_pet.id,
    keycloak_role.pet_api_update_pet.id,
  ]
}

// Consumer client

resource "keycloak_openid_client" "pet_app" {
  realm_id  = keycloak_realm.roles_example.id
  client_id = "pet-app"
  name      = "pet-app"

  enabled = true

  access_type   = "CONFIDENTIAL"
  client_secret = "pet-app-secret"

  // authenticated users - could have many roles
  standard_flow_enabled = true

  // unauthenticated users - needs at least read / list role for browsing
  service_accounts_enabled = true

  valid_redirect_uris = [
    "http://localhost:5555/openid-callback",
  ]

  // disable full scope, roles are assigned via keycloak_generic_role_mapper
  full_scope_allowed = false
}

resource "keycloak_openid_client_optional_scopes" "pet_app_optional_scopes" {
  realm_id       = keycloak_realm.roles_example.id
  client_id      = keycloak_openid_client.pet_app.id

  optional_scopes = [
    keycloak_openid_client_scope.extended_pet_details.name
  ]
}

// The app will always need access to the API, so this audience should be used regardless of auth type
resource "keycloak_openid_audience_protocol_mapper" "pet_app_pet_api_audience_mapper" {
  realm_id  = keycloak_realm.roles_example.id
  client_id = keycloak_openid_client.pet_app.id
  name      = "audience-mapper"

  included_client_audience = keycloak_openid_client.pet_api.client_id
}

// The app will always need to read / list pets regardless of who is logged in
resource "keycloak_openid_hardcoded_role_protocol_mapper" "pet_app_pet_api_read_role" {
  realm_id  = keycloak_realm.roles_example.id
  client_id = keycloak_openid_client.pet_app.id
  name      = "read-pets-role"

  role_id = keycloak_role.pet_api_read_pet.id
}

// Map all roles from the "pet_api" api client to the "pet_app" consumer client, read_pet_details comes via client scope
resource "keycloak_generic_role_mapper" "pet_app_pet_api_read_role_mapping" {
  realm_id  = keycloak_realm.roles_example.id
  client_id = keycloak_openid_client.pet_app.id
  role_id   = keycloak_role.pet_api_read_pet.id
}

resource "keycloak_generic_role_mapper" "pet_app_pet_api_delete_role_mapping" {
  realm_id  = keycloak_realm.roles_example.id
  client_id = keycloak_openid_client.pet_app.id
  role_id   = keycloak_role.pet_api_delete_pet.id
}

resource "keycloak_generic_role_mapper" "pet_app_pet_api_create_role_mapping" {
  realm_id  = keycloak_realm.roles_example.id
  client_id = keycloak_openid_client.pet_app.id
  role_id   = keycloak_role.pet_api_create_pet.id
}

resource "keycloak_generic_role_mapper" "pet_app_pet_api_update_role_mapping" {
  realm_id  = keycloak_realm.roles_example.id
  client_id = keycloak_openid_client.pet_app.id
  role_id   = keycloak_role.pet_api_update_pet.id
}

resource "keycloak_generic_role_mapper" "pet_app_pet_api_admin_role_mapping" {
  realm_id  = keycloak_realm.roles_example.id
  client_id = keycloak_openid_client.pet_app.id
  role_id   = keycloak_role.pet_api_admin.id
}

// Realm roles

resource "keycloak_role" "realm_reader" {
  realm_id  = keycloak_realm.roles_example.id
  name        = "realm_reader"
  description = "Reader realm role"
}

resource "keycloak_role" "realm_writer" {
  realm_id  = keycloak_realm.roles_example.id
  name        = "realm_writer"
  description = "Writer realm role"
}

resource "keycloak_role" "realm_admin" {
  realm_id  = keycloak_realm.roles_example.id
  name        = "realm_admin"
  description = "Admin realm composite role"
  composite_roles = [
    keycloak_role.realm_reader.id,
    keycloak_role.realm_writer.id
  ]
}

// Client scope for realm roles mapping

resource "keycloak_openid_client_scope" "petstore_api_access_scope" {
  realm_id  = keycloak_realm.roles_example.id
  name        = "petstore-api-access"
  description = "Optional scope offering additional information for petstore api access"
}

resource "keycloak_generic_role_mapper" "petstore_api_access_scope_admin" {
    realm_id  = keycloak_realm.roles_example.id
    client_scope_id =  keycloak_openid_client_scope.petstore_api_access_scope.id
    role_id         = keycloak_role.realm_admin.id
}

// Users and groups

resource "keycloak_group" "pet_api_base" {
  realm_id = keycloak_realm.roles_example.id
  name     = "pets"
}

resource "keycloak_group" "pet_api_admins" {
  realm_id  = keycloak_realm.roles_example.id
  parent_id = keycloak_group.pet_api_base.id
  name      = "admins"
}

resource "keycloak_group" "pet_api_front_desk" {
  realm_id  = keycloak_realm.roles_example.id
  parent_id = keycloak_group.pet_api_base.id
  name      = "front-desk"
}

data "keycloak_role" "realm_offline_access" {
  realm_id = keycloak_realm.roles_example.id
  name     = "offline_access"
}

resource "keycloak_group_roles" "admin_group_roles" {
  realm_id = keycloak_realm.roles_example.id
  group_id = keycloak_group.pet_api_admins.id

  role_ids = [
    keycloak_role.pet_api_read_pet.id,
    keycloak_role.pet_api_read_pet_details.id,
    keycloak_role.pet_api_delete_pet.id,
    keycloak_role.pet_api_create_pet.id,
    data.keycloak_role.realm_offline_access.id,
  ]
}

resource "keycloak_group_roles" "front_desk_roles" {
  realm_id = keycloak_realm.roles_example.id
  group_id = keycloak_group.pet_api_front_desk.id

  role_ids = [
    keycloak_role.pet_api_read_pet.id,
    keycloak_role.pet_api_read_pet_details.id,
    keycloak_role.pet_api_create_pet.id,
    data.keycloak_role.realm_offline_access.id,
  ]
}

resource "keycloak_user" "admin_user" {
  realm_id = keycloak_realm.roles_example.id
  username = "admin-user"

  email      = "admin-user@fakedomain.com"
  first_name = "Admin"
  last_name  = "Istrator"
}

resource "keycloak_user_roles" "admin_user_roles" {
  realm_id = keycloak_realm.roles_example.id
  user_id  = keycloak_user.admin_user.id

  role_ids = [
    keycloak_role.pet_api_read_pet.id,
    keycloak_role.pet_api_read_pet_details.id,
    keycloak_role.pet_api_delete_pet.id,
    keycloak_role.pet_api_create_pet.id,
    data.keycloak_role.realm_offline_access.id,
  ]
}
