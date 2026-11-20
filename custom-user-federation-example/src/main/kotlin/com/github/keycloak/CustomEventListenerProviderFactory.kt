package com.github.keycloak

import org.keycloak.Config
import org.keycloak.events.EventListenerProvider
import org.keycloak.events.EventListenerProviderFactory
import org.keycloak.models.KeycloakSession
import org.keycloak.models.KeycloakSessionFactory

class CustomEventListenerProviderFactory : EventListenerProviderFactory {

	override fun create(session: KeycloakSession): EventListenerProvider {
		return CustomEventListenerProvider(session);
	}

	override fun init(config: Config.Scope) {
		// NOOP
	}

	override fun postInit(sessionFactory: KeycloakSessionFactory) {
		// NOOP
	}

	override fun close() {
		// NOOP
	}

	override fun getId(): String {
		return "example-listener";
	}
}
