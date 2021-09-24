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

	"github.com/pghq/go-museum/museum/internal"
)

const (
	// StatusHealthy represents a healthy application state
	StatusHealthy Status = "healthy"
)

// Status is a nice name representing the state of the application
type Status string

// Check is an object representing health of an app component
type Check struct {
	Time   time.Time   `json:"time"`
	Status Status      `json:"status,omitempty"`
	Value  interface{} `json:"observedValue"`
	Unit   string      `json:"observedUnit"`
}

// Client allows interfacing with various health services within the application
type Client struct {
	common service

	Checks *CheckService
}

func (c *Client) Handler() *Handler {
	return &Handler{client: c}
}

// Handler is a http handler for the health client
type Handler struct {
	client *Client
}

// service is a shared service for all health services
type service struct {
	clock internal.Clock
	version   string
}

// New creates a new health client instance
func New(version string, clock internal.Clock) *Client {
	c := &Client{}

	c.common.version = version
	c.common.clock = clock

	c.Checks = (*CheckService)(&c.common)

	return c
}

// NewHealthyCheck creates a check, denoting it as unhealthy
func NewHealthyCheck(observedAt time.Time, value interface{}, unit string) *Check {
	c := &Check{
		Time:   observedAt,
		Status: StatusHealthy,
		Value:  value,
		Unit:   unit,
	}

	return c
}
