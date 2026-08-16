package security

import (
	"context"
	"sync"
)

type FakeSecretStore struct {
	mu     sync.Mutex
	values map[string][]byte
}

func NewFakeSecretStore() *FakeSecretStore { return &FakeSecretStore{values: make(map[string][]byte)} }
func (s *FakeSecretStore) Put(_ context.Context, ref string, value []byte) error {
	if err := ValidateSecretRef(ref); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[ref] = append([]byte(nil), value...)
	return nil
}
func (s *FakeSecretStore) Get(_ context.Context, ref string) ([]byte, error) {
	if err := ValidateSecretRef(ref); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]byte(nil), s.values[ref]...), nil
}
func (s *FakeSecretStore) Delete(_ context.Context, ref string) error {
	if err := ValidateSecretRef(ref); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, ref)
	return nil
}
