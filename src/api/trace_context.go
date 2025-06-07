package api

import (
	"context"
)

// traceIDKey is the context key for trace IDs
type traceIDKey struct{}

// withTraceID adds a trace ID to the context
func withTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

// GetTraceID retrieves the trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey{}).(string); ok {
		return traceID
	}
	return ""
}