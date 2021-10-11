package request

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

// decoder is a global request decoder.
var decoder = NewDecoder()

// Decoder is an instance of a http request decoder
type Decoder struct{
	schema *schema.Decoder
}

// Decode decodes an http request to a value
func (d *Decoder) Decode(r *http.Request, v interface{}) error{
	d.schema.IgnoreUnknownKeys(true)
	d.schema.ZeroEmpty(true)

	// Query parameters and path parameters get decoded
	if err := d.schema.Decode(v, r.URL.Query()); err != nil {
		return errors.Wrap(err)
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
		return errors.Wrap(err)
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
