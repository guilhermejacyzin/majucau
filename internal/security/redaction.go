package security

import (
	"encoding/json"
	"strings"
	"unicode"
)

const maxRedactJSONBytes = 1 << 20

var sensitiveNames = map[string]struct{}{
	"token": {}, "accesstoken": {}, "refreshtoken": {}, "idtoken": {}, "sessiontoken": {},
	"clientsecret": {}, "secret": {}, "secretkey": {}, "privatekey": {}, "apikey": {},
	"password": {}, "passwd": {}, "credential": {}, "credentials": {},
	"authorization": {}, "proxyauthorization": {}, "cookie": {}, "setcookie": {},
}

func RedactValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, value := range x {
			if isSensitiveName(k) {
				out[k] = RedactString(toString(value))
			} else {
				out[k] = RedactValue(value)
			}
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, value := range x {
			out[i] = RedactValue(value)
		}
		return out
	default:
		return v
	}
}

func isSensitiveName(name string) bool {
	normalized := strings.Map(func(r rune) rune {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return -1
		}
		return unicode.ToLower(r)
	}, name)
	_, ok := sensitiveNames[normalized]
	return ok
}

func RedactString(value string) string {
	if value == "" {
		return "[REDACTED]"
	}
	if len(value) <= 4 {
		return "[REDACTED]"
	}
	return "[REDACTED]" + value[len(value)-4:]
}
func RedactJSON(data []byte) []byte {
	if len(data) > maxRedactJSONBytes {
		return []byte(`"[REDACTED]"`)
	}
	var v any
	if json.Unmarshal(data, &v) != nil {
		return []byte(`"[REDACTED]"`)
	}
	b, err := json.Marshal(RedactValue(v))
	if err != nil {
		return []byte(`"[REDACTED]"`)
	}
	return b
}
func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v)
	return string(b)
}
