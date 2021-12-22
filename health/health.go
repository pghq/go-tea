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

// Package health provides resources for interacting with the health domain.
package health

import (
	"time"
)

// Service is a shared service for all health services
type Service struct {
	now     func() time.Time
	start   time.Time
	version string
}

// NewService creates a new health client instance
func NewService(version string) Service {
	return Service{
		version: version,
		now:     time.Now,
		start:   time.Now(),
	}
}
