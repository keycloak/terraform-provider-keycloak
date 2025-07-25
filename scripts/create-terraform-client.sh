#!/usr/bin/env bash

set -e

KEYCLOAK_URL="http://localhost:8080"
KEYCLOAK_USER="keycloak"
KEYCLOAK_PASSWORD="password"
KEYCLOAK_CLIENT_ID="terraform"
KEYCLOAK_CLIENT_JWT_ID="terraform-jwt"
KEYCLOAK_CLIENT_SECRET="884e0f95-0f42-4a63-9b1f-94274655669e"
KEYCLOAK_CLIENT_PUBLIC_CERT="MIICuTCCAaECBgGX+cu3CTANBgkqhkiG9w0BAQsFADAgMR4wHAYDVQQDDBV0ZXJyYWZvcm0tYWRtaW4tY2xpLTIwHhcNMjUwNzExMTQwMTA2WhcNMzUwNzExMTQwMjQ2WjAgMR4wHAYDVQQDDBV0ZXJyYWZvcm0tYWRtaW4tY2xpLTIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDEgtQWZtM/nSIXpuJ9aJ9elurICPD1FEKOaIQBg6MihHYDDQxkIT93FkGLTRPMNFNgSXUKKj7IW9Ihxpv7v4ltsRaNmT0n2CFmVDGQI9404M1vz7V6Gj70feHWtcwzF42kvsCMEETwsW0jkOSX2RXun7RPaLSfXavki0w1ql3/nKVxQMuFgmMZrQzpIGh/EPpRjEWgL9HRYlp4X5wbW/GGzDuUldJoJBWhBWb8uOVSJcmXcgZ45k5LxkGTjTXPlgYJorVSdS8bkoBxBRa3COwilTwUFiBNna4HwnMmLFiKEYMhCGO+HvyFm4AKzhhShHQSm8VrISWsD05Fg0Uo/LEHAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAF+tXQJMAjxx5dvAN43/PB5IzrVjb5MeEZGZfd940V2dtBe+xMT74NmRPgIfTcXavZ7T/LFmRPpt6QZAC6WpjS69gbvCTEmLQo3R5KQTu8C6nVLPp+kVmLKqrB/UN3GxTuojjXiFz9KOdW1kmH5V19asMyDQBrQlrTJV7qoQ8c3wUA0DE1dTVx9xx+x0LeWEGeyO50LmRuLNF1Uuv51fIeHmgTNQCL+DHs/1QHNbz/nga8ruMP00b75PAdv9EWPYaCVeobQYR7tT8k3MfSbcOvkgFqHqOoX6jF3lXssHfbTj6fgJuCYZ0h0kr/JC4oiKmyFElut375+yZ2WynGLkCyA="

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

terraformJwtClient=$(jq -n "{
    id: \"${KEYCLOAK_CLIENT_JWT_ID}\",
    name: \"${KEYCLOAK_CLIENT_JWT_ID}\",
    clientAuthenticatorType: \"client-jwt\",
    enabled: true,
    serviceAccountsEnabled: true,
    directAccessGrantsEnabled: true,
    standardFlowEnabled: false,
    attributes: {\"jwt.credential.certificate\": \"${KEYCLOAK_CLIENT_PUBLIC_CERT}\"}
}")

post "/realms/master/clients" "${terraformClient}"
post "/realms/master/clients" "${terraformJwtClient}"

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
terraformClientJWTServiceAccount=$(get "/realms/master/clients/${KEYCLOAK_CLIENT_JWT_ID}/service-account-user")
terraformClientJWTServiceAccountId=$(echo ${terraformClientJWTServiceAccount} | jq -r '.id')

serviceAccountAdminRoleMapping=$(jq -n "[{
    clientRole: false,
    composite: true,
    containerId: \"master\",
    description: \"\${role_admin}\",
    id: \"${masterRealmAdminRoleId}\",
    name: \"admin\",
}]")

post "/realms/master/users/${terraformClientServiceAccountId}/role-mappings/realm" "${serviceAccountAdminRoleMapping}"
post "/realms/master/users/${terraformClientJWTServiceAccountId}/role-mappings/realm" "${serviceAccountAdminRoleMapping}"


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
