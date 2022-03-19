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
)

func init() {
	sentryOpts := sentry.ClientOptions{
		AttachStacktrace: true,
		Release:          Version,
		Environment:      os.Getenv("APP_ENV"),
	}
	_ = sentry.Init(sentryOpts)
}

// Flush any pending batched
func Flush() {
	sentry.Flush(5 * time.Second)
	logger.Flush()
}

// Span is a trace span
type Span struct {
	context.Context
	sentry *sentry.Span
	logs   Logger
}

// IsTracing checks if the span has tracing enabled
func (s Span) IsTracing() bool {
	return s.sentry != nil
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
		hub := s.sentryHub()
		hub.CaptureException(err)
		s.logs.Error(err)
	}
}

// Recover from panics
func (s Span) Recover(err interface{}) {
	hub := s.sentryHub()
	hub.RecoverWithContext(s, err)
	hub.Flush(5 * time.Second)
	s.logs.ErrorWithStacktrace(err)
}

// Tag sets an attribute value
func (s Span) Tag(key interface{}, v ...interface{}) {
	if s.IsTracing() && len(v) > 0 {
		key, value := fmt.Sprintf("%s", key), fmt.Sprint(v...)
		s.sentry.SetTag(key, value)
		s.logs.Tag(key, value)
	}
}

// End a span
func (s Span) End() {
	if s.IsTracing() {
		s.sentry.Finish()
		s.logs.Flush()
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
	s := Span{Context: ctx, logs: NewLogger()}
	s.sentry = sentry.StartSpan(s.Context, name)
	return s
}

// Nest creates a named child trace
func Nest(ctx context.Context, name string) Span {
	s, _ := ctx.(Span)
	if s.Context == nil {
		s.Context = ctx
	}

	if s.IsTracing() {
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
		defer func() {
			if err := recover(); err != nil {
				span.Recover(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
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
