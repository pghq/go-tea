package tea

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/klauspost/compress/zstd"
	"github.com/rs/cors"
)

const (
	// maxUploadSize is the max http body size that can be sent to the app (~64 MB)
	maxUploadSize = 64 << 20
)

var (
	// enc is a global request encoder.
	enc Encoder

	// dec is a global request decoder.
	dec Decoder
)

func init() {
	enc = defaultEncoder()
	dec = defaultDecoder()
}

// Encoder for requests
type Encoder struct {
	schema *schema.Encoder
	ztsd   *zstd.Encoder
}

// defaultEncoder creates a new request encoder
func defaultEncoder() Encoder {
	enc := Encoder{schema: schema.NewEncoder()}
	enc.ztsd, _ = zstd.NewWriter(nil)
	return enc
}

// Attach a compressed header
// commonly used for sending messages between internal services
func Attach(w http.ResponseWriter, k, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return Stacktrace(err)
	}

	w.Header().Set(fmt.Sprintf("%s", k), base64.StdEncoding.EncodeToString(enc.ztsd.EncodeAll(b, make([]byte, 0, len(b)))))
	return nil
}

// Decoder is an instance of a http request decoder
type Decoder struct {
	schema *schema.Decoder
	ztsd   *zstd.Decoder
}

// defaultDecoder creates a request decoder with sane defaults.
func defaultDecoder() Decoder {
	dec := Decoder{schema: schema.NewDecoder()}
	dec.schema.ZeroEmpty(true)
	dec.schema.SetAliasTag("json")
	dec.ztsd, _ = zstd.NewReader(nil)
	return dec
}

// Detach a compressed header
func Detach(w http.ResponseWriter, k, v interface{}) error {
	b, _ := base64.StdEncoding.DecodeString(w.Header().Get(fmt.Sprintf("%s", k)))
	b, err := dec.ztsd.DecodeAll(b, nil)
	if err != nil {
		return AsErrBadRequest(err)
	}

	if err := json.Unmarshal(b, v); err != nil {
		return AsErrBadRequest(err)
	}

	return nil
}

// ParseURL is a method to decode a http request query and path into a value
// schema struct tags are supported
//
// Query parameters and path parameters get decoded
func ParseURL(r *http.Request, v interface{}) error {
	if v == nil {
		return Err("no value")
	}

	if err := dec.schema.Decode(v, r.URL.Query()); err != nil {
		return AsErrBadRequest(err)
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

	if err := dec.schema.Decode(v, path(r)); err != nil {
		return AsErrBadRequest(err)
	}

	return nil
}

// Parse is a method to decode a http request body into a value
// JSON and schema struct tags are supported
func Parse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return Err("no value")
	}

	if r.Body == http.NoBody {
		return nil
	}

	b, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, maxUploadSize))
	if err != nil {
		return AsErrBadRequest(err)
	}

	_ = r.Body.Close()
	body := ioutil.NopCloser(bytes.NewBuffer(b))
	r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	ct := r.Header.Get("Content-Type")

	switch {
	case strings.Contains(ct, "application/json"):
		if err := json.NewDecoder(body).Decode(v); err != nil {
			return AsErrBadRequest(err)
		}
	default:
		return ErrBadRequest("content type not supported")
	}

	return nil
}

// Part reads a multipart message by name from a http request and leaves the body intact
func Part(w http.ResponseWriter, r *http.Request, name string) (*multipart.Part, error) {
	b, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, maxUploadSize))
	if err != nil {
		return nil, AsErrBadRequest(err)
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
		return nil, AsErrBadRequest(err)
	}

	for {
		part, err := reader.NextPart()
		if err != nil {
			return nil, AsErrBadRequest(err)
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

// CORS constructs a new middleware that handles CORS
func CORS() CORSMiddleware {
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
