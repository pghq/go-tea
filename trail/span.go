package trail

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
)

// maxSpans is the maximum number of spans to buffer
const maxSpans = 1000

// spanContextKey is the expected context key for spans
type spanContextKey struct{}

// Span for tracing instrumentation
type Span struct {
	SpanId    uuid.UUID `json:"spanId"`
	ParentId  uuid.UUID `json:"parentId"`
	Operation string    `json:"operation"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Tags      Tags      `json:"tags,omitempty"`

	ctx      context.Context
	bundle   *bundle
	sentry   *sentry.Span
	*Request `json:"-"`
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

// Finish the span
func (s *Span) Finish() {
	if s.EndTime.IsZero() {
		s.EndTime = time.Now()
	}
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
func StartSpan(ctx context.Context, operation string) *Span {
	parent, hasParent := ctx.Value(spanContextKey{}).(*Span)
	var node Span
	node = Span{
		Operation: operation,
		StartTime: time.Now(),

		ctx: context.WithValue(ctx, spanContextKey{}, &node),
	}

	if hasParent {
		node.ParentId = parent.SpanId
		node.bundle = parent.bundle
		node.Request = parent.Request
	} else {
		node.bundle = &bundle{}
	}

	node.sentry = sentry.StartSpan(ctx, operation)
	node.bundle.add(&node)

	return &node
}

func (s *Span) SetRequest(r *Request) {
	s.sentryHub().Scope().SetRequest(r.Origin())
	s.Request = r
}

func (s *Span) SetResponse(w http.ResponseWriter) {
	s.Request.AddResponse(w)
}

// Tags for custom data
type Tags map[string]string

// Set a tag
func (t *Tags) Set(key, value string) {
	if t == nil || *t == nil {
		*t = make(map[string]string)
	}

	data := *t
	if value == "" {
		delete(data, key)
	}

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
