//go:build !windows

package main

import "context"

func tryRunAsService(_ string, _ func(context.Context) error) (bool, error) { return false, nil }
