package trail

import (
	"bytes"
	"net/http"
)

type httpSpanWriter struct {
	statusCode int
	r          *Request
	w          http.ResponseWriter
	bytes      bytes.Buffer
}

func (w *httpSpanWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.r.SetStatus(statusCode)
}

func (w *httpSpanWriter) Header() http.Header {
	return w.w.Header()
}

func (w *httpSpanWriter) Write(b []byte) (int, error) {
	w.bytes.Write(b)
	return len(b), nil
}

func (w *httpSpanWriter) Send() {
	if w.statusCode != 0 {
		w.r.AddResponseHeaders(w.Header())
		w.Header().Set("Request-Trail", w.r.Trail())
		w.Header().Set("Request-Id", w.r.RequestId().String())
		w.w.WriteHeader(w.statusCode)
		_, _ = w.w.Write(w.bytes.Bytes())
	}
}
