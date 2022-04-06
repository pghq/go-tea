package trail

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
)

// TraceMiddleware is an implementation of the sentry middleware
type TraceMiddleware struct {
	version   string
	collector FiberCollectorFunc
}

// Handle provides an http handler for handling exceptions
func (m TraceMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := httpSpanWriter{ResponseWriter: w}
		ctx := r.Context()
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		span := StartSpan(ctx, "http.request",
			WithSpanRequest(r),
			WithSpanWriter(&sw),
			WithSpanVersion(m.version),
			WithSpanCollector(m.collector),
		)

		defer span.Finish()
		defer func() {
			if err := recover(); err != nil {
				span.Recover(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		r = r.WithContext(span.Context())
		next.ServeHTTP(&sw, r)
	})
}

// Collect sets a custom collector for fibers
func (m *TraceMiddleware) Collect(fn FiberCollectorFunc) {
	m.collector = fn
}

// NewTraceMiddleware constructs a new middleware that handles exceptions
func NewTraceMiddleware(version string) *TraceMiddleware {
	return &TraceMiddleware{
		version: version,
	}
}

type httpSpanWriter struct {
	statusCode int
	http.ResponseWriter
}

func (w *httpSpanWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *httpSpanWriter) WriteSpan(s *Span) {
	if w.statusCode == 0 {
		return
	}

	var spans []*Span
	root := s
	func() {
		for {
			select {
			case span := <-s.bundle.spans:
				if !span.EndTime.IsZero() {
					spans = append(spans, span)
				}
			default:
				return
			}
		}
	}()

	fiber := Fiber{
		FiberId:   uuid.New(),
		RequestId: root.requestId,
		UserAgent: root.userAgent,
		IP:        root.ip,
		Version:   s.version,
		URL:       root.url.String(),
		Status:    w.statusCode,
		Spans:     spans,
	}

	if s.collector != nil {
		fibers := []Fiber{fiber}
		for {
			var fiber Fiber
			if err := decompressHeader(w.ResponseWriter, "Trail-Fiber", &fiber); err != nil {
				break
			}
			fibers = append(fibers, fiber)
		}
		s.collector(fibers)
	} else {
		_ = compressHeader(w.ResponseWriter, "Trail-Fiber", &fiber)
	}

	w.ResponseWriter.WriteHeader(w.statusCode)
}

// compressHeader compresses a header
// commonly used for sending messages in a distributed architecture
func compressHeader(w http.ResponseWriter, k, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return Stacktrace(err)
	}

	w.Header().Add(fmt.Sprintf("%s", k), base64.StdEncoding.EncodeToString(enc.EncodeAll(b, make([]byte, 0, len(b)))))
	return nil
}

// decompressHeader a compressed header
func decompressHeader(w http.ResponseWriter, k, v interface{}) error {
	key := fmt.Sprintf("%s", k)
	values := w.Header().Values(key)
	var value string
	if len(values) > 0 {
		value = values[0]
	}

	b, _ := base64.StdEncoding.DecodeString(value)
	b, err := dec.DecodeAll(b, nil)
	if err != nil {
		return ErrorBadRequest(err)
	}

	if err := json.Unmarshal(b, v); err != nil {
		return ErrorBadRequest(err)
	}

	w.Header().Del(key)
	for _, value := range values[1:] {
		w.Header().Add(key, value)
	}

	return nil
}
