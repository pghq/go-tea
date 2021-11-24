package internal

import "net/http"

// NoopHandler is a http handler that does nothing.
var NoopHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
