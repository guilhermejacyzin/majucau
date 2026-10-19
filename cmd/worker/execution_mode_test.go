package main

import (
	"context"
	"testing"
)

func TestWindowsServiceExecutionModeIsExplicit(t *testing.T) {
	if usesWindowsServiceExecution(context.Background()) {
		t.Fatal("console worker must not select Windows service secret storage")
	}
	if !usesWindowsServiceExecution(withWindowsServiceExecution(context.Background())) {
		t.Fatal("Windows service context must select service secret storage")
	}
	if usesWindowsServiceExecution(nil) {
		t.Fatal("nil context must not select Windows service secret storage")
	}
}
