package internal

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TelemetryLibraryName is the name of the library in various observability metadata
	TelemetryLibraryName = "github.com/lusis/prototokens"
)

// StartSpan starts a trace span
// just to cut down on boilerplate a bit
func StartSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return otel.Tracer(TelemetryLibraryName).Start(ctx, spanName)
}
