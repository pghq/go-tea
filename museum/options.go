package museum

import (
	"github.com/pghq/go-museum/museum/internal"
)

// versionOption is an option for specifying the app version.
type versionOption string

func (o versionOption) Apply(conf *internal.AppConfig) {
	if conf != nil {
		conf.Version = string(o)
	}
}

// Version creates a new version option for the app.
func Version(v string) internal.AppOption{
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
func Environment(environment string) internal.AppOption{
	return environmentOption(environment)
}
