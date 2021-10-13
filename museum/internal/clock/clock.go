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

// Package clock provides resources for dealing with internal time.
package clock

import (
	"time"

	"github.com/pghq/go-museum/museum/internal"
)

// clock is an instance of the internal Clock
type clock struct {
	start time.Time
	now   func() time.Time
}

// Now gets the current time
func (c *clock) Now() time.Time {
	return c.now()
}

// Start gets the clock start time
func (c *clock) Start() time.Time {
	return c.start
}

// From sets the source for getting the current time
func (c *clock) From(now func() time.Time) internal.Clock {
	c.now = now
	return c
}

// New creates a new instance of the internal Clock
func New(start time.Time) internal.Clock {
	return &clock{
		start: start,
		now: func() time.Time {
			return time.Now()
		},
	}
}
