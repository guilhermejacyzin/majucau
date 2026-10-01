package main

import (
	"reflect"
	"strings"
	"testing"

	"majucau.local/financial-intelligence/internal/application"
)

func TestBlingOAuthIPCResponseContractsNeverExposeCredentials(t *testing.T) {
	responses := []struct {
		name string
		typeOf reflect.Type
	}{
		{name: "start", typeOf: reflect.TypeOf(application.BlingOAuthStartResponse{})},
		{name: "status", typeOf: reflect.TypeOf(application.BlingOAuthStatusResponse{})},
		{name: "test", typeOf: reflect.TypeOf(application.BlingOAuthTestResponse{})},
		{name: "disconnect", typeOf: reflect.TypeOf(application.BlingOAuthDisconnectResponse{})},
	}

	for _, response := range responses {
		t.Run(response.name, func(t *testing.T) {
			for index := 0; index < response.typeOf.NumField(); index++ {
				field := response.typeOf.Field(index)
				jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
				name := strings.ToLower(field.Name + " " + jsonName)
				if strings.Contains(name, "token") || strings.Contains(name, "secret") || strings.Contains(name, "credential") {
					t.Fatalf("OAuth IPC response exposes credential field %q", field.Name)
				}
			}
		})
	}
}
