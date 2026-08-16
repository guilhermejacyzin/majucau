package security

import (
	"context"
	"errors"
	"strings"
)

var ErrUnsupportedPlatform = errors.New("secret protection is unsupported on this platform")
var ErrInvalidSecretRef = errors.New("invalid secret reference")

type SecretStore interface {
	Put(context.Context, string, []byte) error
	Get(context.Context, string) ([]byte, error)
	Delete(context.Context, string) error
}

func ValidateSecretRef(ref string) error {
	ref = strings.TrimSpace(ref)
	if ref == "" || len(ref) > 128 || strings.ContainsAny(ref, `/\\:.`) {
		return ErrInvalidSecretRef
	}
	for _, r := range ref {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return ErrInvalidSecretRef
		}
	}
	return nil
}
