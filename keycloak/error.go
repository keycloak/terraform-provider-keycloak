package keycloak

import (
	"net/http"

	"github.com/hashicorp/errwrap"
)

type ApiError struct {
	Code    int
	Message string
}

func (e *ApiError) Error() string {
	return e.Message
}

func ErrorIs404(err error) bool {
	keycloakError, ok := errwrap.GetType(err, &ApiError{}).(*ApiError)

	return ok && keycloakError != nil && keycloakError.Code == http.StatusNotFound
}

func ErrorIs409(err error) bool {
	keycloakError, ok := errwrap.GetType(err, &ApiError{}).(*ApiError)

	return ok && keycloakError != nil && keycloakError.Code == http.StatusConflict
}
