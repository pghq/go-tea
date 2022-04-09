package trail

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/klauspost/compress/zstd"
)

var (
	// enc is a global request encoder.
	enc *zstd.Encoder

	// dec is a global request decoder.
	dec *zstd.Decoder
)

func init() {
	enc, _ = zstd.NewWriter(nil)
	dec, _ = zstd.NewReader(nil)
}

// maxSpans is the maximum number of spans to buffer
const maxSpans = 1000

// spanContextKey is the expected context key for spans
type spanContextKey struct{}

// Tags for custom data
type Tags map[string]string

// Set a tag
func (t *Tags) Set(key, value string) {
	if t == nil || *t == nil {
		*t = make(map[string]string)
	}

	data := *t
	data[key] = value
}

// SetJSON sets a json value
func (t *Tags) SetJSON(key string, value interface{}) {
	if b, err := json.Marshal(value); err == nil {
		if t == nil || *t == nil {
			*t = make(map[string]string)
		}

		data := *t
		data[key] = string(b)
	}
}

// Get a field by key
func (t Tags) Get(key string) string {
	tag, _ := t[key]
	return tag
}

// Span for tracing instrumentation
type Span struct {
	SpanId    uuid.UUID `json:"spanId"`
	ParentId  uuid.UUID `json:"parentId"`
	Scope     string    `json:"scope"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Tags      Tags      `json:"tags,omitempty"`

	ip        net.IP
	requestId uuid.UUID
	userAgent string
	url       *url.URL
	version   string
	root      bool
	ctx       context.Context
	bundle    *bundle
	writer    SpanWriter
	collector FiberCollectorFunc
	sentry    *sentry.Span
}

// Context gets the original context of the span
func (s *Span) Context() context.Context {
	return s.ctx
}

// Capture sets an error for the span
func (s *Span) Capture(err error) {
	if IsFatal(err) {
		hub := s.sentryHub()
		hub.CaptureException(err)
		globalLogger.Error(err)
	}
}

// Recover from panics and fatal errors
func (s *Span) Recover(err interface{}) {
	hub := s.sentryHub()
	hub.RecoverWithContext(s.Context(), err)
	hub.Flush(5 * time.Second)
	globalLogger.ErrorWithStacktrace(err)
	globalLogger.Flush()
}

// Finish the span and write if root
func (s *Span) Finish() {
	if s.EndTime.IsZero() {
		s.EndTime = time.Now()
	}

	if !s.root || s.writer == nil {
		return
	}

	s.writer.WriteSpan(s)
}

// sentryHub gets the closest matching hub to the span or defaults to current hub
func (s *Span) sentryHub() *sentry.Hub {
	for k, value := range s.Tags {
		s.sentry.SetTag(k, value)
	}

	hub := sentry.GetHubFromContext(s.Context())
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	return hub
}

// StartSpan starts a new span instance (or continues from a parent)
func StartSpan(ctx context.Context, scope string, opts ...SpanOption) *Span {
	parent, hasParent := ctx.Value(spanContextKey{}).(*Span)
	var node Span
	node = Span{
		Scope:     scope,
		StartTime: time.Now(),

		ctx: context.WithValue(ctx, spanContextKey{}, &node),
	}

	if hasParent {
		node.ParentId = parent.SpanId
		node.bundle = parent.bundle
	} else {
		node.bundle = &bundle{}
	}

	node.sentry = sentry.StartSpan(ctx, scope)
	node.root = !hasParent
	node.bundle.add(&node)

	for _, opt := range opts {
		opt(&node)
	}

	return &node
}

// SpanOption is an option for configuring the span.
type SpanOption func(span *Span)

// WithSpanRequest is a span option for configuring the span with a http request
func WithSpanRequest(r *http.Request) SpanOption {
	return func(span *Span) {
		requestId := uuid.New()
		if rawId := r.Header.Get("Request-Id"); rawId != "" {
			if id, err := uuid.Parse(rawId); err == nil {
				requestId = id
			}
		}

		span.requestId = requestId
		span.ip = net.ParseIP(r.Header.Get("X-Forwarded-For"))
		span.userAgent = r.UserAgent()
		span.url = r.URL
		span.Tags.Set("request.id", requestId.String())
		span.sentryHub().Scope().SetRequest(r)
		r.Header.Set("Request-Id", requestId.String())
	}
}

// WithSpanWriter is a span option for configuring the span with a writer.
func WithSpanWriter(w SpanWriter) SpanOption {
	return func(span *Span) {
		span.writer = w
	}
}

// WithSpanVersion is a span option for configuring the span with an application version.
func WithSpanVersion(version string) SpanOption {
	return func(span *Span) {
		span.version = version
		span.Tags.Set("app.version", version)
	}
}

// WithSpanCollector is a span option for configuring the span with a custom collector.
func WithSpanCollector(collector FiberCollectorFunc) SpanOption {
	return func(span *Span) {
		span.collector = collector
	}
}

// Fiber is a local root span.
type Fiber struct {
	FiberId   uuid.UUID `json:"fiberId"`
	RequestId uuid.UUID `json:"requestId"`
	UserAgent string    `json:"userAgent,omitempty"`
	Version   string    `json:"version"`
	Status    int       `json:"status"`
	IP        net.IP    `json:"ip"`
	URL       string    `json:"url"`
	Spans     []*Span   `json:"spans"`
}

// bundle buffers the span to prevent unbound memory allocation.
type bundle struct {
	spans chan *Span
}

func (b *bundle) add(s *Span) {
	if b.spans == nil {
		b.spans = make(chan *Span, maxSpans)
	}

	select {
	case b.spans <- s:
	default:
		Warnf("tea.trail: dropping span %s with parent %d", s.SpanId, s.ParentId)
	}
}

// SpanWriter is a writer for handling finished local spans.
type SpanWriter interface {
	WriteSpan(span *Span)
}

// FiberCollectorFunc is a writer for handling aggregated span fibers.
type FiberCollectorFunc func(bundle []Fiber)
