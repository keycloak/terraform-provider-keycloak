#!/usr/bin/env bash

set -xe

KEYCLOAK_URL="http://localhost:8080"
KEYCLOAK_USER="keycloak"
KEYCLOAK_PASSWORD="password"
KEYCLOAK_CLIENT_ID="terraform"
KEYCLOAK_CLIENT_SECRET="884e0f95-0f42-4a63-9b1f-94274655669e"

echo "Creating initial terraform client"

accessToken=$(
    curl -s --fail \
        -d "username=${KEYCLOAK_USER}" \
        -d "password=${KEYCLOAK_PASSWORD}" \
        -d "client_id=admin-cli" \
        -d "grant_type=password" \
        "${KEYCLOAK_URL}/realms/master/protocol/openid-connect/token" \
        | jq -r '.access_token'
)

function post() {
    curl -s --fail \
        -H "Authorization: bearer ${accessToken}" \
        -H "Content-Type: application/json" \
        -d "${2}" \
        "${KEYCLOAK_URL}/admin${1}"
}

function put() {
    curl -s --fail \
        -X PUT \
        -H "Authorization: bearer ${accessToken}" \
        -H "Content-Type: application/json" \
        -d "${2}" \
        "${KEYCLOAK_URL}/admin${1}"
}

function get() {
    curl --fail --silent \
        -H "Authorization: bearer ${accessToken}" \
        -H "Content-Type: application/json" \
        "${KEYCLOAK_URL}/admin${1}"
}

terraformClient=$(jq -n "{
    id: \"${KEYCLOAK_CLIENT_ID}\",
    name: \"${KEYCLOAK_CLIENT_ID}\",
    secret: \"${KEYCLOAK_CLIENT_SECRET}\",
    clientAuthenticatorType: \"client-secret\",
    enabled: true,
    serviceAccountsEnabled: true,
    directAccessGrantsEnabled: true,
    standardFlowEnabled: false
}")

post "/realms/master/clients" "${terraformClient}"

masterRealmAdminRole=$(get "/realms/master/roles" | jq -r '
    .
    | map(
        select(.name == "admin")
    )
    | .[0]
')
masterRealmAdminRoleId=$(echo ${masterRealmAdminRole} | jq -r '.id')

terraformClientServiceAccount=$(get "/realms/master/clients/${KEYCLOAK_CLIENT_ID}/service-account-user")
terraformClientServiceAccountId=$(echo ${terraformClientServiceAccount} | jq -r '.id')

serviceAccountAdminRoleMapping=$(jq -n "[{
    clientRole: false,
    composite: true,
    containerId: \"master\",
    description: \"\${role_admin}\",
    id: \"${masterRealmAdminRoleId}\",
    name: \"admin\",
}]")

post "/realms/master/users/${terraformClientServiceAccountId}/role-mappings/realm" "${serviceAccountAdminRoleMapping}"

echo "Extending access token lifespan (don't do this in production)"

masterRealmExtendAccessToken=$(jq -n "{
    accessTokenLifespan: 86400,
    accessTokenLifespanForImplicitFlow: 86400,
    ssoSessionIdleTimeout: 86400,
    ssoSessionMaxLifespan: 86400,
    offlineSessionIdleTimeout: 86400,
    offlineSessionMaxLifespan: 5184000,
    accessCodeLifespan: 86400,
    accessCodeLifespanUserAction: 86400,
    accessCodeLifespanLogin: 86400,
    actionTokenGeneratedByAdminLifespan: 86400,
    actionTokenGeneratedByUserLifespan: 86400,
    oauth2DeviceCodeLifespan: 86400
}")

put "/realms/master" "${masterRealmExtendAccessToken}"

echo "Done"
