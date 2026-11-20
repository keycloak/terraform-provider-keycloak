package types

import (
	"bytes"
	"strings"
)

// KeycloakSliceHashDelimited is a slice of strings that is marshaled to a hash-delimited (##) string
// Example: ["foo", "bar"] -> "foo##bar"
type KeycloakSliceHashDelimited []string

func (s KeycloakSliceHashDelimited) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	if s == nil || len(s) == 0 {
		// Allows omitempty to work
		return []byte("null"), nil
	} else {
		buf.WriteString(`"`)

		for i, v := range s {
			if i > 0 {
				buf.WriteString("##")
			}
			buf.WriteString(v)
		}

		buf.WriteString(`"`)
	}

	return buf.Bytes(), nil
}

func (s *KeycloakSliceHashDelimited) UnmarshalJSON(in []byte) error {
	value := string(in)
	if value == `""` {
		*s = make([]string, 0)
		return nil
	}

	str := string(in)
	parts := strings.Split(str, "##")
	*s = parts

	return nil
}
