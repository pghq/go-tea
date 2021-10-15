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

// Package cache provides a caching for the app.
package cache

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

const (
	// DefaultCacheSize is the default size of the Cache
	DefaultCacheSize int = 1024
)

// encodeKey encodes the key into a format consistent and compatible with the Cache.
func encodeKey(key interface{}) (string, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return "", errors.Wrap(err)
	}

	return fmt.Sprintf("%x", sha1.Sum(buf.Bytes())), nil
}

// Item is an instance of a value in the lru Cache.
type Item struct {
	value    interface{}
	cachedAt time.Time
	ttl      time.Duration
}

// CachedAt gets the time the item was added to the cace
func (i *Item) CachedAt() time.Time {
	return i.cachedAt
}

// Value gets the raw object
func (i *Item) Value() interface{} {
	return i.value
}
