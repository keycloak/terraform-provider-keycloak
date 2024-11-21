package com.github.keycloak

import org.keycloak.broker.oidc.OIDCIdentityProviderConfig
import org.keycloak.broker.provider.AbstractIdentityProviderFactory
import org.keycloak.models.IdentityProviderModel
import org.keycloak.models.KeycloakSession
import org.keycloak.protocol.oidc.representations.OIDCConfigurationRepresentation
import org.keycloak.util.JsonSerialization
import java.io.IOException
import java.io.InputStream

class CustomIdentityProviderFactory : AbstractIdentityProviderFactory<CustomIdentityProvider>() {
	override fun getName(): String = "custom idp"
	override fun getId(): String = "customIdp"

	override fun create(session: KeycloakSession, model: IdentityProviderModel): CustomIdentityProvider {
		return CustomIdentityProvider(session, CustomIdentityProviderConfig(model))
	}

	override fun parseConfig(session: KeycloakSession, inputStream: InputStream): Map<String, String> {
		return parseOIDCConfig(session, inputStream)
	}

	private fun parseOIDCConfig(session: KeycloakSession, inputStream: InputStream): Map<String, String> {
		val rep: OIDCConfigurationRepresentation
		try {
			rep = JsonSerialization.readValue(inputStream, OIDCConfigurationRepresentation::class.java)
		} catch (e: IOException) {
			throw RuntimeException("failed to load openid connect metadata", e)
		}

		val config = OIDCIdentityProviderConfig(IdentityProviderModel())
		config.issuer = rep.issuer
		config.logoutUrl = rep.logoutEndpoint
		config.authorizationUrl = rep.authorizationEndpoint
		config.tokenUrl = rep.tokenEndpoint
		config.userInfoUrl = rep.userinfoEndpoint
		if (rep.jwksUri != null) {
			config.isValidateSignature = true
			config.isUseJwksUrl = true
			config.jwksUrl = rep.jwksUri
		}
		return config.config
	}

	override fun createConfig(): IdentityProviderModel {
		return IdentityProviderModel()
	}
}
