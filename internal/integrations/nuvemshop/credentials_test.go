package nuvemshop

import (
	"context"
	"errors"
	"io/fs"
	"testing"
)

type memoryStore struct{ values map[string][]byte }

func (s *memoryStore) Put(_ context.Context, ref string, value []byte) error {
	if s.values == nil {
		s.values = map[string][]byte{}
	}
	s.values[ref] = append([]byte(nil), value...)
	return nil
}
func (s *memoryStore) Get(_ context.Context, ref string) ([]byte, error) {
	value, ok := s.values[ref]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return append([]byte(nil), value...), nil
}
func (s *memoryStore) Delete(_ context.Context, ref string) error { delete(s.values, ref); return nil }

func TestNormalizeConfigRequiresHTTPSRelay(t *testing.T) {
	valid, err := normalizeConfig(ConfigInput{AppID: "123", RedirectURI: "https://relay.example.test/oauth/callback", ClientSecret: "secret"})
	if err != nil || valid.AppID != "123" {
		t.Fatalf("valid config = %#v, err=%v", valid, err)
	}
	for _, input := range []ConfigInput{
		{AppID: "", RedirectURI: "https://relay.example.test/callback"},
		{AppID: "123", RedirectURI: "http://127.0.0.1/callback"},
		{AppID: "123", RedirectURI: "https://127.0.0.1/callback"},
		{AppID: "123", RedirectURI: "https://localhost/callback"},
		{AppID: "123", RedirectURI: "https://relay.example.test/callback", ClientSecret: string(make([]byte, 2049))},
	} {
		if _, err := normalizeConfig(input); !errors.Is(err, ErrNuvemshopCredentialValidation) {
			t.Fatalf("input %#v error=%v", input, err)
		}
	}
}

func TestCredentialServiceFailsClosedWithoutDatabase(t *testing.T) {
	service := NewCredentialService(nil, &memoryStore{})
	if _, err := service.Save(context.Background(), ConfigInput{AppID: "123", RedirectURI: "https://relay.example.test/callback", ClientSecret: "secret"}); !errors.Is(err, ErrNuvemshopCredentialVault) {
		t.Fatalf("error=%v", err)
	}
}
