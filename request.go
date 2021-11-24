package tea

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

const (
	// maxUploadSize is the max http body size that can be sent to the app (~16 MB)
	maxUploadSize = 16 << 20

	// defaultQueryLimit is the default query limit.
	defaultQueryLimit = 25

	// maxQueryLimit is the max query limit.
	maxQueryLimit = 100
)

// decoder is a global request decoder.
var decoder = NewDecoder()

// Decoder is an instance of a http request decoder
type Decoder struct {
	schema *schema.Decoder
}

// Decode decodes an http request to a value
func (d *Decoder) Decode(r *http.Request, v interface{}) error {
	d.schema.IgnoreUnknownKeys(true)
	d.schema.ZeroEmpty(true)
	d.schema.SetAliasTag("json")

	// Query parameters and path parameters get decoded
	if err := d.schema.Decode(v, r.URL.Query()); err != nil {
		return Error(err)
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

	if err := d.schema.Decode(v, path(r)); err != nil {
		return Error(err)
	}

	return nil
}

// NewDecoder creates a request decoder with sane defaults.
func NewDecoder() *Decoder {
	return &Decoder{
		schema: schema.NewDecoder(),
	}
}

// CurrentDecoder gets an instance of the global request decoder
func CurrentDecoder() *Decoder {
	return decoder
}

// Request is an instance of a http request
type Request struct {
	query interface{}
	body interface{}
}

// Query sets the query to decode to
func (r *Request) Query(v interface{}) *Request{
	r.query = v
	return r
}

// Body sets the body to decode to
func (r *Request) Body(v interface{}) *Request{
	r.body = v
	return r
}

// Decode the request
func (r *Request) Decode(w http.ResponseWriter, req *http.Request) error{
	if r.query != nil{
		if err := DecodeURL(req, r.query); err != nil{
			return Error(err)
		}
	}

	if r.body != nil{
		if err := DecodeBody(w, req, r.body); err != nil{
			return Error(err)
		}
	}

	return nil
}

// NewRequest creates an instance of a http request
func NewRequest() *Request{
	return &Request{}
}

// DecodeBody is a method to decode a http request body into a value
// JSON and schema struct tags are supported
func DecodeBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return NewError("value must be defined")
	}

	if r.Body == http.NoBody {
		return nil
	}

	b, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, maxUploadSize))
	if err != nil {
		return BadRequest(err)
	}

	_ = r.Body.Close()
	body := ioutil.NopCloser(bytes.NewBuffer(b))
	r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	ct := r.Header.Get("Content-Type")

	switch {
	case strings.Contains(ct, "application/json"):
		if err := json.NewDecoder(body).Decode(v); err != nil {
			return BadRequest(err)
		}
	default:
		return NewBadRequest("content type not supported")
	}

	return nil
}

// MultipartPart reads a multipart by name from a http request and leaves the body intact
func MultipartPart(w http.ResponseWriter, r *http.Request, name string) (*multipart.Part, error) {
	b, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, maxUploadSize))
	if err != nil {
		return nil, BadRequest(err)
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
		return nil, BadRequest(err)
	}

	for {
		part, err := reader.NextPart()
		if err != nil {
			return nil, BadRequest(err)
		}

		if part.FormName() == name {
			return part, nil
		}
	}
}

// DecodeURL is a method to decode a http request query and path into a value
// schema struct tags are supported
func DecodeURL(r *http.Request, v interface{}) error {
	if v == nil {
		return NewError("value must be defined")
	}

	rd := CurrentDecoder()
	if err := rd.Decode(r, v); err != nil {
		return BadRequest(err)
	}

	return nil
}

// Authorization reads and parses the authorization header
// from the request if provided
func Authorization(r *http.Request) (string, string) {
	auth := strings.Split(r.Header.Get("Authorization"), " ")

	if len(auth) != 2 {
		return "", ""
	}

	return auth[0], auth[1]
}

// Page gets the first and after queries for pagination
func Page(r *http.Request) (int, *time.Time, error) {
	first, err := First(r)
	if err != nil {
		return 0, nil, Error(err)
	}

	after, err := After(r)
	if err != nil {
		return 0, nil, Error(err)
	}

	return first, after, nil
}

// First gets the first query for pagination
func First(r *http.Request) (int, error) {
	if f := r.URL.Query().Get("first"); f != "" {
		first, err := strconv.ParseInt(f, 10, 64)
		if err != nil {
			return 0, BadRequest(err)
		}

		if first > maxQueryLimit {
			return 0, NewBadRequest("too many results desired")
		}

		return int(first), nil
	}

	return defaultQueryLimit, nil
}

// After gets the after query for pagination
func After(r *http.Request) (*time.Time, error) {
	if a := r.URL.Query().Get("after"); a != "" {
		ds, err := base64.StdEncoding.DecodeString(a)
		if err != nil {
			return nil, BadRequest(err)
		}

		after, err := time.Parse(time.RFC3339Nano, string(ds))
		if err != nil {
			return nil, BadRequest(err)
		}

		return &after, nil
	}

	return &time.Time{}, nil

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
