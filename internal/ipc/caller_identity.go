package ipc

import (
	"context"
	"regexp"
)

type callerWindowsSIDContextKey struct{}

var windowsSIDPattern = regexp.MustCompile(`^S-[0-9]+(-[0-9]+){1,15}$`)

// WithCallerWindowsSID marks a request with the Windows identity authenticated
// by the named-pipe server. Callers must not derive this value from payloads.
func WithCallerWindowsSID(ctx context.Context, sid string) context.Context {
	if ctx == nil || len(sid) > 184 || !windowsSIDPattern.MatchString(sid) {
		return ctx
	}
	return context.WithValue(ctx, callerWindowsSIDContextKey{}, sid)
}

// CallerWindowsSID returns only the identity attached by the authenticated
// local IPC transport; ordinary request fields cannot set it.
func CallerWindowsSID(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	sid, ok := ctx.Value(callerWindowsSIDContextKey{}).(string)
	return sid, ok && windowsSIDPattern.MatchString(sid)
}
