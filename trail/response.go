package trail

import (
	"net/http"
)

type httpSpanWriter struct {
	http.ResponseWriter
	statusCode int
	r          *Request
}

func (w *httpSpanWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.r.SetStatus(statusCode)
}

func (w *httpSpanWriter) Send() {
	if w.statusCode != 0 {
		w.r.AddResponseHeaders(w.Header())
		w.Header().Set("Request-Trail", w.r.Trail())
		w.Header().Set("Request-Id", w.r.RequestId().String())
		w.ResponseWriter.WriteHeader(w.statusCode)
	}
}
