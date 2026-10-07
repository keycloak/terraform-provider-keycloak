package com.github.keycloak

import org.keycloak.component.ComponentModel
import org.keycloak.models.KeycloakSession
import org.keycloak.models.KeycloakSessionFactory
import org.keycloak.models.utils.KeycloakModelUtils
import org.keycloak.provider.ProviderConfigProperty
import org.keycloak.storage.UserStorageProviderFactory
import org.keycloak.storage.UserStoragePrivateUtil
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

			val importUserOnSync = ProviderConfigProperty()
			importUserOnSync.setName(IMPORT_USER_ON_SYNC)
			importUserOnSync.setLabel("Import user on sync")
			importUserOnSync.setDefaultValue("false")
			importUserOnSync.setType(ProviderConfigProperty.BOOLEAN_TYPE)
			importUserOnSync.setHelpText("Imports the user '<federation name>-synced-<full|changed>' on sync, which makes a sync observable for testing")
			configPropertyList.add(importUserOnSync)
		}

		const val IMPORT_USER_ON_SYNC = "importUserOnSync"
	}

	override fun sync(sessionFactory: KeycloakSessionFactory?, realmId: String?, model: UserStorageProviderModel?):
		SynchronizationResult = importUserOnSync(sessionFactory, realmId, model, "full")

	override fun syncSince(
		lastSync: Date?,
		sessionFactory: KeycloakSessionFactory?,
		realmId: String?,
		model: UserStorageProviderModel?
	): SynchronizationResult = importUserOnSync(sessionFactory, realmId, model, "changed")

	// imports the user '<federation name>-synced-<full|changed>' to make a sync observable for testing
	private fun importUserOnSync(
		sessionFactory: KeycloakSessionFactory?,
		realmId: String?,
		model: UserStorageProviderModel?,
		syncMode: String
	): SynchronizationResult {
		val result = SynchronizationResult()

		if (sessionFactory == null || realmId == null || model == null || !model.get(IMPORT_USER_ON_SYNC, false)) {
			return result
		}

		KeycloakModelUtils.runJobInTransaction(sessionFactory) { session ->
			val realm = session.realms().getRealm(realmId)
			session.context.realm = realm

			val localUsers = UserStoragePrivateUtil.userLocalStorage(session)
			// the kind of the sync is part of the username, since other user attributes might be hidden by the user profile
			val username = "${model.name}-synced-$syncMode".lowercase()

			var user = localUsers.getUserByUsername(realm, username)
			if (user == null) {
				user = localUsers.addUser(realm, username)
				user.federationLink = model.id
				user.isEnabled = true
				result.increaseAdded()
			} else {
				result.increaseUpdated()
			}

		}

		return result
	}
}
