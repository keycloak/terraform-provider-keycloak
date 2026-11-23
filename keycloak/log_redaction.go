package keycloak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const redactedLogValue = "[REDACTED]"

// Token request parameters that are safe to log. Every other parameter is redacted, so that a credential added to the
// request later is not logged by default.
var loggableFormParameters = map[string]bool{
	"client_assertion_type": true,
	"client_id":             true,
	"grant_type":            true,
	"username":              true,
}

// A JSON key holds a secret when its name, in lower case and without separators, contains one of these fragments, as
// in clientSecret, bindCredential, keystorePassword or saml.signing.private.key.
var sensitiveLogKeyFragments = []string{"secret", "password", "passwd", "credential", "privatekey", "passphrase", "apikey"}

// Keycloak settings that match the rules for a secret without being one. They stay readable, as they are needed to
// debug realm and user federation issues.
var nonSensitiveLogKeys = map[string]bool{
	"clientcredentialsuserefreshtoken":     true,
	"ldappasswordpolicyhintsenabled":       true,
	"passwordpolicy":                       true,
	"resetcredentialsflow":                 true,
	"secretsize":                           true,
	"usekerberosforpasswordauthentication": true,
	"usepasswordmodifyextendedop":          true,
	"validatepasswordpolicy":               true,
}

const nonSensitiveLogKeyPrefix = "webauthnpolicypasswordless"

// lazyLogValue is redacted only when a log line is written, so that bodies are not parsed while DEBUG logging is off.
// It is a function on purpose: a logger that ignores both methods prints an address instead of the wrapped body.
type lazyLogValue func() string

func (value lazyLogValue) String() string {
	return value()
}

func (value lazyLogValue) MarshalText() ([]byte, error) {
	return []byte(value()), nil
}

// redactedLogBody returns a request or response body for logging, with the values of all secrets replaced.
func redactedLogBody(body []byte, indent bool) lazyLogValue {
	return func() string {
		return redactJSONBody(body, indent)
	}
}

func redactFormData(formData url.Values) string {
	parameters := make([]string, 0, len(formData))
	for name, values := range formData {
		for _, value := range values {
			if !loggableFormParameters[name] {
				value = redactedLogValue
			}
			parameters = append(parameters, name+"="+value)
		}
	}
	sort.Strings(parameters)

	return strings.Join(parameters, "&")
}

func redactJSONBody(body []byte, indent bool) string {
	omitted := fmt.Sprintf("[%d bytes omitted: not a JSON object or array]", len(body))

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()

	var parsed interface{}
	if err := decoder.Decode(&parsed); err != nil {
		return omitted
	}

	// a body without keys cannot be told apart from a secret
	switch parsed.(type) {
	case map[string]interface{}, []interface{}:
	default:
		return omitted
	}

	var redacted bytes.Buffer
	encoder := json.NewEncoder(&redacted)
	encoder.SetEscapeHTML(false)
	if indent {
		encoder.SetIndent("", "    ")
	}
	if err := encoder.Encode(redactLogValue(parsed)); err != nil {
		return omitted
	}

	return strings.TrimSuffix(redacted.String(), "\n")
}

func redactLogValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		// a credential, such as a client secret or a user password, is sent as {"type": ..., "value": ...}
		_, isCredential := typed["type"]

		for key, child := range typed {
			if isSensitiveLogKey(key) || (isCredential && key == "value") ||
				(key == "config" && typed["providerType"] == realmKeystoreProviderType) {
				typed[key] = maskLogValue(child)
			} else {
				typed[key] = redactLogValue(child)
			}
		}
	case []interface{}:
		for i, child := range typed {
			typed[i] = redactLogValue(child)
		}
	}

	return value
}

// maskLogValue replaces every string below a sensitive key. Booleans and numbers are kept, Keycloak has no secrets of
// these types. An empty string is kept as well, it shows that no secret is set.
func maskLogValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case string:
		if typed != "" {
			return redactedLogValue
		}
	case map[string]interface{}:
		for key, child := range typed {
			typed[key] = maskLogValue(child)
		}
	case []interface{}:
		for i, child := range typed {
			typed[i] = maskLogValue(child)
		}
	}

	return value
}

func isSensitiveLogKey(key string) bool {
	normalized := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r >= 'A' && r <= 'Z':
			return r + 'a' - 'A'
		}
		return -1
	}, key)

	if nonSensitiveLogKeys[normalized] || strings.HasPrefix(normalized, nonSensitiveLogKeyPrefix) {
		return false
	}

	// access_token, refresh_token, registrationAccessToken, but not token_type or accessTokenLifespan
	if strings.HasSuffix(normalized, "token") {
		return true
	}

	for _, fragment := range sensitiveLogKeyFragments {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}

	return false
}
