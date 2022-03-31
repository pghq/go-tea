package tea

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/rs/cors"

	"github.com/pghq/go-tea/trail"
)

const (
	// maxUploadSize is the max http body size that can be sent to the app (~64 MB)
	maxUploadSize = 64 << 20
)

var (
	// dec is a global request decoder.
	dec *schema.Decoder
)

func init() {
	dec = schema.NewDecoder()
	dec.ZeroEmpty(true)
	dec.SetAliasTag("json")
}

// ParseURL is a method to decode a http request query and path into a value
// schema struct tags are supported
//
// Query parameters and path parameters get decoded
func ParseURL(r *http.Request, v interface{}) error {
	if v == nil {
		return trail.NewError("no value")
	}

	if err := dec.Decode(v, r.URL.Query()); err != nil {
		return trail.ErrorBadRequest(err)
	}

	// path returns a map of parameters within the path
	path := func(r *http.Request) url.Values {
		vars := mux.Vars(r)
		parameters := make(url.Values)
		for key, value := range vars {
			parameters.Set(key, value)
		}

		return parameters
	}

	if err := dec.Decode(v, path(r)); err != nil {
		return trail.ErrorBadRequest(err)
	}

	return nil
}

// Parse is a method to decode a http request body into a value
// JSON and schema struct tags are supported
func Parse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return trail.NewError("no value")
	}

	if r.Body == http.NoBody {
		return nil
	}

	b, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, maxUploadSize))
	if err != nil {
		return trail.ErrorBadRequest(err)
	}

	_ = r.Body.Close()
	body := ioutil.NopCloser(bytes.NewBuffer(b))
	r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	ct := r.Header.Get("Content-Type")

	switch {
	case strings.Contains(ct, "application/json"):
		if err := json.NewDecoder(body).Decode(v); err != nil {
			return trail.ErrorBadRequest(err)
		}
	default:
		return trail.NewErrorBadRequest("content type not supported")
	}

	return nil
}

// Part reads a multipart message by name from a http request and leaves the body intact
func Part(w http.ResponseWriter, r *http.Request, name string) (*multipart.Part, error) {
	b, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, maxUploadSize))
	if err != nil {
		return nil, trail.ErrorBadRequest(err)
	}

	_ = r.Body.Close()
	body := ioutil.NopCloser(bytes.NewBuffer(b))
	defer func() {
		r.Body = body
		r.MultipartForm = nil
	}()

	r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, trail.ErrorBadRequest(err)
	}

	for {
		part, err := reader.NextPart()
		if err != nil {
			return nil, trail.ErrorBadRequest(err)
		}

		if part.FormName() == name {
			return part, nil
		}
	}
}

// Auth reads and parses the authorization header
// from the request if provided
func Auth(r *http.Request) string {
	auth := strings.Split(r.Header.Get("Authorization"), " ")

	if len(auth) != 2 {
		return ""
	}

	return auth[1]
}

// Accepts checks whether the response type is accepted
func Accepts(r *http.Request, contentType string) bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return true
	}

	if strings.Contains(accept, "*/*") {
		return true
	}

	if strings.Contains(accept, contentType) {
		return true
	}

	return false
}

// CORSMiddleware is an implementation of the CORS middleware
// providing method, origin, and credential allowance
type CORSMiddleware struct{ cors *cors.Cors }

// Handle provides an http handler for handling CORS
func (m CORSMiddleware) Handle(next http.Handler) http.Handler {
	return m.cors.Handler(next)
}

// NewCORSMiddleware constructs a new middleware that handles CORS
func NewCORSMiddleware() CORSMiddleware {
	return CORSMiddleware{
		cors: cors.New(cors.Options{
			AllowedOrigins:   nil,
			AllowCredentials: true,
			AllowedMethods: []string{
				http.MethodOptions,
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodDelete,
				http.MethodPatch,
				http.MethodPut,
			},
			AllowedHeaders: []string{"*"},
		}),
	}
}
