// Copyright 2021 PGHQ. All Rights Reserved.
//
// Licensed under the GNU General Public License, Version 3 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package request provides resources for decoding http requests
package request

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

const (
	// maxUploadSize is the max http body size that can be sent to the app
	maxUploadSize = 16 << 20

	// defaultQueryLimit is the default query limit.
	defaultQueryLimit = 25

	// maxQueryLimit is the max query limit.
	maxQueryLimit = 100
)

// DecodeBody is a method to decode a http request body into a value
// JSON and schema struct tags are supported
func DecodeBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil{
		return errors.New("value must be defined")
	}

	if r.Body == http.NoBody{
		return nil
	}

	b, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, maxUploadSize))
	if err != nil {
		return errors.HTTP(http.StatusBadRequest, err)
	}

	_ = r.Body.Close()
	body := ioutil.NopCloser(bytes.NewBuffer(b))
	r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	ct := r.Header.Get("Content-Type")

	switch {
	case strings.Contains(ct, "application/json"):
		if err := json.NewDecoder(body).Decode(v); err != nil {
			return errors.HTTP(http.StatusBadRequest, err)
		}
	default:
		return errors.NewHTTP(http.StatusBadRequest, "content type not supported")
	}

	return nil
}

// Decode is a method to decode a http request query and path into a value
// schema struct tags are supported
func Decode(r *http.Request, v interface{}) error {
	if v == nil{
		return errors.New("value must be defined")
	}

	rd := CurrentDecoder()
	if err := rd.Decode(r, v); err != nil {
		return errors.HTTP(http.StatusBadRequest, err)
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

// First gets the first query for pagination
func First(r *http.Request) (int, error) {
	if f := r.URL.Query().Get("first"); f != "" {
		first, err := strconv.ParseInt(f, 10, 64)
		if err != nil {
			return 0, errors.BadRequest(err)
		}

		if first > maxQueryLimit {
			return 0, errors.NewBadRequest("too many results desired")
		}

		return int(first), nil
	}

	return defaultQueryLimit, nil
}

// After gets the after query for pagination
func After(r *http.Request) (string, error) {
	if a := r.URL.Query().Get("after"); a != "" {
		after, err := base64.StdEncoding.DecodeString(a)
		if err != nil {
			return "", errors.BadRequest(err)
		}

		return string(after), nil
	}

	return "", nil

}

// Accepts checks whether the response type is accepted
func Accepts(r *http.Request, contentType string) bool {
	accept := r.Header.Get("Accept")

	if accept == "" {
		return true
	}

	if strings.Contains(accept, contentType) {
		return true
	}

	if strings.Contains(accept, "*/*") {
		return true
	}

	return false
}
