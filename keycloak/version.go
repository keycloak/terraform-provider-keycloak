package keycloak

import (
	"context"
	"github.com/hashicorp/go-version"
)

type Version string

const (
	Version_6  Version = "6.0.0"
	Version_7  Version = "7.0.0"
	Version_8  Version = "8.0.0"
	Version_9  Version = "9.0.0"
	Version_10 Version = "10.0.0"
	Version_11 Version = "11.0.0"
	Version_12 Version = "12.0.0"
	Version_13 Version = "13.0.0"
	Version_14 Version = "14.0.0"
	Version_15 Version = "15.0.0"
	Version_16 Version = "16.0.0"
	Version_17 Version = "17.0.0"
	Version_18 Version = "18.0.0"
	Version_19 Version = "19.0.0"
	Version_20 Version = "20.0.0"
	Version_21 Version = "21.0.0"
	Version_22 Version = "22.0.0"
	Version_23 Version = "23.0.0"
	Version_24 Version = "24.0.0"
	Version_25 Version = "25.0.0"
)

func (keycloakClient *KeycloakClient) VersionIsGreaterThanOrEqualTo(ctx context.Context, versionString Version) (bool, error) {
	if keycloakClient.version == nil {
		err := keycloakClient.login(ctx)
		if err != nil {
			return false, err
		}
	}

	v, err := version.NewVersion(string(versionString))
	if err != nil {
		return false, nil
	}

	return keycloakClient.version.GreaterThanOrEqual(v), nil
}

func (keycloakClient *KeycloakClient) VersionIsLessThanOrEqualTo(ctx context.Context, versionString Version) (bool, error) {
	if keycloakClient.version == nil {
		err := keycloakClient.login(ctx)
		if err != nil {
			return false, err
		}
	}

	v, err := version.NewVersion(string(versionString))
	if err != nil {
		return false, nil
	}

	return keycloakClient.version.LessThanOrEqual(v), nil
}
