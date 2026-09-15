package security

import (
	"encoding/json"
	"strings"
)

var sensitiveNames = map[string]bool{"token": true, "access_token": true, "refresh_token": true, "client_secret": true, "secret": true, "password": true, "authorization": true, "cookie": true}

func RedactValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, value := range x {
			if sensitiveNames[strings.ToLower(k)] {
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
