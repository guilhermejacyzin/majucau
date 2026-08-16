//go:build !windows

package security

import (
	"context"
)

// DPAPIStore is intentionally unavailable off Windows. Tests should inject a
// SecretStore fake rather than weakening the production security boundary.
type DPAPIStore struct{}

func NewDPAPIStore(_ string) (*DPAPIStore, error)               { return nil, ErrUnsupportedPlatform }
func (*DPAPIStore) Put(context.Context, string, []byte) error   { return ErrUnsupportedPlatform }
func (*DPAPIStore) Get(context.Context, string) ([]byte, error) { return nil, ErrUnsupportedPlatform }
func (*DPAPIStore) Delete(context.Context, string) error        { return ErrUnsupportedPlatform }
