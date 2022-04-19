package trail

import (
	"net/http"
)

// NewTraceMiddleware constructs a new middleware that handles tracing
func NewTraceMiddleware(version string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req, err := NewRequest(w, r, version)
			if err != nil {
				globalLogger.ErrorWithStacktrace(err)
				globalLogger.Flush()
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			defer req.Finish()
			defer func() {
				if err := recover(); err != nil {
					req.Recover(err)
					req.Response().WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(req.Response(), req.Origin())
		})
	}
}
