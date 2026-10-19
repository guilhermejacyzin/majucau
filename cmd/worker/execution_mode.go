package main

import "context"

type workerExecutionModeKey struct{}

func withWindowsServiceExecution(ctx context.Context) context.Context {
	return context.WithValue(ctx, workerExecutionModeKey{}, true)
}

func usesWindowsServiceExecution(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	serviceMode, _ := ctx.Value(workerExecutionModeKey{}).(bool)
	return serviceMode
}
