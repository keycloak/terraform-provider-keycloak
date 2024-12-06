package com.github.keycloak

import org.keycloak.broker.oidc.OIDCIdentityProvider
import org.keycloak.models.KeycloakSession

class CustomIdentityProvider(session: KeycloakSession, config: CustomIdentityProviderConfig) : OIDCIdentityProvider(session, config) {
	override fun getConfig(): CustomIdentityProviderConfig {
		return super.getConfig() as CustomIdentityProviderConfig
	}
}
