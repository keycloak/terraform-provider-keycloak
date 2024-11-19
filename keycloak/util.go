package keycloak

import (
	"strconv"
	"strings"
	"time"
)

func getIdFromLocationHeader(locationHeader string) string {
	parts := strings.Split(locationHeader, "/")

	return parts[len(parts)-1]
}

// Converts duration string to a string representing the number of milliseconds, which is used by the Keycloak API
// Ex: "1h" => "3600000"
func getMillisecondsFromDurationString(s string) (string, error) {
	duration, err := time.ParseDuration(s)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(int(duration.Seconds() * 1000)), nil
}

// Converts a string representing milliseconds from Keycloak API to a duration string used by the provider
// Ex: "3600000" => "1h0m0s"
func GetDurationStringFromMilliseconds(milliseconds string) (string, error) {
	ms, err := strconv.Atoi(milliseconds)
	if err != nil {
		return "", err
	}

	return (time.Duration(ms) * time.Millisecond).String(), nil
}

func parseBoolAndTreatEmptyStringAsFalse(b string) (bool, error) {
	if b == "" {
		return false, nil
	}

	return strconv.ParseBool(b)
}

func atoiAndTreatEmptyStringAsZero(s string) (int, error) {
	if s == "" {
		return 0, nil
	}

	return strconv.Atoi(s)
}
