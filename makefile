GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
GOOS?=darwin
GOARCH?=arm64

MAKEFLAGS += --silent

VERSION=$$(git describe --tags)

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o terraform-provider-keycloak_$(VERSION)

build-debug:
	# keep debug info in the binary
	CGO_ENABLED=0 go build -gcflags "all=-N -l" -trimpath -ldflags " -X main.version=$(VERSION)" -o terraform-provider-keycloak_$(VERSION)

prepare-example:
	mkdir -p example/.terraform/plugins/terraform.local/keycloak/keycloak/5.3.0/$(GOOS)_$(GOARCH)
	mkdir -p example/terraform.d/plugins/terraform.local/keycloak/keycloak/5.3.0/$(GOOS)_$(GOARCH)
	cp terraform-provider-keycloak_* example/.terraform/plugins/terraform.local/keycloak/keycloak/5.2.0/$(GOOS)_$(GOARCH)/
	cp terraform-provider-keycloak_* example/terraform.d/plugins/terraform.local/keycloak/keycloak/5.2.0/$(GOOS)_$(GOARCH)/

build-example: build prepare-example

build-example-debug: build-debug prepare-example

run-debug:
	echo "Starting delve debugger listening on port 127.0.0.1:58772"
	dlv exec --listen=:58772 --accept-multiclient --headless "./terraform-provider-keycloak_$(VERSION)" -- -debug

local: deps user-federation-example
	echo "Starting local Keycloak environment"
	docker compose up --build -d
	./scripts/wait-for-local-keycloak.sh
	./scripts/create-terraform-client.sh

local-stop:
	echo "Stopping local Keycloak environment"
	docker compose stop

local-down:
	echo "Destroying local Keycloak environment"
	docker compose down

deps:
	./scripts/check-deps.sh

fmt:
	gofmt -w -s $(GOFMT_FILES)

test: fmtcheck vet
	go test $(TEST)

testacc: fmtcheck vet
	go test -v github.com/keycloak/terraform-provider-keycloak/keycloak
	TF_ACC=1 CHECKPOINT_DISABLE=1 go test -v -timeout 60m -parallel 4 github.com/keycloak/terraform-provider-keycloak/provider $(TESTARGS)

fmtcheck:
	lineCount=$(shell gofmt -l -s $(GOFMT_FILES) | wc -l | tr -d ' ') && exit $$lineCount

vet:
	go vet ./...

user-federation-example:
	cd custom-user-federation-example && ./gradlew shadowJar
