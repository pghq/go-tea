package tea

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	// pathDec is a global request path decoder.
	pathDec *schema.Decoder

	// queryDec is a global request query decoder.
	queryDec *schema.Decoder
)

func init() {
	pathDec = schema.NewDecoder()
	pathDec.ZeroEmpty(true)
	pathDec.SetAliasTag("path")

	queryDec = schema.NewDecoder()
	queryDec.ZeroEmpty(true)
	queryDec.IgnoreUnknownKeys(true)
	queryDec.SetAliasTag("query")
}

// Parse is a method to decode a http request into a value
// schema struct tags are supported
func Parse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return trail.NewError("no value")
	}

	if r.Method != http.MethodGet && r.Body != http.NoBody {
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
			if err := newMultipartDecoder(w, r).decode(v); err != nil {
				return trail.ErrorBadRequest(err)
			}
		default:
			return trail.NewErrorBadRequest("content type not supported")
		}
	}

	return parseURL(r, v)
}

// parseURL is a method to decode a http request query and path into a value
// schema struct tags are supported
//
// Query parameters and path parameters get decoded
func parseURL(r *http.Request, v interface{}) error {
	if v == nil {
		return trail.NewError("no value")
	}

	newHeaderDecoder(r).decode(v)

	if err := queryDec.Decode(v, r.URL.Query()); err != nil {
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

	if err := pathDec.Decode(v, path(r)); err != nil {
		return trail.ErrorBadRequest(err)
	}

	return nil
}

// part reads a multipart message by name from a http request and leaves the body intact
func part(w http.ResponseWriter, r *http.Request, name string) (*multipart.Part, error) {
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

// auth reads and parses the authorization header
// from the request if provided
func auth(r *http.Request, scheme string) string {
	auth := strings.Split(r.Header.Get("Authorization"), " ")

	if len(auth) != 2 || strings.ToLower(scheme) != strings.ToLower(auth[0]) {
		return ""
	}

	return auth[1]
}

// accepts checks whether the response type is accepted
func accepts(r *http.Request, contentType string) bool {
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

// multipartDecoder decodes multipart/form-data requests into multipart.Parts
type multipartDecoder struct {
	w http.ResponseWriter
	r *http.Request
}

func (d multipartDecoder) decode(v interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if rv.Kind() != reflect.Struct {
		return nil
	}

	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := t.Field(i)
		v := rv.Field(i)

		if !v.CanSet() {
			continue
		}

		if key := field.Tag.Get("form"); key != "" && v.Type().Implements(reflect.TypeOf(new(io.Reader)).Elem()) {
			part, err := part(d.w, d.r, key)
			if err != nil {
				return trail.ErrorBadRequest(err)
			}

			v.Set(reflect.ValueOf(part))
		}
	}

	trail.OneOff(fmt.Sprintf("%+v", rv))

	return nil
}

// newMultipartDecoder creates a new multipart/form-data decoder
func newMultipartDecoder(w http.ResponseWriter, r *http.Request) *multipartDecoder {
	return &multipartDecoder{
		w: w,
		r: r,
	}
}

// headerDecoder decodes headers into structs
type headerDecoder struct {
	r *http.Request
}

func (d headerDecoder) decode(v interface{})  {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if rv.Kind() != reflect.Struct {
		return
	}

	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := t.Field(i)
		v := rv.Field(i)

		if !v.CanSet() {
			continue
		}

		if field.Anonymous{
			rv := reflect.New(v.Type())
			d.decode(rv.Interface())
			v.Set(reflect.Indirect(rv))
		}

		if key := field.Tag.Get("auth"); key != "" && v.Type().String() == "string" {
			v.Set(reflect.ValueOf(auth(d.r, key)))
		}

		if key := field.Tag.Get("header"); key != "" {
			switch v.Type().String() {
			case "string":
				header := d.r.Header.Get(key)
				if header == "" {
					header = t.Field(i).Tag.Get("default")
				}
				v.Set(reflect.ValueOf(header))
			case "[]string":
				v.Set(reflect.ValueOf(d.r.Header.Values(key)))
			}
		}
	}
}

// newHeaderDecoder creates a new header decoder instance
func newHeaderDecoder(r *http.Request) *headerDecoder {
	return &headerDecoder{
		r: r,
	}
}
