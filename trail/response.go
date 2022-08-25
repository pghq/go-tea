package trail

import (
	"net/http"
)

type httpSpanWriter struct {
	r               *Request
	w               http.ResponseWriter
	withTrailHeader bool
}

func (w *httpSpanWriter) WriteHeader(statusCode int) {
	w.r.AddResponseHeaders(w.Header())
	w.r.SetStatus(statusCode)
	w.r.Finish()

	w.Header().Set("Request-Trail", w.r.Trail())
	if !w.withTrailHeader {
		w.Header().Del("Request-Trail")
	}

	w.Header().Set("Request-Id", w.r.RequestId())
	w.w.WriteHeader(statusCode)
}

func (w *httpSpanWriter) Header() http.Header {
	return w.w.Header()
}

func (w *httpSpanWriter) Write(b []byte) (int, error) {
	return w.w.Write(b)
}
