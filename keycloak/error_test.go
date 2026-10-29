package keycloak

import (
	"errors"
	"net/http"
	"testing"

	"github.com/hashicorp/errwrap"
)

func TestErrorIs5xx(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"non api error", errors.New("boom"), false},
		{"400", &ApiError{Code: http.StatusBadRequest}, false},
		{"404", &ApiError{Code: http.StatusNotFound}, false},
		{"500", &ApiError{Code: http.StatusInternalServerError}, true},
		{"503", &ApiError{Code: http.StatusServiceUnavailable}, true},
		{"wrapped 500", errwrap.Wrapf("context: {{err}}", &ApiError{Code: http.StatusInternalServerError}), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorIs5xx(tt.err); got != tt.want {
				t.Errorf("ErrorIs5xx() = %v, want %v", got, tt.want)
			}
		})
	}
}
