package com.github.keycloak

import org.keycloak.broker.oidc.OIDCIdentityProviderConfig
import org.keycloak.models.IdentityProviderModel

class CustomIdentityProviderConfig(identityProviderModel: IdentityProviderModel) : OIDCIdentityProviderConfig(identityProviderModel) {

	val dummyConfig: String
		get() = getConfig().getOrDefault("dummyConfig", "")

}
