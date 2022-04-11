package tea

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"net/http"

	"github.com/pghq/go-tea/trail"
)

// Send sends an HTTP response based on content type and body
func Send(w http.ResponseWriter, r *http.Request, raw interface{}) {
	if raw == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err, ok := raw.(error); ok {
		sendError(w, r, err)
		return
	}

	newHeaderEncoder(w).encode(raw)

	body, content, err := body(r, raw)
	if err != nil {
		sendError(w, r, err)
		return
	}

	if content != "" {
		w.Header().Set("Content-Type", content)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

// sendError replies to the request with an error
// and emits fatal http errors to global log and monitor.
func sendError(w http.ResponseWriter, r *http.Request, err error) {
	span := trail.StartSpan(r.Context(), "http.error")
	defer span.Finish()

	msg := err.Error()
	status := trail.StatusCode(err)

	span.Tags.Set("error", msg)
	if trail.IsFatal(err) {
		span.Capture(err)
		msg = http.StatusText(status)
	}

	http.Error(w, msg, status)
}

// body gets the response body as bytes based on origin
func body(r *http.Request, body interface{}) ([]byte, string, error) {
	if accepts(r, "*/*") {
		if body, ok := body.([]byte); ok {
			return body, "", nil
		}

		if body, ok := body.(string); ok {
			return []byte(body), "", nil
		}
	}

	switch {
	case accepts(r, "application/json"):
		bytes, err := json.Marshal(body)
		if err != nil {
			return nil, "", trail.Stacktrace(err)
		}

		return bytes, "application/json", nil
	}

	return nil, "", trail.NewErrorBadRequest("unsupported content type")
}

// headerEncoder encodes the value into the headers
type headerEncoder struct {
	w http.ResponseWriter
}

func (e headerEncoder) encode(v interface{}) {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if rv.Kind() != reflect.Struct {
		return
	}

	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		if key := t.Field(i).Tag.Get("header"); key != "" {
			v := rv.Field(i)
			omitempty := strings.HasSuffix(key, ",omitempty")
			key = strings.TrimSuffix(key, ",omitempty")

			switch v.Type().String() {
			case "string":
				if header := v.String(); header != "" || header == "" && !omitempty {
					e.w.Header().Set(key, header)
				}
			case "[]string":
				headers, _ := v.Interface().([]string)
				for _, header := range headers {
					if header != "" || header == "" && !omitempty {
						e.w.Header().Add(key, header)
					}
				}
			default:
				e.w.Header().Set(key, fmt.Sprintf("%s", v.Interface()))
			}
		}
	}
}

// newHeaderEncoder creates a new header decoder instance
func newHeaderEncoder(w http.ResponseWriter) *headerEncoder {
	return &headerEncoder{
		w: w,
	}
}
