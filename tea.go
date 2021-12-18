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

// Package tea provides a starting point for building apps within the org.
package tea

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-version"

	"github.com/pghq/go-tea/health"
	"github.com/pghq/go-tea/internal"
	"github.com/pghq/go-tea/internal/clock"
)

const (
	// DefaultVersion of the application
	DefaultVersion = "0.0.1"

	// DefaultEnvironment of the application
	DefaultEnvironment = "dev"
)

// App provides access to various services within the SDK.
type App struct {
	version     *version.Version
	environment string
}

// NewApp constructs a new application instance
func NewApp(opts ...internal.AppOption) (*App, error) {
	conf := defaultConfig()
	for _, opt := range opts {
		opt.Apply(conf)
	}

	v, err := version.NewVersion(conf.Version)
	if err != nil {
		return nil, Error(err)
	}

	app := &App{
		version:     v,
		environment: conf.Environment,
	}

	err = Init(MonitorConfig{
		Version:     app.version.String(),
		Environment: app.environment,
	})

	if err != nil {
		return nil, err
	}

	return app, nil
}

// Health provides a client for services within the health domain
func (a *App) Health() *health.Client {
	return health.NewClient(a.version.String(), clock.New(time.Now()))
}

// Router provides a Router for serving http traffic.
func (a *App) Router() *Router {
	r := NewRouter(a.version.Segments()[0]).
		Middleware(NewSentryMiddleware()).
		Route("GET", "/health/status", func(w http.ResponseWriter, r *http.Request) {
			status := a.Health().Checks.Status()
			NewResponse().Body(status).Send(w, r)
		})

	return r
}

// defaultConfig provides the default configuration for the application
func defaultConfig() *internal.AppConfig {
	config := &internal.AppConfig{
		Version:     DefaultVersion,
		Environment: DefaultEnvironment,
	}

	return config
}

// versionOption is an option for specifying the app version.
type versionOption string

func (o versionOption) Apply(conf *internal.AppConfig) {
	if conf != nil {
		conf.Version = string(o)
	}
}

// Version creates a new version option for the app.
func Version(v string) internal.AppOption {
	return versionOption(v)
}

// environmentOption is an option for specifying the app environment.
type environmentOption string

func (o environmentOption) Apply(conf *internal.AppConfig) {
	if conf != nil {
		conf.Environment = string(o)
	}
}

// Environment creates a new environment option for the app.
func Environment(environment string) internal.AppOption {
	return environmentOption(environment)
}
