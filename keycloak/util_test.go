package keycloak

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

func TestLocationHeaderParseForClient(t *testing.T) {
	expectedRealm := "terraform-" + acctest.RandString(10)
	expectedId := "aed9f702-c0d3-42cd-91b6-a894da769324"
	locationHeader := fmt.Sprintf("http://localhost:8080/auth/admin/realms/%s/clients/%s", expectedRealm, expectedId)

	actualId := getIdFromLocationHeader(locationHeader)

	if expectedId != actualId {
		t.Fatalf("parsed keycloak client location header did not return correct ID")
	}
}

func TestLocationHeaderParseForComponent(t *testing.T) {
	expectedRealm := "terraform-" + acctest.RandString(10)
	expectedId := "cca9f52a-2659-4cae-996d-bb788cd1d167"
	locationHeader := fmt.Sprintf("http://localhost:8080/auth/admin/realms/%s/components/%s", expectedRealm, expectedId)

	actualId := getIdFromLocationHeader(locationHeader)

	if expectedId != actualId {
		t.Fatalf("parsed keycloak component location header did not return correct ID")
	}
}
