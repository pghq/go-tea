package tea

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/hashicorp/go-version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TraceVersion is instrumentation tracing version
	TraceVersion = "0.0.2"
)

// tracer global tracer for open telemetry instrumentation
var tracer Tracer

func init() {
	tracer = defaultTracer()
	sentryOpts := sentry.ClientOptions{
		AttachStacktrace: true,
		Release:          TraceVersion,
		Environment:      os.Getenv("APP_ENV"),
	}
	_ = sentry.Init(sentryOpts)
}

// Tracer is an open telemetry tracer
type Tracer struct {
	provider *sdktrace.TracerProvider
	otel     trace.Tracer
}

// Flush any pending batched
func Flush() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tracer.provider.ForceFlush(ctx)
	sentry.Flush(5 * time.Second)
}

// defaultTracer gets the default tracer
func defaultTracer() Tracer {
	exporter, _ := stdouttrace.New()
	provider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	return Tracer{provider: provider, otel: provider.Tracer("tea", trace.WithInstrumentationVersion(TraceVersion), trace.WithSchemaURL(semconv.SchemaURL))}
}

// Span is a trace span
type Span struct {
	context.Context
	otel   trace.Span
	sentry *sentry.Span
}

// IsTracing checks if the span has tracing enabled
func (s Span) IsTracing() bool {
	return s.otel != nil && s.sentry != nil
}

// SetVersion sets the version of the span
func (s Span) SetVersion(version *version.Version) {
	if s.IsTracing() && version != nil {
		s.Tag("version", version.String())
	}
}

// SetRequest sets an http request on the span
func (s Span) SetRequest(r *http.Request) {
	if s.IsTracing() && r != nil {
		s.Tag("endpoint", r.URL.Path)
		if query := r.URL.Query(); len(query) > 0 {
			s.Tag("query", query.Encode())
		}
		s.Tag("method", r.Method)
		hub := s.sentryHub()
		hub.Scope().SetRequest(r)
	}
}

// Capture sets an error for the span
func (s Span) Capture(err error) {
	if s.IsTracing() && IsFatal(err) {
		s.otel.RecordError(err)
		hub := s.sentryHub()
		hub.CaptureException(err)
	}
}

// Tag sets an attribute value
func (s Span) Tag(key interface{}, v ...interface{}) {
	if s.IsTracing() && len(v) > 0 {
		key, value := fmt.Sprintf("%s", key), fmt.Sprint(v...)
		s.otel.SetAttributes(attribute.String(key, value))
		s.sentry.SetTag(key, value)
	}
}

// End a span
func (s Span) End() {
	if s.IsTracing() {
		if err := recover(); err != nil {
			hub := s.sentryHub()
			hub.RecoverWithContext(s, err)
			hub.Flush(5 * time.Second)
			s.otel.RecordError(Err(err))
		}

		s.otel.End(trace.WithStackTrace(true))
		s.sentry.Finish()
	}
}

// sentryHub gets the sentry hub of the span
func (s Span) sentryHub() *sentry.Hub {
	hub := sentry.GetHubFromContext(s)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	return hub
}

// Start is a named context providing a trace span
func Start(ctx context.Context, name string) Span {
	s := Span{Context: ctx}
	if Verbosity() == "trace" {
		s.Context, s.otel = tracer.otel.Start(ctx, name)
		s.sentry = sentry.StartSpan(s.Context, name)
	}

	return s
}

// Nest creates a named child trace
func Nest(ctx context.Context, name string) Span {
	s, _ := ctx.(Span)
	if s.Context == nil {
		s.Context = ctx
	}

	if s.IsTracing() {
		s.Context, s.otel = tracer.otel.Start(ctx, name)
		s.sentry = sentry.StartSpan(s.Context, name)
	}

	return s
}

// TraceMiddleware is an implementation of the sentry middleware
type TraceMiddleware struct {
	version *version.Version
	sentry  *sentryhttp.Handler
}

// Handle provides an http handler for handling exceptions
func (m TraceMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}
		span := Start(ctx, "http")
		defer span.End()
		span.SetRequest(r)
		span.SetVersion(m.version)
		r = r.WithContext(span)
		hub.Scope().SetRequest(r)
		next.ServeHTTP(w, r)
	})
}

// Trace constructs a new middleware that handles exceptions
func Trace(version *version.Version) TraceMiddleware {
	return TraceMiddleware{
		version: version,
		sentry:  sentryhttp.New(sentryhttp.Options{}),
	}
}
