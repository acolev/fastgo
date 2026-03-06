package logger

import (
	"context"
	"strings"
	"sync"
)

type requestScopeKey struct{}
type requestLogBufferKey struct{}

type RequestScope struct {
	Method    string
	Path      string
	IP        string
	RequestID string
}

type requestLogBuffer struct {
	mu     sync.Mutex
	events []string
}

func WithRequestScope(ctx context.Context, scope RequestScope) context.Context {
	return context.WithValue(ctx, requestScopeKey{}, scope)
}

func WithRequestLogBuffer(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestLogBufferKey{}, &requestLogBuffer{})
}

func ScopeMessage(ctx context.Context, msg string) string {
	prefix := ScopePrefix(ctx)
	if prefix == "" {
		return msg
	}

	return prefix + " " + msg
}

func ScopePrefix(ctx context.Context) string {
	scope, ok := ScopeFromContext(ctx)
	if !ok {
		return ""
	}

	method := strings.TrimSpace(scope.Method)
	path := strings.TrimSpace(scope.Path)
	if method == "" && path == "" {
		return ""
	}

	return "[" + strings.TrimSpace(method+" "+path) + "]"
}

func ScopeAttrs(ctx context.Context) []any {
	scope, ok := ScopeFromContext(ctx)
	if !ok {
		return nil
	}

	attrs := make([]any, 0, 6)
	if scope.Method != "" {
		attrs = append(attrs, "method", scope.Method)
	}
	if scope.Path != "" {
		attrs = append(attrs, "path", scope.Path)
	}
	if scope.RequestID != "" {
		attrs = append(attrs, "request_id", scope.RequestID)
	}

	return attrs
}

func ScopeFromContext(ctx context.Context) (RequestScope, bool) {
	if ctx == nil {
		return RequestScope{}, false
	}

	scope, ok := ctx.Value(requestScopeKey{}).(RequestScope)
	return scope, ok
}

func AppendRequestEvent(ctx context.Context, line string) bool {
	buffer, ok := requestLogBufferFromContext(ctx)
	if !ok {
		return false
	}

	buffer.mu.Lock()
	buffer.events = append(buffer.events, strings.TrimSpace(line))
	buffer.mu.Unlock()

	return true
}

func DrainRequestEvents(ctx context.Context) []string {
	buffer, ok := requestLogBufferFromContext(ctx)
	if !ok {
		return nil
	}

	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	if len(buffer.events) == 0 {
		return nil
	}

	events := append([]string(nil), buffer.events...)
	buffer.events = buffer.events[:0]
	return events
}

func requestLogBufferFromContext(ctx context.Context) (*requestLogBuffer, bool) {
	if ctx == nil {
		return nil, false
	}

	buffer, ok := ctx.Value(requestLogBufferKey{}).(*requestLogBuffer)
	return buffer, ok
}
