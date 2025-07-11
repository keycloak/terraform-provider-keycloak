package keycloak

import (
	"context"
	"github.com/hashicorp/go-version"
)

type Version string

const (
	Version_6    Version = "6.0.0"
	Version_7    Version = "7.0.0"
	Version_8    Version = "8.0.0"
	Version_9    Version = "9.0.0"
	Version_10   Version = "10.0.0"
	Version_11   Version = "11.0.0"
	Version_12   Version = "12.0.0"
	Version_13   Version = "13.0.0"
	Version_14   Version = "14.0.0"
	Version_15   Version = "15.0.0"
	Version_16   Version = "16.0.0"
	Version_17   Version = "17.0.0"
	Version_18   Version = "18.0.0"
	Version_19   Version = "19.0.0"
	Version_20   Version = "20.0.0"
	Version_21   Version = "21.0.0"
	Version_22   Version = "22.0.0"
	Version_23   Version = "23.0.0"
	Version_24   Version = "24.0.0"
	Version_25   Version = "25.0.0"
	Version_26   Version = "26.0.0"
	Version_26_1 Version = "26.1.0"
	Version_26_2 Version = "26.2.0"
	Version_26_3 Version = "26.3.0"
)

func (v Version) AsVersion() *version.Version {
	vv, err := version.NewVersion(string(v))
	if err != nil {
		return nil
	}
	return vv
}

func (KeycloakClient *KeycloakClient) Version(ctx context.Context) (*version.Version, error) {
	if KeycloakClient.version == nil {
		err := KeycloakClient.login(ctx)
		if err != nil {
			return nil, err
		}
	}
	return KeycloakClient.version, nil
}

func (keycloakClient *KeycloakClient) VersionIsGreaterThanOrEqualTo(ctx context.Context, versionString Version) (bool, error) {
	version, err := keycloakClient.Version(ctx)
	if err != nil {
		return false, err
	}
	return version.GreaterThanOrEqual(versionString.AsVersion()), nil
}

func (keycloakClient *KeycloakClient) VersionIsLessThanOrEqualTo(ctx context.Context, versionString Version) (bool, error) {
	version, err := keycloakClient.Version(ctx)
	if err != nil {
		return false, err
	}
	return version.LessThanOrEqual(versionString.AsVersion()), nil
}

func (keycloakClient *KeycloakClient) VersionIsLessThan(ctx context.Context, versionString Version) (bool, error) {
	version, err := keycloakClient.Version(ctx)
	if err != nil {
		return false, err
	}
	return version.LessThan(versionString.AsVersion()), nil
}
