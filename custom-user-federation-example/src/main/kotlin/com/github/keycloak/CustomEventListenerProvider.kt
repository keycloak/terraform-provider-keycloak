package com.github.keycloak

import org.keycloak.events.Event
import org.keycloak.events.EventListenerProvider
import org.keycloak.events.admin.AdminEvent
import org.keycloak.models.KeycloakSession

class CustomEventListenerProvider(session: KeycloakSession) : EventListenerProvider {

	override fun onEvent(event: Event) {
		//
	}

	override fun onEvent(adminEvent: AdminEvent, includeRep: Boolean) {
		//
	}

	override fun close() {
		// NOOP
	}

}
