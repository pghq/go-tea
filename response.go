package tea

import (
	"encoding/json"

	"net/http"

	"github.com/pghq/go-tea/trail"
)

// Send sends an HTTP response based on content type and body
func Send(w http.ResponseWriter, r *http.Request, raw interface{}) {
	if raw == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body, content, err := Body(r, raw)
	if err != nil {
		SendError(w, r, err)
		return
	}

	if content != "" {
		w.Header().Set("Content-Type", content)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

// SendError replies to the request with an error
// and emits fatal http errors to global log and monitor.
func SendError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	span := trail.StartSpan(r.Context(), "http.error")
	defer span.Finish()

	msg := err.Error()
	status := trail.StatusCode(err)

	span.Tag("error", msg)
	if trail.IsFatal(err) {
		span.Capture(err)
		msg = http.StatusText(status)
	}

	http.Error(w, msg, status)
}

// SendNotAuthorized sends a not authorized error
func SendNotAuthorized(w http.ResponseWriter, r *http.Request, err error, force ...bool) {
	if (len(force) == 0 || !force[0]) && trail.IsFatal(err) {
		SendError(w, r, err)
		return
	}

	SendError(w, r, trail.NewErrorWithCode(err.Error(), http.StatusUnauthorized))
}

// Body gets the response body as bytes based on origin
func Body(r *http.Request, body interface{}) ([]byte, string, error) {
	if Accepts(r, "*/*") {
		if body, ok := body.([]byte); ok {
			return body, "", nil
		}

		if body, ok := body.(string); ok {
			return []byte(body), "", nil
		}
	}

	switch {
	case Accepts(r, "application/json"):
		bytes, err := json.Marshal(body)
		if err != nil {
			return nil, "", trail.Stacktrace(err)
		}

		return bytes, "application/json", nil
	}

	return nil, "", trail.NewErrorBadRequest("unsupported content type")
}
