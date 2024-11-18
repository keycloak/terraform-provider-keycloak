# terraform-provider-keycloak
Terraform provider for [Keycloak](https://www.keycloak.org/).

> [!NOTE]
> Please note that this is currently work-in-progress, please be patient while we are preparing for the first release in the new organization.

## Docs

All documentation for this provider can now be found on the Terraform Registry: https://registry.terraform.io/providers/keycloak/keycloak/latest/docs

## Installation

This provider can be installed automatically using Terraform >=0.13 by using the `terraform` configuration block:

```hcl
terraform {
  required_providers {
    keycloak = {
      source = "keycloak/keycloak"
      version = ">= 4.0.0"
    }
  }
}
```

If you are using Terraform 0.12, you can use this provider by downloading it and placing it within
one of the [implied local mirror directories](https://www.terraform.io/docs/commands/cli-config.html#implied-local-mirror-directories).
Or, follow the [old instructions for installing third-party plugins](https://www.terraform.io/docs/configuration-0-11/providers.html#third-party-plugins).

If you are using any provider version below v2.0.0, you can also follow the [old instructions for installing third-party plugins](https://www.terraform.io/docs/configuration-0-11/providers.html#third-party-plugins).

## A note for users of the legacy Wildfly distribution

Recently, Keycloak has been updated to use Quarkus over the legacy Wildfly distribution. The only significant change here
that affects this Terraform provider is the removal of `/auth` from the default context path for the Keycloak API.

If you are using the legacy Wildfly distribution of Keycloak, you will need to set the `base_path` provider argument to
`/auth`. This can also be done by using the `KEYCLOAK_BASE_PATH` environment variable.

## Supported Versions

This provider will officially support the latest three major versions of Keycloak, although older versions may still work.

The following versions are used when running acceptance tests in CI:

- 21.0.1 (latest)
- 20.0.5
- 19.0.2

## Releases

This provider uses [GoReleaser](https://goreleaser.com/) to build and publish releases. Each release published to GitHub
contains binary files for Linux, macOS (darwin), and Windows, as configured within the [`.goreleaser.yml`](https://github.com/keycloak/terraform-provider-keycloak/blob/master/.goreleaser.yml)
file.

Each release also contains a `terraform-provider-keycloak_${RELEASE_VERSION}_SHA256SUMS` file, accompanied by a signature
created by a PGP key with the fingerprint `C508 6791 5E11 6CD2`. This key can be found on my Keybase account at https://keybase.io/mrparkers.

You can find the list of releases [here](https://github.com/keycloak/terraform-provider-keycloak/releases).
You can find the changelog for each version [here](https://github.com/keycloak/terraform-provider-keycloak/blob/master/CHANGELOG.md).

Note: Prior to v2.0.0, a statically linked build for use within Alpine linux was included with each release. This is no longer
done due to [GoReleaser not supporting CGO](https://goreleaser.com/limitations/cgo/). Instead of using a statically linked,
build you can use the `linux_amd64` build as long as `libc6-compat` is installed.

## Development

This project requires Go 1.19 and Terraform 1.4.1.
This project uses [Go Modules](https://github.com/golang/go/wiki/Modules) for dependency management, which allows this project to exist outside of an existing GOPATH.

After cloning the repository, you can build the project by running `make build`.

### Local Environment

You can spin up a local developer environment via [Docker Compose](https://docs.docker.com/compose/) by running `make local`.
This will spin up a few containers for Keycloak, PostgreSQL, and OpenLDAP, which can be used for testing the provider.
This environment and its setup via `make local` is not intended for production use.

Note: The setup scripts require the [jq](https://stedolan.github.io/jq/) command line utility.

### Tests

Every resource supported by this provider will have a reasonable amount of acceptance test coverage.

You can run acceptance tests against a Keycloak instance by running `make testacc`. You will need to supply some environment
variables in order to set up the provider during tests. Here is an example for running tests against a local environment
that was created via `make local`:

```
KEYCLOAK_CLIENT_ID=terraform \
KEYCLOAK_CLIENT_SECRET=884e0f95-0f42-4a63-9b1f-94274655669e \
KEYCLOAK_CLIENT_TIMEOUT=5 \
KEYCLOAK_REALM=master \
KEYCLOAK_URL="http://localhost:8080" \
make testacc
```

## Acknowledgments

The Keycloak Terraform Provider was originally created by [Michael Parker](https://github.com/mrparkers). Many thanks for the hard work and dedication in building the foundation for this project.
Also, many thanks to all the contributors extending it and approving the license change to maintain it as part of the [Keycloak](https://www.keycloak.org/) project.

## License

This software is licensed under Apache License, Version 2.0, (LICENSE-APACHE-2.0 or https://www.apache.org/licenses/LICENSE-2.0)

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in this software by you shall be licensed under the Apache License, Version 2.0, without any additional terms or conditions.
