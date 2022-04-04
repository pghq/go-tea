package tea

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
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

// Parse is a method to decode a http request into a value
// schema struct tags are supported
func Parse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return trail.NewError("no value")
	}

	if r.Body != http.NoBody {
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
		case strings.Contains(ct, "multipart/form-data"):
			if err := NewMultipartDecoder(w, r).Decode(v); err != nil {
				return trail.ErrorBadRequest(err)
			}
		default:
			return trail.NewErrorBadRequest("content type not supported")
		}
	}

	if err := NewHeaderDecoder(r).Decode(v); err != nil {
		return trail.ErrorBadRequest(err)
	}

	return ParseURL(r, v)
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
func Auth(r *http.Request, scheme string) string {
	auth := strings.Split(r.Header.Get("Authorization"), " ")

	if len(auth) != 2 || strings.ToLower(scheme) != strings.ToLower(auth[0]) {
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

// MultipartDecoder decodes multipart/form-data requests into multipart.Parts
type MultipartDecoder struct {
	w http.ResponseWriter
	r *http.Request
}

func (d MultipartDecoder) Decode(v interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if rv.Kind() != reflect.Struct {
		return trail.NewErrorf("item of type %T is not a struct", v)
	}

	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		if key := t.Field(i).Tag.Get("form"); key != "" {
			v := rv.Field(i)
			if v.CanSet() && v.Type().Implements(reflect.TypeOf(new(io.Reader)).Elem()) {
				part, err := Part(d.w, d.r, key)
				if err != nil {
					return trail.ErrorBadRequest(err)
				}

				v.Set(reflect.ValueOf(part))
			}
		}
	}

	return nil
}

// NewMultipartDecoder creates a new multipart/form-data decoder
func NewMultipartDecoder(w http.ResponseWriter, r *http.Request) *MultipartDecoder {
	return &MultipartDecoder{
		w: w,
		r: r,
	}
}

// HeaderDecoder decodes headers into structs
type HeaderDecoder struct {
	r *http.Request
}

func (d HeaderDecoder) Decode(v interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if rv.Kind() != reflect.Struct {
		return trail.NewErrorf("item of type %T is not a struct", v)
	}

	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		if key := t.Field(i).Tag.Get("auth"); key != "" {
			v := rv.Field(i)
			if v.CanSet() && v.Type().String() == "string" {
				v.Set(reflect.ValueOf(Auth(d.r, key)))
			}
		}

		if key := t.Field(i).Tag.Get("header"); key != "" {
			v := rv.Field(i)
			if v.CanSet() {
				headers := d.r.Header
				if v.Type().String() == "string" {
					header := headers.Get(key)
					if header == "" {
						header = t.Field(i).Tag.Get("default")
					}
					v.Set(reflect.ValueOf(header))
				}
				if v.Type().String() == "[]string" {
					v.Set(reflect.ValueOf(headers.Values(key)))
				}
			}
		}
	}

	return nil
}

// NewHeaderDecoder creates a new header decoder instance
func NewHeaderDecoder(r *http.Request) *HeaderDecoder {
	return &HeaderDecoder{
		r: r,
	}
}
