package com.github.keycloak

import org.keycloak.component.ComponentModel
import org.keycloak.models.KeycloakSession
import org.keycloak.models.KeycloakSessionFactory
import org.keycloak.provider.ProviderConfigProperty
import org.keycloak.storage.UserStorageProviderFactory
import org.keycloak.storage.UserStorageProviderModel
import org.keycloak.storage.user.ImportSynchronization
import org.keycloak.storage.user.SynchronizationResult
import java.util.Date

class CustomUserStorageProviderFactory : UserStorageProviderFactory<CustomUserStorageProvider>, ImportSynchronization {
	override fun getId(): String = "custom"

	override fun init(config: org.keycloak.Config.Scope) {
		super.init(config)
	}

	override fun create(session: KeycloakSession, model: ComponentModel): CustomUserStorageProvider =
		CustomUserStorageProvider(session, model)

	override fun getConfigProperties(): List<ProviderConfigProperty> = configPropertyList

	companion object {
		private val configPropertyList = ArrayList<ProviderConfigProperty>()

		init {
			val property = ProviderConfigProperty()
			property.setName("dummyConfig")
			property.setLabel("Dummy Config")
			property.setDefaultValue("")
			property.setType(ProviderConfigProperty.STRING_TYPE)
			property.setHelpText("Dummy config for testing")
			configPropertyList.add(property)
		}
	}

	override fun sync(sessionFactory: KeycloakSessionFactory?, realmId: String?, model: UserStorageProviderModel?):
		SynchronizationResult = SynchronizationResult()

	override fun syncSince(
		lastSync: Date?,
		sessionFactory: KeycloakSessionFactory?,
		realmId: String?,
		model: UserStorageProviderModel?
	): SynchronizationResult = SynchronizationResult()
}
